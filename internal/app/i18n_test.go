package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func localeRequest(a http.Handler, method, path, preference string, form url.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept-Language", preference)
	got := httptest.NewRecorder()
	a.ServeHTTP(got, req)
	return got
}

// @test INT-001-AC1 INT-001-AC2 INT-002-AC1 INT-002-AC2 INT-002-AC3 INT-002-AC4 INT-002-AC5 INT-002-AC6 INT-003-AC1
func TestLanguageSetup(t *testing.T) {
	for _, tc := range []struct{ header, lang, heading string }{{"fr-FR,de-DE;q=0.9,en-US;q=0.8", "de-CH", "Mach Platz für Wünsche."}, {"de;q=0.2,en-GB;q=0.9", "en-US", "Make room for wishes."}, {"fr-FR", "en-US", "Make room for wishes."}} {
		a := newLegacy(t.TempDir())
		got := localeRequest(a, "GET", "/", tc.header, nil)
		if !strings.Contains(got.Body.String(), `lang="`+tc.lang+`"`) || !strings.Contains(got.Body.String(), tc.heading) {
			t.Fatalf("locale setup missing %s", tc.lang)
		}
	}
	dir := t.TempDir()
	a := newLegacy(dir)
	form := url.Values{"household": {"Family"}, "timezone": {"Europe/Zurich"}, "language": {"de-CH"}, "setup_language": {"en-US"}}
	preview := localeRequest(a, "POST", "/setup/prepare", "en-US", form)
	if !strings.Contains(preview.Body.String(), "Nearly ready") {
		t.Fatal("setup switched before finishing")
	}
	parent := field(t, preview.Body.String(), "parent_link")
	finish := localeRequest(a, "POST", "/setup/finish", "en-US", url.Values{"ticket": {field(t, preview.Body.String(), "ticket")}})
	if finish.Code != 303 {
		t.Fatalf("finish %d", finish.Code)
	}
	home := localeRequest(a, "GET", parent, "en-US", nil)
	if !strings.Contains(home.Body.String(), "Weihnachten bei Family.") {
		t.Fatal("saved language not applied")
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "households", "Family", "household.yaml"))
	if !strings.Contains(string(raw), "language: de-CH") {
		t.Fatal("language not persisted")
	}
	for _, bad := range []string{"<missing>", "", "de", "de-DE", "en", "fr-FR"} {
		emptyDir := t.TempDir()
		fresh := newLegacy(emptyDir)
		form.Set("language", bad)
		if bad == "<missing>" {
			form.Del("language")
		}
		got := localeRequest(fresh, "POST", "/setup/prepare", "en-US", form)
		if got.Code != 400 || !strings.Contains(got.Body.String(), "Choose English (US) or Deutsch (Schweiz).") {
			t.Fatalf("invalid language %q: %d", bad, got.Code)
		}
		if _, err := os.Stat(filepath.Join(emptyDir, "households")); !os.IsNotExist(err) {
			t.Fatal("invalid setup saved data")
		}
	}
}

