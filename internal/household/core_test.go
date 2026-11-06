package household_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/pheelee/wishmas/internal/household"
	"go.yaml.in/yaml/v3"
)

func sample() household.Household {
	return household.Household{SchemaVersion: 2, Name: "Family", Language: "en-US", Appearance: household.DefaultAppearance(), AdminPasswordHash: fixtureHash(), Users: []household.Person{{ID: "u1", Name: "Grand Dad"}}, Recipients: []household.Person{{ID: "r1", Name: "Grand Dad"}}, Lists: []household.List{{ID: "l1", Name: "Christmas 2026!", RecipientIDs: []string{"r1"}, UserLinks: []household.UserLink{{UserID: "u1", Code: strings.Repeat("a", 64)}}}}, Wishes: []household.Wish{{ID: "w1", Title: "Train engine", Price: "0", Links: []string{"https://shop.example/train"}, Claims: []household.Claim{{UserID: "u1"}}, Assignment: &household.Assignment{ListID: "l1", RecipientID: "r1"}}}}
}

// @test CON-007-AC1 CON-007-AC2 ACC-005-AC2
func TestPasswordPolicyHashAndBoundedVerification(t *testing.T) {
	for _, password := range []string{"Abcdef1!", "Äbcdef١!", "Long mixed 1! passphrase"} {
		if err := household.ValidatePassword(password); err != nil {
			t.Fatal("valid password rejected")
		}
	}
	for _, password := range []string{"Ab1!", "abcdef1!", "ABCDEF1!", "Abcdefgh!", "Abcdef1 ", "Abcdef١\u0001"} {
		if household.ValidatePassword(password) == nil {
			t.Fatal("invalid password accepted")
		}
	}
	a, err := household.HashPassword("PublicFixture1!")
	if err != nil {
		t.Fatal(err)
	}
	b, err := household.HashPassword("PublicFixture1!")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("salt reused")
	}
	ok, err := household.CheckPassword(a, "PublicFixture1!")
	if err != nil || !ok {
		t.Fatal("correct password rejected")
	}
	ok, err = household.CheckPassword(a, "WrongFixture1!")
	if err != nil || ok {
		t.Fatal("wrong password accepted")
	}
	for _, bad := range []string{"", strings.Repeat("A", 20000), strings.Replace(a, "m=19456", "m=4294967295", 1), strings.Replace(a, "t=2", "t=999999999", 1), strings.Replace(a, "argon2id", "argon2i", 1), a + "=", strings.Replace(a, "v=19", "v=16", 1)} {
		if _, err := household.CheckPassword(bad, "PublicFixture1!"); err == nil {
			t.Fatal("unsupported/unbounded hash accepted")
		}
	}
}

// @test CON-004-AC2 CON-004-AC3 WLS-010-AC2 ACC-007-AC2
func TestModelRejectsInvalidReferencesAndNames(t *testing.T) {
	cases := map[string]func(*household.Household){
		"unknown schema": func(h *household.Household) { h.SchemaVersion = 1 },
		"password hash":  func(h *household.Household) { h.AdminPasswordHash = "not-a-hash" },
		"duplicate IDs":  func(h *household.Household) { h.Wishes[0].ID = "u1" },
		"duplicate names": func(h *household.Household) {
			h.Users = append(h.Users, household.Person{ID: "u2", Name: " grand dad "})
		},
		"short name":            func(h *household.Household) { h.Recipients[0].Name = " A " },
		"punctuation in roster": func(h *household.Household) { h.Recipients[0].Name = "Dad!" },
		"unknown claimant":      func(h *household.Household) { h.Wishes[0].Claims[0].UserID = "absent" },
		"duplicate claim":       func(h *household.Household) { h.Wishes[0].Claims = append(h.Wishes[0].Claims, h.Wishes[0].Claims[0]) },
		"unselected recipient":  func(h *household.Household) { h.Lists[0].RecipientIDs = nil },
		"unknown list":          func(h *household.Household) { h.Wishes[0].Assignment.ListID = "missing" },
		"missing user link":     func(h *household.Household) { h.Lists[0].UserLinks = nil },
		"duplicate code": func(h *household.Household) {
			h.Lists = append(h.Lists, household.List{ID: "l2", Name: "Birthday party", UserLinks: []household.UserLink{h.Lists[0].UserLinks[0]}})
		},
		"duplicate list name": func(h *household.Household) {
			h.Lists = append(h.Lists, household.List{ID: "l2", Name: " christmas 2026! "})
		},
		"bad source URL":      func(h *household.Household) { h.Wishes[0].PictureSource = "not a URL" },
		"traversal image":     func(h *household.Household) { h.Wishes[0].Picture = "../other.webp" },
		"short title":         func(h *household.Household) { h.Wishes[0].Title = " abc " },
		"negative price":      func(h *household.Household) { h.Wishes[0].Price = "-0.01" },
		"three decimal price": func(h *household.Household) { h.Wishes[0].Price = "1.001" },
		"bare link":           func(h *household.Household) { h.Wishes[0].Links = []string{"shop.example/a"} },
		"invalid session":     func(h *household.Household) { h.Sessions = []household.Session{{TokenHash: "raw-token"}} },
	}
	if err := sample().Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			h := sample()
			mutate(&h)
			if h.Validate() == nil {
				t.Fatal("invalid household accepted")
			}
		})
	}
	h := sample()
	h.Wishes[0].Assignment = nil
	if err := h.Validate(); err != nil {
		t.Fatal("unassigned wish with claims rejected")
	}
}

