package app

import (
	"bytes"
	"crypto/rand"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	_ "time/tzdata"
	"unicode"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"
)

//go:embed web/*
var web embed.FS
var pages = template.Must(template.New("").Funcs(template.FuncMap{"t": func(string, ...any) string { return "" }, "price": func(string) string { return "" }}).ParseFS(web, "web/*.html"))

// @impl INT-003-AC1 INT-003-AC6: only the language configuration is added to storage.
type household struct {
	Appearance    *appearance `yaml:"appearance,omitempty"`
	Language      string      `yaml:"language,omitempty"`
	SetupLanguage string      `yaml:"-"`
	dir           string
	Name          string `yaml:"name"`
	Timezone      string `yaml:"timezone"`
	ParentCode    string `yaml:"parent_code"`
	FamilyCode    string `yaml:"family_code"`
}

type application struct {
	dataDir string
	// Request-local render locale, accessed only under mu.
	language      string
	now           func() time.Time
	mu            sync.Mutex
	drafts        map[string]household
	assets        http.Handler
	pictureClient *http.Client
}

type view struct {
	Locale, Language string
	ReferrerPolicy   string

	ClaimPerson, ViewerName, NameDraft, NameError, ClaimError, PendingWish string
	NamePrompt, EditClaimName                                              bool
	Appearance                                                             *appearance
	Bulbs                                                                  []bulbDecoration
	LeftSnow, RightSnow                                                    []snowDecoration
	Wishes                                                                 []wishCard
	Draft                                                                  wishCard
	CreateDraft                                                            wishCard
	WishError                                                              string
	Name, Timezone, Error, Ticket, ParentLink, FamilyLink, Role, Code      string
	Year                                                                   int
	People                                                                 []personCard
	PersonName, HomeLink                                                   string
	Titles                                                                 []string
	Years                                                                  []int
	IsCurrent                                                              bool
	RenameName                                                             string
}

func newLegacy(dataDir string) http.Handler {
	assets, err := fs.Sub(web, "web")
	if err != nil {
		panic(err)
	}
	return &application{dataDir: dataDir, now: time.Now, drafts: make(map[string]household), pictureClient: newPictureClient(), assets: http.StripPrefix("/assets/", http.FileServer(http.FS(assets)))}
}

