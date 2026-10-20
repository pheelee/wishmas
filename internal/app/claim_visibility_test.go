package app

import (
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
)

// @test CLM-003-AC1 CLM-003-AC2 CLM-003-AC5 CLM-003-AC6
func TestClaimVisibilityAndPageLoadRefresh(t *testing.T) {
	a, dir, parent, family := newPeopleApp(t)
	addTestPerson(t, a, parent, "Anna")
	path := listFile(dir, "2026", "Anna")
	writeTestYAML(t, path, "wishes:\n  - title: Wooden train\n    claims: [{name: Auntie}, {name: Uncle}]\n  - title: Another train\n")
	before := snapshotFiles(t, dir)
	for _, role := range []struct {
		link  string
		names bool
	}{{parent, true}, {family, false}} {
		page := request(t, a, "GET", wishLink(t, role.link, "Anna", "2026"), nil)
		if page.Code != 200 {
			t.Fatal(page.Code)
		}
		body := page.Body.String()
		for _, want := range []string{"Claimed · 2 claimants", "Unclaimed"} {
			if !strings.Contains(body, want) {
				t.Fatalf("missing %s", want)
			}
		}
		for _, name := range []string{"Auntie", "Uncle"} {
			if strings.Contains(body, name) != role.names {
				t.Fatalf("incorrect visibility for %s", name)
			}
		}
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("viewing changed files")
	}
	writeTestYAML(t, path, "wishes:\n  - title: Wooden train\n    claims: [{name: Cousin}]\n  - title: Another train\n    claims: []\n")
	for _, link := range []string{parent, family} {
		page := request(t, a, "GET", wishLink(t, link, "Anna", "2026"), nil)
		if !strings.Contains(page.Body.String(), "Claimed · 1 claimant") || strings.Contains(page.Body.String(), "Auntie") {
			t.Fatal("page load did not refresh claims")
		}
	}
	// Historical claims remain visible to parents, without claim controls.
	writeTestYAML(t, listFile(dir, "2025", "Anna"), "wishes: [{title: Historic train, claims: [{name: Grandma}]}]\n")
	page := request(t, a, "GET", wishLink(t, parent, "Anna", "2025"), nil)
	if page.Code != 200 || !strings.Contains(page.Body.String(), "Grandma") || !strings.Contains(page.Body.String(), "Claimed · 1 claimant") {
		t.Fatal("historical claim missing")
	}
}

// @test CLM-003-AC5 CON-004
func TestInvalidManualClaimsLeaveFilesUntouched(t *testing.T) {
	for _, claims := range []string{"broken", "[broken]", "[{name: ''}]", "[{name: A!unt}]", "[{name: Al}]", "[{name: Auntie}, {name: ' auntie '}]", "[{name: Auntie}, {name: AUNTIE}]", "[{name: '" + strings.Repeat("𐐀", 51) + "'}]", "[{name: 𐐀𐐀𐐀}, {name: 𐐨𐐨𐐨}]"} {
		t.Run(claims, func(t *testing.T) {
			a, dir, parent, family := newPeopleApp(t)
			addTestPerson(t, a, parent, "Anna")
			writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes: [{title: Wooden train, claims: "+claims+"}]\n")
			before := snapshotFiles(t, dir)
			for _, link := range []string{parent, family} {
				got := request(t, a, "GET", wishLink(t, link, "Anna", "2026"), nil)
				if got.Code != 500 {
					t.Fatalf("invalid claims returned %d", got.Code)
				}
			}
			if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
				t.Fatal("invalid claims rewritten")
			}
		})
	}
}

// @test CLM-003-AC1 CLM-003-AC5 WSH-002-AC3 CON-004
func TestUnicodeClaimsAndDetailsEditPreserveClaims(t *testing.T) {
	a, dir, parent, family := newPeopleApp(t)
	addTestPerson(t, a, parent, "Anna")
	longName := strings.Repeat("𐐀", 50)
	path := listFile(dir, "2026", "Anna")
	writeTestYAML(t, path, "wishes:\n  - title: Wooden train\n    claims:\n      - name: '  Zoë  '\n        future: retained\n      - name: '"+longName+"'\n")
	page := request(t, a, "GET", wishLink(t, parent, "Anna", "2026"), nil)
	if page.Code != 200 || !strings.Contains(page.Body.String(), "Zoë") || !strings.Contains(page.Body.String(), longName) {
		t.Fatal("valid Unicode claims rejected")
	}
	token := field(t, page.Body.String(), "wish")
	got := request(t, a, "POST", personAction(t, parent, "/wishes/save"), url.Values{"person": {"Anna"}, "title": {"Updated train"}, "wish": {token}})
	if got.Code != 303 {
		t.Fatal(got.Code)
	}
	raw, _ := os.ReadFile(path)
	var saved wishlist
	if err := readYAML(path, &saved); err != nil {
		t.Fatal(err)
	}
	cards, err := wishCards(saved)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "future: retained") || len(cards) != 1 || len(cards[0].Claims) != 2 || cards[0].Claims[1].Name != longName {
		t.Fatal("claim details lost")
	}
	page = request(t, a, "GET", wishLink(t, family, "Anna", "2026"), nil)
	if strings.Contains(page.Body.String(), "Zoë") || strings.Contains(page.Body.String(), longName) {
		t.Fatal("claim names leaked to family source")
	}
}
