package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

func namedRequest(t *testing.T, a http.Handler, method, path string, form url.Values, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	a.ServeHTTP(w, r)
	return w
}
func nameCookie(name string) *http.Cookie {
	return &http.Cookie{Name: "wishmas_claim_name", Value: url.QueryEscape(name)}
}
func claimFixture(t *testing.T) (*application, string, string, string) {
	t.Helper()
	a, dir, parent, family := newPeopleApp(t)
	addTestPerson(t, a, parent, "Anna")
	writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes: [{title: Wooden train}, {title: Another train}]\n")
	return a, dir, parent, family
}
func claimForm(t *testing.T, a http.Handler, link string) url.Values {
	t.Helper()
	page := request(t, a, "GET", wishLink(t, link, "Anna", "2026"), nil)
	if page.Code != 200 {
		t.Fatalf("read list: %d", page.Code)
	}
	match := regexp.MustCompile(`name="wish" value="([^"]+)"`).FindStringSubmatch(page.Body.String())
	if len(match) != 2 {
		t.Fatal("missing wish action on current-year list")
	}
	return url.Values{"person": {"Anna"}, "year": {"2026"}, "wish": {match[1]}, "mode": {"add"}}
}
func responseName(t *testing.T, w *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, c := range w.Result().Cookies() {
		if c.Name == "wishmas_claim_name" {
			return c
		}
	}
	t.Fatal("missing remembered name")
	return nil
}

// @test CLM-001-AC1 CLM-001-AC5 CLM-001-AC6 CLM-002-AC1 CLM-002-AC6
func TestFirstClaimPromptsAndRemembersValidatedName(t *testing.T) {
	a, dir, _, family := claimFixture(t)
	form := claimForm(t, a, family)
	before := snapshotFiles(t, dir)
	if page := request(t, a, "GET", wishLink(t, family, "Anna", "2026"), nil); strings.Contains(page.Body.String(), "Save name and claim") {
		t.Fatal("name prompted while browsing")
	}
	prompt := namedRequest(t, a, "POST", personAction(t, family, "/claims/toggle"), form, nil)
	if prompt.Code != 200 || !strings.Contains(prompt.Body.String(), "Save name and claim") {
		t.Fatalf("prompt: %d", prompt.Code)
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("prompt changed data")
	}
	for _, bad := range []string{"", "Al", "Aunt!", strings.Repeat("é", 51)} {
		form.Set("name", bad)
		got := namedRequest(t, a, "POST", personAction(t, family, "/claims/toggle"), form, nil)
		if got.Code != 400 || len(got.Result().Cookies()) != 0 || !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
			t.Fatalf("invalid name %q: %d", bad, got.Code)
		}
	}
	form.Set("name", "  "+strings.Repeat("𐐀", 50)+"  ")
	got := namedRequest(t, a, "POST", personAction(t, family, "/claims/toggle"), form, nil)
	if got.Code != 303 {
		t.Fatalf("claim: %d", got.Code)
	}
	cookie := responseName(t, got)
	name, _ := url.QueryUnescape(cookie.Value)
	if name != strings.Repeat("𐐀", 50) || cookie.MaxAge <= 0 || !cookie.HttpOnly {
		t.Fatal("name not trimmed/persisted")
	}
	page := namedRequest(t, a, "GET", wishLink(t, family, "Anna", "2026"), nil, cookie)
	if !strings.Contains(page.Body.String(), name) || !strings.Contains(page.Body.String(), "Claimed by you") {
		t.Fatal("remembered name/ownership missing")
	}
}

