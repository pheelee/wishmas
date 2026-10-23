package app

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"html/template"
	"io/fs"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	data "github.com/pheelee/wishmas/internal/household"
)

const adminCookie = "__Host-wishmas-admin"
const csrfCookie = "__Host-wishmas-csrf"
const persistentCookieAge = 400 * 24 * 60 * 60

type Options struct{ TrustedProxies []netip.Prefix }

// ParseTrustedProxies accepts explicit Traefik peer addresses/CIDRs. No header
// trust is inferred from loopback/private addresses or enabled by default.
func ParseTrustedProxies(raw string) ([]netip.Prefix, error) {
	result := []netip.Prefix{}
	if strings.TrimSpace(raw) == "" {
		return result, nil
	}
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		prefix, err := netip.ParsePrefix(entry)
		if err != nil {
			addr, e := netip.ParseAddr(entry)
			if e != nil {
				return nil, errors.New("invalid trusted proxy address/CIDR")
			}
			addr = addr.Unmap()
			prefix = netip.PrefixFrom(addr, addr.BitLen())
		}
		if prefix.Addr().Is4In6() {
			if prefix.Bits() < 96 {
				return nil, errors.New("invalid mapped proxy network")
			}
			prefix = netip.PrefixFrom(prefix.Addr().Unmap(), prefix.Bits()-96)
		}
		if prefix.Bits() == 0 {
			return nil, errors.New("trust-all proxy networks are not permitted")
		}
		result = append(result, prefix.Masked())
	}
	return result, nil
}

type loginAttempts struct {
	Failures     []time.Time
	BlockedUntil time.Time
}
type householdApp struct {
	pictureClient *http.Client
	dataDir       string
	now           func() time.Time
	mu            sync.Mutex
	assets        http.Handler
	proxies       []netip.Prefix
	attempts      map[netip.Addr]loginAttempts
	lastSweep     time.Time
}
type adminView struct {
	Admin                                                                  bool
	ContextList, ContextRecipient, ContextURL, FilterList, FilterRecipient string
	FilterUnassigned                                                       bool
	CurrentRecipient                                                       *data.Person
	AssignmentChoices                                                      []assignmentChoice
	InputAppearance                                                        *appearance
	view
	Revision, DraftKind, DraftName, OpenModal, Code       string
	CurrentList                                           *data.List
	ListRows, RecipientRows, UserRows, SelectedRecipients []managementRow
	AvailableRecipients                                   []data.Person
	ListLinks                                             []listLinkView
	ReadRecipients                                        []recipientReadView
	CSRF                                                  string
	Lists                                                 []data.List
	Recipients, Users                                     []data.Person
}

// @impl ACC-005 ACC-006 CON-007: production defaults to password access only.
func New(dataDir string) http.Handler { return NewWithOptions(dataDir, Options{}) }
func NewWithOptions(dataDir string, options Options) http.Handler {
	assets, err := fs.Sub(web, "web")
	if err != nil {
		panic(err)
	}
	return &householdApp{pictureClient: newPictureClient(), dataDir: dataDir, now: time.Now, assets: http.StripPrefix("/assets/", http.FileServer(http.FS(assets))), proxies: append([]netip.Prefix(nil), options.TrustedProxies...), attempts: map[netip.Addr]loginAttempts{}}
}

