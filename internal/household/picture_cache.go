package household

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

func referenced(h Household, name string) bool {
	for _, w := range h.Wishes {
		if w.Picture == name {
			return true
		}
	}
	return false
}
func (s *Store) cleanupPath() string {
	return filepath.Join(filepath.Dir(s.path), "picture-cleanup.json")
}
func (s *Store) cleanupRecords() (map[string]retryRecord, error) {
	raw, err := readRegular(s.cleanupPath())
	if os.IsNotExist(err) {
		return map[string]retryRecord{}, nil
	}
	if err != nil {
		return nil, err
	}
	var records map[string]retryRecord
	if json.Unmarshal(raw, &records) != nil || records == nil {
		return nil, errors.New("invalid cleanup state")
	}
	for name := range records {
		if !SafePicture(name) {
			return nil, errors.New("invalid cleanup filename")
		}
	}
	return records, nil
}
func (s *Store) saveCleanup(records map[string]retryRecord) error {
	raw, err := json.Marshal(records)
	if err != nil {
		return err
	}
	return atomicPrivate(s.cleanupPath(), raw)
}
func (s *Store) queueCleanup(name string) error {
	if name == "" {
		return nil
	}
	records, err := s.cleanupRecords()
	if err != nil {
		return err
	}
	if _, ok := records[name]; !ok {
		records[name] = retryRecord{}
	}
	return s.saveCleanup(records)
}

// @impl WSH-008-AC9 WSH-008-AC10 WSH-008-AC11
// Record cleanup intentions before PUT: an uncertain save or process interruption
// retains old-image cleanup across interruption. A fresh remote reference check protects files
// belonging to writes that did not complete. Local transactions are unchanged.
func (s *Store) remotePicture(revision, oldPicture, newPicture string, raw []byte, change func(*Household) error, remove func(string) error) error {
	h, current, err := s.load()
	if err != nil {
		return err
	}
	if s.readOnly {
		return ErrReadOnly
	}
	if revision == "" || current != revision {
		return ErrStale
	}
	if err = change(&h); err != nil {
		return err
	}
	encoded, err := Encode(h)
	if err != nil {
		return err
	}
	if oldPicture != "" && !SafePicture(oldPicture) || newPicture != "" && !SafePicture(newPicture) {
		return ErrPictureRemove
	}
	if oldPicture != newPicture {
		if err = s.queueCleanup(oldPicture); err != nil {
			return err
		}
	}
	// Retain a newly published file if PUT is uncertain; remote data may reference it.
	if len(raw) > 0 {
		if err = publishPicture(s.PictureDirectory(), newPicture, raw); err != nil {
			return err
		}
	}
	if err = s.write(encoded, false, current); err != nil {
		return err
	}
	// Attempt immediately, retaining durable intent and backoff on failure. Never
	// report a confirmed remote save as rolled back because local cleanup failed.
	_ = s.cleanupPictures(time.Now(), remove)
	return nil
}

// publishPicture creates a fully synced regular file, atomically without replacing
// any existing path (including a symlink). Parent symlinks are rejected.
func publishPicture(dir, name string, raw []byte) error {
	if !SafePicture(name) {
		return errors.New("invalid picture name")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	if err := RegularPictureDirectory(dir); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".restore-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if err = file.Chmod(0600); err == nil {
		_, err = file.Write(raw)
	}
	if err == nil {
		err = file.Sync()
	}
	if e := file.Close(); err == nil {
		err = e
	}
	if err != nil {
		return err
	}
	if err = os.Link(file.Name(), filepath.Join(dir, name)); err != nil {
		return err
	}
	// @impl WSH-008-AC13: sync through a root-relative directory handle.
	if root, e := os.OpenRoot(dir); e == nil {
		defer root.Close()
		if parent, e := root.Open("."); e == nil {
			_ = parent.Sync()
			_ = parent.Close()
		}
	}
	return nil
}

// @impl WSH-008-AC6 WSH-008-AC7 WSH-008-AC12
func (s *Store) RestorePicture(wishID, picture, source string, raw []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	h, _, err := s.load()
	if err != nil {
		return err
	}
	current := false
	for _, w := range h.Wishes {
		if w.ID == wishID && w.Picture == picture && w.PictureSource == source {
			current = true
			break
		}
	}
	if !current {
		return nil
	}
	if _, err = os.Lstat(filepath.Join(s.PictureDirectory(), picture)); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	err = publishPicture(s.PictureDirectory(), picture, raw)
	if os.IsExist(err) {
		return nil
	}
	return err
}

// @impl WSH-008-AC10 WSH-008-AC12
func (s *Store) CleanupPictures(now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cleanupPictures(now, os.Remove)
}
func (s *Store) cleanupPictures(now time.Time, remove func(string) error) (resultErr error) {
	defer func() {
		if resultErr != nil {
			slog.Error("picture cleanup failed", "error_kind", storageErrorKind(resultErr))
		}
	}()
	if s.remote == nil {
		return nil
	}
	records, err := s.cleanupRecords()
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}
	h, _, err := s.load()
	if err != nil || s.readOnly {
		for name, record := range records {
			if !now.Before(record.Next) {
				records[name] = deferRetry(record, now)
			}
		}
		_ = s.saveCleanup(records)
		if err != nil {
			return err
		}
		return ErrReadOnly
	}
	changed := false
	for name, record := range records {
		if now.Before(record.Next) {
			continue
		}
		if referenced(h, name) {
			delete(records, name)
			changed = true
			continue
		}
		err = RegularPictureDirectory(s.PictureDirectory())
		if err == nil {
			path := filepath.Join(s.PictureDirectory(), name)
			info, e := os.Lstat(path)
			if os.IsNotExist(e) {
				err = nil
			} else if e != nil {
				err = e
			} else if !info.Mode().IsRegular() {
				err = ErrPictureRemove
			} else {
				err = remove(path)
			}
		} else if os.IsNotExist(err) {
			err = nil
		}
		changed = true
		if err == nil || os.IsNotExist(err) {
			delete(records, name)
		} else {
			record.Next = now.Add(RetryDelay(record.Attempts))
			slog.Error("picture cleanup failed", "error_kind", storageErrorKind(err), "retry_at", record.Next)
			if record.Attempts < 5 {
				record.Attempts++
			}
			records[name] = record
		}
	}
	if changed {
		return s.saveCleanup(records)
	}
	return nil
}