// @test CLM-001-AC2 CLM-001-AC3 CLM-001-AC4 CLM-001-AC7 CLM-002-AC2 CLM-002-AC3 CLM-002-AC5
// @test CLM-003-AC3 CLM-003-AC4 CLM-003-AC6 CLM-004-AC1 CLM-004-AC2 CLM-004-AC3 CLM-004-AC4 CLM-004-AC5 CLM-004-AC7
func TestClaimOwnershipRenamingAndRemoval(t *testing.T) {
	a, dir, parent, family := claimFixture(t)
	post := func(link string, f url.Values, c *http.Cookie) *httptest.ResponseRecorder {
		return namedRequest(t, a, "POST", personAction(t, link, "/claims/toggle"), f, c)
	}
	oldForm := claimForm(t, a, family)
	if got := post(family, oldForm, nameCookie("Auntie")); got.Code != 303 {
		t.Fatalf("first %d", got.Code)
	}
	if got := post(parent, claimForm(t, a, parent), nameCookie("Uncle")); got.Code != 303 {
		t.Fatalf("second %d", got.Code)
	}
	before := snapshotFiles(t, dir)
	stale := post(family, oldForm, nameCookie(" auntie "))
	if stale.Code != 409 || !strings.Contains(stale.Body.String(), "role=\"alert\"") || !strings.Contains(stale.Body.String(), "Claimed · 2 claimants") {
		t.Fatalf("stale refresh: %d", stale.Code)
	}
	duplicate := post(family, claimForm(t, a, family), nameCookie("AUNTIE"))
	if duplicate.Code != 409 || !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("duplicate/stale modified claims")
	}
	page := namedRequest(t, a, "GET", wishLink(t, family, "Anna", "2026"), nil, nameCookie(" auntie "))
	if !strings.Contains(page.Body.String(), "Claimed by you") || strings.Contains(page.Body.String(), "Uncle") {
		t.Fatal("ownership/privacy incorrect")
	}
	rename := namedRequest(t, a, "POST", personAction(t, family, "/claim-name"), url.Values{"person": {"Anna"}, "year": {"2026"}, "name": {" Cousin "}}, nameCookie("Auntie"))
	if rename.Code != 303 || !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("rename modified old claims")
	}
	cousin := responseName(t, rename)
	page = namedRequest(t, a, "GET", wishLink(t, family, "Anna", "2026"), nil, cousin)
	if strings.Contains(page.Body.String(), "Claimed by you") {
		t.Fatal("rename retained old ownership")
	}
	remove := claimForm(t, a, family)
	remove.Set("mode", "remove")
	remove.Set("name", "Auntie")
	remove.Set("claimant", "Auntie")
	if got := post(family, remove, cousin); got.Code != 403 || !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatalf("old-name removal: %d", got.Code)
	}
	remove.Set("mode", "remove-other")
	if got := post(family, remove, cousin); got.Code != 403 {
		t.Fatal("family removed another claim")
	}
	if got := post(family, claimForm(t, a, family), cousin); got.Code != 303 {
		t.Fatal(got.Code)
	}
	remove = claimForm(t, a, family)
	remove.Set("mode", "remove")
	if got := post(family, remove, cousin); got.Code != 303 {
		t.Fatal(got.Code)
	}
	remove = claimForm(t, a, parent)
	remove.Set("mode", "remove-other")
	remove.Set("claimant", " auntie ")
	if got := post(parent, remove, nil); got.Code != 303 {
		t.Fatal(got.Code)
	}
	page = request(t, a, "GET", wishLink(t, parent, "Anna", "2026"), nil)
	if !strings.Contains(page.Body.String(), "Uncle") || strings.Contains(page.Body.String(), ">Auntie<") {
		t.Fatal("wrong claim removed")
	}
	remove = claimForm(t, a, parent)
	remove.Set("mode", "remove-other")
	remove.Set("claimant", "Uncle")
	if got := post(parent, remove, nil); got.Code != 303 {
		t.Fatal(got.Code)
	}
	page = request(t, a, "GET", wishLink(t, family, "Anna", "2026"), nil)
	if strings.Contains(page.Body.String(), "Claimed ·") {
		t.Fatal("last claim remained")
	}
}

