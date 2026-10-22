package household

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrPictureRemove = errors.New("cannot remove old picture")

func (s *Store) PictureDirectory() string { return filepath.Join(filepath.Dir(s.path), "pictures") }
func RegularPictureDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("invalid picture directory")
	}
	return nil
}

// @impl CON-004-AC5 WSH-006-AC7 WSH-006-AC12 WSH-006-AC14
// Serialize YAML publication and image cleanup with every repository writer.
func (s *Store) UpdatePicture(revision, oldPicture, newPicture string, raw []byte, change func(*Household) error) error {
	return s.updatePicture(revision, oldPicture, newPicture, raw, change, os.Remove)
}
func (s *Store) updatePicture(revision, oldPicture, newPicture string, raw []byte, change func(*Household) error, remove func(string) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	h, current, err := s.load()
	if err != nil {
		return err
	}
	if revision == "" || revision != current {
		return ErrStale
	}
	original, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	if err = change(&h); err != nil {
		return err
	}
	encoded, err := Encode(h)
	if err != nil {
		return err
	}
	dir := s.PictureDirectory()
	oldPath := ""
	if oldPicture != "" && oldPicture != newPicture {
		if !SafePicture(oldPicture) {
			return ErrPictureRemove
		}
		for _, w := range h.Wishes {
			if w.Picture == oldPicture {
				return ErrPictureRemove
			}
		}
		if err := RegularPictureDirectory(dir); err != nil && !os.IsNotExist(err) {
			return ErrPictureRemove
		}
		oldPath = filepath.Join(dir, oldPicture)
		if info, err := os.Lstat(oldPath); err == nil && !info.Mode().IsRegular() {
			return ErrPictureRemove
		} else if err != nil && !os.IsNotExist(err) {
			return ErrPictureRemove
		}
	}
	newPath := ""
	committed := false
	defer func() {
		if !committed && newPath != "" {
			os.Remove(newPath)
		}
	}()
	if len(raw) > 0 {
		if !SafePicture(newPicture) {
			return errors.New("invalid new picture")
		}
		if err = os.MkdirAll(dir, 0700); err != nil {
			return err
		}
		if err = RegularPictureDirectory(dir); err != nil {
			return err
		}
		newPath = filepath.Join(dir, newPicture)
		f, err := os.OpenFile(newPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			newPath = ""
			return err
		}
		_, err = f.Write(raw)
		if err == nil {
			err = f.Sync()
		}
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}
	backup, err := os.CreateTemp(filepath.Dir(s.path), ".picture-rollback-*")
	if err != nil {
		return err
	}
	defer os.Remove(backup.Name())
	_, err = backup.Write(original)
	if err == nil {
		err = backup.Sync()
	}
	closeErr := backup.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err = s.write(encoded, false, current); err != nil {
		return err
	}
	if oldPath != "" {
		if err = remove(oldPath); err != nil && !os.IsNotExist(err) {
			if rollbackErr := os.Rename(backup.Name(), s.path); rollbackErr != nil {
				return rollbackErr
			}
			return ErrPictureRemove
		}
	}
	committed = true
	return nil
}
