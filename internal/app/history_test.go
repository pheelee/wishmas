package app

import (
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func personForm(name, newName string) url.Values {
	return url.Values{"person": {name}, "name": {newName}}
}
func yearHome(t *testing.T, link, year string) string {
	t.Helper()
	u, err := url.Parse(link)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("year", year)
	return "/?" + q.Encode()
}
func seedHistory(t *testing.T) (*application, string, string, string, map[string]string) {
	t.Helper()
	h, dir, parent, family := newPeopleApp(t)
	addTestPerson(t, h, parent, "Anna")
	addTestPerson(t, h, parent, "Bobby")
	data := map[string]string{
		"2024": "# Keep this comment\nwishes:\n  - title: Wooden train\n    claims:\n      - name: Auntie\n      - name: Uncle\n",
		"2025": "wishes:\n  - title: Story books\n    note: Preserve this note\n",
		"2026": "wishes:\n  - title: Building blocks\n    claims:\n      - name: Grandma\n",
	}
	for year, raw := range data {
		writeTestYAML(t, listFile(dir, year, "Anna"), raw)
	}
	return h, dir, parent, family, data
}

// @test WLS-003-AC3 WLS-001-AC8 CON-004
func TestRenamePreservesAllYearsAndClaims(t *testing.T) {
	h, dir, parent, family, originals := seedHistory(t)
	birthday := filepath.Join(dir, "households", "Family", "2025", "birthday", "Anna", "wishlist.yaml")
	writeTestYAML(t, birthday, "wishes:\n  - title: Birthday keepsake\n")
	got := request(t, h, "POST", personAction(t, parent, "/people/rename"), personForm("Anna", "  Alicia  "))
	if got.Code != 303 {
		t.Fatalf("rename returned %d", got.Code)
	}
	for year, original := range originals {
		raw, err := os.ReadFile(listFile(dir, year, "Alicia"))
		if err != nil || string(raw) != original {
			t.Fatalf("lost data or claims in %s: %v", year, err)
		}
		if _, err := os.Stat(listFile(dir, year, "Anna")); !os.IsNotExist(err) {
			t.Fatal("old name directory still present")
		}
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(filepath.Dir(birthday)), "Alicia", "wishlist.yaml")); err != nil {
		t.Fatal("other event data was not preserved")
	}
	restarted := newLegacy(dir).(*application)
	restarted.now = h.now
	for _, handler := range []*application{h, restarted} {
		current := request(t, handler, "GET", wishLink(t, family, "Alicia", "2026"), nil)
		if current.Code != 200 || !strings.Contains(current.Body.String(), "Building blocks") {
			t.Fatal("renamed current list unavailable")
		}
		old := request(t, handler, "GET", wishLink(t, parent, "Alicia", "2024"), nil)
		if old.Code != 200 || !strings.Contains(old.Body.String(), "Wooden train") {
			t.Fatal("renamed historical list unavailable")
		}
		if got := request(t, handler, "GET", wishLink(t, parent, "Anna", "2026"), nil); got.Code != 404 {
			t.Fatal("old name still exposed")
		}
		home := request(t, handler, "GET", parent, nil).Body.String()
		if strings.Index(home, ">Alicia</") < 0 || strings.Index(home, ">Alicia</") > strings.Index(home, ">Bobby</") {
			t.Fatal("rename changed added order")
		}
	}
}