// @test CON-004-AC3 INT-007-AC2
func TestDecodeLanguageAndStrictYAML(t *testing.T) {
	raw, err := household.Encode(sample())
	if err != nil {
		t.Fatal(err)
	}
	missing := bytes.Replace(raw, []byte("language: en-US\n"), nil, 1)
	h, err := household.Decode(missing)
	if err != nil || h.Language != "en-US" {
		t.Fatal("missing-language fallback failed")
	}
	for _, replacement := range []string{"language:\n", "language: null\n", "language: ''\n", "language: fr-FR\n"} {
		if _, err := household.Decode(bytes.Replace(raw, []byte("language: en-US\n"), []byte(replacement), 1)); err == nil {
			t.Fatal("invalid language accepted")
		}
	}
	for _, invalid := range [][]byte{append(append([]byte{}, raw...), []byte("---\nname: second\n")...), append(append([]byte{}, raw...), []byte("schema_version: 2\n")...), append(append([]byte{}, raw...), []byte("parent_code: legacy\n")...)} {
		if _, err := household.Decode(invalid); err == nil {
			t.Fatal("invalid/multiple/legacy YAML accepted")
		}
	}
}

// @test CON-004-AC3 CON-004-AC4
func TestStoreAtomicValidationStaleSavesAndManualEdits(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "household.yaml")
	store := household.NewStore(path)
	h := sample()
	h.Wishes[0].Extra = map[string]yaml.Node{"future_field": {Kind: yaml.ScalarNode, Tag: "!!str", Value: "retain me"}}
	if err := store.Create(h); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)
	_, revision, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Create(sample()); err == nil {
		t.Fatal("create overwrote existing data")
	}
	if err = store.Update(revision, func(h *household.Household) error { h.Wishes[0].Claims[0].UserID = "absent"; return nil }); err == nil {
		t.Fatal("invalid mutation saved")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("failed mutation changed file")
	}
	if err = store.Update(revision, func(h *household.Household) error { h.Wishes[0].Note = "changed"; return nil }); err != nil {
		t.Fatal(err)
	}
	loaded, _, err := store.Load()
	if err != nil || loaded.Wishes[0].Extra["future_field"].Value != "retain me" || len(loaded.Wishes[0].Claims) != 1 {
		t.Fatal("unrelated data lost")
	}
	if err = store.Update(revision, func(h *household.Household) error { return nil }); !errors.Is(err, household.ErrStale) {
		t.Fatal("stale mutation accepted")
	}
	_, fresh, _ := store.Load()
	manual := append([]byte("# manual edit\n"), mustRead(t, path)...)
	if err = store.Update(fresh, func(h *household.Household) error { mustWrite(t, path, string(manual)); h.Wishes = nil; return nil }); !errors.Is(err, household.ErrStale) {
		t.Fatal("concurrent manual edit overwritten")
	}
	if !bytes.Equal(manual, mustRead(t, path)) {
		t.Fatal("manual edit changed")
	}
	mustWrite(t, path, "not: [valid YAML")
	if _, _, err = store.Load(); err == nil {
		t.Fatal("invalid manual YAML accepted")
	}
	if string(mustRead(t, path)) != "not: [valid YAML" {
		t.Fatal("invalid manual YAML rewritten")
	}
}

// @test CON-004-AC4
func TestStoreSerializesIndependentInstances(t *testing.T) {
	path := filepath.Join(t.TempDir(), "household.yaml")
	store := household.NewStore(path)
	if err := store.Create(sample()); err != nil {
		t.Fatal(err)
	}
	_, revision, _ := store.Load()
	var successes atomic.Int32
	var stale atomic.Int32
	var wg sync.WaitGroup
	for range 12 {
		wg.Go(func() {
			err := household.NewStore(path).Update(revision, func(h *household.Household) error { h.Wishes[0].Note = "saved"; return nil })
			if err == nil {
				successes.Add(1)
			} else if errors.Is(err, household.ErrStale) {
				stale.Add(1)
			} else {
				t.Errorf("save: %v", err)
			}
		})
	}
	wg.Wait()
	if successes.Load() != 1 || stale.Load() != 11 {
		t.Fatal("concurrent stale saves were not rejected")
	}
}
func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// @test CON-004-AC3 CON-004-AC4
func TestEncodingConflictingFutureFieldsReturnsError(t *testing.T) {
	h := sample()
	h.Wishes[0].Extra = map[string]yaml.Node{"title": {Kind: yaml.ScalarNode, Tag: "!!str", Value: "collision"}}
	if _, err := household.Encode(h); err == nil {
		t.Fatal("conflicting future field accepted")
	}
}

// @test CON-004-AC3 CON-004-AC4
func TestStoreDoesNotFollowSymlink(t *testing.T) {
	dir := t.TempDir()
	original := filepath.Join(dir, "original.yaml")
	store := household.NewStore(original)
	if err := store.Create(sample()); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "household.yaml")
	if err := os.Symlink(original, link); err != nil {
		t.Fatal(err)
	}
	if _, _, err := household.NewStore(link).Load(); err == nil {
		t.Fatal("symlink household accepted")
	}
}

// @test CON-004-AC3 INT-007-AC2
func TestMergedLanguageUsesPresentValue(t *testing.T) {
	raw, err := household.Encode(sample())
	if err != nil {
		t.Fatal(err)
	}
	raw = bytes.Replace(raw, []byte("language: en-US\n"), nil, 1)
	raw = append([]byte("<<: {language: de-CH}\n"), raw...)
	h, err := household.Decode(raw)
	if err != nil || h.Language != "de-CH" {
		t.Fatal("present merged language incorrectly received missing-field fallback")
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
