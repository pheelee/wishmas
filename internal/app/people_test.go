package app

import (
	"io/fs"
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

func personAction(t *testing.T, link, path string) string {
	t.Helper()
	parsed, err := url.Parse(link)
	if err != nil {
		t.Fatal(err)
	}
	return path + "?" + parsed.RawQuery
}
func newPeopleApp(t *testing.T) (*application, string, string, string) {
	t.Helper()
	dir := t.TempDir()
	handler := newLegacy(dir).(*application)
	handler.now = func() time.Time { return time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC) }
	parent, family := setupHousehold(t, handler, "Europe/Zurich")
	return handler, dir, parent, family
}
func listFile(dir, year, name string) string {
	return filepath.Join(dir, "households", "Family", year, "christmas", name, "wishlist.yaml")
}
func peopleFile(dir string) string { return filepath.Join(dir, "households", "Family", "people.yaml") }
func writeTestYAML(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
}
func snapshotFiles(t *testing.T, dir string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		result[relative] = string(raw)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func addTestPerson(t *testing.T, h http.Handler, parent, name string) {
	t.Helper()
	got := request(t, h, "POST", personAction(t, parent, "/people/add"), url.Values{"name": {name}})
	if got.Code != 303 || got.Header().Get("Location") != parent {
		t.Fatalf("add person returned %d, expected 303", got.Code)
	}
}
func wishLink(t *testing.T, link, name, year string) string {
	t.Helper()
	parsed, err := url.Parse(link)
	if err != nil {
		t.Fatal(err)
	}
	q := parsed.Query()
	q.Set("person", name)
	q.Set("year", year)
	return "/wishlist?" + q.Encode()
}

// @test WLS-001-AC1 WLS-001-AC2 WLS-001-AC8 CON-004
func TestParentsCreatePeopleAndPersistentEmptyListsInAddedOrder(t *testing.T) {
	handler, dir, parent, family := newPeopleApp(t)
	names := []string{"Zora", "Émilie", "Liam"}
	for _, name := range names {
		addTestPerson(t, handler, parent, "  "+name+"  ")
		raw, err := os.ReadFile(listFile(dir, "2026", name))
		if err != nil {
			t.Fatal(err)
		}
		var data struct {
			Wishes []any `yaml:"wishes"`
		}
		if err := yaml.Unmarshal(raw, &data); err != nil {
			t.Fatal(err)
		}
		if len(data.Wishes) != 0 {
			t.Fatal("new wishlist is not empty")
		}
	}
	var roster struct {
		People []struct {
			Name string `yaml:"name"`
		} `yaml:"people"`
	}
	raw, err := os.ReadFile(peopleFile(dir))
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(raw, &roster); err != nil {
		t.Fatal(err)
	}
	if len(roster.People) != len(names) {
		t.Fatal("people were not saved")
	}
	for i, p := range roster.People {
		if p.Name != names[i] {
			t.Fatal("roster not in added order")
		}
	}
	restarted := newLegacy(dir).(*application)
	restarted.now = handler.now
	for _, h := range []http.Handler{handler, restarted} {
		for _, link := range []string{parent, family} {
			page := request(t, h, "GET", link, nil)
			if page.Code != 200 {
				t.Fatalf("people page: %d", page.Code)
			}
			last := -1
			for _, name := range names {
				position := strings.Index(page.Body.String(), ">"+name+"</")
				if position <= last {
					t.Fatalf("missing or unordered %s: %s", name, page.Body.String())
				}
				last = position
				list := request(t, h, "GET", wishLink(t, link, name, "2026"), nil)
				if list.Code != 200 || !strings.Contains(list.Body.String(), "No wishes yet.") {
					t.Fatalf("empty list unavailable: %d %s", list.Code, list.Body.String())
				}
			}
		}
	}
}

// @test WLS-003-AC1 WLS-003-AC2 WLS-003-AC4 WLS-003-AC5 WLS-003-AC6
func TestPersonValidationAndArchivedNameReservations(t *testing.T) {
	handler, dir, parent, _ := newPeopleApp(t)
	writeTestYAML(t, peopleFile(dir), "people:\n  - name: Anna\n  - name: Samuel\n    archived: true\n")
	for _, name := range []string{"", "  ", "ab", strings.Repeat("a", 51), "Anne Marie", "../Anna", "Anna!", "anna", " Anna ", "SAMUEL"} {
		t.Run(name, func(t *testing.T) {
			before := snapshotFiles(t, dir)
			got := request(t, handler, "POST", personAction(t, parent, "/people/add"), url.Values{"name": {name}})
			if got.Code != 400 {
				t.Fatalf("invalid/duplicate name returned %d, expected 400", got.Code)
			}
			if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
				t.Fatal("rejected name changed files")
			}
		})
	}
	for _, name := range []string{"Éø1", strings.Repeat("É", 50), strings.Repeat("𐐀", 50), "ſam"} {
		addTestPerson(t, handler, parent, name)
	}
	got := request(t, handler, "POST", personAction(t, parent, "/people/add"), url.Values{"name": {"Sam"}})
	if got.Code != 400 {
		t.Fatal("Unicode-equivalent case collision accepted")
	}
}

