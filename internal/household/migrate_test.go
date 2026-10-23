package household_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pheelee/wishmas/internal/household"
)

func legacyFixture(t *testing.T) string {
	t.Helper()
	source := filepath.Join(t.TempDir(), "legacy")
	mustWrite(t, filepath.Join(source, "household.yaml"), "name: Family\nadmin_password_hash: "+fixtureHash()+"\nappearance: {bulb_size: 150, snow_size: 80, snow_amount: 24}\n")
	mustWrite(t, filepath.Join(source, "people.yaml"), "people:\n - name: Leon\n - name: Anna\n   archived: true\n")
	mustWrite(t, filepath.Join(source, "2026/christmas/Leon/wishlist.yaml"), "wishes:\n - title: Model train\n   price: '00.10'\n   note: Keep this note\n   links: [https://shop.example/one, https://shop.example/two]\n   claims: [{name: Dad, future_claim_field: retained}]\n")
	return source
}

// @test CON-008-AC1 CON-008-AC2 CON-008-AC3 CON-008-AC4 CON-008-AC5
func TestMigrationCopiesImagesAndReportsMissingImages(t *testing.T) {
	source := legacyFixture(t)
	path := filepath.Join(source, "2026/christmas/Leon/wishlist.yaml")
	wish := string(mustRead(t, path)) + "   picture: same.webp\n   picture_source: https://images.example/a.png\n"
	mustWrite(t, path, wish)
	// Already unreadable/invalid image bytes must remain byte-identical, so future
	// browsing can retain its unavailable-image behavior rather than re-fetching.
	image := []byte("RIFF public fixture bytes, never a remote download")
	mustWrite(t, filepath.Join(filepath.Dir(path), "pictures/same.webp"), string(image))
	mustWrite(t, filepath.Join(source, "2026/christmas/Anna/wishlist.yaml"), "wishes:\n - title: Another present\n   picture: same.webp\n   picture_source: https://images.example/missing.png\n")
	before := treeHashes(t, source)
	dest := filepath.Join(t.TempDir(), "new")
	report, err := household.Migrate(source, dest)
	if err != nil {
		t.Fatal(err)
	}
	if report.Wishes != 2 || report.Claims != 1 || report.Users != 1 || report.Recipients != 2 || report.MissingPictures != 1 || len(report.Pictures) != 1 || !report.SourceUnchanged || !report.DestinationVerified {
		t.Fatal("incorrect comparison report")
	}
	if report.SourceCounts != report.DestinationCounts || report.SourceCounts.Wishes != 2 {
		t.Fatal("source/destination count comparison missing")
	}
	if report.Pictures[0].SourceSHA256 != report.Pictures[0].DestinationSHA256 {
		t.Fatal("checksum mismatch")
	}
	h, _, err := household.NewStore(filepath.Join(dest, "household.yaml")).Load()
	if err != nil {
		t.Fatal(err)
	}
	if h.Language != "en-US" || h.Appearance.BulbSize != 150 || h.Wishes[0].Price != "00.10" || h.Wishes[0].Note != "Keep this note" || len(h.Wishes[0].Links) != 2 || h.Wishes[0].Claims[0].Extra["future_claim_field"].Value != "retained" {
		t.Fatal("migration altered configuration/details")
	}
	if !bytes.Equal(image, mustRead(t, filepath.Join(dest, "pictures", h.Wishes[0].Picture))) {
		t.Fatal("image bytes altered")
	}
	if h.Wishes[0].Picture == h.Wishes[1].Picture {
		t.Fatal("duplicate original picture filenames collided")
	}
	if _, err := os.Stat(filepath.Join(dest, "pictures", h.Wishes[1].Picture)); !os.IsNotExist(err) {
		t.Fatal("missing image unexpectedly materialized")
	}
	after := treeHashes(t, source)
	if len(before) != len(after) {
		t.Fatal("source changed")
	}
	for path, digest := range before {
		if after[path] != digest {
			t.Fatal("source changed")
		}
	}
}

