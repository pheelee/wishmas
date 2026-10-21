package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func request(t *testing.T, handler http.Handler, method, path string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	return response
}

func field(t *testing.T, body, name string) string {
	t.Helper()
	match := regexp.MustCompile(`name="` + name + `" value="([^"]+)"`).FindStringSubmatch(body)
	if len(match) != 2 {
		t.Fatalf("missing field %s in %s", name, body)
	}
	return match[1]
}

// @test ACC-001-AC1 ACC-001-AC2 ACC-001-AC3 ACC-001-AC4 ACC-001-AC6 ACC-001-AC7 ACC-001-AC8 CON-004
func TestSetupPersistsAcrossRestart(t *testing.T) {
	dir := t.TempDir()
	handler := newLegacy(dir)
	first := request(t, handler, "GET", "/", nil)
	if first.Code != 200 || !strings.Contains(first.Body.String(), `name="household"`) || !strings.Contains(first.Body.String(), `name="timezone"`) {
		t.Fatalf("expected open household/timezone wizard: %d %s", first.Code, first.Body.String())
	}
	preview := request(t, handler, "POST", "/setup/prepare", url.Values{"language": {"en-US"}, "household": {"  Noël2026  "}, "timezone": {"Europe/Zurich"}})
	if preview.Code != 200 {
		t.Fatalf("prepare: %d %s", preview.Code, preview.Body.String())
	}
	body := preview.Body.String()
	parent := field(t, body, "parent_link")
	family := field(t, body, "family_link")
	if parent == family || !strings.Contains(parent, "code=") || !strings.Contains(family, "code=") {
		t.Fatal("expected distinct code links")
	}
	if strings.Count(body, "data-copy=") != 2 {
		t.Fatal("expected copy buttons for both links")
	}
	// Setup remains open until explicitly completed.
	if got := request(t, handler, "GET", "/", nil); !strings.Contains(got.Body.String(), `name="household"`) {
		t.Fatal("preview closed setup prematurely")
	}
	finish := request(t, handler, "POST", "/setup/finish", url.Values{"ticket": {field(t, body, "ticket")}})
	if finish.Code != http.StatusSeeOther || finish.Header().Get("Location") != parent {
		t.Fatalf("finish: %d %s", finish.Code, finish.Header().Get("Location"))
	}
	file := filepath.Join(dir, "households", "Noël2026", "household.yaml")
	saved, err := os.ReadFile(file)
	if err != nil || !strings.Contains(string(saved), "Europe/Zurich") {
		t.Fatalf("persisted YAML: %s, %v", saved, err)
	}
	restarted := newLegacy(dir)
	for _, link := range []string{parent, family} {
		got := request(t, restarted, "GET", link, nil)
		if got.Code != 200 || !strings.Contains(got.Body.String(), "Noël2026") {
			t.Fatalf("restart link: %d %s", got.Code, got.Body.String())
		}
	}
	for _, path := range []string{"/", "/?code=invalid", "/setup"} {
		if got := request(t, restarted, "GET", path, nil); got.Code != 404 {
			t.Fatalf("%s after setup: %d", path, got.Code)
		}
	}
	if got := request(t, restarted, "POST", "/setup/finish", url.Values{"ticket": {field(t, body, "ticket")}}); got.Code != 404 {
		t.Fatal("completed setup was reusable")
	}
}

// @test ACC-001-AC9
func TestHouseholdNameValidation(t *testing.T) {
	for _, name := range []string{"", "  ", "ab", strings.Repeat("a", 51), "Family Name", "../Family", "Family!"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			got := request(t, newLegacy(dir), "POST", "/setup/prepare", url.Values{"language": {"en-US"}, "household": {name}, "timezone": {"Europe/Zurich"}})
			if got.Code != 400 {
				t.Fatalf("invalid name accepted: %d", got.Code)
			}
			entries, _ := os.ReadDir(dir)
			if len(entries) != 0 {
				t.Fatal("invalid input wrote files")
			}
		})
	}
}

