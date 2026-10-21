package household

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

type MigrationCounts struct {
	Lists      int `json:"lists"`
	Recipients int `json:"recipients"`
	Wishes     int `json:"wishes"`
	Claims     int `json:"claims"`
}
type MigrationReport struct {
	SourceCounts      MigrationCounts `json:"source_counts"`
	DestinationCounts MigrationCounts `json:"destination_counts"`

	Lists               int                 `json:"lists"`
	Recipients          int                 `json:"recipients"`
	Users               int                 `json:"users"`
	Wishes              int                 `json:"wishes"`
	Claims              int                 `json:"claims"`
	Pictures            []PictureComparison `json:"pictures"`
	MissingPictures     int                 `json:"missing_pictures"`
	SourceUnchanged     bool                `json:"source_unchanged"`
	DestinationVerified bool                `json:"destination_verified"`
}
type PictureComparison struct {
	SourceSHA256      string `json:"source_sha256"`
	DestinationSHA256 string `json:"destination_sha256"`
}
type legacyConfig struct {
	Name              string      `yaml:"name"`
	Language          yaml.Node   `yaml:"language"`
	Appearance        *Appearance `yaml:"appearance"`
	AdminPasswordHash string      `yaml:"admin_password_hash"`
	Timezone          string      `yaml:"timezone"`
	ParentCode        string      `yaml:"parent_code"`
	FamilyCode        string      `yaml:"family_code"`
}
type legacyPerson struct {
	Name     string `yaml:"name"`
	Archived bool   `yaml:"archived"`
}
type legacyClaim struct {
	Name  string               `yaml:"name"`
	Extra map[string]yaml.Node `yaml:",inline"`
}
type legacyWish struct {
	Title         string               `yaml:"title"`
	Price         string               `yaml:"price"`
	Links         []string             `yaml:"links"`
	Note          string               `yaml:"note"`
	Picture       string               `yaml:"picture"`
	PictureSource string               `yaml:"picture_source"`
	Claims        []legacyClaim        `yaml:"claims"`
	Extra         map[string]yaml.Node `yaml:",inline"`
}

func decodeLegacy(raw []byte, target any) error {
	d := yaml.NewDecoder(bytes.NewReader(raw))
	d.KnownFields(true)
	if err := d.Decode(target); err != nil {
		return errors.New("invalid legacy YAML fields")
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return errors.New("expected one legacy YAML document")
	}
	return nil
}