// @test CON-008-AC1 CON-008-AC5
func TestMigrationRefusesUnsafeOrUnrepresentableSources(t *testing.T) {
	cases := map[string]func(*testing.T, string, string){
		"existing destination": func(t *testing.T, s, d string) {
			if err := os.Mkdir(d, 0700); err != nil {
				t.Fatal(err)
			}
		},
		"nonempty destination": func(t *testing.T, s, d string) { mustWrite(t, filepath.Join(d, "sentinel"), "unchanged") },
		"missing hash":         func(t *testing.T, s, d string) { mustWrite(t, filepath.Join(s, "household.yaml"), "name: Family\n") },
		"invalid hash": func(t *testing.T, s, d string) {
			mustWrite(t, filepath.Join(s, "household.yaml"), "name: Family\nadmin_password_hash: invalid-secret-hash\n")
		},
		"invalid language": func(t *testing.T, s, d string) {
			mustWrite(t, filepath.Join(s, "household.yaml"), "name: Family\nlanguage: null\nadmin_password_hash: "+fixtureHash()+"\n")
		},
		"unknown file": func(t *testing.T, s, d string) {
			mustWrite(t, filepath.Join(s, "2026/birthday/Leon/wishlist.yaml"), "wishes: []\n")
		},
		"invalid wish": func(t *testing.T, s, d string) {
			mustWrite(t, filepath.Join(s, "2026/christmas/Leon/wishlist.yaml"), "wishes: [{title: bad}]\n")
		},
		"duplicate claim": func(t *testing.T, s, d string) {
			mustWrite(t, filepath.Join(s, "2026/christmas/Leon/wishlist.yaml"), "wishes: [{title: Model train, claims: [{name: Dad}, {name: dad}]}]\n")
		},
		"reserved future field": func(t *testing.T, s, d string) {
			mustWrite(t, filepath.Join(s, "2026/christmas/Leon/wishlist.yaml"), "wishes: [{title: Model train, id: manually-added}]\n")
		},
		"duplicate recipient": func(t *testing.T, s, d string) {
			mustWrite(t, filepath.Join(s, "people.yaml"), "people: [{name: Leon}, {name: leon}]\n")
		},
		"traversal picture": func(t *testing.T, s, d string) {
			mustWrite(t, filepath.Join(s, "2026/christmas/Leon/wishlist.yaml"), "wishes: [{title: Model train, picture: ../secret.webp}]\n")
		},
		"multiple documents": func(t *testing.T, s, d string) {
			mustWrite(t, filepath.Join(s, "2026/christmas/Leon/wishlist.yaml"), "wishes: []\n---\nwishes: []\n")
		},
		"symlink": func(t *testing.T, s, d string) {
			if err := os.Symlink("household.yaml", filepath.Join(s, "alias.yaml")); err != nil {
				t.Fatal(err)
			}
		},
	}
	for name, prepare := range cases {
		t.Run(name, func(t *testing.T) {
			source := legacyFixture(t)
			destination := filepath.Join(t.TempDir(), "new")
			prepare(t, source, destination)
			// Snapshot symlink metadata safely rather than reading through it.
			config := mustRead(t, filepath.Join(source, "household.yaml"))
			_, err := household.Migrate(source, destination)
			if err == nil {
				t.Fatal("unsafe migration succeeded")
			}
			if strings.Contains(err.Error(), fixtureHash()) || strings.Contains(err.Error(), "invalid-secret-hash") {
				t.Fatal("error exposed hash")
			}
			if !bytes.Equal(config, mustRead(t, filepath.Join(source, "household.yaml"))) {
				t.Fatal("failure changed source")
			}
			entries, _ := os.ReadDir(filepath.Dir(destination))
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".wishmas-migration-") {
					t.Fatal("staging directory leaked")
				}
			}
			switch name {
			case "existing destination":
				entries, _ := os.ReadDir(destination)
				if len(entries) != 0 {
					t.Fatal("existing destination changed")
				}
			case "nonempty destination":
				if string(mustRead(t, filepath.Join(destination, "sentinel"))) != "unchanged" {
					t.Fatal("destination overwritten")
				}
			default:
				if _, err := os.Stat(destination); !os.IsNotExist(err) {
					t.Fatal("failed migration left destination")
				}
			}
		})
	}
	source := legacyFixture(t)
	if _, err := household.Migrate(source, filepath.Join(source, "nested")); err == nil {
		t.Fatal("source-nested destination accepted")
	}
	parent := t.TempDir()
	if err := os.Symlink(source, filepath.Join(parent, "alias")); err != nil {
		t.Fatal(err)
	}
	if _, err := household.Migrate(source, filepath.Join(parent, "alias/nested")); err == nil {
		t.Fatal("source-nested symlink destination accepted")
	}
}

// @test CON-008-AC2
func TestMigrationYearOrderAndEmptyHousehold(t *testing.T) {
	source := legacyFixture(t)
	mustWrite(t, filepath.Join(source, "2024/christmas/Leon/wishlist.yaml"), "wishes: [{title: Model train}]\n")
	dest := filepath.Join(t.TempDir(), "new")
	_, err := household.Migrate(source, dest)
	if err != nil {
		t.Fatal(err)
	}
	h, _, err := household.NewStore(filepath.Join(dest, "household.yaml")).Load()
	if err != nil || h.Lists[0].Name != "Christmas 2024" || h.Lists[1].Name != "Christmas 2026" || len(h.Wishes) != 2 {
		t.Fatal("year order/duplicate occurrences lost")
	}
	source = filepath.Join(t.TempDir(), "empty")
	mustWrite(t, filepath.Join(source, "household.yaml"), "name: Family\nadmin_password_hash: "+fixtureHash()+"\n")
	dest = filepath.Join(t.TempDir(), "new")
	report, err := household.Migrate(source, dest)
	if err != nil || report.Lists != 0 || report.Wishes != 0 || report.Users != 0 {
		t.Fatal("empty household migration failed")
	}
}

// @test CON-004-AC4
func TestFailedSaveLeavesOriginalBytes(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permission failure requires a non-root user")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "household.yaml")
	store := household.NewStore(path)
	if err := store.Create(sample()); err != nil {
		t.Fatal(err)
	}
	before := mustRead(t, path)
	_, rev, _ := store.Load()
	if err := os.Chmod(dir, 0500); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(dir, 0700)
	if err := store.Update(rev, func(h *household.Household) error { h.Wishes[0].Note = "should not save"; return nil }); err == nil {
		t.Fatal("read-only-directory save unexpectedly succeeded")
	}
	if !bytes.Equal(before, mustRead(t, path)) {
		t.Fatal("failed save altered data")
	}
}