// @test ACC-001-AC2 CON-004
func TestInvalidTimezoneDoesNotCompleteSetup(t *testing.T) {
	dir := t.TempDir()
	got := request(t, newLegacy(dir), "POST", "/setup/prepare", url.Values{"language": {"en-US"}, "household": {"Family"}, "timezone": {"No/SuchTimezone"}})
	if got.Code != 400 {
		t.Fatalf("expected validation error: %d", got.Code)
	}
	if len(mustReadDir(t, dir)) != 0 {
		t.Fatal("invalid timezone wrote files")
	}
}

func mustReadDir(t *testing.T, dir string) []os.DirEntry {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	return entries
}

// @test CON-004 ACC-001-AC5 (timezone reload; calendar-boundary verification remains in milestone 2)
func TestManualYAMLReloadAndInvalidDataLeftUntouched(t *testing.T) {
	dir := t.TempDir()
	handler := newLegacy(dir)
	preview := request(t, handler, "POST", "/setup/prepare", url.Values{"language": {"en-US"}, "household": {"Family"}, "timezone": {"Europe/Zurich"}})
	parent := field(t, preview.Body.String(), "parent_link")
	request(t, handler, "POST", "/setup/finish", url.Values{"ticket": {field(t, preview.Body.String(), "ticket")}})
	file := filepath.Join(dir, "households", "Family", "household.yaml")
	contents, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	changed := strings.ReplaceAll(string(contents), "Europe/Zurich", "Pacific/Auckland")
	if err := os.WriteFile(file, []byte(changed), 0600); err != nil {
		t.Fatal(err)
	}
	got := request(t, handler, "GET", parent, nil)
	if got.Code != 200 || !strings.Contains(got.Body.String(), "Pacific/Auckland") {
		t.Fatal("manual timezone change was not loaded")
	}
	invalid := []byte("timezone: [broken")
	if err := os.WriteFile(file, invalid, 0600); err != nil {
		t.Fatal(err)
	}
	got = request(t, handler, "GET", parent, nil)
	if got.Code != 500 {
		t.Fatalf("invalid YAML status: %d", got.Code)
	}
	after, _ := os.ReadFile(file)
	if string(after) != string(invalid) {
		t.Fatal("invalid YAML was modified")
	}
}

// @test ACC-001-AC6 (browser copy interaction requires browser verification)
func TestLocalAssetsAvailable(t *testing.T) {
	handler := newLegacy(t.TempDir())
	for _, path := range []string{"/assets/app.css", "/assets/app.js"} {
		got := request(t, handler, "GET", path, nil)
		contents, _ := io.ReadAll(got.Result().Body)
		if got.Code != 200 || len(contents) == 0 {
			t.Fatalf("missing asset %s", path)
		}
	}
}

// @test ACC-001-AC9
func TestHouseholdNameLengthBoundaries(t *testing.T) {
	for _, name := range []string{"Éø1", strings.Repeat("É", 50)} {
		t.Run(name, func(t *testing.T) {
			got := request(t, newLegacy(t.TempDir()), "POST", "/setup/prepare", url.Values{"language": {"en-US"}, "household": {name}, "timezone": {"Europe/Zurich"}})
			if got.Code != 200 {
				t.Fatalf("valid boundary name rejected: %d %s", got.Code, got.Body.String())
			}
		})
	}
}

// @test CON-003-AC1 CON-003-AC5 CON-004
func TestInvalidYAMLUsesTheSharedErrorTheme(t *testing.T) {
	dir := t.TempDir()
	handler := newLegacy(dir)
	parent, _ := setupHousehold(t, handler, "Europe/Zurich")
	file := filepath.Join(dir, "households", "Family", "household.yaml")
	invalid := []byte("timezone: [broken")
	if err := os.WriteFile(file, invalid, 0600); err != nil {
		t.Fatal(err)
	}
	got := request(t, handler, "GET", parent, nil)
	if got.Code != 500 || !strings.Contains(got.Body.String(), `href="/assets/app.css"`) {
		t.Fatalf("expected shared themed error page: %d %s", got.Code, got.Body.String())
	}
	after, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(invalid) {
		t.Fatal("error page modified invalid YAML")
	}
}
