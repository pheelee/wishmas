package app

import (
	"strings"
	"testing"
)

// @test WSH-005-AC6 WSH-005-AC7: observe status on real parent/family responses.
func TestWishStatusColors(t *testing.T) {
	a, dir, parent, family := newPeopleApp(t)
	addTestPerson(t, a, parent, "Anna")
	for _, tc := range []struct{ name, details, accent string }{
		{"title only", "", "red"},
		{"missing price", "    links: [https://shop.example/train]\n", "red"},
		{"missing links", "    price: '12.50'\n", "red"},
		{"claimed incomplete", "    price: '12.50'\n    claims: [{name: Auntie}]\n", "red"},
		{"optional details not completeness", "    note: Some details\n    picture: unavailable.webp\n", "red"},
		{"complete unclaimed", "    price: '12.50'\n    links: [https://shop.example/train]\n", "yellow"},
		{"provided zero price", "    price: '0'\n    links: [https://shop.example/train]\n", "yellow"},
		{"complete claimed", "    price: '12.50'\n    links: [https://shop.example/train]\n    claims: [{name: Auntie}]\n", "green"},
		{"multiple claims and shops", "    price: '12.50'\n    links: [https://shop.example/train, https://other.example/train]\n    claims: [{name: Auntie}, {name: Uncle}]\n", "green"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes:\n  - title: Wooden train\n"+tc.details)
			for _, link := range []string{parent, family} {
				got := request(t, a, "GET", wishLink(t, link, "Anna", "2026"), nil)
				if got.Code != 200 || !strings.Contains(got.Body.String(), "wish-accent-"+tc.accent) {
					t.Fatalf("expected %s card for %s; status %d", tc.accent, tc.name, got.Code)
				}
			}
		})
	}
}
