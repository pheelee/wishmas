package household_test

import (
	"crypto/sha256"
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// This syntactically valid public fixture hash is not a household credential.
func fixtureHash() string {
	return "$argon2id$v=19$m=19456,t=2,p=1$" + base64.RawStdEncoding.EncodeToString(make([]byte, 16)) + "$" + base64.RawStdEncoding.EncodeToString(make([]byte, 32))
}

// @test CON-004-AC1 CON-008-AC1 CON-008-AC2 CON-008-AC3 CON-008-AC5
func TestMigrationCommandPreservesSourceAndClaims(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "legacy")
	destination := filepath.Join(root, "review")
	mustWrite(t, filepath.Join(source, "household.yaml"), "name: Family\ntimezone: Europe/Zurich\nparent_code: old-parent\nfamily_code: old-family\nlanguage: de-CH\nadmin_password_hash: "+fixtureHash()+"\n")
	mustWrite(t, filepath.Join(source, "people.yaml"), "people:\n - name: Leon\n - name: Anna\n   archived: true\n")
	mustWrite(t, filepath.Join(source, "2026/christmas/Leon/wishlist.yaml"), "wishes:\n - title: Model train\n   price: '12.30'\n   links: [https://shop.example/train]\n   claims: [{name: Dad}]\n   future_field: preserve me\n - title: Model train\n   claims: [{name: ' dad '}]\n")
	mustWrite(t, filepath.Join(source, "2026/christmas/Anna/wishlist.yaml"), "wishes:\n - title: Other present\n   claims: [{name: Dád}]\n")
	before := treeHashes(t, source)
	command := exec.Command("go", "run", "../../cmd/wishmas-admin", "migrate", "--source", source, "--destination", destination)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("migration command failed: %v: %s", err, output)
	}
	var doc struct {
		Version    int `yaml:"schema_version"`
		Users      []struct{ ID, Name string }
		Recipients []struct{ ID, Name string }
		Lists      []struct{ ID, Name string }
		Wishes     []struct {
			Title  string
			Claims []struct {
				UserID string `yaml:"user_id"`
			}
			Future string `yaml:"future_field"`
		}
	}
	raw, err := os.ReadFile(filepath.Join(destination, "household.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err = yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Version != 2 || len(doc.Users) != 2 || len(doc.Recipients) != 2 || len(doc.Lists) != 1 || doc.Lists[0].Name != "Christmas 2026" || len(doc.Wishes) != 3 {
		t.Fatal("migration lost records or failed to merge case/space claimant names")
	}
	if doc.Wishes[0].Future != "preserve me" || doc.Wishes[0].Claims[0].UserID != doc.Wishes[1].Claims[0].UserID || doc.Users[0].Name != "Dad" {
		t.Fatal("claim identity/spelling or future fields lost")
	}
	for path, digest := range before {
		if treeHashes(t, source)[path] != digest {
			t.Fatal("source changed")
		}
	}
	for _, secret := range []string{fixtureHash(), "old-parent", "old-family"} {
		if strings.Contains(string(output), secret) {
			t.Fatal("report exposed credentials")
		}
	}
	if strings.Contains(string(raw), "parent_code:") || strings.Contains(string(raw), "timezone:") || strings.Contains(string(raw), "archived:") {
		t.Fatal("obsolete fields persisted")
	}
}
func mustWrite(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
}
func treeHashes(t *testing.T, dir string) map[string][32]byte {
	t.Helper()
	out := map[string][32]byte{}
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			b, e := os.ReadFile(path)
			if e != nil {
				return e
			}
			out[path] = sha256.Sum256(b)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