// @test WLS-003-AC1 WLS-003-AC2 WLS-003-AC4 WLS-003-AC5 WLS-003-AC6
func TestRenameUsesConfirmedValidationAndArchivedReservations(t *testing.T) {
	h, dir, parent, _, originals := seedHistory(t)
	writeTestYAML(t, peopleFile(dir), "people:\n  - name: Anna\n  - name: Bobby\n  - name: Samuel\n    archived: true\n")
	for _, name := range []string{"", "  ", "ab", strings.Repeat("a", 51), "Bad Name", "Anna!", "bobby", " SAMUEL "} {
		before := snapshotFiles(t, dir)
		got := request(t, h, "POST", personAction(t, parent, "/people/rename"), personForm("Anna", name))
		if got.Code != 400 {
			t.Fatalf("invalid rename %q returned %d", name, got.Code)
		}
		if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
			t.Fatal("rejected rename changed files")
		}
	}
	for _, name := range []string{"ANNA", "Élodie", strings.Repeat("𐐀", 50)} {
		got := request(t, h, "POST", personAction(t, parent, "/people/rename"), personForm("Anna", name))
		if got.Code != 303 {
			t.Fatalf("valid rename returned %d", got.Code)
		}
		for year, raw := range originals {
			got, err := os.ReadFile(listFile(dir, year, name))
			if err != nil || string(got) != raw {
				t.Fatal("valid rename lost data")
			}
		}
		got = request(t, h, "POST", personAction(t, parent, "/people/rename"), personForm(name, "Anna"))
		if got.Code != 303 {
			t.Fatal("rename back failed")
		}
	}
	before := snapshotFiles(t, dir)
	if got := request(t, h, "POST", personAction(t, parent, "/people/rename"), personForm("Anna", " Anna ")); got.Code != 303 {
		t.Fatal("unchanged trimmed name rejected")
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("unchanged name altered files")
	}
}

// @test WLS-002-AC3 WLS-002-AC4 WLS-002-AC5 ACC-002-AC2 ACC-002-AC3
func TestParentsBrowseReadOnlyHistoryAndFamilyCannot(t *testing.T) {
	h, dir, parent, family, _ := seedHistory(t)
	before := snapshotFiles(t, dir)
	for _, year := range []string{"2024", "2025"} {
		home := request(t, h, "GET", yearHome(t, parent, year), nil)
		if home.Code != 200 || !strings.Contains(home.Body.String(), `value="`+year+`"`) {
			t.Fatalf("year navigation unavailable: %d", home.Code)
		}
		if strings.Contains(home.Body.String(), `action="/people/add`) {
			t.Fatal("history exposes creation controls")
		}
		list := request(t, h, "GET", wishLink(t, parent, "Anna", year), nil)
		if list.Code != 200 || !strings.Contains(list.Body.String(), "Read-only history") {
			t.Fatalf("historical list unavailable: %d", list.Code)
		}
		if strings.Contains(list.Body.String(), "/people/rename") || strings.Contains(list.Body.String(), "/people/archive") {
			t.Fatal("historical list exposes mutation controls")
		}
		for _, path := range []string{yearHome(t, family, year), wishLink(t, family, "Anna", year)} {
			if got := request(t, h, "GET", path, nil); got.Code != 404 {
				t.Fatal("family accessed history")
			}
		}
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("browsing history changed data")
	}
	writeTestYAML(t, listFile(dir, "2025", "Anna"), "wishes:\n  - title: Updated story books\n")
	got := request(t, h, "GET", wishLink(t, parent, "Anna", "2025"), nil)
	if got.Code != 200 || !strings.Contains(got.Body.String(), "Updated story books") {
		t.Fatal("manual history edit not reloaded")
	}
	before = snapshotFiles(t, dir)
	if got := request(t, h, "GET", yearHome(t, parent, "1900"), nil); got.Code != 404 {
		t.Fatal("unsaved year did not return not found")
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("unsaved historical lists were created")
	}
	familyPage := request(t, h, "GET", family, nil).Body.String()
	if strings.Contains(familyPage, `name="year"`) || strings.Contains(familyPage, "Updated story books") {
		t.Fatal("family page exposed history")
	}
}

