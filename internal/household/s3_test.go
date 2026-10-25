package household

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// @test CON-009-AC1 CON-009-AC2
func TestStorageEnvironment(t *testing.T) {
	env := map[string]string{}
	get := func(k string) string { return env[k] }
	c, err := StorageConfigFromEnv(get)
	if err != nil || c.Backend != "local" {
		t.Fatal("default backend must be local")
	}
	env["WISHMAS_STORAGE"] = "other"
	if _, err = StorageConfigFromEnv(get); err == nil {
		t.Fatal("unknown backend accepted")
	}
	env["WISHMAS_STORAGE"] = "s3"
	if _, err = StorageConfigFromEnv(get); err == nil {
		t.Fatal("missing S3 settings accepted")
	}
	for k, v := range map[string]string{"ENDPOINT": "https://s3.test", "BUCKET": "wishmas", "ACCESS_KEY_ID": "fixture", "SECRET_ACCESS_KEY": "fixture-secret"} {
		env["WISHMAS_S3_"+k] = v
	}
	c, err = StorageConfigFromEnv(get)
	if err != nil || c.Region != "us-east-1" || c.Timeout != 5*time.Second {
		t.Fatal("S3 defaults not applied", err)
	}
	for _, v := range []string{"0s", "-1s", "nonsense"} {
		env["WISHMAS_S3_TIMEOUT"] = v
		if _, err = StorageConfigFromEnv(get); err == nil {
			t.Fatal("bad timeout accepted")
		}
	}
	delete(env, "WISHMAS_S3_TIMEOUT")
	for _, v := range []string{"ftp://s3.test", "https://user:secret@s3.test", "https://s3.test?secret=yes", "https:///"} {
		env["WISHMAS_S3_ENDPOINT"] = v
		if _, err = StorageConfigFromEnv(get); err == nil {
			t.Fatal("bad endpoint accepted")
		}
	}
}

type s3Fixture struct {
	mu        sync.Mutex
	objects   map[string][]byte
	status    int
	puts      int
	uncertain bool
	signed    bool
}