// @test CLM-002-AC4 CLM-002-AC5 CLM-004-AC6 CON-004
func TestClaimMutationGuards(t *testing.T) {
	for _, scenario := range []string{"past", "query-year", "rollover", "archived", "invalid", "stale", "revoked", "missing-person", "get"} {
		t.Run(scenario, func(t *testing.T) {
			a, dir, parent, family := claimFixture(t)
			form := claimForm(t, a, family)
			status := 404
			method := "POST"
			link := family
			switch scenario {
			case "past":
				form.Set("year", "2025")
			case "query-year":
				link += "&year=2025"
			case "rollover":
				a.now = func() time.Time { return time.Date(2027, 1, 2, 0, 0, 0, 0, time.UTC) }
			case "archived":
				writeTestYAML(t, peopleFile(dir), "people: [{name: Anna, archived: true}]\n")
			case "invalid":
				writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes: [invalid]\n")
				status = 500
			case "stale":
				form.Set("wish", "missing")
				status = 409
			case "revoked":
				got := request(t, a, "POST", personAction(t, parent, "/access/regenerate"), url.Values{"role": {"family"}})
				if got.Code != 303 {
					t.Fatal(got.Code)
				}
			case "missing-person":
				form.Set("person", "Missing")
			case "get":
				method = "GET"
			}
			before := snapshotFiles(t, dir)
			for _, mode := range []string{"add", "remove", "remove-other"} {
				form.Set("mode", mode)
				got := namedRequest(t, a, method, personAction(t, link, "/claims/toggle"), form, nameCookie("Auntie"))
				if got.Code != status {
					t.Fatalf("%s: got %d want %d", mode, got.Code, status)
				}
				if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
					t.Fatal("rejected action changed files")
				}
			}
		})
	}
}

// @test CLM-001-AC2 CLM-001-AC3 CLM-001-AC6 CLM-004-AC6
func TestClaimNameValidationAndHistoryDoesNotRewriteClaims(t *testing.T) {
	a, dir, parent, family := claimFixture(t)
	writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes: [{title: Wooden train, claims: [{name: Auntie}]}]\n")
	writeTestYAML(t, listFile(dir, "2025", "Anna"), "wishes: [{title: Historic train, claims: [{name: Auntie}]}]\n")
	before := snapshotFiles(t, dir)
	for _, bad := range []string{"", "Al", "Aunt!", strings.Repeat("𐐀", 51)} {
		got := namedRequest(t, a, "POST", personAction(t, family, "/claim-name"), url.Values{"person": {"Anna"}, "year": {"2026"}, "name": {bad}}, nameCookie("Auntie"))
		if got.Code != 400 || len(got.Result().Cookies()) != 0 || !strings.Contains(got.Body.String(), "role=\"alert\"") {
			t.Fatalf("invalid rename %q: %d", bad, got.Code)
		}
		if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
			t.Fatal("invalid rename rewrote claims")
		}
	}
	got := namedRequest(t, a, "POST", personAction(t, parent, "/claim-name"), url.Values{"person": {"Anna"}, "year": {"2025"}, "name": {"NewName"}}, nameCookie("Auntie"))
	if got.Code != 303 || !strings.Contains(got.Header().Get("Location"), "year=2025") || !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("historical name change modified files/context")
	}
}

// @test CLM-001-AC4 CLM-001-AC7 CLM-002-AC2 CLM-002-AC3 CLM-003-AC3
func TestClaimNamesKeepAccentsAndConcurrentDuplicatesAreSafe(t *testing.T) {
	a, dir, _, family := claimFixture(t)
	form := claimForm(t, a, family)
	results := make(chan int, 2)
	for _, name := range []string{"Auntie", "AUNTIE"} {
		go func(name string) {
			results <- namedRequest(t, a, "POST", personAction(t, family, "/claims/toggle"), form, nameCookie(name)).Code
		}(name)
	}
	first, second := <-results, <-results
	if !((first == 303 && second == 409) || (first == 409 && second == 303)) {
		t.Fatalf("concurrent results %d/%d", first, second)
	}
	writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes: [{title: Wooden train, claims: [{name: Zoë}]}]\n")
	page := namedRequest(t, a, "GET", wishLink(t, family, "Anna", "2026"), nil, nameCookie("Zoe"))
	if strings.Contains(page.Body.String(), "Claimed by you") {
		t.Fatal("ownership incorrectly stripped accents")
	}
	got := namedRequest(t, a, "POST", personAction(t, family, "/claims/toggle"), claimForm(t, a, family), nameCookie("Zoe"))
	if got.Code != 303 {
		t.Fatal(got.Code)
	}
}

