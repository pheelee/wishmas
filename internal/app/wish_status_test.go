package app

import (
	"go.yaml.in/yaml/v3"
	"testing"
	data "wishmas/internal/household"
)

// @test WSH-005-AC6 WSH-005-AC7
func TestWishStatusColors(t *testing.T) {
	for _, tc := range []struct{ name, details, accent string }{
		{"title only", "", "red"},
		{"missing price", "    links: [https://shop.example/train]\n", "red"},
		{"missing links", "    price: '12.50'\n", "red"},
		{"claimed incomplete", "    price: '12.50'\n    claims: [{user_id: aunt}]\n", "red"},
		{"optional details not completeness", "    note: Some details\n    picture: unavailable.webp\n", "red"},
		{"complete unclaimed", "    price: '12.50'\n    links: [https://shop.example/train]\n", "yellow"},
		{"provided zero price", "    price: '0'\n    links: [https://shop.example/train]\n", "yellow"},
		{"complete claimed", "    price: '12.50'\n    links: [https://shop.example/train]\n    claims: [{user_id: aunt}]\n", "green"},
		{"multiple claims and shops", "    price: '12.50'\n    links: [https://shop.example/train, https://other.example/train]\n    claims: [{user_id: aunt}, {user_id: uncle}]\n", "green"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var wish data.Wish
			raw := "title: Wooden train\n" + tc.details
			var wrapped struct {
				Wishes []data.Wish `yaml:"wishes"`
			}
			if err := yaml.Unmarshal([]byte("wishes:\n  - "+raw), &wrapped); err != nil {
				t.Fatal(err)
			}
			wish = wrapped.Wishes[0]
			h := data.Household{Language: "en-US", Users: []data.Person{{ID: "aunt", Name: "Auntie"}, {ID: "uncle", Name: "Uncle"}}}
			card := householdCard(h, wish, data.NewStore(t.TempDir()), "", "", true)
			if card.Accent != tc.accent {
				t.Fatalf("got %s, want %s", card.Accent, tc.accent)
			}
		})
	}
}