func (f *s3Fixture) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !strings.HasPrefix(r.URL.Path, "/bucket") || !strings.Contains(r.Header.Get("Authorization"), "AWS4-HMAC-SHA256") || !strings.Contains(r.Header.Get("Authorization"), "/test-region/s3/aws4_request") {
		http.Error(w, "path/signature", 400)
		return
	}
	f.signed = true
	if f.status != 0 {
		w.WriteHeader(f.status)
		return
	}
	if r.URL.Query().Get("list-type") == "2" {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(`<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">`))
		for k := range f.objects {
			w.Write([]byte("<Contents><Key>" + k + "</Key></Contents>"))
		}
		w.Write([]byte("</ListBucketResult>"))
		return
	}
	key := strings.TrimPrefix(r.URL.Path, "/bucket/")
	switch r.Method {
	case "GET":
		raw, ok := f.objects[key]
		if !ok {
			http.Error(w, "missing", 404)
			return
		}
		w.Write(raw)
	case "PUT":
		raw, _ := io.ReadAll(r.Body)
		f.objects[key] = raw
		f.puts++
		if f.uncertain {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(200)
		}
	default:
		w.WriteHeader(405)
	}
}
func newS3Fixture(t *testing.T) (*s3Fixture, *S3Backend) {
	t.Helper()
	f := &s3Fixture{objects: map[string][]byte{}}
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	b, err := NewS3Backend(t.TempDir(), StorageConfig{Backend: "s3", Endpoint: srv.URL, Bucket: "bucket", Region: "test-region", AccessKey: "fixture", SecretKey: "fixture-secret", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	return f, b
}

// @test CON-009-AC3 CON-009-AC5 CON-009-AC6 CON-009-AC7 CON-009-AC8 CON-009-AC9 CON-009-AC11 CON-009-AC12 CON-009-AC13 CON-009-AC16
func TestS3AuthorityCacheAndWrites(t *testing.T) {
	f, b := newS3Fixture(t)
	h, s, _, err := b.Load()
	if err != nil || h != nil || s != nil {
		t.Fatal("empty bucket must offer setup", err)
	}
	hash, _ := HashPassword("Fixture123!")
	original := Household{SchemaVersion: 2, Name: "Family", Language: "en-US", AdminPasswordHash: hash}
	s, err = b.Create(original)
	if err != nil {
		t.Fatal(err)
	}
	h, s, rev, err := b.Load()
	if err != nil || h.Name != original.Name || s.ReadOnly() {
		t.Fatal("load failed", err)
	}
	if !f.signed {
		t.Fatal("unsigned request")
	}
	if err = s.Update(rev, func(h *Household) error { h.Language = "de-CH"; return nil }); err != nil {
		t.Fatal(err)
	}
	if err = s.Update(rev, func(h *Household) error { h.Language = "en-US"; return nil }); !errors.Is(err, ErrStale) {
		t.Fatal("stale write accepted", err)
	}
	f.mu.Lock()
	f.status = 503
	f.mu.Unlock()
	h, s, rev, err = b.Load()
	if err != nil || h.Language != "de-CH" || !s.ReadOnly() {
		t.Fatal("cached readonly expected", err)
	}
	if err = s.Update(rev, func(h *Household) error { return nil }); !errors.Is(err, ErrReadOnly) {
		t.Fatal("offline update accepted", err)
	}
	restarted, err := NewS3Backend(b.dataDir, b.config)
	if err != nil {
		t.Fatal(err)
	}
	if h, _, _, err = restarted.Load(); err != nil || h.Language != "de-CH" {
		t.Fatal("cache not durable", err)
	}
	f.mu.Lock()
	f.status = 403
	f.mu.Unlock()
	if _, _, _, err = b.Load(); !errors.Is(err, ErrStorageConfig) {
		t.Fatal("permissions used fallback", err)
	}
	f.mu.Lock()
	f.status = 0
	f.objects[original.Name+"/household.yaml"] = []byte("invalid: yaml")
	f.mu.Unlock()
	if _, _, _, err = b.Load(); !errors.Is(err, ErrStorageConfig) {
		t.Fatal("invalid remote used fallback", err)
	}
	f.mu.Lock()
	f.objects = map[string][]byte{}
	f.mu.Unlock()
	if h, _, _, err = b.Load(); err != nil || h != nil {
		t.Fatal("authoritative missing used old cache", err)
	}
	f.mu.Lock()
	f.status = 503
	f.mu.Unlock()
	if _, _, _, err = b.Load(); !errors.Is(err, ErrStorageUnavailable) {
		t.Fatal("removed household cache revived", err)
	}

}

// @test CON-009-AC4 CON-009-AC8 CON-009-AC13 CON-009-AC14
func TestS3FailureBoundaries(t *testing.T) {
	f, b := newS3Fixture(t)
	hash, _ := HashPassword("Fixture123!")
	h := Household{SchemaVersion: 2, Name: "Family", Language: "en-US", AdminPasswordHash: hash}
	s, err := b.Create(h)
	if err != nil {
		t.Fatal(err)
	}
	_, rev, _ := s.Load()
	f.mu.Lock()
	f.uncertain = true
	f.mu.Unlock()
	if err = s.Update(rev, func(h *Household) error { h.Language = "de-CH"; return nil }); !errors.Is(err, ErrUncertain) {
		t.Fatal("uncertain write claimed success", err)
	}
	f.mu.Lock()
	if f.puts != 2 {
		t.Fatal("uncertain write was retried")
	}
	f.uncertain = false
	f.mu.Unlock()
	_, rev, err = s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(b.cachePath()); err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(b.cachePath(), 0700); err != nil {
		t.Fatal(err)
	}
	if err = s.Update(rev, func(h *Household) error { h.Language = "en-US"; return nil }); err != nil || !s.CacheWarning() {
		t.Fatal("cache failure rolled back remote save or lacked warning", err)
	}
	f.mu.Lock()
	f.objects["Other/household.yaml"] = []byte("invalid")
	f.mu.Unlock()
	if _, _, _, err = b.Load(); !errors.Is(err, ErrMultipleHouseholds) {
		t.Fatal("multiple households accepted", err)
	}
	other, err := NewS3Backend(b.dataDir, StorageConfig{Endpoint: b.config.Endpoint, Bucket: "other", Region: "test-region", AccessKey: "fixture", SecretKey: "fixture-secret", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	f.status = 503
	f.mu.Unlock()
	if h, _, _, e := other.Load(); h != nil || e == nil {
		t.Fatal("cross-bucket cache leak", err)
	}
}

// @test ACC-008-AC4 ACC-008-AC5 ACC-008-AC6 ACC-008-AC7 ACC-008-AC8 ACC-008-AC9 ACC-008-AC10
func TestDurableOfflineLogout(t *testing.T) {
	f, b := newS3Fixture(t)
	hash, _ := HashPassword("Fixture123!")
	token := strings.Repeat("a", 64)
	other := strings.Repeat("b", 64)
	s, err := b.Create(Household{SchemaVersion: 2, Name: "Family", Language: "en-US", AdminPasswordHash: hash, Sessions: []Session{{TokenHash: token, PasswordHashFingerprint: revisionOf([]byte(hash))}, {TokenHash: other, PasswordHashFingerprint: revisionOf([]byte(hash))}}})
	if err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	f.status = 503
	f.mu.Unlock()
	if err = s.RevokeSession(token); err != nil {
		t.Fatal(err)
	}
	restarted, _ := NewS3Backend(b.dataDir, b.config)
	h, s, _, err := restarted.Load()
	if err != nil || h == nil || !s.SessionRevoked(token) || s.SessionRevoked(other) {
		t.Fatal("revocation did not survive restart or affected another session")
	}
	f.mu.Lock()
	f.status = 0
	f.mu.Unlock()
	h, s, _, err = restarted.Load()
	if err != nil || len(h.Sessions) != 1 || h.Sessions[0].TokenHash != other {
		t.Fatal("recovery did not remove exactly revoked session", err)
	}
	raw, _ := os.ReadFile(restarted.cachePath())
	if strings.Contains(string(raw), "raw-browser-token") {
		t.Fatal("raw session token persisted")
	}
}

// @test CON-009-AC2 CON-009-AC5 CON-009-AC10 CON-009-AC11 CON-009-AC13
func TestS3OperationTimeoutIncludesBodyAndNoCacheStartup(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("list-type") == "2" {
			w.Write([]byte(`<ListBucketResult><Contents><Key>Family/household.yaml</Key></Contents></ListBucketResult>`))
			return
		}
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()
	defer close(release)
	b, err := NewS3Backend(t.TempDir(), StorageConfig{Endpoint: srv.URL, Bucket: "bucket", AccessKey: "fixture", SecretKey: "fixture-secret", Timeout: 50 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	if h, _, _, err := b.Load(); h != nil || !errors.Is(err, ErrStorageUnavailable) {
		t.Fatal("unreachable startup became setup or cached data", err)
	}
	if time.Since(started) > time.Second {
		t.Fatal("body read exceeded operation timeout")
	}
}

// @test ACC-008-AC5 ACC-008-AC8 ACC-008-AC9 ACC-008-AC10
func TestRevocationFailureCacheAndRetryDeadlines(t *testing.T) {
	f, b := newS3Fixture(t)
	hash, _ := HashPassword("Fixture123!")
	token := strings.Repeat("a", 64)
	h := Household{SchemaVersion: 2, Name: "Family", Language: "en-US", AdminPasswordHash: hash, Sessions: []Session{{TokenHash: token, PasswordHashFingerprint: revisionOf([]byte(hash))}}}
	s, err := b.Create(h)
	if err != nil {
		t.Fatal(err)
	}
	os.MkdirAll(filepath.Dir(s.revocationPath()), 0700)
	os.Mkdir(s.revocationPath(), 0700)
	if err = s.RevokeSession(token); err == nil {
		t.Fatal("failed local persist acknowledged logout")
	}
	os.Remove(s.revocationPath())
	if err = s.RevokeSession(token); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	f.mu.Lock()
	f.status = 503
	f.mu.Unlock()
	b.Maintenance(now)
	records, _ := s.revocations()
	if records[token].Attempts != 1 || !records[token].Next.Equal(now.Add(time.Minute)) {
		t.Fatal("first retry delay incorrect")
	}
	b.Maintenance(now.Add(30 * time.Second))
	records, _ = s.revocations()
	if records[token].Attempts != 1 {
		t.Fatal("retried before deadline")
	}
	b.Maintenance(now.Add(time.Minute))
	records, _ = s.revocations()
	if records[token].Attempts != 2 || !records[token].Next.Equal(now.Add(3*time.Minute)) {
		t.Fatal("increasing retry delay incorrect")
	}
	// Remote removal may succeed while the fallback write fails; deny must remain.
	os.Remove(b.cachePath())
	os.Mkdir(b.cachePath(), 0700)
	f.mu.Lock()
	f.status = 0
	f.mu.Unlock()
	fresh, _, _, err := b.Load()
	if err != nil || len(fresh.Sessions) != 0 || !s.SessionRevoked(token) {
		t.Fatal("cache-write failure resurrected removed session", err)
	}
	os.Remove(b.cachePath())
	fresh, _, _, err = b.Load()
	if err != nil || len(fresh.Sessions) != 0 {
		t.Fatal(err)
	}
	records, _ = s.revocations()
	if len(records) != 0 {
		t.Fatal("safe reconciliation did not finish")
	}
}

// @test CON-009-AC5 CON-009-AC7 ACC-008-AC7
func TestS3PreservesFutureWishClaimFieldsAndManualEdits(t *testing.T) {
	f, b := newS3Fixture(t)
	hash, _ := HashPassword("Fixture123!")
	h := Household{SchemaVersion: 2, Name: "Family", Language: "en-US", AdminPasswordHash: hash, Wishes: []Wish{{ID: "w1", Title: "Original wish"}}}
	s, err := b.Create(h)
	if err != nil {
		t.Fatal(err)
	}
	_, rev, _ := s.Load()
	f.mu.Lock()
	raw := strings.Replace(string(f.objects["Family/household.yaml"]), "title: Original wish", "title: Manual wish\n      future: retained", 1)
	f.objects["Family/household.yaml"] = []byte(raw)
	f.mu.Unlock()
	if err = s.Update(rev, func(h *Household) error { h.Language = "de-CH"; return nil }); !errors.Is(err, ErrStale) {
		t.Fatal("manual remote change overwritten", err)
	}
	fresh, rev, err := s.Load()
	if err != nil || fresh.Wishes[0].Title != "Manual wish" {
		t.Fatal("manual edit not reloaded", err)
	}
	if err = s.Update(rev, func(h *Household) error { h.Language = "de-CH"; return nil }); err != nil {
		t.Fatal(err)
	}
	saved, _ := b.get("Family/household.yaml")
	if !strings.Contains(string(saved), "future: retained") {
		t.Fatal("future field lost")
	}
}
