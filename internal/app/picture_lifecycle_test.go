package app

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"go.yaml.in/yaml/v3"
)

func storedPicture(t *testing.T, path string) wishCard {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var list wishlist
	if err := yaml.Unmarshal(raw, &list); err != nil {
		t.Fatal(err)
	}
	cards, err := wishCards(list)
	if err != nil || len(cards) == 0 {
		t.Fatalf("cards: %v", err)
	}
	return cards[0]
}

// @test WSH-006-AC1 WSH-006-AC2 WSH-006-AC7 WSH-006-AC12 WSH-006-AC13 WSH-006-AC14
func TestPictureLifecycle(t *testing.T) {
	handler, dir, parent, family := newPeopleApp(t)
	a := handler
	addTestPerson(t, a, parent, "Anna")
	calls := 0
	mockPictureClient(a, pictureFixture(t, 800, 400), &calls)
	form := url.Values{"person": {"Anna"}, "year": {"2026"}, "title": {"Picture train"}, "note": {"Keep this note"}, "picture_url": {"https://images.example/one.png"}}
	save := func() {
		t.Helper()
		if got := request(t, a, "POST", personAction(t, parent, "/wishes/save"), form); got.Code != 303 {
			t.Fatalf("save %d", got.Code)
		}
	}
	save()
	path := listFile(dir, "2026", "Anna")
	card := storedPicture(t, path)
	pic := filepath.Join(filepath.Dir(path), "pictures", card.Picture)
	if _, err := os.Stat(pic); err != nil || card.PictureSource != form.Get("picture_url") {
		t.Fatalf("persisted image: %v", err)
	}
	raw, _ := os.ReadFile(path)
	writeTestYAML(t, path, strings.Replace(string(raw), "title: Picture train", "title: Picture train\n      claims: [{name: Auntie}]\n      future_field: untouched", 1))
	card = storedPicture(t, path)
	form.Set("wish", card.Token)
	form.Set("title", "Edited train")
	save()
	if calls != 1 {
		t.Fatal("unchanged source fetched")
	}
	raw, _ = os.ReadFile(path)
	if !bytes.Contains(raw, []byte("Auntie")) || !bytes.Contains(raw, []byte("future_field: untouched")) {
		t.Fatal("wish data lost")
	}
	// Failed replacement keeps every entered value and original files.
	card = storedPicture(t, path)
	form.Set("wish", card.Token)
	form.Set("picture_url", "http://127.0.0.1/private.png")
	form.Set("price", "12.50")
	form.Set("links", "https://shop.example/one")
	form.Set("note", "Preserved draft")
	before := snapshotFiles(t, dir)
	got := request(t, a, "POST", personAction(t, parent, "/wishes/save"), form)
	if got.Code != 400 || !strings.Contains(got.Body.String(), "Preserved draft") || !strings.Contains(got.Body.String(), "http://127.0.0.1/private.png") {
		t.Fatal("invalid replacement lost draft")
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("failed replacement changed files")
	}
	for _, link := range []string{family, parent} {
		form.Set("year", "2025")
		got := request(t, a, "POST", personAction(t, link, "/wishes/save"), form)
		if got.Code != 404 {
			t.Fatalf("permission %d", got.Code)
		}
	}
	form.Set("year", "2026")
	form.Set("picture_url", "https://images.example/two.png")
	save()
	if _, err := os.Stat(pic); !os.IsNotExist(err) {
		t.Fatal("replacement left old image")
	}
	card = storedPicture(t, path)
	pic = filepath.Join(filepath.Dir(path), "pictures", card.Picture)
	form.Set("wish", card.Token)
	form.Set("picture_url", "")
	save()
	if _, err := os.Stat(pic); !os.IsNotExist(err) {
		t.Fatal("clear left image")
	}
	if c := storedPicture(t, path); c.Picture != "" || c.PictureSource != "" {
		t.Fatal("metadata not cleared")
	}
	card = storedPicture(t, path)
	form.Set("wish", card.Token)
	form.Set("picture_url", "https://images.example/three.png")
	save()
	card = storedPicture(t, path)
	pic = filepath.Join(filepath.Dir(path), "pictures", card.Picture)
	got = request(t, a, "POST", personAction(t, parent, "/wishes/delete"), url.Values{"person": {"Anna"}, "year": {"2026"}, "wish": {card.Token}, "confirm": {"yes"}})
	if got.Code != 303 {
		t.Fatalf("delete: %d", got.Code)
	}
	if _, err := os.Stat(pic); !os.IsNotExist(err) {
		t.Fatal("delete left image")
	}
	files, _ := os.ReadDir(filepath.Dir(pic))
	if len(files) != 0 {
		t.Fatal("orphan images/originals")
	}
}

