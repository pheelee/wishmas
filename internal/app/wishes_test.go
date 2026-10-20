package app

import (
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
)

// @test WLS-001-AC3 WLS-001-AC4
// @test WSH-001-AC1 WSH-001-AC3 WSH-001-AC4 WSH-001-AC5 WSH-001-AC10 WSH-001-AC11 WSH-002-AC3
func TestWishCreateEdit(t *testing.T) {
	a, dir, parent, family := newPeopleApp(t)
	addTestPerson(t, a, parent, "Anna")
	form := url.Values{"person": {"Anna"}, "year": {"2026"}, "title": {"  Wooden train  "}}
	got := request(t, a, "POST", personAction(t, parent, "/wishes/save"), form)
	if got.Code != 303 {
		t.Fatalf("create: %d", got.Code)
	}
	raw, _ := os.ReadFile(listFile(dir, "2026", "Anna"))
	if !strings.Contains(string(raw), "title: Wooden train") {
		t.Fatal("title not trimmed")
	}
	writeTestYAML(t, listFile(dir, "2026", "Anna"), strings.Replace(string(raw), "title: Wooden train", "title: Wooden train\n      claims: [{name: Auntie}]", 1))
	page := request(t, a, "GET", wishLink(t, parent, "Anna", "2026"), nil)
	token := field(t, page.Body.String(), "wish")
	form.Set("wish", token)
	form.Set("title", "Electric train")
	form.Set("price", "12.50")
	form.Set("note", "A lovely gift")
	form.Set("links", "https://first.example/item\nhttps://second.example/product")
	got = request(t, a, "POST", personAction(t, parent, "/wishes/save"), form)
	if got.Code != 303 {
		t.Fatalf("edit: %d", got.Code)
	}
	raw, _ = os.ReadFile(listFile(dir, "2026", "Anna"))
	if !strings.Contains(string(raw), "Auntie") {
		t.Fatal("claims lost")
	}
	page = request(t, a, "GET", wishLink(t, family, "Anna", "2026"), nil)
	body := page.Body.String()
	for _, want := range []string{"Electric train", "12.50", "A lovely gift", `href="https://first.example/item"`, ">first.example</a>"} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %s", want)
		}
	}
	if strings.Index(body, ">first.example</a>") > strings.Index(body, ">second.example</a>") {
		t.Fatal("link order")
	}
	if strings.Contains(body, "Auntie") || strings.Contains(body, "Save wish") {
		t.Fatal("family disclosure")
	}
}

// @test WSH-001-AC2 WSH-001-AC6 WSH-001-AC7 WSH-001-AC8 WSH-001-AC9 WSH-002-AC2 WSH-002-AC5
func TestWishValidationAndGuards(t *testing.T) {
	a, dir, parent, family := newPeopleApp(t)
	addTestPerson(t, a, parent, "Anna")
	for _, entry := range []struct{ key, value string }{{"title", ""}, {"title", "tiny"}, {"title", strings.Repeat("é", 151)}, {"price", "-1"}, {"price", "1.234"}, {"price", "12,50"}, {"price", "1e2"}, {"links", "shop.example/item"}, {"links", "javascript:alert(1)"}, {"links", "https:///item"}} {
		before := snapshotFiles(t, dir)
		form := url.Values{"person": {"Anna"}, "title": {"Valid title"}}
		form.Set(entry.key, entry.value)
		got := request(t, a, "POST", personAction(t, parent, "/wishes/save"), form)
		if got.Code != 400 {
			t.Fatalf("%s=%q: %d", entry.key, entry.value, got.Code)
		}
		if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
			t.Fatal("invalid form changed files")
		}
	}
	for _, link := range []string{family, parent} {
		form := url.Values{"person": {"Anna"}, "title": {"Valid title"}, "year": {"2025"}}
		before := snapshotFiles(t, dir)
		got := request(t, a, "POST", personAction(t, link, "/wishes/save"), form)
		if got.Code != 404 {
			t.Fatalf("guard %d", got.Code)
		}
		if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
			t.Fatal("guard changed files")
		}
	}
	for _, price := range []string{"0", "2", "12.50"} {
		got := request(t, a, "POST", personAction(t, parent, "/wishes/save"), url.Values{"person": {"Anna"}, "title": {strings.Repeat("é", 150)}, "price": {price}})
		if got.Code != 303 {
			t.Fatalf("valid price %s: %d", price, got.Code)
		}
	}
}

