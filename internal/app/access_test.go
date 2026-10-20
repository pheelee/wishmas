package app

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupHousehold(t *testing.T, handler http.Handler, timezone string) (string, string) {
	t.Helper()
	preview := request(t, handler, "POST", "/setup/prepare", url.Values{"language": {"en-US"}, "household": {"Family"}, "timezone": {timezone}})
	if preview.Code != 200 {
		t.Fatalf("setup preview: %d %s", preview.Code, preview.Body.String())
	}
	parent := field(t, preview.Body.String(), "parent_link")
	family := field(t, preview.Body.String(), "family_link")
	finish := request(t, handler, "POST", "/setup/finish", url.Values{"ticket": {field(t, preview.Body.String(), "ticket")}})
	if finish.Code != 303 {
		t.Fatalf("setup finish: %d", finish.Code)
	}
	return parent, family
}

func regeneratePath(t *testing.T, parentLink string) string {
	t.Helper()
	link, err := url.Parse(parentLink)
	if err != nil {
		t.Fatal(err)
	}
	return "/access/regenerate?" + link.RawQuery
}

// @test ACC-003-AC1 ACC-003-AC2 ACC-003-AC3 ACC-003-AC4
func TestFamilyCodeRegenerationIsIndependentAndPersistent(t *testing.T) {
	dir := t.TempDir()
	handler := New(dir)
	parent, oldFamily := setupHousehold(t, handler, "Europe/Zurich")
	got := request(t, handler, "POST", regeneratePath(t, parent), url.Values{"role": {"family"}})
	if got.Code != 303 || got.Header().Get("Location") != parent {
		t.Fatalf("family regeneration: %d %s", got.Code, got.Body.String())
	}
	landing := request(t, handler, "GET", parent, nil)
	family := field(t, landing.Body.String(), "family_link")
	if family == oldFamily {
		t.Fatal("family link did not change")
	}
	if field(t, landing.Body.String(), "parent_link") != parent {
		t.Fatal("parent link changed unexpectedly")
	}
	for _, h := range []http.Handler{handler, New(dir)} {
		if got := request(t, h, "GET", oldFamily, nil); got.Code != 404 {
			t.Fatalf("old family link remains valid: %d", got.Code)
		}
		if got := request(t, h, "GET", family, nil); got.Code != 200 || !strings.Contains(got.Body.String(), "Family access") {
			t.Fatal("new family access failed")
		}
		if got := request(t, h, "GET", parent, nil); got.Code != 200 {
			t.Fatal("parent access lost")
		}
	}
	// A request from a stale page must authenticate against the latest code.
	if got := request(t, handler, "POST", regeneratePath(t, oldFamily), url.Values{"role": {"parent"}}); got.Code != 404 {
		t.Fatal("stale family code retained action access")
	}
}

// @test ACC-003-AC1 ACC-003-AC2 ACC-003-AC3 ACC-003-AC4 ACC-003-AC5
func TestParentRegenerationRedirectsAndRevokesAlreadyOpenPages(t *testing.T) {
	dir := t.TempDir()
	handler := New(dir)
	oldParent, family := setupHousehold(t, handler, "Europe/Zurich")
	staleAction := regeneratePath(t, oldParent)
	got := request(t, handler, "POST", staleAction, url.Values{"role": {"parent"}})
	if got.Code != 303 {
		t.Fatalf("parent regeneration: %d %s", got.Code, got.Body.String())
	}
	parent := got.Header().Get("Location")
	if parent == oldParent || !strings.HasPrefix(parent, "/?code=") {
		t.Fatal("not redirected to a replacement parent link")
	}
	for _, h := range []http.Handler{handler, New(dir)} {
		landing := request(t, h, "GET", parent, nil)
		if landing.Code != 200 || !strings.Contains(landing.Body.String(), "Parent access") {
			t.Fatal("replacement parent access failed")
		}
		if field(t, landing.Body.String(), "family_link") != family {
			t.Fatal("family link changed unexpectedly")
		}
		if got := request(t, h, "GET", oldParent, nil); got.Code != 404 {
			t.Fatal("old parent code retained read access")
		}
	}
	if got := request(t, handler, "POST", staleAction, url.Values{"role": {"family"}}); got.Code != 404 {
		t.Fatal("already-open parent page retained action access")
	}
	if got := request(t, handler, "GET", family, nil); got.Code != 200 {
		t.Fatal("family code lost access")
	}
}