func (a *householdApp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "same-origin")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", "frame-ancestors 'none'")
	if r.URL.Path == "/assets/app.css" || r.URL.Path == "/assets/app.js" {
		a.assets.ServeHTTP(w, r)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	h, store, revision, err := a.load()
	if err != nil {
		a.render(w, 500, "error.html", adminView{view: view{Locale: "en-US", Error: "HR103"}})
		return
	}
	locale := browserLocale(r.Header.Get("Accept-Language"))
	v := adminView{view: view{Locale: locale, Language: locale, Appearance: defaultAppearance()}}
	if h != nil {
		v = a.householdView(*h)
		if err = a.pruneSessions(store, h, &revision); err != nil {
			v.Error = "HR105"
			a.render(w, 500, "error.html", v)
			return
		}
	}
	if r.Method == "POST" {
		r.Body = http.MaxBytesReader(w, r.Body, 16384)
		if err = r.ParseForm(); err != nil {
			v.Error = "HR112"
			a.render(w, 400, "error.html", v)
			return
		}
		if !a.validMutation(r) {
			v.Error = "HR112"
			a.render(w, 403, "error.html", v)
			return
		}
	}
	nonce, err := a.csrf(w, r)
	if err != nil {
		v.Error = "E07"
		a.render(w, 500, "error.html", v)
		return
	}
	v.CSRF = nonce
	v.Revision = revision
	if h == nil {
		switch {
		case r.Method == "GET" && (r.URL.Path == "/" || r.URL.Path == "/setup"):
			a.render(w, 200, "admin-setup.html", v)
		case r.Method == "POST" && r.URL.Path == "/setup/finish":
			a.setup(w, r, v)
		default:
			v.Error = "E02"
			a.render(w, 404, "error.html", v)
		}
		return
	}
	// Explicit link identity always wins over an unrelated admin cookie.
	if r.URL.Query().Has("code") {
		code := r.URL.Query().Get("code")
		for _, list := range h.Lists {
			for _, link := range list.UserLinks {
				if constantEqual(link.Code, code) {
					if r.Method == "POST" && r.URL.Path == "/claims/toggle" {
						a.householdClaim(w, r, *h, store, revision, v, list.ID, link.UserID, code)
						return
					}
					if r.Method == "GET" && r.URL.Path == "/wish-picture" {
						a.householdPicture(w, r, *h, store, list.ID, v)
						return
					}
					if r.Method != "GET" || r.URL.Path != "/" {
						v.Error = "HR112"
						a.render(w, 403, "error.html", v)
						return
					}
					a.userList(w, r, *h, list, link.UserID, code, store, v)
					return
				}
			}
		}
		// Invalid/replaced link never falls through into privileged session access.
		if r.Method == "POST" && r.URL.Path == "/login" {
			a.login(w, r, *h, store, revision, v)
			return
		}
		a.render(w, 200, "admin-login.html", v)
		return
	}
	session, ok := a.session(r, *h)
	if r.Method == "POST" && r.URL.Path == "/login" {
		a.login(w, r, *h, store, revision, v)
		return
	}
	if !ok {
		if _, err := r.Cookie(adminCookie); err == nil {
			a.clearCookie(w, adminCookie)
			v.Error = "HR12"
		}
		status := 200
		if r.Method != "GET" {
			status = 401
		}
		a.render(w, status, "admin-login.html", v)
		return
	}
	if r.Method == "GET" && r.URL.Path == "/wish-picture" {
		a.householdPicture(w, r, *h, store, "", v)
		return
	}
	if r.Method == "POST" && r.URL.Path == "/claims/toggle" {
		a.householdClaim(w, r, *h, store, revision, v, "", "", "")
		return
	}
	if r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/wishes/") {
		a.householdWishAction(w, r, *h, store, revision, v)
		return
	}
	// @impl WLS-007-AC4 WLS-008-AC1 ACC-007-AC1: these routes require the verified admin session, with no date gates.
	populateManagement(&v, *h, revision, "")
	if r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/manage/") {
		a.manage(w, r, store, revision, v)
		return
	}
	if r.Method == "GET" && r.URL.Path == "/list" {
		if !populateManagement(&v, *h, revision, r.URL.Query().Get("id")) || v.CurrentList == nil {
			v.Error = "HR111"
			a.render(w, 404, "error.html", v)
			return
		}
		if !prepareHouseholdWishes(&v, *h, store, r) {
			v.Error = "HR111"
			a.render(w, 404, "error.html", v)
			return
		}
		a.render(w, 200, "admin-list.html", v)
		return
	}
	if r.Method == "POST" && r.URL.Path == "/logout" {
		if err = store.Update(revision, func(current *data.Household) error {
			kept := current.Sessions[:0]
			for _, record := range current.Sessions {
				if !constantEqual(record.TokenHash, session) {
					kept = append(kept, record)
				}
			}
			current.Sessions = kept
			return nil
		}); err != nil {
			v.Error = "HR17"
			a.render(w, 500, "admin-home.html", v)
			return
		}
		a.clearCookie(w, adminCookie)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if r.Method == "GET" && (r.URL.Path == "/" || r.URL.Path == "/login") {
		if cookie, err := r.Cookie(adminCookie); err == nil {
			a.setCookie(w, adminCookie, cookie.Value)
		}
		if !prepareHouseholdWishes(&v, *h, store, r) {
			v.Error = "HR111"
			a.render(w, 404, "error.html", v)
			return
		}
		a.render(w, 200, "admin-home.html", v)
		return
	}
	v.Error = "E02"
	a.render(w, 404, "error.html", v)
}