// @test INT-003-AC2 INT-003-AC3 INT-003-AC4 INT-003-AC5 INT-004-AC1 INT-004-AC2 INT-004-AC3
func TestLanguageYAMLAndErrors(t *testing.T) {
	dir := t.TempDir()
	a := newLegacy(dir)
	parent, _ := setupHousehold(t, a, "Europe/Zurich")
	file := filepath.Join(dir, "households", "Family", "household.yaml")
	original, _ := os.ReadFile(file)
	// Legacy absence is English. No implicit migration writes.
	legacy := string(original)
	for _, line := range strings.Split(legacy, "\n") {
		if strings.HasPrefix(line, "language:") {
			legacy = strings.ReplaceAll(legacy, line+"\n", "")
		}
	}
	for _, value := range []string{"\"\"", "null", "", "de-DE", "en", "[de-CH]", "{}", "42", "true"} {
		raw := legacy + "language: " + value + "\n"
		os.WriteFile(file, []byte(raw), 0600)
		got := localeRequest(a, "GET", parent, "de-CH", nil)
		if got.Code != 500 || !strings.Contains(got.Body.String(), "Unable to read household data.") {
			t.Fatalf("bad YAML language %q accepted", value)
		}
		after, _ := os.ReadFile(file)
		if string(after) != raw {
			t.Fatal("invalid YAML rewritten")
		}
	}
	for _, raw := range []string{legacy + "<<: {language: null}\n", legacy + "<<: {language: \"\"}\n"} {
		os.WriteFile(file, []byte(raw), 0600)
		if got := localeRequest(a, "GET", parent, "de-CH", nil); got.Code != 500 {
			t.Fatal("merged empty language received legacy fallback")
		}
		after, _ := os.ReadFile(file)
		if string(after) != raw {
			t.Fatal("merged invalid language rewritten")
		}
	}
	os.WriteFile(file, []byte(legacy), 0600)
	if got := localeRequest(a, "GET", parent, "de-CH", nil); !strings.Contains(got.Body.String(), `lang="en-US"`) {
		t.Fatal("legacy fallback not English")
	}
	os.WriteFile(file, []byte(legacy+"language: de-CH\n"), 0600)
	if got := localeRequest(a, "GET", parent, "en-US", nil); !strings.Contains(got.Body.String(), "Weihnachten bei Family.") {
		t.Fatal("manual reload missing")
	}
	for _, path := range []string{"/", "/?code=invalid"} {
		got := localeRequest(a, "GET", path, "de-CH", nil)
		if !strings.Contains(got.Body.String(), `lang="en-US"`) {
			t.Fatal("invalid access leaked household language")
		}
	}
	got := localeRequest(a, "GET", "/unknown"+parent, "en-US", nil)
	if !strings.Contains(got.Body.String(), "Seite nicht gefunden") {
		t.Fatal("authenticated not-found not localized")
	}
}

