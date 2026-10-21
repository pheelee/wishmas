package app

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
	data "wishmas/internal/household"
)

// @test ACC-005-AC1 ACC-005-AC4
func TestDefaultSetupUsesPasswordInsteadOfLinks(t *testing.T) {
	a := New(t.TempDir())
	r := httptest.NewRequest("GET", "https://wishmas.test/", nil)
	w := httptest.NewRecorder()
	a.ServeHTTP(w, r)
	body := w.Body.String()
	if w.Code != 200 || !strings.Contains(body, `name="password"`) || !strings.Contains(body, `name="password_confirm"`) {
		t.Fatal("setup does not offer password and confirmation")
	}
	for _, obsolete := range []string{`name="timezone"`, `/setup/prepare`, "Create access links"} {
		if strings.Contains(body, obsolete) {
			t.Fatal("obsolete setup remains")
		}
	}
}

type authBrowser struct {
	handler                  http.Handler
	cookies                  map[string]string
	language, remote, scheme string
	headers                  http.Header
}

func newAuthBrowser(handler http.Handler) *authBrowser {
	return &authBrowser{handler: handler, cookies: map[string]string{}, language: "en-US", remote: "127.0.0.1:12345", scheme: "https", headers: http.Header{}}
}
func (b *authBrowser) send(t *testing.T, method, path string, values url.Values) *httptest.ResponseRecorder {
	t.Helper()
	var body io.Reader
	if values != nil {
		body = strings.NewReader(values.Encode())
	}
	r := httptest.NewRequest(method, b.scheme+"://wishmas.test"+path, body)
	r.RemoteAddr = b.remote
	r.Header.Set("Accept-Language", b.language)
	if values != nil {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for name, values := range b.headers {
		r.Header[name] = values
	}
	for name, value := range b.cookies {
		r.AddCookie(&http.Cookie{Name: name, Value: value})
	}
	w := httptest.NewRecorder()
	b.handler.ServeHTTP(w, r)
	for _, cookie := range w.Result().Cookies() {
		if cookie.MaxAge < 0 {
			delete(b.cookies, cookie.Name)
		} else {
			b.cookies[cookie.Name] = cookie.Value
		}
	}
	return w
}
func (b *authBrowser) get(t *testing.T, path string) *httptest.ResponseRecorder {
	return b.send(t, "GET", path, nil)
}
func (b *authBrowser) post(t *testing.T, path string, values url.Values) *httptest.ResponseRecorder {
	t.Helper()
	if values == nil {
		values = url.Values{}
	}
	values.Set("csrf", b.cookies[csrfCookie])
	return b.send(t, "POST", path, values)
}

const publicAuthPassword = "PublicFixture1!"

func authSetupValues() url.Values {
	return url.Values{"household": {"Family"}, "language": {"de-CH"}, "setup_language": {"en-US"}, "password": {publicAuthPassword}, "password_confirm": {publicAuthPassword}, "bulb_size": {"150"}, "snow_size": {"80"}, "snow_amount": {"24"}}
}
func authSetup(t *testing.T, dir string) (*householdApp, *authBrowser) {
	t.Helper()
	a := New(dir).(*householdApp)
	b := newAuthBrowser(a)
	if w := b.get(t, "/"); w.Code != 200 {
		t.Fatal("setup GET")
	}
	if w := b.post(t, "/setup/finish", authSetupValues()); w.Code != 303 {
		t.Fatalf("setup failed: %d", w.Code)
	}
	return a, b
}
func authStore(dir string) *data.Store {
	return data.NewStore(filepath.Join(dir, "households/Family/household.yaml"))
}

// @test ACC-005-AC1 ACC-005-AC2 ACC-005-AC3 ACC-005-AC4 ACC-005-AC5 ACC-005-AC6 CON-007-AC3 CON-007-AC4 INT-006
func TestSetupPersistsOnlyHashAndSecureSession(t *testing.T) {
	dir := t.TempDir()
	_, b := authSetup(t, dir)
	w := b.get(t, "/")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Admin-Zugang") {
		t.Fatal("setup did not auto-sign in using selected household language")
	}
	h, _, err := authStore(dir).Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(h.Users) != 0 || len(h.Lists) != 0 || len(h.Recipients) != 0 || len(h.Wishes) != 0 || len(h.Sessions) != 1 || h.Appearance.BulbSize != 150 || h.Language != "de-CH" {
		t.Fatal("setup state incorrect")
	}
	ok, err := data.CheckPassword(h.AdminPasswordHash, publicAuthPassword)
	if err != nil || !ok {
		t.Fatal("password hash not usable")
	}
	raw, err := os.ReadFile(filepath.Join(dir, "households/Family/household.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{publicAuthPassword, b.cookies[adminCookie], b.cookies[csrfCookie], "parent_code:", "family_code:", "timezone:"} {
		if forbidden != "" && strings.Contains(string(raw), forbidden) {
			t.Fatal("plaintext/obsolete setup data persisted")
		}
	}
	if h.Sessions[0].TokenHash != digest(b.cookies[adminCookie]) || h.Sessions[0].PasswordHashFingerprint != digest(h.AdminPasswordHash) {
		t.Fatal("session is not hashed and password-bound")
	}
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == adminCookie && (!cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.MaxAge <= 0 || cookie.Path != "/" || cookie.Domain != "") {
			t.Fatal("unsafe/nonpersistent admin cookie")
		}
	}
	for _, path := range []string{"/?code=invalid", "/?code=old-parent", "/?code=old-family"} {
		response := b.get(t, path)
		if response.Code != 200 || !strings.Contains(response.Body.String(), "Admin-Anmeldung") || strings.Contains(response.Body.String(), "Abmelden") {
			t.Fatal("invalid link must show household-language login even with admin cookie")
		}
	}
}