// @test WSH-006-AC12 WSH-006-AC14
func TestPictureCleanupRollbackAndMissingOld(t *testing.T) {
	handler, dir, parent, _ := newPeopleApp(t)
	a := handler
	addTestPerson(t, a, parent, "Anna")
	path := listFile(dir, "2026", "Anna")
	writeTestYAML(t, path, "# preserve exact bytes\nwishes:\n  - title: Picture train\n    picture: old.webp\n    picture_source: https://images.example/old\n")
	folder := filepath.Join(filepath.Dir(path), "pictures")
	os.MkdirAll(folder, 0700)
	old := pictureFixture(t, 8, 8)
	os.WriteFile(filepath.Join(folder, "old.webp"), old, 0600)
	before := snapshotFiles(t, dir)
	var list wishlist
	raw, _ := os.ReadFile(path)
	yaml.Unmarshal(raw, &list)
	list.Wishes = nil
	if err := pictureTransaction(path, list, "old.webp", "new.webp", []byte("new"), func(string) error { return errors.New("delete denied") }); !errors.Is(err, errPictureRemove) {
		t.Fatalf("delete failure: %v", err)
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("rollback changed files")
	}
	os.Remove(filepath.Join(folder, "old.webp"))
	if err := savePictureWish(path, list, "old.webp", "", nil); err != nil {
		t.Fatalf("missing old blocked save: %v", err)
	}
	// A malformed old path or a directory cannot delete unrelated files.
	if err := savePictureWish(path, list, "../household.yaml", "", nil); !errors.Is(err, errPictureRemove) {
		t.Fatal("unsafe removal accepted")
	}
	os.Mkdir(filepath.Join(folder, "blocked.webp"), 0700)
	before = snapshotFiles(t, dir)
	if err := savePictureWish(path, list, "blocked.webp", "new.webp", []byte("new")); !errors.Is(err, errPictureRemove) {
		t.Fatal("nonregular old file accepted")
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("blocked cleanup changed files")
	}
}

// @test WSH-006-AC11 WSH-006-AC12: revalidate after releasing the lock for network work.
func TestPictureDownloadRevalidatesAccessAndSnapshot(t *testing.T) {
	for _, action := range []string{"revoke", "archive", "rename", "stale", "year"} {
		t.Run(action, func(t *testing.T) {
			handler, dir, parent, _ := newPeopleApp(t)
			a := handler
			addTestPerson(t, a, parent, "Anna")
			path := listFile(dir, "2026", "Anna")
			writeTestYAML(t, path, "wishes:\n - title: Picture train\n")
			card := storedPicture(t, path)
			started, release := make(chan struct{}), make(chan struct{})
			source := pictureFixture(t, 2, 2)
			a.pictureClient = &http.Client{Transport: pictureRoundTrip(func(r *http.Request) (*http.Response, error) {
				close(started)
				<-release
				return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(source)), Header: make(http.Header)}, nil
			})}
			done := make(chan int, 1)
			go func() {
				got := request(t, a, "POST", personAction(t, parent, "/wishes/save"), url.Values{"person": {"Anna"}, "year": {"2026"}, "wish": {card.Token}, "title": {"Changed train"}, "picture_url": {"https://images.example/p.png"}})
				done <- got.Code
			}()
			<-started
			// A GET must complete while the source is blocked.
			if got := request(t, a, "GET", parent, nil); got.Code != 200 {
				t.Fatal("requests blocked")
			}
			switch action {
			case "revoke":
				request(t, a, "POST", personAction(t, parent, "/access/regenerate"), url.Values{"role": {"parent"}})
			case "archive":
				request(t, a, "POST", personAction(t, parent, "/people/archive"), url.Values{"person": {"Anna"}, "year": {"2026"}})
			case "rename":
				request(t, a, "POST", personAction(t, parent, "/people/rename"), url.Values{"person": {"Anna"}, "name": {"ANNA"}, "year": {"2026"}})
			case "stale":
				writeTestYAML(t, path, "wishes:\n - title: Manually changed\n")
			case "year":
				a.mu.Lock()
				a.now = func() time.Time { return time.Date(2027, 1, 1, 12, 0, 0, 0, time.UTC) }
				a.mu.Unlock()
			}
			before := snapshotFiles(t, dir)
			close(release)
			status := <-done
			if status != 404 && status != 409 {
				t.Fatalf("stale download saved: %d", status)
			}
			if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
				t.Fatal("stale download changed files")
			}
		})
	}
}

// @test WSH-006-AC12 WSH-006-AC14: real filesystem deletion failure rolls back web saves/deletes.
func TestPictureFilesystemFailure(t *testing.T) {
	a, dir, parent, _ := newPeopleApp(t)
	addTestPerson(t, a, parent, "Anna")
	path := listFile(dir, "2026", "Anna")
	writeTestYAML(t, path, "# existing wish\nwishes:\n - title: Picture train\n   picture: old.webp\n   picture_source: https://images.example/old.png\n")
	folder := filepath.Join(filepath.Dir(path), "pictures")
	os.MkdirAll(folder, 0700)
	raw, err := optimizePicture(pictureFixture(t, 4, 4))
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(folder, "old.webp"), raw, 0600)
	os.Chmod(folder, 0500)
	defer os.Chmod(folder, 0700)
	before := snapshotFiles(t, dir)
	card := storedPicture(t, path)
	got := request(t, a, "POST", personAction(t, parent, "/wishes/save"), url.Values{"person": {"Anna"}, "year": {"2026"}, "wish": {card.Token}, "title": {"Updated train"}, "picture_url": {""}})
	if got.Code != 400 || !strings.Contains(got.Body.String(), "old image could not be removed") {
		t.Fatalf("image removal failure: %d", got.Code)
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("failed web removal changed files")
	}
	got = request(t, a, "POST", personAction(t, parent, "/wishes/delete"), url.Values{"person": {"Anna"}, "year": {"2026"}, "wish": {card.Token}, "confirm": {"yes"}})
	if got.Code != 500 {
		t.Fatalf("failed deletion: %d", got.Code)
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("failed web deletion changed files")
	}
	// An unreadable configured file displays a placeholder without re-fetching.
	os.Chmod(filepath.Join(folder, "old.webp"), 0000)
	defer os.Chmod(filepath.Join(folder, "old.webp"), 0600)
	got = request(t, a, "GET", wishLink(t, parent, "Anna", "2026"), nil)
	if !strings.Contains(got.Body.String(), `class="picture-unavailable" >Image unavailable`) {
		t.Fatal("unreadable image placeholder missing")
	}
}