// @test INT-001-AC1 INT-001-AC2 INT-001-AC4 INT-001-AC5 INT-004-AC4 INT-004-AC5
func TestTranslationInventoryAndFallback(t *testing.T) {
	// Remaining legacy domains retain their baseline wording while the new
	// default server uses approved HR text. Validate both exact inventories.
	approved, err := os.ReadFile("testdata/legacy-translations.md")
	if err != nil {
		t.Fatal(err)
	}
	target, err := os.ReadFile("../../sdd/translations.md")
	if err != nil {
		t.Fatal(err)
	}
	// The target table owns all retained/current wording, including approved
	// updates to shared image messages. Keep baseline-only retired entries until
	// milestone 6 removes the legacy harness; do not rewrite its historical table.
	row := regexp.MustCompile(`(?m)^\| ((?:[GSAPWCEI]|HR)[0-9]+) \| (.*?) \| (.*?) \|$`)
	inventory := map[string][]string{}
	for _, source := range [][]byte{approved, target} {
		seen := map[string]bool{}
		for _, r := range row.FindAllStringSubmatch(string(source), -1) {
			if seen[r[1]] {
				t.Errorf("duplicate approved row %s", r[1])
			}
			seen[r[1]] = true
			inventory[r[1]] = r
		}
	}
	if len(inventory) != len(messages) {
		t.Fatalf("approved rows %d; catalog %d", len(inventory), len(messages))
	}
	keysInOrder := make([]string, 0, len(inventory))
	for key := range inventory {
		keysInOrder = append(keysInOrder, key)
	}
	sort.Strings(keysInOrder)
	rows := make([][]string, 0, len(inventory))
	for _, key := range keysInOrder {
		rows = append(rows, inventory[key])
	}
	for _, r := range rows {
		en := regexp.MustCompile(`\s*\(\*\*new.*?\*\*\)`).ReplaceAllString(r[2], "")
		m, ok := messages[r[1]]
		if !ok || m.EN != en || m.DE != r[3] || m.EN == "" || m.DE == "" {
			t.Errorf("catalog differs from approved row %s", r[1])
		}
		if strings.Contains(m.DE, "ß") {
			t.Errorf("Swiss spelling %s", r[1])
		}
		english := placeholders.FindAllString(m.EN, -1)
		german := placeholders.FindAllString(m.DE, -1)
		sort.Strings(english)
		sort.Strings(german)
		if !reflect.DeepEqual(english, german) {
			t.Errorf("placeholder mismatch %s", r[1])
		}
	}
	original := messages["W23"]
	messages["W23"] = message{EN: original.EN}
	if got := translate("de-CH", "W23"); got != "Add wish" {
		t.Fatal("missing German translation did not fall back")
	}
	messages["W23"] = original
	if got := translate("de-CH", "W02", "{wish} <b>Anna</b>"); got != "Wunschliste von {wish} <b>Anna</b>" {
		t.Fatal("user content changed")
	}
	// Catalog completeness also covers newly added template labels and UI errors,
	// rather than merely comparing existing dictionary key counts.
	actions := regexp.MustCompile(`(?s){{.*?}}`)
	literal := regexp.MustCompile(`>([^<>]+)<`)
	keys := regexp.MustCompile(`{{t "((?:[A-Z]|HR)[0-9]+)"`)
	files, _ := filepath.Glob("web/*.html")
	for _, file := range files {
		raw, _ := os.ReadFile(file)
		for _, key := range keys.FindAllStringSubmatch(string(raw), -1) {
			if m, ok := messages[key[1]]; !ok || m.EN == "" || m.DE == "" {
				t.Errorf("missing template key %s in %s", key[1], file)
			}
		}
		for _, part := range literal.FindAllStringSubmatch(actions.ReplaceAllString(string(raw), ""), -1) {
			text := strings.TrimSpace(part[1])
			if regexp.MustCompile(`[A-Za-z]`).MatchString(text) && text != "wishmas" && text != "Deutsch (Schweiz)" && text != "English (US)" {
				t.Errorf("untranslated template text %q in %s", text, file)
			}
		}
	}
	// Labels/tooltips and client-side feedback must not bypass the catalog.
	for _, file := range files {
		raw, _ := os.ReadFile(file)
		stripped := actions.ReplaceAllString(string(raw), "")
		for _, attribute := range regexp.MustCompile(`(?:aria-label|title)="([^"]+)"`).FindAllStringSubmatch(stripped, -1) {
			if regexp.MustCompile(`[A-Za-z]`).MatchString(attribute[1]) {
				t.Errorf("untranslated accessible text %q in %s", attribute[1], file)
			}
		}
	}
	script, _ := os.ReadFile("web/app.js")
	if regexp.MustCompile(`\.textContent\s*=\s*['"][A-Za-z]`).Match(script) {
		t.Error("client text bypasses translated message data")
	}
	files, _ = filepath.Glob("*.go")
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		raw, _ := os.ReadFile(file)
		for _, entry := range regexp.MustCompile(`"([^"\n]+)"`).FindAllStringSubmatch(string(raw), -1) {
			text := entry[1]
			if (regexp.MustCompile(`^[A-Z][a-z]+ `).MatchString(text) || text == "404 page not found") && strings.Contains(text, " ") {
				found := false
				for _, m := range messages {
					if m.EN == text {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("UI message absent from catalogs: %q in %s", text, file)
				}
			}
		}
	}
}

// @test INT-005-AC1 INT-005-AC2 INT-005-AC3 INT-005-AC4
func TestLocalizedPricePrecision(t *testing.T) {
	for _, tc := range []struct{ raw, de, en string }{{"1234.5", "1’234.50", "1,234.50"}, {"12", "12.00", "12.00"}, {"0.01", "0.01", "0.01"}, {"00001234.00", "1’234.00", "1,234.00"}, {"9007199254740992.01", "9’007’199’254’740’992.01", "9,007,199,254,740,992.01"}, {"", "", ""}} {
		if formatPrice("de-CH", tc.raw) != tc.de || formatPrice("en-US", tc.raw) != tc.en {
			t.Fatalf("price precision/format changed for %q", tc.raw)
		}
	}
}
