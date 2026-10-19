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
var pages = template.Must(template.ParseFS(web, "web/*.html"))

type household struct {
	Name       string `yaml:"name"`
	Timezone   string `yaml:"timezone"`
	ParentCode string `yaml:"parent_code"`
	FamilyCode string `yaml:"family_code"`
}

type application struct {
	dataDir string
	mu      sync.Mutex
	drafts  map[string]household
	assets  http.Handler
}

type view struct {
	Name, Timezone, Error, Ticket, ParentLink, FamilyLink, Role string
	Year                                                        int
}

func New(dataDir string) http.Handler {
	assets, err := fs.Sub(web, "web")
	if err != nil {
		panic(err)
	}
	return &application{dataDir: dataDir, drafts: make(map[string]household), assets: http.StripPrefix("/assets/", http.FileServer(http.FS(assets)))}
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
	h, err := a.load()
	if err != nil {
		http.Error(w, "Unable to read household data. Check the YAML files; no files were changed.", 500)
		return
	}
	if h != nil {
		if r.Method != "GET" || r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		code := r.URL.Query().Get("code")
		role := ""
		switch code {
		case h.ParentCode:
			role = "Parent"
		case h.FamilyCode:
			role = "Family"
		}
		if role == "" {
			http.NotFound(w, r)
			return
		}
		zone, _ := time.LoadLocation(h.Timezone)
		a.render(w, 200, "home.html", view{Name: h.Name, Timezone: h.Timezone, Role: role, Year: time.Now().In(zone).Year()})
		return
	}
	switch {
	case r.Method == "GET" && (r.URL.Path == "/" || r.URL.Path == "/setup"):
		a.render(w, 200, "setup.html", view{})
	case r.Method == "POST" && r.URL.Path == "/setup/prepare":
		a.prepare(w, r)
	case r.Method == "POST" && r.URL.Path == "/setup/finish":
		a.finish(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (a *application) render(w http.ResponseWriter, status int, name string, data view) {
	var body bytes.Buffer
	if err := pages.ExecuteTemplate(&body, name, data); err != nil {
		http.Error(w, "Unable to display this page", 500)
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
		http.Error(w, "Invalid setup form", 400)
		return
	}
	name := strings.TrimSpace(r.PostForm.Get("household"))
	timezone := r.PostForm.Get("timezone")
	data := view{Name: name, Timezone: timezone}
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
		http.Error(w, "Unable to generate access codes", 500)
		return
	}
	family, err := token()
	if err != nil {
		http.Error(w, "Unable to generate access codes", 500)
		return
	}
	ticket, err := token()
	if err != nil {
		http.Error(w, "Unable to prepare setup", 500)
		return
	}
	a.drafts[ticket] = household{Name: name, Timezone: timezone, ParentCode: parent, FamilyCode: family}
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

// @impl ACC-001-AC7 CON-004
func (a *application) finish(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid setup form", 400)
		return
	}
	h, ok := a.drafts[r.PostForm.Get("ticket")]
	if !ok {
		http.Error(w, "Setup preview is no longer available. Open setup again.", 400)
		return
	}
	if err := a.save(h); err != nil {
		http.Error(w, "Unable to save household setup. No setup was completed.", 500)
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
		decoder := yaml.NewDecoder(raw)
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
		if !validName(h.Name) || h.ParentCode == "" || h.FamilyCode == "" || h.ParentCode == h.FamilyCode {
			return nil, fmt.Errorf("invalid household settings")
		}
		if _, err := time.LoadLocation(h.Timezone); err != nil || h.Timezone == "" || h.Timezone == "Local" {
			return nil, fmt.Errorf("invalid household timezone")
		}
		if found != nil {
			return nil, fmt.Errorf("v1 supports one household")
		}
		found = &h
	}
	return found, nil
}

func (a *application) save(h household) error {
	dir := filepath.Join(a.dataDir, "households", h.Name)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	raw, err := yaml.Marshal(h)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".household-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(raw); err != nil {
		file.Close()
		return err
	}
	if err = file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), filepath.Join(dir, "household.yaml"))
}