// @test ACC-005-AC7 INT-006-AC3 INT-006-AC4
func TestSetupErrorsClearPasswordsAndRetainOtherValues(t *testing.T) {
	for _, kind := range []string{"confirmation", "policy", "language", "appearance", "name", "noninteger"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			b := newAuthBrowser(New(dir))
			b.language = "de-DE"
			b.get(t, "/")
			values := authSetupValues()
			values.Set("setup_language", "de-CH")
			switch kind {
			case "confirmation":
				values.Set("password_confirm", "DifferentFixture1!")
			case "policy":
				values.Set("password", "weak")
				values.Set("password_confirm", "weak")
			case "language":
				values.Set("language", "fr-FR")
			case "appearance":
				values.Set("bulb_size", "201")
			case "noninteger":
				values.Set("bulb_size", "bad")
			case "name":
				values.Set("household", "bad name")
			}
			w := b.post(t, "/setup/finish", values)
			if w.Code != 400 || !strings.Contains(w.Body.String(), `lang="de-CH"`) {
				t.Fatal("invalid setup not localized")
			}
			for _, secret := range []string{publicAuthPassword, "DifferentFixture1!", `value="weak"`} {
				if strings.Contains(w.Body.String(), secret) {
					t.Fatal("password redisplayed")
				}
			}
			if !strings.Contains(w.Body.String(), `value="`+html.EscapeString(values.Get("household"))+`"`) || !strings.Contains(w.Body.String(), `name="snow_amount" type="range" min="8" max="48" value="24"`) {
				t.Fatal("non-password draft lost")
			}
			if _, err := os.Stat(filepath.Join(dir, "households/Family/household.yaml")); !os.IsNotExist(err) {
				t.Fatal("invalid setup saved")
			}
		})
	}
}

// @test ACC-006-AC4 ACC-006-AC5 CON-001-AC2 CON-007-AC3
func TestAdminSessionsSurviveRestartAndLogoutIsBrowserLocal(t *testing.T) {
	dir := t.TempDir()
	_, first := authSetup(t, dir)
	second := newAuthBrowser(New(dir))
	second.get(t, "/")
	if w := second.post(t, "/login", url.Values{"password": {publicAuthPassword}}); w.Code != 303 {
		t.Fatal("second login failed")
	}
	if first.cookies[adminCookie] == second.cookies[adminCookie] {
		t.Fatal("tokens shared")
	}
	first.handler = New(dir)
	second.handler = New(dir)
	if w := first.get(t, "/"); !strings.Contains(w.Body.String(), "Admin-Zugang") {
		t.Fatal("session lost on restart")
	}
	if w := second.post(t, "/logout", nil); w.Code != 303 {
		t.Fatal("logout failed")
	}
	if w := second.get(t, "/"); !strings.Contains(w.Body.String(), "Admin-Anmeldung") {
		t.Fatal("logged-out session retained")
	}
	if w := first.get(t, "/"); !strings.Contains(w.Body.String(), "Admin-Zugang") {
		t.Fatal("other browser logged out")
	}
	h, _, _ := authStore(dir).Load()
	if len(h.Sessions) != 1 {
		t.Fatal("logout did not remove exactly its session record")
	}
}

