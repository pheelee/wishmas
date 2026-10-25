package app

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	data "github.com/pheelee/wishmas/internal/household"
)

// @test WSH-008-AC1 WSH-008-AC2 WSH-008-AC3 WSH-008-AC4 WSH-008-AC5 WSH-008-AC6 WSH-008-AC7 WSH-008-AC8 WSH-008-AC12
func TestBackgroundRecoveryIsBoundedAndPreservesYAML(t *testing.T) {
	for _, remote := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "s3"}[remote], func(t *testing.T) { testBackgroundRecovery(t, remote) })
	}
}
func testBackgroundRecovery(t *testing.T, remote bool) {
	dir := t.TempDir()
	a := New(dir).(*householdApp)
	var f *appS3Fixture
	var backend *data.S3Backend
	if remote {
		f, a, backend = s3App(t)
	}
	hash, _ := data.HashPassword("Fixture123!")
	token, _ := data.NewToken()
	h := data.Household{SchemaVersion: 2, Name: "Family", Language: "en-US", AdminPasswordHash: hash, Sessions: []data.Session{{TokenHash: digest(token), PasswordHashFingerprint: digest(hash)}}}
	for _, id := range []string{"w1", "w2", "w3", "w4", "w5"} {
		h.Wishes = append(h.Wishes, data.Wish{ID: id, Title: "Train picture", Picture: id + ".webp", PictureSource: "https://images.example/" + id + ".png"})
	}
	path := filepath.Join(dir, "households", "Family", "household.yaml")
	os.MkdirAll(filepath.Dir(path), 0700)
	s := data.NewStore(path)
	if remote {
		var err error
		s, err = backend.Create(h)
		if err != nil {
			t.Fatal(err)
		}
	} else if err := s.Create(h); err != nil {
		t.Fatal(err)
	}
	saved := func() []byte {
		if remote {
			f.mu.Lock()
			defer f.mu.Unlock()
			return bytes.Clone(f.objects["Family/household.yaml"])
		}
		raw, _ := os.ReadFile(path)
		return raw
	}
	before := saved()
	var encoded bytes.Buffer
	im := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	im.Set(1, 1, color.NRGBA{R: 255, A: 128})
	png.Encode(&encoded, im)
	release := make(chan struct{})
	started := make(chan struct{}, 10)
	var active, peak atomic.Int32
	a.pictureClient = &http.Client{Transport: pictureRoundTrip(func(r *http.Request) (*http.Response, error) {
		n := active.Add(1)
		defer active.Add(-1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		started <- struct{}{}
		select {
		case <-release:
		case <-r.Context().Done():
			return nil, r.Context().Err()
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(encoded.Bytes())), Header: http.Header{}}, nil
	})}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := StartBackground(ctx, a)
	defer func() { cancel(); <-done }()
	for range 3 {
		select {
		case <-started:
		case <-time.After(3 * time.Second):
			t.Fatal("startup did not schedule three background restores")
		}
	}
	select {
	case <-started:
		t.Fatal("more than three concurrent restores")
	default:
	}
	// Pages stay available while all workers wait on their sources.
	browser := newAuthBrowser(a)
	browser.cookies[adminCookie] = token
	if w := browser.get(t, "/"); w.Code != 200 || !strings.Contains(w.Body.String(), translate("en-US", "I04")) {
		t.Fatal("recovery blocked pages")
	}
	close(release)
	deadline := time.Now().Add(5 * time.Second)
	for {
		all := true
		for _, w := range h.Wishes {
			if _, err := readHouseholdPicture(s, w.Picture); err != nil {
				all = false
			}
		}
		if all {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("images were not restored")
		}
		time.Sleep(20 * time.Millisecond)
	}
	if w := browser.get(t, "/"); w.Code != 200 || !strings.Contains(w.Body.String(), `class="wish-picture"`) {
		t.Fatal("restored image not visible on next page load")
	}
	after := saved()
	if !bytes.Equal(before, after) || peak.Load() > 3 {
		t.Fatal("recovery changed YAML or exceeded concurrency")
	}
}

// @test WSH-008-AC5 WSH-008-AC7
func TestBackgroundRecoveryRetriesAtIncreasingDeadlines(t *testing.T) {
	a := New(t.TempDir()).(*householdApp)
	hash, _ := data.HashPassword("Fixture123!")
	h := data.Household{SchemaVersion: 2, Name: "Family", Language: "en-US", AdminPasswordHash: hash, Wishes: []data.Wish{{ID: "w1", Title: "Retry picture", Picture: "w1.webp", PictureSource: "https://images.example/retry.png"}}}
	path := filepath.Join(a.dataDir, "households", "Family", "household.yaml")
	os.MkdirAll(filepath.Dir(path), 0700)
	s := data.NewStore(path)
	if err := s.Create(h); err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	now := time.Now()
	a.now = func() time.Time { mu.Lock(); defer mu.Unlock(); return now }
	advance := func(d time.Duration) { mu.Lock(); now = now.Add(d); mu.Unlock() }
	var calls atomic.Int32
	a.pictureClient = &http.Client{Transport: pictureRoundTrip(func(*http.Request) (*http.Response, error) { calls.Add(1); return nil, errPictureDownload })}
	ctx, cancel := context.WithCancel(context.Background())
	done := StartBackground(ctx, a)
	defer func() { cancel(); <-done }()
	await := func(want int32) {
		t.Helper()
		deadline := time.Now().Add(3 * time.Second)
		for calls.Load() != want {
			if time.Now().After(deadline) {
				t.Fatal("retry was not scheduled", want, calls.Load())
			}
			time.Sleep(10 * time.Millisecond)
		}
		time.Sleep(150 * time.Millisecond)
	}
	await(1)
	advance(30 * time.Second)
	time.Sleep(150 * time.Millisecond)
	if calls.Load() != 1 {
		t.Fatal("retried before one minute")
	}
	advance(30 * time.Second)
	await(2)
	advance(time.Minute)
	time.Sleep(150 * time.Millisecond)
	if calls.Load() != 2 {
		t.Fatal("second retry ignored two-minute delay")
	}
	advance(time.Minute)
	await(3)
	// A changed reference cancels the old recovery without another source request.
	_, rev, _ := s.Load()
	if err := s.Update(rev, func(h *data.Household) error { h.Wishes[0].Picture = ""; h.Wishes[0].PictureSource = ""; return nil }); err != nil {
		t.Fatal(err)
	}
	advance(4 * time.Minute)
	time.Sleep(250 * time.Millisecond)
	if calls.Load() != 3 {
		t.Fatal("removed image reference downloaded again")
	}
}