func (a *householdApp) householdView(h data.Household) adminView {
	theme := defaultAppearance()
	if h.Appearance != nil {
		theme = &appearance{h.Appearance.BulbSize, h.Appearance.SnowSize, h.Appearance.SnowAmount}
	}
	return adminView{view: view{Locale: h.Language, Language: h.Language, Name: h.Name, Appearance: theme}, Lists: h.Lists, Recipients: h.Recipients, Users: h.Users}
}

// @impl CON-004-AC6 INT-007: legacy configuration is rejected, never migrated.
func (a *householdApp) load() (*data.Household, *data.Store, string, error) {
	root := filepath.Join(a.dataDir, "households")
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return nil, nil, "", nil
	}
	if err != nil {
		return nil, nil, "", err
	}
	var found *data.Household
	var store *data.Store
	var revision string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(root, entry.Name(), "household.yaml")
		if _, err := os.Lstat(path); os.IsNotExist(err) {
			continue
		}
		if found != nil {
			return nil, nil, "", errors.New("one household supported")
		}
		candidate := data.NewStore(path)
		h, rev, err := candidate.Load()
		if err != nil {
			return nil, nil, "", err
		}
		found = &h
		store = candidate
		revision = rev
	}
	return found, store, revision, nil
}

// @impl ACC-005 ACC-005-AC7 INT-006
func (a *householdApp) setup(w http.ResponseWriter, r *http.Request, v adminView) {
	if language := r.PostForm.Get("setup_language"); supportedLocale(language) {
		v.Locale = language
	}
	v.Name = strings.TrimSpace(r.PostForm.Get("household"))
	v.Language = r.PostForm.Get("language")
	theme, themeErr := parseAppearance(r.PostForm)
	v.Appearance = theme
	v.InputAppearance = theme
	reject := func(message string) { v.Error = message; a.render(w, 400, "admin-setup.html", v) }
	if !validName(v.Name) {
		reject("E09")
		return
	}
	if !supportedLocale(v.Language) {
		reject("E47")
		return
	}
	if themeErr != nil {
		reject(themeErr.Error())
		return
	}
	password := r.PostForm.Get("password")
	if password != r.PostForm.Get("password_confirm") {
		reject("HR05")
		return
	}
	if data.ValidatePassword(password) != nil {
		reject("HR06")
		return
	}
	encoded, err := data.HashPassword(password)
	if err != nil {
		v.Error = "E14"
		a.render(w, 500, "admin-setup.html", v)
		return
	}
	token, err := data.NewToken()
	if err != nil {
		v.Error = "E14"
		a.render(w, 500, "admin-setup.html", v)
		return
	}
	newCSRF, err := data.NewToken()
	if err != nil {
		v.Error = "E14"
		a.render(w, 500, "admin-setup.html", v)
		return
	}
	h := data.Household{SchemaVersion: 2, Name: v.Name, Language: v.Language, Appearance: &data.Appearance{BulbSize: theme.BulbSize, SnowSize: theme.SnowSize, SnowAmount: theme.SnowAmount}, AdminPasswordHash: encoded, Sessions: []data.Session{{TokenHash: digest(token), PasswordHashFingerprint: digest(encoded)}}, Users: []data.Person{}, Recipients: []data.Person{}, Lists: []data.List{}, Wishes: []data.Wish{}}
	root := filepath.Join(a.dataDir, "households")
	if err = os.MkdirAll(root, 0700); err != nil {
		v.Error = "E14"
		a.render(w, 500, "admin-setup.html", v)
		return
	}
	dir := filepath.Join(root, h.Name)
	if err = os.Mkdir(dir, 0700); err != nil {
		v.Error = "E14"
		a.render(w, 500, "admin-setup.html", v)
		return
	}
	if err = data.NewStore(filepath.Join(dir, "household.yaml")).Create(h); err != nil {
		_ = os.Remove(dir)
		v.Error = "E14"
		a.render(w, 500, "admin-setup.html", v)
		return
	}
	a.setCookie(w, adminCookie, token)
	a.setCookie(w, csrfCookie, newCSRF)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// @impl ACC-006-AC1 ACC-006-AC4 ACC-006-AC7 CON-007
func (a *householdApp) login(w http.ResponseWriter, r *http.Request, h data.Household, store *data.Store, revision string, v adminView) {
	now := a.now()
	if a.lastSweep.IsZero() || now.Sub(a.lastSweep) >= time.Minute {
		cutoff := now.Add(-15 * time.Minute)
		for ip, old := range a.attempts {
			if !now.Before(old.BlockedUntil) && (len(old.Failures) == 0 || !old.Failures[len(old.Failures)-1].After(cutoff)) {
				delete(a.attempts, ip)
			}
		}
		a.lastSweep = now
	}
	ip := a.clientIP(r)
	attempt := a.attempts[ip]
	if now.Before(attempt.BlockedUntil) {
		w.Header().Set("Retry-After", strconv.Itoa(int((attempt.BlockedUntil.Sub(now)+time.Second-1)/time.Second)))
		v.Error = "HR11"
		a.render(w, 429, "admin-login.html", v)
		return
	}
	kept := attempt.Failures[:0]
	for _, failure := range attempt.Failures {
		if failure.After(now.Add(-15 * time.Minute)) {
			kept = append(kept, failure)
		}
	}
	attempt.Failures = kept
	ok, err := data.CheckPassword(h.AdminPasswordHash, r.PostForm.Get("password"))
	if err != nil {
		v.Error = "HR13"
		a.render(w, 500, "admin-login.html", v)
		return
	}
	if !ok {
		attempt.Failures = append(attempt.Failures, now)
		if len(attempt.Failures) >= 5 {
			attempt.BlockedUntil = now.Add(15 * time.Minute)
		}
		a.attempts[ip] = attempt
		v.Error = "HR10"
		a.render(w, 401, "admin-login.html", v)
		return
	}
	token, err := data.NewToken()
	if err != nil {
		v.Error = "HR13"
		a.render(w, 500, "admin-login.html", v)
		return
	}
	newCSRF, err := data.NewToken()
	if err != nil {
		v.Error = "HR13"
		a.render(w, 500, "admin-login.html", v)
		return
	}
	old := ""
	if cookie, e := r.Cookie(adminCookie); e == nil {
		old = digest(cookie.Value)
	}
	err = store.Update(revision, func(current *data.Household) error {
		sessions := current.Sessions[:0]
		for _, record := range current.Sessions {
			if !constantEqual(record.TokenHash, old) {
				sessions = append(sessions, record)
			}
		}
		current.Sessions = append(sessions, data.Session{TokenHash: digest(token), PasswordHashFingerprint: digest(current.AdminPasswordHash)})
		return nil
	})
	if err != nil {
		v.Error = "HR13"
		a.render(w, 500, "admin-login.html", v)
		return
	}
	a.setCookie(w, adminCookie, token)
	a.setCookie(w, csrfCookie, newCSRF)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// @impl ACC-006-AC6: persisted invalidation also prevents reviving old sessions
// by subsequently restoring a previous hash after a change has been observed.
func (a *householdApp) pruneSessions(store *data.Store, h *data.Household, revision *string) error {
	fingerprint := digest(h.AdminPasswordHash)
	sessions := []data.Session{}
	for _, s := range h.Sessions {
		if constantEqual(s.PasswordHashFingerprint, fingerprint) {
			sessions = append(sessions, s)
		}
	}
	if len(sessions) == len(h.Sessions) {
		return nil
	}
	if err := store.Update(*revision, func(current *data.Household) error { current.Sessions = sessions; return nil }); err != nil {
		return err
	}
	fresh, rev, err := store.Load()
	if err != nil {
		return err
	}
	*h = fresh
	*revision = rev
	return nil
}
func (a *householdApp) session(r *http.Request, h data.Household) (string, bool) {
	cookie, err := r.Cookie(adminCookie)
	if err != nil || !validToken(cookie.Value) {
		return "", false
	}
	hashed := digest(cookie.Value)
	for _, s := range h.Sessions {
		if constantEqual(s.TokenHash, hashed) && constantEqual(s.PasswordHashFingerprint, digest(h.AdminPasswordHash)) {
			return hashed, true
		}
	}
	return "", false
}
func digest(s string) string { sum := sha256.Sum256([]byte(s)); return hex.EncodeToString(sum[:]) }
func validToken(s string) bool {
	if len(s) != 64 {
		return false
	}
	raw, err := hex.DecodeString(s)
	return err == nil && len(raw) == 32
}
func constantEqual(a, b string) bool { return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1 }
func (a *householdApp) setCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{Name: name, Value: value, Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: persistentCookieAge, Expires: a.now().Add(persistentCookieAge * time.Second)})
}
func (a *householdApp) clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{Name: name, Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: -1, Expires: time.Unix(1, 0)})
}
func (a *householdApp) csrf(w http.ResponseWriter, r *http.Request) (string, error) {
	if cookie, err := r.Cookie(csrfCookie); err == nil && validToken(cookie.Value) {
		return cookie.Value, nil
	}
	token, err := data.NewToken()
	if err != nil {
		return "", err
	}
	a.setCookie(w, csrfCookie, token)
	return token, nil
}

