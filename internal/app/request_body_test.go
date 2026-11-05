package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type stalledRequestBody struct {
	started, release    chan struct{}
	readOnce, closeOnce sync.Once
}

func (b *stalledRequestBody) Read([]byte) (int, error) {
	b.readOnce.Do(func() { close(b.started) })
	<-b.release
	return 0, io.EOF
}
func (b *stalledRequestBody) Close() error {
	b.closeOnce.Do(func() { close(b.release) })
	return nil
}

// @test OPS-001-AC7: a body stalled in Read must not own the household lock.
func TestSlowPOSTDoesNotBlockPages(t *testing.T) {
	for _, initialized := range []bool{false, true} {
		t.Run(map[bool]string{false: "before setup", true: "after setup"}[initialized], func(t *testing.T) {
			dir := t.TempDir()
			a := New(dir)
			if initialized {
				app, _ := authSetup(t, dir)
				a = app
			}
			body := &stalledRequestBody{started: make(chan struct{}), release: make(chan struct{})}
			request := httptest.NewRequest(http.MethodPost, "https://wishmas.test/login", body)
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.ContentLength = 100
			post := httptest.NewRecorder()
			postDone := make(chan struct{})
			go func() { defer close(postDone); a.ServeHTTP(post, request) }()
			t.Cleanup(func() { _ = body.Close(); <-postDone })
			select {
			case <-body.started:
			case <-time.After(2 * time.Second):
				t.Fatal("POST did not begin reading its body")
			}
			health := httptest.NewRecorder()
			a.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
			if health.Code != http.StatusOK {
				t.Fatal("health probe failed")
			}
			page := httptest.NewRecorder()
			pageDone := make(chan struct{})
			go func() {
				defer close(pageDone)
				a.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "https://wishmas.test/", nil))
			}()
			t.Cleanup(func() { _ = body.Close(); <-pageDone })
			select {
			case <-pageDone:
			case <-time.After(2 * time.Second):
				t.Fatal("incomplete unauthenticated POST blocked ordinary page")
			}
			form := `action="/setup/finish"`
			if initialized {
				form = `action="/login"`
			}
			if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), form) {
				t.Fatal("expected functional page was not served")
			}
			_ = body.Close()
			<-postDone
			if post.Code != http.StatusForbidden {
				t.Fatalf("unauthenticated POST: %d", post.Code)
			}
		})
	}
}

// @test OPS-001-AC7: retain parse-error precedence and household localization.
func TestPOSTBodyValidationRetainsLocalizedErrors(t *testing.T) {
	a, _ := authSetup(t, t.TempDir())
	for name, body := range map[string]string{"malformed": "x=%zz", "oversized": "x=" + strings.Repeat("a", 16384)} {
		t.Run(name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "https://wishmas.test/login", strings.NewReader(body))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			a.ServeHTTP(w, r)
			if w.Code != http.StatusBadRequest || w.Header().Get("Content-Language") != "de-CH" {
				t.Fatalf("parse failure changed: %d %s", w.Code, w.Header().Get("Content-Language"))
			}
		})
	}
}