// @test WSH-002-AC2 WSH-002-AC3 WSH-002-AC5 CON-004
func TestWishStaleArchivedAndInvalidData(t *testing.T) {
	a, dir, parent, family := newPeopleApp(t)
	addTestPerson(t, a, parent, "Anna")
	form := url.Values{"person": {"Anna"}, "year": {"2026"}, "title": {"Wooden train"}}
	if got := request(t, a, "POST", personAction(t, parent, "/wishes/save"), form); got.Code != 303 {
		t.Fatal(got.Code)
	}
	page := request(t, a, "GET", wishLink(t, parent, "Anna", "2026"), nil)
	token := field(t, page.Body.String(), "wish")
	for _, scenario := range []string{"family", "stale", "archived", "invalid"} {
		t.Run(scenario, func(t *testing.T) {
			raw, _ := os.ReadFile(listFile(dir, "2026", "Anna"))
			people, _ := os.ReadFile(peopleFile(dir))
			defer os.WriteFile(listFile(dir, "2026", "Anna"), raw, 0600)
			defer os.WriteFile(peopleFile(dir), people, 0600)
			link := parent
			expected := 409
			form.Set("wish", token)
			switch scenario {
			case "family":
				link = family
				expected = 404
			case "stale":
				writeTestYAML(t, listFile(dir, "2026", "Anna"), strings.Replace(string(raw), "Wooden train", "Changed train", 1))
			case "archived":
				writeTestYAML(t, peopleFile(dir), "people: [{name: Anna, archived: true}]\n")
				expected = 404
			case "invalid":
				writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes: [{title: Valid title, price: -1}]\n")
				expected = 500
			}
			before := snapshotFiles(t, dir)
			got := request(t, a, "POST", personAction(t, link, "/wishes/save"), form)
			if got.Code != expected {
				t.Fatalf("status %d, want %d", got.Code, expected)
			}
			if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
				t.Fatal("rejected request changed files")
			}
		})
	}
}

// @test WSH-001-AC4 WSH-001-AC5 WSH-002-AC3
func TestWishEditClearsOptionalDetailsPreservingAttachedData(t *testing.T) {
	a, dir, parent, _ := newPeopleApp(t)
	addTestPerson(t, a, parent, "Anna")
	writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes:\n  - title: Wooden train\n    price: 12.50\n    links: [https://shop.example/train]\n    note: Old note\n    claims: [{name: Auntie}, {name: Uncle}]\n    future: retained\n  - title: Another train\n")
	page := request(t, a, "GET", wishLink(t, parent, "Anna", "2026"), nil)
	token := field(t, page.Body.String(), "wish")
	got := request(t, a, "POST", personAction(t, parent, "/wishes/save"), url.Values{"person": {"Anna"}, "wish": {token}, "title": {"  Updated train  "}})
	if got.Code != 303 {
		t.Fatal(got.Code)
	}
	raw, _ := os.ReadFile(listFile(dir, "2026", "Anna"))
	text := string(raw)
	for _, want := range []string{"Updated train", "Auntie", "Uncle", "future: retained", "Another train"} {
		if !strings.Contains(text, want) {
			t.Fatalf("lost %s", want)
		}
	}
	for _, gone := range []string{"price:", "links:", "note:"} {
		if strings.Contains(text, gone) {
			t.Fatalf("did not clear %s", gone)
		}
	}
}
