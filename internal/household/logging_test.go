package household

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// @test CON-004-AC9 OPS-001-AC3
func TestPictureRollbackCleanupDiagnostics(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission failure requires a non-root user")
	}
	for _, missing := range []bool{false, true} {
		t.Run(map[bool]string{false: "permission denied", true: "already missing"}[missing], func(t *testing.T) {
			s := NewStore(filepath.Join(t.TempDir(), "household.yaml"))
			hash, err := HashPassword("PublicFixture1!")
			if err != nil {
				t.Fatal(err)
			}
			h := Household{SchemaVersion: 2, Name: "Family", Language: "en-US", AdminPasswordHash: hash, Wishes: []Wish{{ID: "w1", Title: "Picture title", Picture: "w1-old.webp"}}}
			if err := s.Create(h); err != nil {
				t.Fatal(err)
			}
			dir := s.PictureDirectory()
			if err := os.Mkdir(dir, 0700); err != nil {
				t.Fatal(err)
			}
			defer os.Chmod(dir, 0700)
			oldPath := filepath.Join(dir, "w1-old.webp")
			if err := os.WriteFile(oldPath, []byte("old"), 0600); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(s.path)
			if err != nil {
				t.Fatal(err)
			}
			_, rev, err := s.Load()
			if err != nil {
				t.Fatal(err)
			}
			var logs bytes.Buffer
			previous := slog.Default()
			slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
			defer slog.SetDefault(previous)
			err = s.updatePicture(rev, "w1-old.webp", "w1-new.webp", []byte("new"), func(h *Household) error {
				h.Wishes[0].Picture = "w1-new.webp"
				return nil
			}, func(string) error {
				if missing {
					if err := os.Remove(filepath.Join(dir, "w1-new.webp")); err != nil {
						t.Fatal(err)
					}
				} else if err := os.Chmod(dir, 0500); err != nil {
					t.Fatal(err)
				}
				return errors.New("private failure details")
			})
			if !errors.Is(err, ErrPictureRemove) {
				t.Fatalf("original transaction error lost: %v", err)
			}
			after, err := os.ReadFile(s.path)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("YAML rollback failed")
			}
			if raw, err := os.ReadFile(oldPath); err != nil || string(raw) != "old" {
				t.Fatal("old picture changed")
			}
			logged := strings.Contains(logs.String(), "picture rollback cleanup failed")
			if logged == missing {
				t.Fatalf("unexpected cleanup diagnostic: %s", &logs)
			}
			if !missing && !strings.Contains(logs.String(), "permission denied") {
				t.Fatal("error category missing")
			}
			for _, private := range []string{dir, "w1-new.webp", "private failure details"} {
				if strings.Contains(logs.String(), private) {
					t.Fatal("private diagnostic leaked")
				}
			}
		})
	}
}

// @test OPS-001-AC3
func TestFailedSaveLogsWithoutHouseholdOrPath(t *testing.T) {
	var logs bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	defer slog.SetDefault(old)
	s := NewStore(filepath.Join(t.TempDir(), "PrivateHouseholdName", "household.yaml"))
	if err := s.write([]byte("SecretHouseholdContent"), true, ""); err == nil {
		t.Fatal("write should fail")
	}
	if !strings.Contains(logs.String(), "household save failed") || !strings.Contains(logs.String(), "file does not exist") {
		t.Fatal(logs.String())
	}
	for _, secret := range []string{"PrivateHouseholdName", "SecretHouseholdContent", "household.yaml"} {
		if strings.Contains(logs.String(), secret) {
			t.Fatal("private data leaked")
		}
	}
}
