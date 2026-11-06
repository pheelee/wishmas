package app

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	data "github.com/pheelee/wishmas/internal/household"
)

type pictureResponseWriter struct {
	header               http.Header
	fail                 bool
	writes, statusWrites int
	body                 bytes.Buffer
}

func (w *pictureResponseWriter) Header() http.Header { return w.header }
func (w *pictureResponseWriter) WriteHeader(int)     { w.statusWrites++ }
func (w *pictureResponseWriter) Write(raw []byte) (int, error) {
	w.writes++
	if w.fail {
		return 0, errors.New("private transport failure")
	}
	return w.body.Write(raw)
}

// @test WSH-006-AC17 OPS-001-AC3
func TestPictureResponseWriteDiagnostics(t *testing.T) {
	dir := t.TempDir()
	store := data.NewStore(filepath.Join(dir, "household.yaml"))
	if err := os.Mkdir(store.PictureDirectory(), 0700); err != nil {
		t.Fatal(err)
	}
	raw, err := optimizePicture(pictureFixture(t, 2, 2))
	if err != nil {
		t.Fatal(err)
	}
	const filename = "private-picture.webp"
	if err := os.WriteFile(filepath.Join(store.PictureDirectory(), filename), raw, 0600); err != nil {
		t.Fatal(err)
	}
	h := data.Household{Wishes: []data.Wish{{ID: "w1", Picture: filename}}}
	app := New(dir).(*householdApp)
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "write failure"}[fail], func(t *testing.T) {
			var logs bytes.Buffer
			previous := slog.Default()
			slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
			defer slog.SetDefault(previous)
			response := &pictureResponseWriter{header: make(http.Header), fail: fail}
			request := httptest.NewRequest("GET", "https://wishmas.test/wish-picture?wish=w1&picture="+filename+"&code=PrivateShareCode", nil)
			app.householdPicture(response, request, h, store, "", adminView{})
			if response.writes != 1 || response.statusWrites != 0 {
				t.Fatal("picture response was retried or replaced")
			}
			if response.Header().Get("Content-Type") != "image/webp" {
				t.Fatal("incorrect picture content type")
			}
			if fail {
				if !strings.Contains(logs.String(), "picture response write failed") || !strings.Contains(logs.String(), "error_kind=") {
					t.Fatalf("missing diagnostic: %s", &logs)
				}
			} else if logs.Len() != 0 || !bytes.Equal(response.body.Bytes(), raw) {
				t.Fatal("successful response changed or logged")
			}
			for _, private := range []string{"private transport failure", filename, "PrivateShareCode", "wish-picture", dir} {
				if strings.Contains(logs.String(), private) {
					t.Fatal("private response data leaked")
				}
			}
		})
	}
}