// @test ACC-006-AC6 CON-007-AC3
func TestPasswordHashEditPermanentlyInvalidatesOldSessions(t *testing.T) {
	dir := t.TempDir()
	_, first := authSetup(t, dir)
	second := newAuthBrowser(New(dir))
	second.get(t, "/")
	second.post(t, "/login", url.Values{"password": {publicAuthPassword}})
	oldToken := first.cookies[adminCookie]
	store := authStore(dir)
	h, revision, _ := store.Load()
	oldHash := h.AdminPasswordHash
	replacement, err := data.HashPassword("ReplacementFixture2!")
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Update(revision, func(h *data.Household) error { h.AdminPasswordHash = replacement; return nil }); err != nil {
		t.Fatal(err)
	}
	first.handler = New(dir)
	for _, b := range []*authBrowser{first, second} {
		if w := b.get(t, "/"); !strings.Contains(w.Body.String(), "Admin-Anmeldung") {
			t.Fatal("old session survived password change")
		}
	}
	h, revision, _ = store.Load()
	if len(h.Sessions) != 0 {
		t.Fatal("old session records not invalidated")
	}
	if err = store.Update(revision, func(h *data.Household) error { h.AdminPasswordHash = oldHash; return nil }); err != nil {
		t.Fatal(err)
	}
	first.cookies[adminCookie] = oldToken
	if w := first.get(t, "/"); !strings.Contains(w.Body.String(), "Admin-Anmeldung") {
		t.Fatal("old token revived after restoring hash")
	}
}

// @test CON-007-AC4 ACC-005-AC1 ACC-006-AC5
func TestCSRFOriginHTTPSAndStaleSetupProtection(t *testing.T) {
	dir := t.TempDir()
	a := New(dir)
	b := newAuthBrowser(a)
	b.get(t, "/")
	if w := b.send(t, "POST", "/setup/finish", authSetupValues()); w.Code != 403 {
		t.Fatal("missing CSRF accepted")
	}
	b.headers.Set("Origin", "https://attacker.example")
	if w := b.post(t, "/setup/finish", authSetupValues()); w.Code != 403 {
		t.Fatal("cross-origin setup accepted")
	}
	b.headers.Del("Origin")
	b.scheme = "http"
	if w := b.post(t, "/setup/finish", authSetupValues()); w.Code != 403 {
		t.Fatal("insecure setup accepted")
	}
	b.scheme = "https"
	b.headers.Set("Sec-Fetch-Site", "cross-site")
	if w := b.post(t, "/setup/finish", authSetupValues()); w.Code != 403 {
		t.Fatal("cross-site setup accepted")
	}
	b.headers.Del("Sec-Fetch-Site")
	if w := b.post(t, "/setup/finish", authSetupValues()); w.Code != 303 {
		t.Fatal("valid setup failed")
	}
	path := filepath.Join(dir, "households/Family/household.yaml")
	before, _ := os.ReadFile(path)
	for _, path := range []string{"/setup/finish", "/access/regenerate", "/people/add", "/wishes/save"} {
		if w := b.post(t, path, authSetupValues()); w.Code == 200 || w.Code == 303 {
			t.Fatal("obsolete/unimplemented mutation accepted")
		}
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("obsolete routes changed data")
	}
	b.headers.Set("Origin", "https://attacker.example")
	if w := b.post(t, "/logout", nil); w.Code != 403 {
		t.Fatal("cross-origin logout accepted")
	}
}

