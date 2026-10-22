package household

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// @test CON-004-AC5 WSH-006-AC14
func TestPictureTransactionRollsBackOnRemovalFailure(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(filepath.Join(dir, "household.yaml"))
	hash, hashErr := HashPassword("PublicFixture1!")
	if hashErr != nil {
		t.Fatal(hashErr)
	}
	h := Household{SchemaVersion: 2, Name: "Family", Language: "en-US", AdminPasswordHash: hash}
	h.Wishes = []Wish{{ID: "w1", Title: "Train picture", Picture: "w1-old.webp"}}
	if err := s.Create(h); err != nil {
		t.Fatal(err)
	}
	os.MkdirAll(s.PictureDirectory(), 0700)
	old := filepath.Join(s.PictureDirectory(), "w1-old.webp")
	os.WriteFile(old, []byte("old"), 0600)
	original, _ := os.ReadFile(s.path)
	_, rev, _ := s.Load()
	err := s.updatePicture(rev, "w1-old.webp", "w1-new.webp", []byte("new"), func(h *Household) error { h.Wishes[0].Picture = "w1-new.webp"; return nil }, func(string) error { return errors.New("blocked") })
	after, _ := os.ReadFile(s.path)
	if !errors.Is(err, ErrPictureRemove) || !bytes.Equal(original, after) {
		t.Fatal("failed removal changed YAML", err)
	}
	if raw, _ := os.ReadFile(old); string(raw) != "old" {
		t.Fatal("old picture lost")
	}
	if _, err := os.Stat(filepath.Join(s.PictureDirectory(), "w1-new.webp")); !os.IsNotExist(err) {
		t.Fatal("staged image leaked")
	}
}

// @test CON-004-AC5 WSH-006-AC7 WSH-006-AC14
func TestPictureTransactionRejectsSymlinkDirectory(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(filepath.Join(dir, "household.yaml"))
	hash, err := HashPassword("PublicFixture1!")
	if err != nil {
		t.Fatal(err)
	}
	h := Household{SchemaVersion: 2, Name: "Family", Language: "en-US", AdminPasswordHash: hash}
	if err = s.Create(h); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err = os.Symlink(outside, s.PictureDirectory()); err != nil {
		t.Fatal(err)
	}
	_, rev, _ := s.Load()
	err = s.UpdatePicture(rev, "", "w1-new.webp", []byte("new"), func(h *Household) error {
		h.Wishes = append(h.Wishes, Wish{ID: "w1", Title: "Picture title", Picture: "w1-new.webp"})
		return nil
	})
	if err == nil {
		t.Fatal("symlink picture directory accepted")
	}
	_, after, _ := s.Load()
	entries, _ := os.ReadDir(outside)
	if rev != after || len(entries) != 0 {
		t.Fatal("symlink failure modified files")
	}
}