// @test WLS-001-AC6 ACC-003-AC4
func TestFamilyAndRevokedLinksCannotAddPeople(t *testing.T) {
	handler, dir, parent, family := newPeopleApp(t)
	before := snapshotFiles(t, dir)
	for _, link := range []string{family, "/?code=invalid", "/"} {
		got := request(t, handler, "POST", personAction(t, link, "/people/add"), url.Values{"name": {"Anna"}})
		if got.Code != 404 {
			t.Fatalf("unauthorized creation: %d", got.Code)
		}
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("unauthorized action changed files")
	}
	if got := request(t, handler, "GET", parent, nil); !strings.Contains(got.Body.String(), `name="name"`) {
		t.Fatal("parent creation form missing")
	}
	if got := request(t, handler, "GET", family, nil); strings.Contains(got.Body.String(), `name="name"`) {
		t.Fatal("family sees creation form")
	}
	rotated := request(t, handler, "POST", regeneratePath(t, parent), url.Values{"role": {"parent"}})
	if rotated.Code != 303 {
		t.Fatal("parent regeneration failed")
	}
	got := request(t, handler, "POST", personAction(t, parent, "/people/add"), url.Values{"name": {"Anna"}})
	if got.Code != 404 {
		t.Fatal("stale parent page retained creation permission")
	}
	addTestPerson(t, handler, rotated.Header().Get("Location"), "Anna")
}

// @test WLS-001-AC5 WLS-001-AC7 CON-004
func TestManualPeopleAndWishlistEditsReloadWithoutOverwritingInvalidFiles(t *testing.T) {
	handler, dir, parent, family := newPeopleApp(t)
	writeTestYAML(t, peopleFile(dir), "people:\n  - name: Anna\n  - name: Bobby\n")
	got := request(t, handler, "GET", family, nil)
	if got.Code != 200 || !strings.Contains(got.Body.String(), ">Bobby</") {
		t.Fatal("manual person not loaded")
	}
	writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes:\n  - title: Train set\n")
	for _, title := range []string{"Train set", "Story books"} {
		writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes:\n  - title: "+title+"\n")
		page := request(t, handler, "GET", wishLink(t, family, "Anna", "2026"), nil)
		if page.Code != 200 || !strings.Contains(page.Body.String(), title) {
			t.Fatalf("manual wish edit not loaded: %d %s", page.Code, page.Body.String())
		}
	}
	for _, invalid := range []string{"people: [broken", "people:\n  - name: Anna\n  - name: anna\n", "people:\n  - name: Samuel\n    archived: true\n  - name: SAMUEL\n"} {
		writeTestYAML(t, peopleFile(dir), invalid)
		before := snapshotFiles(t, dir)
		if got := request(t, handler, "GET", parent, nil); got.Code != 500 {
			t.Fatalf("invalid people not rejected: %d", got.Code)
		}
		if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
			t.Fatal("invalid roster changed files")
		}
	}
	writeTestYAML(t, peopleFile(dir), "people:\n  - name: Anna\n  - name: Bobby\n")
	writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes: [broken")
	before := snapshotFiles(t, dir)
	got = request(t, handler, "POST", personAction(t, parent, "/people/add"), url.Values{"name": {"Clara"}})
	if got.Code != 500 {
		t.Fatal("invalid wishlist did not block creation")
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("invalid wishlist changed files")
	}
}