// @test WLS-004-AC1 WLS-004-AC2 WLS-004-AC3 WLS-004-AC4 WLS-004-AC5 WLS-004-AC6 WLS-004-AC7 WLS-003-AC4
func TestArchiveHidesAllYearsAndManualRestorePreservesData(t *testing.T) {
	h, dir, parent, family, originals := seedHistory(t)
	got := request(t, h, "POST", personAction(t, parent, "/people/archive"), personForm("Anna", ""))
	if got.Code != 303 {
		t.Fatalf("archive returned %d", got.Code)
	}
	for year, raw := range originals {
		saved, err := os.ReadFile(listFile(dir, year, "Anna"))
		if err != nil || string(saved) != raw {
			t.Fatal("archive deleted or rewrote a list")
		}
		for _, link := range []string{parent, family} {
			if got := request(t, h, "GET", wishLink(t, link, "Anna", year), nil); got.Code != 404 {
				t.Fatal("archived list still visible")
			}
		}
	}
	for _, link := range []string{parent, family} {
		body := request(t, h, "GET", link, nil).Body.String()
		if strings.Contains(body, ">Anna</") || strings.Contains(body, "/people/restore") {
			t.Fatal("archived person or restore action exposed")
		}
	}
	before := snapshotFiles(t, dir)
	if got := request(t, h, "POST", personAction(t, parent, "/people/add"), url.Values{"name": {"anna"}}); got.Code != 400 {
		t.Fatal("archived name was released")
	}
	if got := request(t, h, "POST", personAction(t, parent, "/people/restore"), personForm("Anna", "")); got.Code != 404 {
		t.Fatal("app exposed restoration")
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("rejected action modified data")
	}
	h.now = func() time.Time { return time.Date(2027, 1, 1, 12, 0, 0, 0, time.UTC) }
	if got := request(t, h, "GET", family, nil); got.Code != 200 {
		t.Fatal("new-year page failed")
	}
	if _, err := os.Stat(listFile(dir, "2027", "Anna")); !os.IsNotExist(err) {
		t.Fatal("archived person got a yearly list")
	}
	writeTestYAML(t, peopleFile(dir), "people:\n  - name: Anna\n  - name: Bobby\n")
	if got := request(t, h, "GET", family, nil); got.Code != 200 || !strings.Contains(got.Body.String(), ">Anna</") {
		t.Fatal("manual restore not loaded")
	}
	raw, err := os.ReadFile(listFile(dir, "2027", "Anna"))
	if err != nil || !strings.Contains(string(raw), "wishes: []") {
		t.Fatal("restore did not create a missing current list")
	}
	existing := "wishes:\n  - title: Keep this current wish\n    claims:\n      - name: Auntie\n"
	writeTestYAML(t, listFile(dir, "2027", "Anna"), existing)
	request(t, h, "POST", personAction(t, parent, "/people/archive"), personForm("Anna", ""))
	writeTestYAML(t, peopleFile(dir), "people:\n  - name: Anna\n  - name: Bobby\n")
	if got := request(t, h, "GET", family, nil); got.Code != 200 {
		t.Fatal("second restore failed")
	}
	after, err := os.ReadFile(listFile(dir, "2027", "Anna"))
	if err != nil || string(after) != existing {
		t.Fatal("restore replaced an existing current list")
	}
}

// @test WLS-001-AC6 WLS-002-AC4 ACC-003-AC4
func TestPersonChangesRejectFamilyStaleAndHistoricalActions(t *testing.T) {
	h, dir, parent, family, _ := seedHistory(t)
	actions := []string{"/people/rename", "/people/archive", "/people/add"}
	before := snapshotFiles(t, dir)
	for _, action := range actions {
		for _, link := range []string{family, "/?code=invalid", "/"} {
			if got := request(t, h, "POST", personAction(t, link, action), personForm("Anna", "Alicia")); got.Code != 404 {
				t.Fatal("unauthorized mutation succeeded")
			}
		}
		form := personForm("Anna", "Alicia")
		form.Set("year", "2025")
		if got := request(t, h, "POST", personAction(t, parent, action), form); got.Code != 404 {
			t.Fatalf("historical action returned %d", got.Code)
		}
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("rejected action modified files")
	}
	rotation := request(t, h, "POST", regeneratePath(t, parent), url.Values{"role": {"parent"}})
	if rotation.Code != 303 {
		t.Fatal("parent rotation failed")
	}
	before = snapshotFiles(t, dir)
	for _, action := range actions {
		if got := request(t, h, "POST", personAction(t, parent, action), personForm("Anna", "Alicia")); got.Code != 404 {
			t.Fatal("stale parent retained mutation access")
		}
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("stale action modified files")
	}
}