// @test CLM-002-AC5 CLM-004-AC4 CON-004
func TestClaimWritesPreserveOtherAliasesAndFailClosedOnCycles(t *testing.T) {
	a, dir, _, family := claimFixture(t)
	path := listFile(dir, "2026", "Anna")
	writeTestYAML(t, path, "wishes:\n  - title: Wooden train\n    note: &note Same note\n    claims: &shared\n      - name: Auntie\n        future: retained\n  - title: Another train\n    note: *note\n    claims: *shared\n")
	form := claimForm(t, a, family)
	form.Set("mode", "remove")
	got := namedRequest(t, a, "POST", personAction(t, family, "/claims/toggle"), form, nameCookie("Auntie"))
	if got.Code != 303 {
		t.Fatal(got.Code)
	}
	var list wishlist
	if err := readYAML(path, &list); err != nil {
		t.Fatal(err)
	}
	cards, err := wishCards(list)
	if err != nil || len(cards) != 2 || len(cards[0].Claims) != 0 || len(cards[1].Claims) != 1 || cards[1].Claims[0].Name != "Auntie" || cards[1].Note != "Same note" {
		t.Fatal("alias mutation changed another wish")
	}
	raw, _ := os.ReadFile(path)
	if !strings.Contains(string(raw), "future: retained") {
		t.Fatal("claim future field lost")
	}
	writeTestYAML(t, path, "wishes:\n  - title: Wooden train\n    claims: [{name: Auntie}]\n    future: &cycle [*cycle]\n")
	form = claimForm(t, a, family)
	before := snapshotFiles(t, dir)
	got = namedRequest(t, a, "POST", personAction(t, family, "/claims/toggle"), form, nameCookie("Uncle"))
	if got.Code != 500 || !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("cyclic aliases did not fail without writes")
	}
}

// @test CLM-002-AC5 CLM-004-AC4 CON-004
func TestClaimWriteFailureRefreshesWithoutChangingClaims(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission failure requires an unprivileged user")
	}
	a, dir, _, family := claimFixture(t)
	path := listFile(dir, "2026", "Anna")
	writeTestYAML(t, path, "wishes: [{title: Wooden train, claims: [{name: Auntie}]}]\n")
	form := claimForm(t, a, family)
	folder := filepath.Dir(path)
	if err := os.Chmod(folder, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(folder, 0700) })
	before := snapshotFiles(t, dir)
	got := namedRequest(t, a, "POST", personAction(t, family, "/claims/toggle"), form, nameCookie("Uncle"))
	if got.Code != 500 || !strings.Contains(got.Body.String(), "role=\"alert\"") || !strings.Contains(got.Body.String(), "Claimed · 1 claimant") || strings.Contains(got.Body.String(), "Auntie") {
		t.Fatalf("write failure refresh: %d", got.Code)
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("write failure changed claims")
	}
}

// @test CLM-002-AC4 CLM-004-AC6
func TestParentsCannotMutateHistoricalClaims(t *testing.T) {
	a, dir, parent, _ := claimFixture(t)
	writeTestYAML(t, listFile(dir, "2025", "Anna"), "wishes: [{title: Historic train, claims: [{name: Auntie}]}]\n")
	var historical wishlist
	if err := readYAML(listFile(dir, "2025", "Anna"), &historical); err != nil {
		t.Fatal(err)
	}
	cards, err := wishCards(historical)
	if err != nil {
		t.Fatal(err)
	}
	token := cards[0].Token
	before := snapshotFiles(t, dir)
	for _, mode := range []string{"add", "remove", "remove-other"} {
		got := namedRequest(t, a, "POST", personAction(t, parent, "/claims/toggle"), url.Values{"person": {"Anna"}, "year": {"2025"}, "wish": {token}, "mode": {mode}, "claimant": {"Auntie"}}, nameCookie("Auntie"))
		if got.Code != 404 || !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
			t.Fatalf("historical parent %s: %d", mode, got.Code)
		}
	}
}