// @impl ACC-001-AC1 ACC-001-AC4 ACC-002-AC1 ACC-002-AC5
func (a *application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if r.URL.Path == "/assets/app.css" || r.URL.Path == "/assets/app.js" {
		a.assets.ServeHTTP(w, r)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	// @impl INT-003-AC3 INT-003-AC4 INT-003-AC5 INT-004-AC1 INT-004-AC2 INT-004-AC3
	a.language = "en-US"
	h, err := a.load()
	if err != nil {
		a.renderError(w, 500, "Unable to read household data. Check the YAML files; no files were changed.")
		return
	}
	if h != nil {
		code := r.URL.Query().Get("code")
		role := ""
		switch code {
		case h.ParentCode:
			role = "Parent"
		case h.FamilyCode:
			role = "Family"
		}
		if role == "" {
			a.renderError(w, 404, "404 page not found")
			return
		}
		a.language = householdLocale(h.Language)
		if r.Method == "GET" && r.URL.Path == "/wish-picture" {
			a.serveWishPicture(w, r, *h, role)
			return
		}
		if r.Method == "POST" && r.URL.Path == "/claims/toggle" {
			a.toggleClaim(w, r, *h, role, code)
			return
		}
		if r.Method == "POST" && r.URL.Path == "/claim-name" {
			a.changeClaimName(w, r, *h, role, code)
			return
		}
		if r.Method == "POST" && r.URL.Path == "/access/regenerate" && role == "Parent" {
			a.regenerate(w, r, *h)
			return
		}
		// @impl WLS-001-AC6 WLS-003-AC3 WLS-004-AC1
		if r.Method == "POST" && role == "Parent" {
			switch r.URL.Path {
			case "/wishes/delete":
				a.deleteWish(w, r, *h)
				return
			case "/people/rename":
				a.renamePerson(w, r, *h)
				return
			case "/people/archive":
				a.archivePerson(w, r, *h)
				return
			}
		}
		if r.Method == "POST" && r.URL.Path == "/wishes/save" && role == "Parent" {
			a.saveWish(w, r, *h)
			return
		}
		if r.Method == "POST" && r.URL.Path == "/people/add" && role == "Parent" {
			a.addPerson(w, r, *h)
			return
		}
		if r.Method == "GET" && (r.URL.Path == "/" || r.URL.Path == "/wishlist") {
			a.showPeople(w, r, *h, role, code)
			return
		}
		a.renderError(w, 404, "404 page not found")
		return
	}
	a.language = browserLocale(r.Header.Get("Accept-Language"))
	switch {
	case r.Method == "GET" && (r.URL.Path == "/" || r.URL.Path == "/setup"):
		a.render(w, 200, "setup.html", view{Language: a.language})
	case r.Method == "POST" && r.URL.Path == "/setup/prepare":
		a.prepare(w, r)
	case r.Method == "POST" && r.URL.Path == "/setup/finish":
		a.finish(w, r)
	default:
		a.language = "en-US"
		a.renderError(w, 404, "404 page not found")
	}
}

// @impl ACC-003-AC1 ACC-003-AC2 ACC-003-AC3 ACC-003-AC4 ACC-003-AC5 ACC-003-AC6
// ServeHTTP reloads YAML and validates the current parent code before each action.
// The request mutex keeps rotations and their authorization checks in one sequence.
func (a *application) regenerate(w http.ResponseWriter, r *http.Request, h household) {
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	if err := r.ParseForm(); err != nil {
		a.renderError(w, 400, "Invalid access form")
		return
	}
	role := r.PostForm.Get("role")
	if role != "parent" && role != "family" {
		a.renderError(w, 400, "Choose parent or family access")
		return
	}
	replacement, err := token()
	if err != nil {
		a.renderError(w, 500, "Unable to generate an access code")
		return
	}
	if role == "parent" {
		h.ParentCode = replacement
	} else {
		h.FamilyCode = replacement
	}
	if err := a.save(h); err != nil {
		a.renderError(w, 500, "Unable to save the replacement access code")
		return
	}
	http.Redirect(w, r, accessLink(h.ParentCode), http.StatusSeeOther)
}

// @impl CON-003-AC1 CON-003-AC5
func (a *application) renderError(w http.ResponseWriter, status int, message string) {
	a.render(w, status, "error.html", view{Error: message})
}

func (a *application) render(w http.ResponseWriter, status int, name string, data view) {
	// @impl INT-001-AC2 INT-001-AC4 INT-004-AC4 INT-005-AC1
	data.Locale = householdLocale(a.language)
	data.Error = translate(data.Locale, data.Error)
	data.NameError = translate(data.Locale, data.NameError)
	data.ClaimError = translate(data.Locale, data.ClaimError)
	data.WishError = translate(data.Locale, data.WishError)
	a.prepareWishPictures(&data)
	a.decorate(&data)
	localized, err := pages.Clone()
	if err != nil {
		http.Error(w, translate(data.Locale, "E07"), 500)
		return
	}
	localized = localized.Funcs(template.FuncMap{
		"t":     func(key string, args ...any) string { return translate(data.Locale, key, args...) },
		"price": func(raw string) string { return formatPrice(data.Locale, raw) },
	})
	w.Header().Set("Content-Language", data.Locale)
	var body bytes.Buffer
	if err := localized.ExecuteTemplate(&body, name, data); err != nil {
		http.Error(w, translate(data.Locale, "E07"), 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = body.WriteTo(w)
}

// @impl ACC-001-AC2 ACC-001-AC3 ACC-001-AC6 ACC-001-AC8 ACC-001-AC9
func (a *application) prepare(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	if err := r.ParseForm(); err != nil {
		a.renderError(w, 400, "Invalid setup form")
		return
	}
	// @impl INT-002-AC3 INT-002-AC4 INT-002-AC5
	if setupLanguage := r.PostForm.Get("setup_language"); supportedLocale(setupLanguage) {
		a.language = setupLanguage
	}
	language := r.PostForm.Get("language")
	name := strings.TrimSpace(r.PostForm.Get("household"))
	timezone := r.PostForm.Get("timezone")
	theme, themeErr := parseAppearance(r.PostForm)
	data := view{Name: name, Timezone: timezone, Appearance: theme, Language: language}
	if !supportedLocale(language) {
		data.Error = translate("en-US", "E47")
		a.render(w, 400, "setup.html", data)
		return
	}
	if themeErr != nil {
		data.Error = themeErr.Error()
		a.render(w, 400, "setup.html", data)
		return
	}
	if !validName(name) {
		data.Error = "Use 3–50 Unicode letters and digits for your household name."
		a.render(w, 400, "setup.html", data)
		return
	}
	if _, err := time.LoadLocation(timezone); err != nil || timezone == "" || timezone == "Local" {
		data.Error = "Choose a valid timezone, such as Europe/Zurich."
		a.render(w, 400, "setup.html", data)
		return
	}
	parent, err := token()
	if err != nil {
		a.renderError(w, 500, "Unable to generate access codes")
		return
	}
	family, err := token()
	if err != nil {
		a.renderError(w, 500, "Unable to generate access codes")
		return
	}
	ticket, err := token()
	if err != nil {
		a.renderError(w, 500, "Unable to prepare setup")
		return
	}
	a.drafts[ticket] = household{Name: name, Timezone: timezone, ParentCode: parent, FamilyCode: family, Appearance: theme, Language: language, SetupLanguage: a.language}
	data.Ticket = ticket
	data.ParentLink = accessLink(parent)
	data.FamilyLink = accessLink(family)
	a.render(w, 200, "links.html", data)
}

func validName(name string) bool {
	size := utf8.RuneCountInString(name)
	if size < 3 || size > 50 {
		return false
	}
	for _, char := range name {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

func token() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", bytes), nil
}
func accessLink(code string) string { return "/?code=" + url.QueryEscape(code) }

// @impl ACC-001-AC7 CON-004 INT-002-AC6 INT-003-AC1
func (a *application) finish(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	if err := r.ParseForm(); err != nil {
		a.renderError(w, 400, "Invalid setup form")
		return
	}
	h, ok := a.drafts[r.PostForm.Get("ticket")]
	if !ok {
		a.renderError(w, 400, "Setup preview is no longer available. Open setup again.")
		return
	}
	a.language = householdLocale(h.SetupLanguage)
	if err := a.save(h); err != nil {
		a.renderError(w, 500, "Unable to save household setup. No setup was completed.")
		return
	}
	clear(a.drafts)
	http.Redirect(w, r, accessLink(h.ParentCode), http.StatusSeeOther)
}

// @impl CON-004 ACC-001-AC5
// Load on every page request so direct YAML changes are observable without a restart.
func (a *application) load() (*household, error) {
	root := filepath.Join(a.dataDir, "households")
	dirs, err := os.ReadDir(root)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var found *household
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		file := filepath.Join(root, dir.Name(), "household.yaml")
		raw, err := os.Open(file)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		contents, readErr := io.ReadAll(raw)
		if readErr != nil {
			raw.Close()
			return nil, readErr
		}
		decoder := yaml.NewDecoder(bytes.NewReader(contents))
		decoder.KnownFields(true)
		var h household
		err = decoder.Decode(&h)
		if err == nil {
			var extra any
			if next := decoder.Decode(&extra); next != io.EOF {
				err = fmt.Errorf("expected a single YAML document")
			}
		}
		raw.Close()
		if err != nil {
			return nil, err
		}
		// @impl INT-003-AC1 INT-003-AC2 INT-003-AC3
		if h.Language == "" {
			// Resolve YAML merges/aliases too: an explicitly null language is
			// invalid even when inherited from a merged mapping.
			var fields map[string]any
			if err := yaml.Unmarshal(contents, &fields); err != nil {
				return nil, err
			}
			if _, present := fields["language"]; present {
				return nil, fmt.Errorf("invalid household language")
			}
		} else if !supportedLocale(h.Language) {
			return nil, fmt.Errorf("invalid household language")
		}
		if h.Appearance != nil && !h.Appearance.valid() {
			return nil, fmt.Errorf("invalid appearance settings")
		}
		if !validName(h.Name) || h.ParentCode == "" || h.FamilyCode == "" || h.ParentCode == h.FamilyCode {
			return nil, fmt.Errorf("invalid household settings")
		}
		if _, err := time.LoadLocation(h.Timezone); err != nil || h.Timezone == "" || h.Timezone == "Local" {
			return nil, fmt.Errorf("invalid household timezone")
		}
		if found != nil {
			return nil, fmt.Errorf("v1 supports one household")
		}
		h.dir = filepath.Join(root, dir.Name())
		found = &h
	}
	return found, nil
}

func (a *application) save(h household) error {
	return writeYAML(filepath.Join(a.householdDir(h), "household.yaml"), h, false)
}

// @impl ACC-001-AC5 ACC-002-AC4 ACC-002-AC6
func (a *application) currentYear(h household) int {
	zone, _ := time.LoadLocation(h.Timezone)
	return a.now().In(zone).Year()
}