// @test WLS-002-AC1 WLS-002-AC2 WLS-004-AC4 ACC-002-AC3 ACC-002-AC4
func TestNewCalendarYearCreatesEmptyListsOnlyForActivePeople(t *testing.T) {
	handler, dir, parent, family := newPeopleApp(t)
	writeTestYAML(t, peopleFile(dir), "people:\n  - name: Anna\n  - name: Bobby\n    archived: true\n")
	previous := "wishes:\n  - title: Train set\n    claims:\n      - name: Auntie\n"
	writeTestYAML(t, listFile(dir, "2026", "Anna"), previous)
	handler.now = func() time.Time { return time.Date(2026, 12, 31, 23, 0, 0, 0, time.UTC) }
	page := request(t, handler, "GET", family, nil)
	if page.Code != 200 || !strings.Contains(page.Body.String(), "Family access · 2027") {
		t.Fatal("new household year not selected")
	}
	raw, err := os.ReadFile(listFile(dir, "2027", "Anna"))
	if err != nil {
		t.Fatal(err)
	}
	var data struct {
		Wishes []any `yaml:"wishes"`
	}
	if err := yaml.Unmarshal(raw, &data); err != nil || len(data.Wishes) != 0 {
		t.Fatal("new list copied previous wishes/claims")
	}
	if _, err := os.Stat(listFile(dir, "2027", "Bobby")); !os.IsNotExist(err) {
		t.Fatal("archived person received a new list")
	}
	old, err := os.ReadFile(listFile(dir, "2026", "Anna"))
	if err != nil || string(old) != previous {
		t.Fatal("history changed")
	}
	if got := request(t, handler, "GET", wishLink(t, family, "Anna", "2027"), nil); got.Code != 200 || !strings.Contains(got.Body.String(), "No wishes yet.") {
		t.Fatal("family link did not open the new empty current-year list")
	}
	if got := request(t, handler, "GET", wishLink(t, family, "Anna", "2026"), nil); got.Code != 404 {
		t.Fatal("family could access past year")
	}
	if got := request(t, handler, "GET", wishLink(t, parent, "Bobby", "2027"), nil); got.Code != 404 {
		t.Fatal("archived list was exposed")
	}
}

// @test WLS-003-AC1 WLS-003-AC2
func TestConcurrentDuplicatePeopleAreNotCreated(t *testing.T) {
	handler, dir, parent, _ := newPeopleApp(t)
	results := make(chan int, 2)
	start := make(chan struct{})
	path := personAction(t, parent, "/people/add")
	for _, name := range []string{"Anna", "anna"} {
		go func(name string) {
			<-start
			got := request(t, handler, "POST", path, url.Values{"name": {name}})
			results <- got.Code
		}(name)
	}
	close(start)
	successes, rejected := 0, 0
	for range 2 {
		switch status := <-results; status {
		case 303:
			successes++
		case 400:
			rejected++
		default:
			t.Fatalf("unexpected creation status: %d", status)
		}
	}
	if successes != 1 || rejected != 1 {
		t.Fatal("duplicate concurrent creations both succeeded")
	}
	raw, err := os.ReadFile(peopleFile(dir))
	if err != nil {
		t.Fatal(err)
	}
	var data struct {
		People []any `yaml:"people"`
	}
	if err := yaml.Unmarshal(raw, &data); err != nil || len(data.People) != 1 {
		t.Fatal("duplicate people persisted")
	}
}

// @test CON-004 WLS-001-AC7
func TestCreationDoesNotOverwriteAnExistingManualList(t *testing.T) {
	handler, dir, parent, _ := newPeopleApp(t)
	writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes:\n  - title: Train set\n")
	before := snapshotFiles(t, dir)
	got := request(t, handler, "POST", personAction(t, parent, "/people/add"), url.Values{"name": {"Anna"}})
	if got.Code != 500 {
		t.Fatalf("expected an existing-data error, got %d", got.Code)
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("manual list overwritten or person partially created")
	}
}

// @test ACC-002-AC3 ACC-003-AC2
func TestReplacementLinksRetainAllCurrentPeopleAndYearRestrictions(t *testing.T) {
	handler, _, parent, oldFamily := newPeopleApp(t)
	for _, name := range []string{"Anna", "Bobby"} {
		addTestPerson(t, handler, parent, name)
	}
	rotation := request(t, handler, "POST", regeneratePath(t, parent), url.Values{"role": {"family"}})
	if rotation.Code != 303 {
		t.Fatal("family regeneration failed")
	}
	home := request(t, handler, "GET", parent, nil)
	family := field(t, home.Body.String(), "family_link")
	rotation = request(t, handler, "POST", regeneratePath(t, parent), url.Values{"role": {"parent"}})
	if rotation.Code != 303 {
		t.Fatal("parent regeneration failed")
	}
	newParent := rotation.Header().Get("Location")
	for _, name := range []string{"Anna", "Bobby"} {
		for _, link := range []string{family, newParent} {
			if got := request(t, handler, "GET", wishLink(t, link, name, "2026"), nil); got.Code != 200 {
				t.Fatalf("replacement lost current-year %s access: %d", name, got.Code)
			}
		}
		if got := request(t, handler, "GET", wishLink(t, family, name, "2025"), nil); got.Code != 404 {
			t.Fatal("replacement family code widened year access")
		}
	}
	for _, link := range []string{oldFamily, parent} {
		if got := request(t, handler, "GET", wishLink(t, link, "Anna", "2026"), nil); got.Code != 404 {
			t.Fatal("old code retained list access")
		}
	}
}