// @test ACC-002-AC2 ACC-003-AC2
func TestRegeneratedParentLinksRetainHistoryScope(t *testing.T) {
	h, _, parent, family, _ := seedHistory(t)
	for _, role := range []string{"family", "parent"} {
		rotation := request(t, h, "POST", regeneratePath(t, parent), url.Values{"role": {role}})
		if rotation.Code != 303 {
			t.Fatal("rotation failed")
		}
		parent = rotation.Header().Get("Location")
		home := request(t, h, "GET", parent, nil).Body.String()
		family = field(t, home, "family_link")
		for _, year := range []string{"2024", "2025", "2026"} {
			if got := request(t, h, "GET", wishLink(t, parent, "Anna", year), nil); got.Code != 200 {
				t.Fatalf("replacement parent lost %s", year)
			}
			expected := 404
			if year == "2026" {
				expected = 200
			}
			if got := request(t, h, "GET", wishLink(t, family, "Anna", year), nil); got.Code != expected {
				t.Fatal("replacement family scope changed")
			}
		}
	}
}

// @test WLS-001-AC7 WLS-003-AC3 CON-004
func TestInvalidHistoricalDataAndRenameConflictsLeaveFilesUntouched(t *testing.T) {
	h, dir, parent, _, _ := seedHistory(t)
	writeTestYAML(t, listFile(dir, "2025", "Anna"), "wishes: [broken")
	before := snapshotFiles(t, dir)
	if got := request(t, h, "GET", wishLink(t, parent, "Anna", "2025"), nil); got.Code != 500 {
		t.Fatal("invalid history did not show an error")
	}
	if got := request(t, h, "POST", personAction(t, parent, "/people/rename"), personForm("Anna", "Alicia")); got.Code != 500 {
		t.Fatal("invalid history was renamed")
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("invalid history changed files")
	}
	writeTestYAML(t, listFile(dir, "2025", "Anna"), "wishes:\n  - title: Story books\n")
	writeTestYAML(t, listFile(dir, "2026", "Alicia"), "wishes:\n  - title: Preserve manual data\n")
	before = snapshotFiles(t, dir)
	if got := request(t, h, "POST", personAction(t, parent, "/people/rename"), personForm("Anna", "Alicia")); got.Code == 303 {
		t.Fatal("conflicting directory was overwritten")
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("rename conflict partially moved data")
	}
}

// @test WLS-003-AC3 CON-004
func TestFailedRenameRollsBackEarlierDirectoryMoves(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission failure requires an unprivileged user")
	}
	h, dir, parent, _, _ := seedHistory(t)
	blocked := filepath.Dir(filepath.Dir(listFile(dir, "2026", "Anna")))
	if err := os.Chmod(blocked, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(blocked, 0700) })
	before := snapshotFiles(t, dir)
	got := request(t, h, "POST", personAction(t, parent, "/people/rename"), personForm("Anna", "Alicia"))
	if got.Code != 500 {
		t.Fatalf("expected failed rename, got %d", got.Code)
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("failed rename left earlier years moved")
	}
}

// @test WLS-003-AC3 CON-004
func TestFailedRosterSaveRollsBackRenamedLists(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission failure requires an unprivileged user")
	}
	h, dir, parent, _, _ := seedHistory(t)
	root := filepath.Join(dir, "households", "Family")
	if err := os.Chmod(root, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(root, 0700) })
	before := snapshotFiles(t, dir)
	got := request(t, h, "POST", personAction(t, parent, "/people/rename"), personForm("Anna", "Alicia"))
	if got.Code != 500 {
		t.Fatalf("expected roster save failure, got %d", got.Code)
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("failed roster save did not restore all lists")
	}
}
