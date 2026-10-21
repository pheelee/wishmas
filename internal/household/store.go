package household

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

var ErrStale = errors.New("household changed; reload before saving")
var storeLocks sync.Map

// Store serializes app mutations across repository instances for the same path.
// Manual writers are detected by revision comparison; they do not share this lock.
type Store struct {
	path string
	mu   *sync.Mutex
}

func NewStore(path string) *Store {
	absolute, _ := filepath.Abs(path)
	mu, _ := storeLocks.LoadOrStore(absolute, &sync.Mutex{})
	return &Store{path: absolute, mu: mu.(*sync.Mutex)}
}

// @impl CON-004-AC1 CON-004-AC3
func (s *Store) Load() (Household, string, error) { s.mu.Lock(); defer s.mu.Unlock(); return s.load() }
func (s *Store) load() (Household, string, error) {
	info, err := os.Lstat(s.path)
	if err != nil {
		return Household{}, "", err
	}
	if !info.Mode().IsRegular() {
		return Household{}, "", errors.New("household must be a regular file")
	}
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return Household{}, "", err
	}
	h, err := Decode(raw)
	if err != nil {
		return Household{}, "", err
	}
	sum := sha256.Sum256(raw)
	return h, hex.EncodeToString(sum[:]), nil
}

// @impl CON-004-AC4: callback edits an isolated fresh snapshot, never cached data.
func (s *Store) Update(revision string, change func(*Household) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	h, current, err := s.load()
	if err != nil {
		return err
	}
	if revision == "" || revision != current {
		return ErrStale
	}
	if err = change(&h); err != nil {
		return err
	}
	raw, err := Encode(h)
	if err != nil {
		return err
	}
	return s.write(raw, false, current)
}

// Create publishes without overwriting an existing file, even an invalid one.
func (s *Store) Create(h Household) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	raw, err := Encode(h)
	if err != nil {
		return err
	}
	return s.write(raw, true, "")
}
func (s *Store) write(raw []byte, create bool, revision string) error {
	if info, err := os.Lstat(s.path); err == nil && (!info.Mode().IsRegular() || create) {
		return errors.New("destination exists or is not a regular file")
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	parent := filepath.Dir(s.path)
	file, err := os.CreateTemp(parent, ".household-*")
	if err != nil {
		return err
	}
	tmp := file.Name()
	defer os.Remove(tmp)
	defer file.Close()
	if err = file.Chmod(0600); err != nil {
		return err
	}
	if _, err = file.Write(raw); err != nil {
		return err
	}
	if err = file.Sync(); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	if create {
		if err = os.Link(tmp, s.path); err != nil {
			return err
		}
	} else {
		current, err := os.ReadFile(s.path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(current)
		if hex.EncodeToString(sum[:]) != revision {
			return ErrStale
		}
		if err = os.Rename(tmp, s.path); err != nil {
			return err
		}
	}
	// The new file is already committed. A directory sync failure is not reported
	// as a rolled-back save; the best-effort sync improves crash durability.
	if dir, err := os.Open(parent); err == nil {
		_ = dir.Sync()
		_ = dir.Close()
	}
	return nil
}
