package app

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// @test INT-001-AC1 INT-001-AC2 INT-001-AC4 INT-001-AC5 INT-007-AC7 INT-007-AC8
func TestTranslationInventoryAndFallback(t *testing.T) {
	// @test INT-007-AC8: final delivery contains only approved current wording.
	target, err := os.ReadFile("../../sdd/translations.md")
	if err != nil {
		t.Fatal(err)
	}
	row := regexp.MustCompile(`(?m)^\| ((?:[GSAPWCEI]|HR)[0-9]+) \| (.*?) \| (.*?) \|$`)
	inventory := map[string][]string{}
	for _, r := range row.FindAllStringSubmatch(string(target), -1) {
		if _, exists := inventory[r[1]]; exists {
			t.Errorf("duplicate approved row %s", r[1])
		}
		inventory[r[1]] = r
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
	if got := translate("de-CH", "W13", "{wish} <b>Anna</b>"); got != "Wunsch bearbeiten: {wish} <b>Anna</b>" {
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
