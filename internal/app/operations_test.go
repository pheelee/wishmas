package app

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// @test OPS-001-AC1 OPS-001-AC3
func TestHealthIsIndependentAndQuiet(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "households"), []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	var logs bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	defer slog.SetDefault(old)
	a := New(dir).(*householdApp)
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, method := range []string{"GET", "HEAD", "POST"} {
		w := httptest.NewRecorder()
		a.ServeHTTP(w, httptest.NewRequest(method, "/healthz?code=secret", nil))
		want := 200
		if method == "POST" {
			want = 405
		}
		if w.Code != want {
			t.Fatalf("%s: %d", method, w.Code)
		}
	}
	if logs.Len() != 0 {
		t.Fatal(logs.String())
	}
}

// @test OPS-001-AC3
func TestImageFailureLoggingRedactsURL(t *testing.T) {
	var logs bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	defer slog.SetDefault(old)
	_, err := fetchPictureWithClient(t.Context(), http.DefaultClient, "https://user:secret@127.0.0.1/private?token=hidden")
	if err == nil {
		t.Fatal("invalid URL accepted")
	}
	if !strings.Contains(logs.String(), "image download failed") {
		t.Fatal(logs.String())
	}
	for _, secret := range []string{"secret", "hidden", "/private"} {
		if strings.Contains(logs.String(), secret) {
			t.Fatal("URL leaked")
		}
	}
}

// @test OPS-001-AC3
func TestMutationLogsAndQuietReads(t *testing.T) {
	dir := t.TempDir()
	_, b := authSetup(t, dir)
	var logs bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	defer slog.SetDefault(old)
	b.get(t, "/")
	if logs.Len() != 0 {
		t.Fatalf("ordinary read logged: %s", logs.String())
	}
	_, rev, err := authStore(dir).Load()
	if err != nil {
		t.Fatal(err)
	}
	w := b.post(t, "/manage/list/create", url.Values{"name": {"PrivateListName"}, "revision": {rev}})
	if w.Code != 303 {
		t.Fatal(w.Code)
	}
	for _, want := range []string{"household saved", "action completed", "/manage/list/create"} {
		if !strings.Contains(logs.String(), want) {
			t.Fatalf("missing %s: %s", want, logs.String())
		}
	}
	if strings.Contains(logs.String(), "PrivateListName") {
		t.Fatal("content leaked")
	}
	logs.Reset()
	b.get(t, "/unknown?code=SecretShareCode")
	if strings.Contains(logs.String(), "SecretShareCode") || strings.Contains(logs.String(), "/unknown") {
		t.Fatal("URL leaked")
	}
}

// @test OPS-001-AC3
func TestImageCompletionLogging(t *testing.T) {
	var logs bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	defer slog.SetDefault(old)
	raw := pictureFixture(t, 2, 2)
	client := &http.Client{Transport: pictureRoundTrip(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(raw)), Header: make(http.Header)}, nil
	})}
	result, err := fetchPictureWithClient(t.Context(), client, "https://images.example/private?token=SecretToken")
	if err != nil || len(result) == 0 {
		t.Fatal(err)
	}
	for _, want := range []string{"image download started", "image download completed", "bytes="} {
		if !strings.Contains(logs.String(), want) {
			t.Fatal(logs.String())
		}
	}
	if strings.Contains(logs.String(), "SecretToken") || strings.Contains(logs.String(), "images.example") {
		t.Fatal("URL leaked")
	}
}