// @impl CON-008-AC1 CON-008-AC2 CON-008-AC3 CON-008-AC4 CON-008-AC5
// Migrate reads a copied legacy HOUSEHOLD directory and publishes one new,
// previously absent HOUSEHOLD directory. It never updates the HTTP app's data.
func Migrate(source, destination string) (MigrationReport, error) {
	report := MigrationReport{}
	source, err := filepath.Abs(source)
	if err != nil {
		return report, err
	}
	source, err = filepath.EvalSymlinks(source)
	if err != nil {
		return report, err
	}
	destination, err = filepath.Abs(destination)
	if err != nil {
		return report, err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(destination))
	if err != nil {
		return report, err
	}
	destination = filepath.Join(parent, filepath.Base(destination))
	if rel, err := filepath.Rel(source, destination); err != nil || rel == "." || rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return report, errors.New("destination must be outside source")
	}
	if _, err := os.Lstat(destination); !os.IsNotExist(err) {
		return report, errors.New("destination must not exist")
	}
	snapshot, err := sourceSnapshot(source)
	if err != nil {
		return report, err
	}
	used := map[string]bool{}
	read := func(relative string) ([]byte, error) {
		used[relative] = true
		raw, err := os.ReadFile(filepath.Join(source, relative))
		return raw, err
	}
	raw, err := read("household.yaml")
	if err != nil {
		return report, errors.New("missing copied household configuration")
	}
	var config legacyConfig
	if err = decodeLegacy(raw, &config); err != nil {
		return report, err
	}
	if err = ValidatePasswordHash(config.AdminPasswordHash); err != nil {
		return report, errors.New("provide a supported admin_password_hash in the copied legacy household.yaml")
	}
	language := "en-US"
	if config.Language.Kind != 0 {
		if config.Language.Kind != yaml.ScalarNode || config.Language.Tag != "!!str" {
			return report, errors.New("invalid legacy household language")
		}
		language = config.Language.Value
	}
	h := Household{SchemaVersion: 2, Name: config.Name, Language: language, Appearance: config.Appearance, AdminPasswordHash: config.AdminPasswordHash, Sessions: []Session{}, Users: []Person{}, Recipients: []Person{}, Lists: []List{}, Wishes: []Wish{}}
	if h.Appearance == nil {
		h.Appearance = DefaultAppearance()
	}
	var roster struct {
		People []legacyPerson `yaml:"people"`
	}
	raw, err = read("people.yaml")
	if os.IsNotExist(err) {
		delete(used, "people.yaml")
	} else if err != nil {
		return report, err
	} else if err = decodeLegacy(raw, &roster); err != nil {
		return report, err
	}
	for _, p := range roster.People {
		if !validName(p.Name, 3, 50, false) {
			return report, errors.New("invalid legacy recipient name")
		}
		id, err := NewID()
		if err != nil {
			return report, err
		}
		h.Recipients = append(h.Recipients, Person{ID: id, Name: strings.TrimSpace(p.Name)})
	}
	entries, err := os.ReadDir(source)
	if err != nil {
		return report, err
	}
	years := []int{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		year, err := strconv.Atoi(entry.Name())
		if err != nil || year < 1 || strconv.Itoa(year) != entry.Name() {
			return report, errors.New("unrepresentable legacy directory")
		}
		years = append(years, year)
	}
	sort.Ints(years)
	// Stage everything outside the source, then publish only after complete validation.
	stage, err := os.MkdirTemp(parent, ".wishmas-migration-*")
	if err != nil {
		return report, err
	}
	defer os.RemoveAll(stage)
	if err = os.Mkdir(filepath.Join(stage, "pictures"), 0700); err != nil {
		return report, err
	}
	userID := func(name string) (string, error) {
		name = strings.TrimSpace(name)
		if !validName(name, 3, 50, false) {
			return "", errors.New("invalid legacy claimant name")
		}
		for _, user := range h.Users {
			if strings.EqualFold(user.Name, name) {
				return user.ID, nil
			}
		}
		id, err := NewID()
		if err != nil {
			return "", err
		}
		h.Users = append(h.Users, Person{ID: id, Name: name})
		return id, nil
	}
	for _, year := range years {
		listID, err := NewID()
		if err != nil {
			return report, err
		}
		list := List{ID: listID, Name: fmt.Sprintf("Christmas %d", year), RecipientIDs: []string{}, UserLinks: []UserLink{}}
		for _, recipient := range h.Recipients {
			list.RecipientIDs = append(list.RecipientIDs, recipient.ID)
			relative := filepath.Join(strconv.Itoa(year), "christmas", recipient.Name, "wishlist.yaml")
			raw, err := read(relative)
			if os.IsNotExist(err) {
				delete(used, relative)
				continue
			}
			if err != nil {
				return report, err
			}
			var saved struct {
				Wishes []legacyWish `yaml:"wishes"`
			}
			if err = decodeLegacy(raw, &saved); err != nil {
				return report, validationError("invalid legacy wishlist", err)
			}
			report.SourceCounts.Wishes += len(saved.Wishes)
			for _, old := range saved.Wishes {
				if _, collision := old.Extra["id"]; collision {
					return report, errors.New("legacy field conflicts with new wish identity")
				}
				if _, collision := old.Extra["assignment"]; collision {
					return report, errors.New("legacy field conflicts with new assignment")
				}
				id, err := NewID()
				if err != nil {
					return report, err
				}
				wish := Wish{ID: id, Title: old.Title, Price: old.Price, Links: old.Links, Note: old.Note, PictureSource: old.PictureSource, Extra: old.Extra, Assignment: &Assignment{ListID: list.ID, RecipientID: recipient.ID}}
				for _, claim := range old.Claims {
					if _, collision := claim.Extra["user_id"]; collision {
						return report, errors.New("legacy claim field conflicts with new identity")
					}
					user, err := userID(claim.Name)
					if err != nil {
						return report, err
					}
					wish.Claims = append(wish.Claims, Claim{UserID: user, Extra: claim.Extra})
					report.Claims++
					report.SourceCounts.Claims++
				}
				if old.Picture != "" {
					if !SafePicture(old.Picture) {
						return report, errors.New("unsafe legacy picture filename")
					}
					version, err := NewID()
					if err != nil {
						return report, err
					}
					wish.Picture = id + "-" + version + ".webp"
					imageRelative := filepath.Join(filepath.Dir(relative), "pictures", old.Picture)
					image, err := read(imageRelative)
					if os.IsNotExist(err) {
						delete(used, imageRelative)
						report.MissingPictures++
					} else if err != nil {
						return report, errors.New("cannot read legacy picture")
					} else {
						target := filepath.Join(stage, "pictures", wish.Picture)
						if err = os.WriteFile(target, image, 0600); err != nil {
							return report, err
						}
						copyBytes, err := os.ReadFile(target)
						if err != nil {
							return report, err
						}
						a, b := sha256.Sum256(image), sha256.Sum256(copyBytes)
						if a != b {
							return report, errors.New("picture verification failed")
						}
						report.Pictures = append(report.Pictures, PictureComparison{hex.EncodeToString(a[:]), hex.EncodeToString(b[:])})
					}
				}
				h.Wishes = append(h.Wishes, wish)
			}
		}
		h.Lists = append(h.Lists, list)
	}
	for i := range h.Lists {
		for _, user := range h.Users {
			code, err := NewToken()
			if err != nil {
				return report, err
			}
			h.Lists[i].UserLinks = append(h.Lists[i].UserLinks, UserLink{UserID: user.ID, Code: code})
		}
	}
	for path := range snapshot {
		if !used[path] {
			return report, errors.New("unrepresentable legacy file; migration refused without discarding data")
		}
	}
	if err = h.Validate(); err != nil {
		return report, err
	}
	store := NewStore(filepath.Join(stage, "household.yaml"))
	if err = store.Create(h); err != nil {
		return report, err
	}
	loaded, _, err := store.Load()
	if err != nil {
		return report, err
	}
	original, err := Encode(h)
	if err != nil {
		return report, err
	}
	roundtrip, err := Encode(loaded)
	if err != nil || !bytes.Equal(original, roundtrip) {
		return report, errors.New("destination round-trip verification failed")
	}
	after, err := sourceSnapshot(source)
	if err != nil || !sameSnapshot(snapshot, after) {
		return report, errors.New("source changed during migration; retry from a consistent copy")
	}
	report.SourceCounts.Lists = len(years)
	report.SourceCounts.Recipients = len(roster.People)
	report.DestinationCounts = MigrationCounts{Lists: len(loaded.Lists), Recipients: len(loaded.Recipients), Wishes: len(loaded.Wishes)}
	for _, wish := range loaded.Wishes {
		report.DestinationCounts.Claims += len(wish.Claims)
	}
	if report.SourceCounts != report.DestinationCounts {
		return report, errors.New("source/destination record counts differ")
	}
	// Reserve the absent destination exclusively; rollback our directory on failure.
	// A regular rename onto an empty concurrent directory could otherwise clobber it.
	if err = os.Mkdir(destination, 0700); err != nil {
		return report, errors.New("destination appeared during migration")
	}
	committed := false
	defer func() {
		if !committed {
			_ = os.RemoveAll(destination)
		}
	}()
	if err = os.Rename(filepath.Join(stage, "pictures"), filepath.Join(destination, "pictures")); err != nil {
		return report, err
	}
	if err = os.Rename(filepath.Join(stage, "household.yaml"), filepath.Join(destination, "household.yaml")); err != nil {
		return report, err
	}
	committed = true
	report.Lists = len(h.Lists)
	report.Recipients = len(h.Recipients)
	report.Users = len(h.Users)
	report.Wishes = len(h.Wishes)
	report.SourceUnchanged = true
	report.DestinationVerified = true
	return report, nil
}
func sourceSnapshot(root string) (map[string][32]byte, error) {
	result := map[string][32]byte{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return errors.New("legacy source contains symlink")
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return errors.New("legacy source contains non-regular file")
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, _ := filepath.Rel(root, path)
		result[relative] = sha256.Sum256(raw)
		return nil
	})
	return result, err
}
func sameSnapshot(a, b map[string][32]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for path, sum := range a {
		if other, ok := b[path]; !ok || other != sum {
			return false
		}
	}
	return true
}
