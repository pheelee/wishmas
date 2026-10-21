package app

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// @test WSH-006-AC9 WSH-006-AC11 WSH-006-AC15
func TestPictureAccess(t *testing.T) {
	dir := t.TempDir()
	a := newLegacy(dir)
	parent, family := setupHousehold(t, a, "Europe/Zurich")
	addTestPerson(t, a, parent, "Anna")
	file := listFile(dir, "2026", "Anna")
	writeTestYAML(t, file, "wishes:\n  - title: Picture train\n    picture: sample.webp\n")
	pictures := filepath.Join(filepath.Dir(file), "pictures")
	os.MkdirAll(pictures, 0700)
	optimized, err := optimizePicture(pictureFixture(t, 8, 8))
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(pictures, "sample.webp"), optimized, 0600)
	for _, link := range []string{parent, family} {
		code, _ := url.Parse(link)
		q := code.Query()
		q.Set("person", "Anna")
		q.Set("year", "2026")
		q.Set("picture", "sample.webp")
		path := "/wish-picture?" + q.Encode()
		got := request(t, a, "GET", path, nil)
		if got.Code != 200 || !bytes.Equal(got.Body.Bytes(), optimized) {
			t.Fatalf("authorized picture: %d", got.Code)
		}
		q.Set("picture", "../household.yaml")
		if got := request(t, a, "GET", "/wish-picture?"+q.Encode(), nil); got.Code != 404 {
			t.Fatal("path traversal allowed")
		}
		q.Set("picture", "sample.webp")
		q.Set("code", "invalid")
		if got := request(t, a, "GET", "/wish-picture?"+q.Encode(), nil); got.Code != 404 {
			t.Fatal("unauthed image access")
		}
	}
	got := request(t, a, "GET", wishLink(t, parent, "Anna", "2026")+"", nil)
	if !strings.Contains(got.Body.String(), "/wish-picture?") {
		t.Fatal("picture missing from wishlist preview")
	}
}

// @test WSH-006-AC11 WSH-006-AC15
func TestPictureHistoryArchivesAndMissing(t *testing.T) {
	a, dir, parent, family := newPeopleApp(t)
	addTestPerson(t, a, parent, "Anna")
	path := listFile(dir, "2026", "Anna")
	writeTestYAML(t, path, "wishes:\n - title: Picture train\n   picture: sample.webp\n   picture_source: https://images.example/p.png\n - title: Plain title\n")
	page := request(t, a, "GET", wishLink(t, parent, "Anna", "2026"), nil)
	if strings.Count(page.Body.String(), "data-picture-shelf") != 1 || !strings.Contains(page.Body.String(), "Image unavailable") {
		t.Fatal("missing/no-image placeholders")
	}
	raw, err := optimizePicture(pictureFixture(t, 8, 8))
	if err != nil {
		t.Fatal(err)
	}
	folder := filepath.Join(filepath.Dir(path), "pictures")
	os.MkdirAll(folder, 0700)
	os.WriteFile(filepath.Join(folder, "sample.webp"), raw, 0600)
	old := listFile(dir, "2025", "Anna")
	os.MkdirAll(filepath.Dir(old), 0700)
	contents, _ := os.ReadFile(path)
	os.WriteFile(old, contents, 0600)
	oldFolder := filepath.Join(filepath.Dir(old), "pictures")
	os.MkdirAll(oldFolder, 0700)
	os.WriteFile(filepath.Join(oldFolder, "sample.webp"), raw, 0600)
	pictureLink := func(link, year, picture string) string {
		u, _ := url.Parse(link)
		q := u.Query()
		q.Set("year", year)
		q.Set("person", "Anna")
		q.Set("picture", picture)
		return "/wish-picture?" + q.Encode()
	}
	if got := request(t, a, "GET", pictureLink(parent, "2025", "sample.webp"), nil); got.Code != 200 {
		t.Fatalf("parent history %d", got.Code)
	}
	if got := request(t, a, "GET", pictureLink(family, "2025", "sample.webp"), nil); got.Code != 404 {
		t.Fatal("family history visible")
	}
	os.WriteFile(filepath.Join(folder, "orphan.webp"), raw, 0600)
	if got := request(t, a, "GET", pictureLink(parent, "2026", "orphan.webp"), nil); got.Code != 404 {
		t.Fatal("unattached image served")
	}
	os.Remove(filepath.Join(folder, "sample.webp"))
	os.Symlink(filepath.Join(oldFolder, "sample.webp"), filepath.Join(folder, "sample.webp"))
	if got := request(t, a, "GET", pictureLink(parent, "2026", "sample.webp"), nil); got.Code != 404 {
		t.Fatal("symlink image served")
	}
	os.Remove(filepath.Join(folder, "sample.webp"))
	os.WriteFile(filepath.Join(folder, "sample.webp"), raw, 0600)
	request(t, a, "POST", personAction(t, parent, "/people/archive"), url.Values{"person": {"Anna"}, "year": {"2026"}})
	for _, link := range []string{parent, family} {
		if got := request(t, a, "GET", pictureLink(link, "2026", "sample.webp"), nil); got.Code != 404 {
			t.Fatal("archived image visible")
		}
	}
}
