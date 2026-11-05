package app

import (
	"encoding/xml"
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

	data "github.com/pheelee/wishmas/internal/household"
)

type appS3Fixture struct {
	mu        sync.Mutex
	status    int
	objects   map[string][]byte
	puts      int
	putStatus int
}

func (f *appS3Fixture) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.status != 0 {
		w.WriteHeader(f.status)
		return
	}
	if r.URL.Query().Get("list-type") == "2" {
		type item struct{ Key string }
		out := struct {
			XMLName  xml.Name `xml:"ListBucketResult"`
			Contents []item
		}{}
		for k := range f.objects {
			out.Contents = append(out.Contents, item{k})
		}
		xml.NewEncoder(w).Encode(out)
		return
	}
	key := strings.TrimPrefix(r.URL.Path, "/bucket/")
	if r.Method == "PUT" {
		f.objects[key], _ = io.ReadAll(r.Body)
		f.puts++
		if f.putStatus != 0 {
			w.WriteHeader(f.putStatus)
		}
		return
	}
	raw, ok := f.objects[key]
	if !ok {
		w.WriteHeader(404)
		return
	}
	w.Write(raw)
}
func s3App(t *testing.T) (*appS3Fixture, *householdApp, *data.S3Backend) {
	t.Helper()
	f := &appS3Fixture{objects: map[string][]byte{}}
	srv := httptest.NewTLSServer(f)
	t.Cleanup(srv.Close)
	b, err := data.NewS3Backend(t.TempDir(), data.StorageConfig{Endpoint: srv.URL, Bucket: "bucket", AccessKey: "fixture", SecretKey: "fixture-secret", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	a := NewWithOptions(t.TempDir(), Options{Storage: b}).(*householdApp)
	return f, a, b
}

// @test CON-009-AC6 CON-009-AC9 CON-009-AC10 CON-009-AC11 CON-009-AC12 ACC-008-AC1 ACC-008-AC2 ACC-008-AC3 ACC-008-AC4 ACC-008-AC5
func TestS3SetupCachedAccessAndLogout(t *testing.T) {
	f, a, b := s3App(t)
	admin := newAuthBrowser(a)
	admin.get(t, "/")
	values := url.Values{"household": {"Family"}, "language": {"de-CH"}, "setup_language": {"en-US"}, "password": {"PublicFixture1!"}, "password_confirm": {"PublicFixture1!"}, "bulb_size": {"120"}, "snow_size": {"100"}, "snow_amount": {"16"}}
	if w := admin.post(t, "/setup/finish", values); w.Code != 303 {
		t.Fatal("remote setup failed", w.Code)
	}
	token := admin.cookies[adminCookie]
	if w := admin.get(t, "/"); w.Code != 200 || !strings.Contains(w.Body.String(), "/logout") {
		t.Fatal("remote admin access failed")
	}
	f.mu.Lock()
	f.status = 503
	f.mu.Unlock()
	w := admin.get(t, "/")
	if w.Code != 200 || !strings.Contains(w.Body.String(), translate("de-CH", "HR113")) || !strings.Contains(w.Body.String(), "disabled") {
		t.Fatal("cached readonly UI missing")
	}
	h, s, rev, err := b.Load()
	if err != nil || h == nil {
		t.Fatal(err)
	}
	w = admin.post(t, "/manage/list/create", url.Values{"revision": {rev}, "name": {"Offline list"}})
	if w.Code == 303 || !strings.Contains(w.Body.String(), translate("de-CH", "HR113")) {
		t.Fatal("offline mutation accepted or not explained", w.Code)
	}
	other := newAuthBrowser(a)
	other.get(t, "/")
	w = other.post(t, "/login", url.Values{"password": {"PublicFixture1!"}})
	if w.Code == 303 || !strings.Contains(w.Body.String(), translate("de-CH", "HR114")) {
		t.Fatal("offline login accepted")
	}
	w = admin.post(t, "/logout", nil)
	if w.Code != 303 || admin.cookies[adminCookie] != "" || !s.SessionRevoked(digest(token)) {
		t.Fatal("offline logout failed", w.Code)
	}
	admin.cookies[adminCookie] = token
	if w = admin.get(t, "/"); strings.Contains(w.Body.String(), "action=\"/logout\"") {
		t.Fatal("revoked session reused")
	}
	restarted := NewWithOptions(t.TempDir(), Options{Storage: b})
	admin.handler = restarted
	if w = admin.get(t, "/"); strings.Contains(w.Body.String(), "action=\"/logout\"") {
		t.Fatal("revoked session survived restart")
	}
	f.mu.Lock()
	f.status = 0
	f.mu.Unlock()
	h, _, _, err = b.Load()
	if err != nil || len(h.Sessions) != 0 || len(h.Lists) != 0 {
		t.Fatal("reconciliation or blocked data integrity failed", err)
	}
}

// @test ACC-008-AC1 ACC-008-AC3 ACC-008-AC10 CON-009-AC12 CON-009-AC13
func TestS3CachedLinkPrivacyMutationsAndRecovery(t *testing.T) {
	f, a, b := s3App(t)
	hash, _ := data.HashPassword("Fixture123!")
	token, _ := data.NewToken()
	code := strings.Repeat("a", 64)
	h := data.Household{SchemaVersion: 2, Name: "Family", Language: "en-US", AdminPasswordHash: hash, Sessions: []data.Session{{TokenHash: digest(token), PasswordHashFingerprint: digest(hash)}}, Users: []data.Person{{ID: "u1", Name: "Visible User"}, {ID: "u2", Name: "Hidden Claimant"}}, Recipients: []data.Person{{ID: "r1", Name: "Child"}}, Lists: []data.List{{ID: "l1", Name: "Christmas", RecipientIDs: []string{"r1"}, UserLinks: []data.UserLink{{UserID: "u1", Code: code}, {UserID: "u2", Code: strings.Repeat("b", 64)}}}}, Wishes: []data.Wish{{ID: "w1", Title: "Assigned train", Assignment: &data.Assignment{ListID: "l1", RecipientID: "r1"}, Claims: []data.Claim{{UserID: "u2"}}}, {ID: "w2", Title: "Unassigned secret"}}}
	_, err := b.Create(h)
	if err != nil {
		t.Fatal(err)
	}
	user := newAuthBrowser(a)
	user.get(t, "/?code="+code)
	admin := newAuthBrowser(a)
	admin.cookies[adminCookie] = token
	admin.get(t, "/")
	f.mu.Lock()
	f.status = 503
	before := f.puts
	f.mu.Unlock()
	w := user.get(t, "/?code="+code)
	body := w.Body.String()
	if w.Code != 200 || !strings.Contains(body, "Assigned train") || strings.Contains(body, "Hidden Claimant") || strings.Contains(body, "Unassigned secret") || strings.Contains(body, "name=\"user\"") {
		t.Fatal("cached user scope/privacy failed")
	}
	_, _, rev, _ := b.Load()
	w = user.post(t, "/claims/toggle?code="+code, url.Values{"revision": {rev}, "wish": {"w1"}, "mode": {"add"}})
	if w.Code == 303 || !strings.Contains(w.Body.String(), translate("en-US", "HR113")) {
		t.Fatal("cached claim accepted or missing error")
	}
	f.mu.Lock()
	if f.puts != before {
		t.Fatal("offline mutation queued/wrote")
	}
	f.status = 0
	f.mu.Unlock()
	if w = user.get(t, "/?code="+code); w.Code != 200 || strings.Contains(w.Body.String(), "data-storage-readonly") {
		t.Fatal("recovery did not restore normal UI")
	}
	_, _, rev, _ = b.Load()
	w = user.post(t, "/claims/toggle?code="+code, url.Values{"revision": {rev}, "wish": {"w1"}, "mode": {"add"}})
	if w.Code != 303 {
		t.Fatal("recovered claim failed", w.Code)
	}
	// A 5xx PUT may have committed. Show uncertainty, preserving the user's draft.
	_, _, rev, _ = b.Load()
	f.mu.Lock()
	f.putStatus = 503
	f.mu.Unlock()
	w = admin.post(t, "/wishes/save", url.Values{"revision": {rev}, "wish": {"w2"}, "title": {"Updated secret"}, "price": {"12.34"}, "links": {"https://shop.example/item"}, "note": {"Retain draft"}})
	if w.Code == 303 || !strings.Contains(w.Body.String(), translate("en-US", "HR119")) || !strings.Contains(w.Body.String(), "Retain draft") {
		t.Fatal("uncertain save misreported or draft lost", w.Code)
	}
}

// @test ACC-008-AC6 CON-009-AC10 CON-009-AC11
func TestS3ErrorsAndFailedLogout(t *testing.T) {
	f, a, b := s3App(t)
	guest := newAuthBrowser(a)
	f.mu.Lock()
	f.status = 503
	f.mu.Unlock()
	w := guest.get(t, "/")
	if w.Code != 500 || !strings.Contains(w.Body.String(), translate("en-US", "HR115")) || strings.Contains(w.Body.String(), "name=\"household\"") {
		t.Fatal("unavailable storage opened setup")
	}
	f.mu.Lock()
	f.status = 403
	f.mu.Unlock()
	w = guest.get(t, "/")
	if !strings.Contains(w.Body.String(), translate("en-US", "HR116")) {
		t.Fatal("access rejection did not show config error")
	}
	f.mu.Lock()
	f.status = 0
	f.mu.Unlock()
	guest.get(t, "/")
	if w = guest.post(t, "/setup/finish", authSetupValues()); w.Code != 303 {
		t.Fatal("setup failed")
	}
	guest.get(t, "/")
	token := guest.cookies[adminCookie]
	// Make the local directory unwritable after authentication, without corrupting
	// an existing deny record. chmod failure under root is explicitly skipped.
	_, s, _, err := b.Load()
	if err != nil {
		t.Fatal(err)
	}
	parent := filepath.Dir(s.PictureDirectory())
	os.MkdirAll(parent, 0700)
	os.Chmod(parent, 0500)
	defer os.Chmod(parent, 0700)
	probe, err := os.CreateTemp(parent, ".permission-probe-")
	if err == nil {
		probe.Close()
		os.Remove(probe.Name())
		t.Skip("process bypasses directory permissions")
	}
	w = guest.post(t, "/logout", nil)
	if w.Code != 500 || guest.cookies[adminCookie] != token || !strings.Contains(w.Body.String(), translate("de-CH", "HR120")) {
		t.Fatal("failed durable logout cleared cookie or lacked error", w.Code)
	}
	os.Chmod(parent, 0700)
	if w = guest.get(t, "/"); !strings.Contains(w.Body.String(), "action=\"/logout\"") {
		t.Fatal("failed logout ended session")
	}
}

// @test CON-009-AC14: a confirmed save survives a broken fallback and the next
// page explains the stale offline copy without disabling healthy remote writes.
func TestS3ConfirmedSaveShowsCacheWarning(t *testing.T) {
	_, a, b := s3App(t)
	admin := newAuthBrowser(a)
	admin.get(t, "/")
	if w := admin.post(t, "/setup/finish", authSetupValues()); w.Code != 303 {
		t.Fatal("setup failed")
	}
	_, s, rev, err := b.Load()
	if err != nil {
		t.Fatal(err)
	}
	fallback := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(s.PictureDirectory()))), "fallback.json")
	if err = os.Remove(fallback); err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(fallback, 0700); err != nil {
		t.Fatal(err)
	}
	if w := admin.post(t, "/manage/list/create", url.Values{"revision": {rev}, "name": {"Confirmed remote list"}}); w.Code != 303 {
		t.Fatal("cache failure rolled back confirmed save", w.Code)
	}
	w := admin.get(t, "/")
	if w.Code != 200 || !strings.Contains(w.Body.String(), translate("de-CH", "HR118")) || strings.Contains(w.Body.String(), "data-storage-readonly") || !strings.Contains(w.Body.String(), "Confirmed remote list") {
		t.Fatal("confirmed save warning or remote data missing")
	}
}
