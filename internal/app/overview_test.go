package app

import (
	"go.yaml.in/yaml/v3"
	"strconv"
	"testing"
)

// @test WLS-011-AC1 WLS-011-AC2 WLS-011-AC3 WLS-011-AC6
func TestOverviewTotalsAndCompleteWishCounts(t *testing.T) {
	cases := []struct{ name, wishes, total, claimed, unclaimed string }{
		{"empty", "wishes: []\n", "0.00", "0", "0"},
		{"mixed", "wishes:\n - title: Claimed train\n   price: '10.10'\n   links: [https://shop.example/train]\n   claims: [{name: Auntie}, {name: Uncle}]\n - title: Zero price train\n   price: '0'\n   links: [https://shop.example/zero]\n - title: Unclaimed train\n   price: '0.20'\n   links: [https://shop.example/unclaimed]\n - title: Incomplete priced train\n   price: '5.05'\n   claims: [{name: Hiddenclaimant}]\n - title: Unpriced linked train\n   links: [https://shop.example/unpriced]\n   claims: [{name: Anotherhidden}]\n - title: Title only train\n", "15.35", "1", "2"},
		{"exact decimals", "wishes:\n - title: First priced wish\n   price: '0.10'\n - title: Second priced wish\n   price: '0.20'\n", "0.30", "0", "0"},
		{"large exact amount", "wishes:\n - title: Large priced wish\n   price: '9007199254740992.01'\n - title: Small priced wish\n   price: '0.02'\n", "9007199254740992.03", "0", "0"},
		{"carry and leading zeros", "wishes:\n - title: First priced wish\n   price: '000099.99'\n - title: Second priced wish\n   price: '0.01'\n", "100.00", "0", "0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var list struct {
				Wishes []wishCard `yaml:"wishes"`
			}
			if err := yaml.Unmarshal([]byte(tc.wishes), &list); err != nil {
				t.Fatal(err)
			}
			total, claimed, unclaimed := summarizeWishes(list.Wishes)
			if total != tc.total || strconv.Itoa(claimed) != tc.claimed || strconv.Itoa(unclaimed) != tc.unclaimed {
				t.Fatalf("got %s/%d/%d", total, claimed, unclaimed)
			}
		})
	}
}