// @impl CON-007-AC4: double-submit nonce plus origin/fetch-site checks.
func (a *householdApp) validMutation(r *http.Request) bool {
	if !a.https(r) {
		return false
	}
	cookie, err := r.Cookie(csrfCookie)
	if err != nil || !validToken(cookie.Value) || !constantEqual(cookie.Value, r.PostForm.Get("csrf")) {
		return false
	}
	if site := r.Header.Get("Sec-Fetch-Site"); site == "cross-site" {
		return false
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		u, err := url.Parse(origin)
		if err != nil || u.Scheme != "https" || !strings.EqualFold(u.Host, r.Host) || u.User != nil || u.Path != "" {
			return false
		}
	}
	return true
}
func (a *householdApp) trusted(ip netip.Addr) bool {
	for _, prefix := range a.proxies {
		if prefix.Bits() > 0 && prefix.Contains(ip.Unmap()) {
			return true
		}
	}
	return false
}
func peerIP(r *http.Request) netip.Addr {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip, _ := netip.ParseAddr(host)
	return ip.Unmap()
}

// @impl ACC-006-AC10 CON-007-AC5: walk from the trusted peer toward the client.
func (a *householdApp) clientIP(r *http.Request) netip.Addr {
	ip := peerIP(r)
	if !a.trusted(ip) {
		return ip
	}
	forwarded := strings.Split(strings.Join(r.Header.Values("X-Forwarded-For"), ","), ",")
	for i := len(forwarded) - 1; i >= 0; i-- {
		next, err := netip.ParseAddr(strings.TrimSpace(forwarded[i]))
		if err != nil {
			return peerIP(r)
		}
		ip = next.Unmap()
		if !a.trusted(ip) {
			return ip
		}
	}
	return ip
}
func (a *householdApp) https(r *http.Request) bool {
	return r.TLS != nil || (a.trusted(peerIP(r)) && len(r.Header.Values("X-Forwarded-Proto")) == 1 && r.Header.Get("X-Forwarded-Proto") == "https")
}
func (a *householdApp) render(w http.ResponseWriter, status int, name string, v adminView) {
	if v.Locale == "" {
		v.Locale = "en-US"
	}
	v.ReferrerPolicy = "same-origin"
	v.Error = translate(v.Locale, v.Error)
	if v.Appearance == nil {
		v.Appearance = defaultAppearance()
	}
	if v.InputAppearance == nil {
		v.InputAppearance = v.Appearance
	}
	decorateView(&v.view)
	localized, err := pages.Clone()
	if err != nil {
		http.Error(w, translate(v.Locale, "E07"), 500)
		return
	}
	localized = localized.Funcs(template.FuncMap{"t": func(key string, args ...any) string { return translate(v.Locale, key, args...) }, "price": func(raw string) string { return formatPrice(v.Locale, raw) }})
	var body bytes.Buffer
	if err = localized.ExecuteTemplate(&body, name, v); err != nil {
		http.Error(w, translate(v.Locale, "E07"), 500)
		return
	}
	w.Header().Set("Content-Language", v.Locale)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = body.WriteTo(w)
}
