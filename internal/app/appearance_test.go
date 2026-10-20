package app

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
)

// @test CON-003-AC6 CON-003-AC7 CON-003-AC8
func TestAppearanceSetupPersistenceAndValidation(t *testing.T) {
	for _, tc := range []struct {
		name, bulbs, size, count string
		status                   int
	}{{"chosen", "180", "150", "32", 200}, {"minimum", "80", "50", "8", 200}, {"maximum", "200", "200", "48", 200}, {"bad bulbs", "201", "100", "16", 400}, {"bad size", "120", "49", "16", 400}, {"bad count", "120", "100", "0", 400}, {"noninteger", "120.5", "100", "16", 400}} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			a := New(dir)
			before := snapshotFiles(t, dir)
			got := request(t, a, "POST", "/setup/prepare", url.Values{"household": {"Family"}, "timezone": {"Europe/Zurich"}, "bulb_size": {tc.bulbs}, "snow_size": {tc.size}, "snow_amount": {tc.count}})
			if got.Code != tc.status {
				t.Fatalf("got %d want %d", got.Code, tc.status)
			}
			if tc.status == 400 {
				if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
					t.Fatal("invalid setup wrote data")
				}
				return
			}
			family := field(t, got.Body.String(), "family_link")
			ticket := field(t, got.Body.String(), "ticket")
			got = request(t, a, "POST", "/setup/finish", url.Values{"ticket": {ticket}})
			if got.Code != 303 {
				t.Fatal(got.Code)
			}
			restarted := New(dir)
			page := request(t, restarted, "GET", family, nil)
			for _, want := range []string{`data-bulb-size="` + tc.bulbs + `"`, `data-snow-size="` + tc.size + `"`, `data-snow-amount="` + tc.count + `"`} {
				if !strings.Contains(page.Body.String(), want) {
					t.Fatalf("missing %s after restart", want)
				}
			}
		})
	}
}

// @test CON-003-AC6 CON-003-AC8 CON-004 ACC-003-AC1
func TestAppearanceLegacyAndManualData(t *testing.T) {
	a, dir, parent, _ := newPeopleApp(t)
	h, err := a.load()
	if err != nil {
		t.Fatal(err)
	}
	h.Appearance = nil
	if err := a.save(*h); err != nil {
		t.Fatal(err)
	}
	got := request(t, New(dir), "GET", parent, nil)
	for _, want := range []string{`data-bulb-size="120"`, `data-snow-size="100"`, `data-snow-amount="16"`} {
		if !strings.Contains(got.Body.String(), want) {
			t.Fatalf("legacy defaults missing %s", want)
		}
	}
	h.Appearance = &appearance{180, 150, 32}
	if err := a.save(*h); err != nil {
		t.Fatal(err)
	}
	got = request(t, a, "POST", personAction(t, parent, "/access/regenerate"), url.Values{"role": {"family"}})
	if got.Code != 303 {
		t.Fatal(got.Code)
	}
	saved, err := a.load()
	if err != nil || saved.Appearance == nil || *saved.Appearance != *h.Appearance {
		t.Fatal("regeneration lost appearance")
	}
	for _, bad := range []appearance{{0, 100, 16}, {120, 201, 16}, {120, 100, 49}} {
		h.Appearance = &bad
		if err := a.save(*h); err != nil {
			t.Fatal(err)
		}
		before := snapshotFiles(t, dir)
		got = request(t, a, "GET", parent, nil)
		if got.Code != 500 {
			t.Fatalf("invalid manual settings returned %d", got.Code)
		}
		if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
			t.Fatal("invalid data was rewritten")
		}
	}
}
