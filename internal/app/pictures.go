package app

import (
	"bytes"
	"image"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func safePictureName(name string) bool {
	return name != "" && filepath.Base(name) == name && !strings.ContainsAny(name, `/\\`) && strings.HasSuffix(name, ".webp") && !strings.HasPrefix(name, ".")
}

func readPicture(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxPictureBytes {
		return nil, errPictureFormat
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || format != "webp" || config.Width > 640 || config.Height > 640 {
		return nil, errPictureFormat
	}
	return raw, nil
}

// @impl WSH-006-AC7 WSH-006-AC9 WSH-006-AC11 WSH-006-AC15
func (a *application) prepareWishPictures(data *view) {
	if data.PersonName == "" {
		return
	}
	h, err := a.load()
	if err != nil || h == nil {
		return
	}
	for i := range data.Wishes {
		card := &data.Wishes[i]
		if card.Picture == "" {
			continue
		}
		card.PictureUnavailable = true
		if !safePictureName(card.Picture) {
			continue
		}
		path := filepath.Join(filepath.Dir(a.wishlistPath(*h, data.Year, data.PersonName)), "pictures", card.Picture)
		if _, err := readPicture(path); err != nil {
			continue
		}
		query := url.Values{"code": {data.Code}, "year": {strconv.Itoa(data.Year)}, "person": {data.PersonName}, "picture": {card.Picture}}
		card.PictureURL = "/wish-picture?" + query.Encode()
		card.PictureUnavailable = false
	}
}

// @impl WSH-006-AC11 WSH-006-AC15
func (a *application) serveWishPicture(w http.ResponseWriter, r *http.Request, h household, role string) {
	year, err := strconv.Atoi(r.URL.Query().Get("year"))
	if err != nil || year < 1 || strconv.Itoa(year) != r.URL.Query().Get("year") || (role != "Parent" && year != a.currentYear(h)) {
		a.renderError(w, 404, "404 page not found")
		return
	}
	name, picture := r.URL.Query().Get("person"), r.URL.Query().Get("picture")
	if !safePictureName(picture) {
		a.renderError(w, 404, "404 page not found")
		return
	}
	state, err := a.readPeopleYear(h, year)
	if err != nil || activePersonIndex(state.People, name) < 0 {
		a.renderError(w, 404, "404 page not found")
		return
	}
	list, exists := state.Lists[name]
	if !exists {
		a.renderError(w, 404, "404 page not found")
		return
	}
	cards, err := wishCards(list)
	if err != nil {
		a.renderError(w, 404, "404 page not found")
		return
	}
	attached := false
	for _, card := range cards {
		if card.Picture == picture {
			attached = true
		}
	}
	if !attached {
		a.renderError(w, 404, "404 page not found")
		return
	}
	raw, err := readPicture(filepath.Join(filepath.Dir(a.wishlistPath(h, year, name)), "pictures", picture))
	if err != nil {
		a.renderError(w, 404, "404 page not found")
		return
	}
	w.Header().Set("Content-Type", "image/webp")
	w.Write(raw)
}

// @impl WSH-006-AC7 WSH-006-AC12 WSH-006-AC13 WSH-006-AC14
// A staged YAML file allows rollback by rename without allocating another file.
// Requests hold the application mutex throughout publication and cleanup.
func savePictureWish(path string, list wishlist, oldPicture, newPicture string, raw []byte) error {
	return pictureTransaction(path, list, oldPicture, newPicture, raw, os.Remove)
}
func pictureTransaction(path string, list wishlist, oldPicture, newPicture string, raw []byte, remove func(string) error) error {
	original, err := os.ReadFile(path)
	if err != nil {
		return errPictureSave
	}
	rollback, err := os.CreateTemp(filepath.Dir(path), ".picture-rollback-*")
	if err != nil {
		return errPictureSave
	}
	defer os.Remove(rollback.Name())
	if _, err = rollback.Write(original); err == nil {
		err = rollback.Sync()
	}
	closeErr := rollback.Close()
	if err != nil || closeErr != nil {
		return errPictureSave
	}
	pictures := filepath.Join(filepath.Dir(path), "pictures")
	oldPath := ""
	if oldPicture != "" && oldPicture != newPicture {
		if !safePictureName(oldPicture) {
			return errPictureRemove
		}
		// Never remove a file still referenced by another wish after a manual edit.
		for _, node := range list.Wishes {
			var card wishCard
			if node.Decode(&card) != nil || card.Picture == oldPicture {
				return errPictureRemove
			}
		}
		oldPath = filepath.Join(pictures, oldPicture)
		if info, err := os.Lstat(oldPath); err == nil && !info.Mode().IsRegular() {
			return errPictureRemove
		} else if err != nil && !os.IsNotExist(err) {
			return errPictureRemove
		}
	}
	newPath := ""
	if len(raw) != 0 {
		if !safePictureName(newPicture) || os.MkdirAll(pictures, 0700) != nil {
			return errPictureSave
		}
		newPath = filepath.Join(pictures, newPicture)
		f, err := os.OpenFile(newPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return errPictureSave
		}
		_, err = f.Write(raw)
		if err == nil {
			err = f.Sync()
		}
		closeErr := f.Close()
		if err != nil || closeErr != nil {
			os.Remove(newPath)
			return errPictureSave
		}
	}
	if err := writeYAML(path, list, false); err != nil {
		if newPath != "" {
			os.Remove(newPath)
		}
		return errPictureSave
	}
	if oldPath != "" {
		if err := remove(oldPath); err != nil && !os.IsNotExist(err) {
			// Same-directory rename restores the exact previous bytes, including comments.
			if err := os.Rename(rollback.Name(), path); err != nil {
				return errPictureSave
			}
			if newPath != "" {
				os.Remove(newPath)
			}
			return errPictureRemove
		}
	}
	return nil
}