// @test ACC-006-AC8 ACC-006-AC9
func TestUserLinkIdentityCannotAcquireAdminSessionPrivileges(t *testing.T) {
	dir := t.TempDir()
	_, b := authSetup(t, dir)
	store := authStore(dir)
	_, rev, _ := store.Load()
	code := strings.Repeat("a", 64)
	if err := store.Update(rev, func(h *data.Household) error {
		h.Users = []data.Person{{ID: "u1", Name: "Grand Dad"}}
		h.Lists = []data.List{{ID: "l1", Name: "Birthday party", UserLinks: []data.UserLink{{UserID: "u1", Code: code}}}}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	path := "/?code=" + code
	w := b.get(t, path)
	body := w.Body.String()
	if w.Code != 200 || !strings.Contains(body, "Grand Dad") || !strings.Contains(body, "Birthday party") || strings.Contains(body, "Admin-Zugang") || strings.Contains(body, "Abmelden") {
		t.Fatal("link identity did not override admin session")
	}
	before, _, _ := store.Load()
	if w := b.post(t, "/logout?code="+code, nil); w.Code != 403 {
		t.Fatal("user link authorized admin logout")
	}
	after, _, _ := store.Load()
	if len(before.Sessions) != len(after.Sessions) {
		t.Fatal("user action changed admin sessions")
	}
	if w := b.get(t, "/"); !strings.Contains(w.Body.String(), "Admin-Zugang") {
		t.Fatal("separate admin session lost")
	}
}

// @test ACC-006-AC7 ACC-006-AC10 CON-007-AC5
func TestLoginLimitUsesTrustedSideOfForwardedChainAndResetsOnRestart(t *testing.T) {
	dir := t.TempDir()
	_, _ = authSetup(t, dir)
	proxies, err := ParseTrustedProxies("10.0.0.2/32,10.0.0.3")
	if err != nil {
		t.Fatal(err)
	}
	a := NewWithOptions(dir, Options{TrustedProxies: proxies}).(*householdApp)
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
	a.now = func() time.Time { return now }
	b := newAuthBrowser(a)
	b.remote = "10.0.0.2:1234"
	b.scheme = "http"
	b.headers.Set("X-Forwarded-Proto", "https")
	b.headers.Set("X-Forwarded-For", "203.0.113.99, 198.51.100.7, 10.0.0.3")
	b.get(t, "/")
	for range 5 {
		if w := b.post(t, "/login", url.Values{"password": {"WrongFixture1!"}}); w.Code != 401 {
			t.Fatal("expected five failed attempts")
		}
	}
	b.headers.Set("X-Forwarded-For", "203.0.113.88, 198.51.100.7, 10.0.0.3")
	if w := b.post(t, "/login", url.Values{"password": {publicAuthPassword}}); w.Code != 429 {
		t.Fatal("spoofed leftmost IP bypassed login limit")
	}
	b.headers.Set("X-Forwarded-For", "198.51.100.8, 10.0.0.3")
	if w := b.post(t, "/login", url.Values{"password": {publicAuthPassword}}); w.Code != 303 {
		t.Fatal("other client IP blocked")
	}
	b.headers.Set("X-Forwarded-For", "198.51.100.7, 10.0.0.3")
	now = now.Add(15 * time.Minute)
	if w := b.post(t, "/login", url.Values{"password": {publicAuthPassword}}); w.Code != 303 {
		t.Fatal("limit did not expire")
	}
	other := newAuthBrowser(a)
	other.remote = "192.0.2.1:99"
	other.get(t, "/")
	for range 5 {
		other.headers.Set("X-Forwarded-For", fmt.Sprintf("198.51.100.%d", len(a.attempts)))
		if w := other.post(t, "/login", url.Values{"password": {"WrongFixture1!"}}); w.Code != 401 {
			t.Fatal("untrusted failure counting")
		}
	}
	if w := other.post(t, "/login", url.Values{"password": {publicAuthPassword}}); w.Code != 429 {
		t.Fatal("untrusted forwarded headers bypassed limit")
	}
	other.handler = New(dir)
	if w := other.post(t, "/login", url.Values{"password": {publicAuthPassword}}); w.Code != 303 {
		t.Fatal("restart did not reset in-memory block")
	}
}

// @test CON-004-AC6 INT-007-AC2
func TestNewRuntimeRejectsLegacyAndInvalidDataWithoutWrites(t *testing.T) {
	for _, raw := range []string{"name: Family\ntimezone: Europe/Zurich\nparent_code: old-parent\nfamily_code: old-family\n", "schema_version: 2\nname: [broken"} {
		dir := t.TempDir()
		path := filepath.Join(dir, "households/Family/household.yaml")
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		b := newAuthBrowser(New(dir))
		b.language = "de-CH"
		w := b.get(t, "/")
		after, _ := os.ReadFile(path)
		if w.Code != 500 || !strings.Contains(w.Body.String(), `lang="en-US"`) || !bytes.Equal(after, []byte(raw)) {
			t.Fatal("legacy/invalid data accepted or rewritten")
		}
	}
}

// @test CON-007-AC5 ACC-006-AC10
func TestProxyTrustContractRejectsSpoofingAndInvalidConfiguration(t *testing.T) {
	for _, bad := range []string{"not-a-proxy", "0.0.0.0/0", "::/0", "10.0.0.2,", "::ffff:0.0.0.0/96"} {
		if _, err := ParseTrustedProxies(bad); err == nil {
			t.Fatal("invalid/trust-all proxy configuration accepted")
		}
	}
	proxies, err := ParseTrustedProxies("10.0.0.2/32, fd00::2, ::ffff:10.0.0.3/128")
	if err != nil {
		t.Fatal(err)
	}
	a := NewWithOptions(t.TempDir(), Options{TrustedProxies: proxies}).(*householdApp)
	cases := []struct{ remote, chain, want string }{{"192.0.2.1:10", "203.0.113.4", "192.0.2.1"}, {"10.0.0.2:10", "forged, 198.51.100.1", "198.51.100.1"}, {"10.0.0.2:10", "203.0.113.4, 198.51.100.1, 10.0.0.3", "198.51.100.1"}, {"[fd00::2]:10", "2001:db8::1", "2001:db8::1"}, {"10.0.0.2:10", "invalid", "10.0.0.2"}}
	for _, tc := range cases {
		r := httptest.NewRequest("GET", "http://wishmas.test/", nil)
		r.RemoteAddr = tc.remote
		r.Header.Set("X-Forwarded-For", tc.chain)
		r.Header.Set("X-Forwarded-Proto", "https")
		if a.clientIP(r).String() != tc.want {
			t.Fatal("forwarded client chain trusted incorrectly")
		}
		if a.https(r) == (tc.remote == "192.0.2.1:10") {
			t.Fatal("forwarded HTTPS trusted incorrectly")
		}
	}
}

// @test ACC-005-AC1 ACC-005-AC6 CON-004-AC4
func TestConcurrentSetupCannotOverwriteOtherSetup(t *testing.T) {
	dir := t.TempDir()
	first := newAuthBrowser(New(dir))
	second := newAuthBrowser(New(dir))
	first.get(t, "/")
	second.get(t, "/")
	var wg sync.WaitGroup
	codes := make([]int, 2)
	for i, b := range []*authBrowser{first, second} {
		i, b := i, b
		wg.Go(func() { codes[i] = b.post(t, "/setup/finish", authSetupValues()).Code })
	}
	wg.Wait()
	wins := 0
	for _, code := range codes {
		if code == 303 {
			wins++
		}
	}
	if wins != 1 {
		t.Fatal("concurrent setup did not have exactly one winner")
	}
	h, _, err := authStore(dir).Load()
	if err != nil || len(h.Sessions) != 1 {
		t.Fatal("concurrent setup corrupted household")
	}
}

// @test ACC-006-AC5 CON-004-AC4 CON-007-AC3
func TestSessionWriteFailuresPreserveSessionsAndCookies(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission-failure test requires non-root")
	}
	dir := t.TempDir()
	_, first := authSetup(t, dir)
	second := newAuthBrowser(New(dir))
	second.get(t, "/")
	path := filepath.Join(dir, "households/Family/household.yaml")
	before, _ := os.ReadFile(path)
	parent := filepath.Dir(path)
	if err := os.Chmod(parent, 0500); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(parent, 0700)
	if w := second.post(t, "/login", url.Values{"password": {publicAuthPassword}}); w.Code != 500 {
		t.Fatal("failed session write reported successful login")
	}
	if second.cookies[adminCookie] != "" {
		t.Fatal("failed login issued token")
	}
	token := first.cookies[adminCookie]
	if w := first.post(t, "/logout", nil); w.Code != 500 {
		t.Fatal("failed logout reported success")
	}
	if first.cookies[adminCookie] != token {
		t.Fatal("failed logout cleared browser cookie")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("failed session write changed stored data")
	}
}

// @test CON-007-AC5 ACC-006-AC10
func TestForwardedHeadersUseAllValuesAndRejectAmbiguousProtocol(t *testing.T) {
	proxies, _ := ParseTrustedProxies("10.0.0.2")
	a := NewWithOptions(t.TempDir(), Options{TrustedProxies: proxies}).(*householdApp)
	r := httptest.NewRequest("GET", "http://wishmas.test/", nil)
	r.RemoteAddr = "10.0.0.2:12"
	r.Header["X-Forwarded-For"] = []string{"203.0.113.99", "198.51.100.4"}
	if got := a.clientIP(r).String(); got != "198.51.100.4" {
		t.Fatal("duplicate forwarded header bypasses trusted-chain parsing")
	}
	r.Header["X-Forwarded-Proto"] = []string{"https", "http"}
	if a.https(r) {
		t.Fatal("ambiguous protocol header trusted")
	}
}