// @test ACC-003-AC6 ACC-003-AC1
func TestFamilyAndUnauthenticatedRequestsCannotRegenerateCodes(t *testing.T) {
	dir := t.TempDir()
	handler := New(dir)
	parent, family := setupHousehold(t, handler, "Europe/Zurich")
	file := filepath.Join(dir, "households", "Family", "household.yaml")
	original, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	for _, link := range []string{family, "/?code=invalid", "/"} {
		for _, role := range []string{"family", "parent"} {
			if got := request(t, handler, "POST", regeneratePath(t, link), url.Values{"role": {role}}); got.Code != 404 {
				t.Fatalf("unauthorized regeneration: %d", got.Code)
			}
		}
	}
	for _, role := range []string{"", "unknown"} {
		if got := request(t, handler, "POST", regeneratePath(t, parent), url.Values{"role": {role}}); got.Code != 400 {
			t.Fatalf("invalid role accepted: %d", got.Code)
		}
	}
	// Merely visiting an action URL must not regenerate a code.
	if got := request(t, handler, "GET", regeneratePath(t, parent), nil); got.Code != 404 {
		t.Fatal("GET action should not change codes")
	}
	after, _ := os.ReadFile(file)
	if string(after) != string(original) {
		t.Fatal("rejected action changed files")
	}
	if got := request(t, handler, "GET", family, nil); strings.Contains(got.Body.String(), "parent_link") || strings.Contains(got.Body.String(), "family_link") || strings.Contains(got.Body.String(), "Regenerate") {
		t.Fatal("family page exposed parent controls or access links")
	}
}

// @test ACC-001-AC5 ACC-002-AC4 ACC-002-AC6
func TestCalendarYearUsesHouseholdTimezoneAndCodesDoNotExpire(t *testing.T) {
	dir := t.TempDir()
	handler := New(dir).(*application)
	now := time.Date(2026, time.December, 31, 22, 59, 59, 0, time.UTC)
	handler.now = func() time.Time { return now }
	parent, family := setupHousehold(t, handler, "Europe/Zurich")
	file := filepath.Join(dir, "households", "Family", "household.yaml")
	original, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	assertYear := func(link string, year string) {
		t.Helper()
		got := request(t, handler, "GET", link, nil)
		if got.Code != 200 || !strings.Contains(got.Body.String(), "access · "+year) {
			t.Fatalf("expected household year %s: %d %s", year, got.Code, got.Body.String())
		}
	}
	assertYear(family, "2026")
	now = now.Add(time.Second)
	assertYear(family, "2027") // Still December 31 UTC; already January 1 in Zurich.
	assertYear(parent, "2027")
	changed := strings.ReplaceAll(string(original), "Europe/Zurich", "America/Los_Angeles")
	if err := os.WriteFile(file, []byte(changed), 0600); err != nil {
		t.Fatal(err)
	}
	assertYear(family, "2026") // Manual timezone edit applies on next load.
	now = time.Date(2036, time.January, 1, 8, 0, 0, 0, time.UTC)
	assertYear(family, "2036")
	assertYear(parent, "2036")
	after, _ := os.ReadFile(file)
	if string(after) != changed {
		t.Fatal("time passing modified access codes or configuration")
	}
}

// @test ACC-003-AC3 ACC-003-AC4
func TestConcurrentParentRotationsRevokeTheOtherStaleAction(t *testing.T) {
	handler := New(t.TempDir())
	parent, _ := setupHousehold(t, handler, "Europe/Zurich")
	type result struct {
		status   int
		location string
	}
	results := make(chan result, 2)
	start := make(chan struct{})
	for range 2 {
		go func() {
			<-start
			got := request(t, handler, "POST", regeneratePath(t, parent), url.Values{"role": {"parent"}})
			results <- result{got.Code, got.Header().Get("Location")}
		}()
	}
	close(start)
	successes, revoked := 0, 0
	var replacement string
	for range 2 {
		got := <-results
		switch got.status {
		case 303:
			successes++
			replacement = got.location
		case 404:
			revoked++
		default:
			t.Fatalf("unexpected rotation status: %d", got.status)
		}
	}
	if successes != 1 || revoked != 1 {
		t.Fatalf("expected one success and one revoked stale action; got %d successes, %d revoked", successes, revoked)
	}
	if got := request(t, handler, "GET", replacement, nil); got.Code != 200 {
		t.Fatal("successful replacement lost access")
	}
}
