package app

import (
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
)

// @test WLS-006-AC1 WLS-006-AC2 WLS-006-AC3 WLS-006-AC6
func TestOverviewTotalsAndCompleteWishCounts(t *testing.T) {
	a, dir, parent, family := newPeopleApp(t)
	addTestPerson(t, a, parent, "Anna")
	cases := []struct{ name, wishes, total, claimed, unclaimed string }{
		{"empty", "wishes: []\n", "0.00", "0", "0"},
		{"mixed", "wishes:\n - title: Claimed train\n   price: '10.10'\n   links: [https://shop.example/train]\n   claims: [{name: Auntie}, {name: Uncle}]\n - title: Zero price train\n   price: '0'\n   links: [https://shop.example/zero]\n - title: Unclaimed train\n   price: '0.20'\n   links: [https://shop.example/unclaimed]\n - title: Incomplete priced train\n   price: '5.05'\n   claims: [{name: Hiddenclaimant}]\n - title: Unpriced linked train\n   links: [https://shop.example/unpriced]\n   claims: [{name: Anotherhidden}]\n - title: Title only train\n", "15.35", "1", "2"},
		{"exact decimals", "wishes:\n - title: First priced wish\n   price: '0.10'\n - title: Second priced wish\n   price: '0.20'\n", "0.30", "0", "0"},
		{"large exact amount", "wishes:\n - title: Large priced wish\n   price: '9007199254740992.01'\n - title: Small priced wish\n   price: '0.02'\n", "9007199254740992.03", "0", "0"},
		{"carry and leading zeros", "wishes:\n - title: First priced wish\n   price: '000099.99'\n - title: Second priced wish\n   price: '0.01'\n", "100.00", "0", "0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := listFile(dir, "2026", "Anna")
			writeTestYAML(t, path, tc.wishes)
			before := snapshotFiles(t, dir)
			for _, link := range []string{parent, family} {
				page := request(t, a, "GET", link, nil)
				body := page.Body.String()
				if page.Code != 200 || !strings.Contains(body, `class="person-summary-total"`) {
					t.Fatalf("overview summary missing: status %d", page.Code)
				}
				for _, text := range []string{">" + tc.total + "</strong>", ">" + tc.claimed + " claimed</span>", ">" + tc.unclaimed + " unclaimed</span>"} {
					if !strings.Contains(body, text) {
						t.Fatalf("expected total/count missing for %s", tc.name)
					}
				}
				if strings.Contains(body, "Hiddenclaimant") || strings.Contains(body, "Anotherhidden") || strings.Contains(body, "Uncle") {
					t.Fatal("overview leaked claimant names")
				}
			}
			if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
				t.Fatal("viewing overview changed stored data")
			}
			raw, _ := os.ReadFile(path)
			if string(raw) != tc.wishes {
				t.Fatal("summary rewrote wishlist")
			}
		})
	}
}

// @test WLS-006-AC1 WLS-006-AC2 WLS-006-AC3 WLS-006-AC6
func TestOverviewScopeReloadAndLocalization(t *testing.T) {
	a, dir, parent, family := newPeopleApp(t)
	for _, name := range []string{"Anna", "Ben", "Hiddenperson"} {
		addTestPerson(t, a, parent, name)
	}
	writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes:\n - title: Current train\n   price: '1024.50'\n   links: [https://shop.example/train]\n   claims: [{name: Privateclaimant}, {name: Anotherclaimant}]\n")
	writeTestYAML(t, listFile(dir, "2026", "Ben"), "wishes:\n - title: Incomplete train\n   price: '20'\n")
	writeTestYAML(t, listFile(dir, "2025", "Anna"), "wishes:\n - title: Historical train\n   price: '99.99'\n   links: [https://shop.example/history]\n")
	// Preserve normal archive behavior; summaries must never expose retained data.
	response := request(t, a, "POST", personAction(t, parent, "/people/archive"), url.Values{"person": {"Hiddenperson"}, "year": {"2026"}})
	if response.Code != 303 {
		t.Fatalf("archive: %d", response.Code)
	}
	writeTestYAML(t, listFile(dir, "2026", "Hiddenperson"), "invalid: [\n")
	for _, locale := range []string{"en-US", "de-CH"} {
		h, err := a.load()
		if err != nil {
			t.Fatal(err)
		}
		h.Language = locale
		if err := a.save(*h); err != nil {
			t.Fatal(err)
		}
		for _, link := range []string{parent, family} {
			before := snapshotFiles(t, dir)
			page := request(t, a, "GET", link, nil)
			body := page.Body.String()
			if page.Code != 200 {
				t.Fatalf("overview: %d", page.Code)
			}
			for _, want := range []string{">1024.50</strong>", ">20.00</strong>"} {
				if !strings.Contains(body, want) {
					t.Fatal("totals not scoped to each person")
				}
			}
			claimed := "1 claimed"
			unclaimed := "0 unclaimed"
			if locale == "de-CH" {
				claimed = "1 reserviert"
				unclaimed = "0 nicht reserviert"
			}
			if !strings.Contains(body, ">"+claimed+"</span>") || !strings.Contains(body, ">"+unclaimed+"</span>") {
				t.Fatal("localized summary missing")
			}
			for _, hidden := range []string{"Privateclaimant", "Anotherclaimant", "Hiddenperson", "99.99"} {
				if strings.Contains(body, hidden) {
					t.Fatal("overview leaked names, archived data or another year")
				}
			}
			if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
				t.Fatal("overview changed files")
			}
		}
		history := request(t, a, "GET", parent+"&year=2025", nil)
		if history.Code != 200 || !strings.Contains(history.Body.String(), ">99.99</strong>") || strings.Contains(history.Body.String(), ">1024.50</strong>") {
			t.Fatal("historical totals incorrect")
		}
	}
	before := snapshotFiles(t, dir)
	if got := request(t, a, "GET", family+"&year=2025", nil); got.Code != 404 {
		t.Fatal("family historical summary allowed")
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("denied view changed files")
	}
	writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes:\n - title: Updated train\n   price: '0.10'\n")
	if got := request(t, a, "GET", family, nil); !strings.Contains(got.Body.String(), ">0.10</strong>") {
		t.Fatal("manual changes not reloaded")
	}
	writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes:\n - title: Invalid price train\n   price: '1,000'\n")
	before = snapshotFiles(t, dir)
	if got := request(t, a, "GET", parent, nil); got.Code != 500 {
		t.Fatal("invalid data received a summary")
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("invalid data overwritten")
	}
}
