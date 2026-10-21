// Package household implements the version-2 data contract independently of the
// legacy HTTP application. Its only legacy reader is the explicit migrator.
package household

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"
)

// @impl CON-004-AC1 CON-004-AC2
// Slice position is added order; assignment never moves a wish within Wishes.
type Household struct {
	SchemaVersion     int         `yaml:"schema_version"`
	Name              string      `yaml:"name"`
	Language          string      `yaml:"language"`
	Appearance        *Appearance `yaml:"appearance,omitempty"`
	AdminPasswordHash string      `yaml:"admin_password_hash"`
	Sessions          []Session   `yaml:"sessions"`
	Users             []Person    `yaml:"users"`
	Recipients        []Person    `yaml:"recipients"`
	Lists             []List      `yaml:"lists"`
	Wishes            []Wish      `yaml:"wishes"`
}
type Appearance struct {
	BulbSize   int `yaml:"bulb_size"`
	SnowSize   int `yaml:"snow_size"`
	SnowAmount int `yaml:"snow_amount"`
}
type Session struct {
	TokenHash               string `yaml:"token_hash"`
	PasswordHashFingerprint string `yaml:"password_hash_fingerprint"`
}
type Person struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}
type List struct {
	ID           string     `yaml:"id"`
	Name         string     `yaml:"name"`
	RecipientIDs []string   `yaml:"recipient_ids"`
	UserLinks    []UserLink `yaml:"user_links"`
}
type UserLink struct {
	UserID string `yaml:"user_id"`
	Code   string `yaml:"code"`
}
type Assignment struct {
	ListID      string `yaml:"list_id"`
	RecipientID string `yaml:"recipient_id"`
}
type Claim struct {
	UserID string               `yaml:"user_id"`
	Extra  map[string]yaml.Node `yaml:",inline"`
}
type Wish struct {
	ID            string               `yaml:"id"`
	Title         string               `yaml:"title"`
	Price         string               `yaml:"price,omitempty"`
	Links         []string             `yaml:"links,omitempty"`
	Note          string               `yaml:"note,omitempty"`
	PictureSource string               `yaml:"picture_source,omitempty"`
	Picture       string               `yaml:"picture,omitempty"`
	Assignment    *Assignment          `yaml:"assignment,omitempty"`
	Claims        []Claim              `yaml:"claims,omitempty"`
	Extra         map[string]yaml.Node `yaml:",inline"`
}

var safeID = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
var hexDigest = regexp.MustCompile(`^[0-9a-f]{64}$`)
var decimal = regexp.MustCompile(`^[0-9]+(\.[0-9]{1,2})?$`)

func validName(name string, min, max int, spaces bool) bool {
	n := strings.TrimSpace(name)
	if !utf8.ValidString(n) || utf8.RuneCountInString(n) < min || utf8.RuneCountInString(n) > max {
		return false
	}
	for _, r := range n {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !(spaces && r == ' ') {
			return false
		}
	}
	return true
}
func HTTPURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Hostname() != ""
}
func SafePicture(s string) bool {
	return s != "" && filepath.Base(s) == s && !strings.ContainsAny(s, `/\`) && !strings.HasPrefix(s, ".") && strings.HasSuffix(s, ".webp")
}
func DefaultAppearance() *Appearance { return &Appearance{120, 100, 16} }
func (a *Appearance) valid() bool {
	return a == nil || (a.BulbSize >= 80 && a.BulbSize <= 200 && a.SnowSize >= 50 && a.SnowSize <= 200 && a.SnowAmount >= 8 && a.SnowAmount <= 48)
}

// @impl CON-004-AC3 INT-007-AC1 INT-007-AC2
func Decode(raw []byte) (Household, error) {
	var h Household
	// Inspect presence separately: null/empty must never get missing-field fallback.
	var root yaml.Node
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return h, errors.New("invalid household YAML")
	}
	if len(root.Content) != 1 || root.Content[0].Kind != yaml.MappingNode {
		return h, errors.New("household must be a mapping")
	}
	var fields map[string]yaml.Node
	if err := root.Content[0].Decode(&fields); err != nil {
		return h, errors.New("invalid household mapping")
	}
	_, present := fields["language"]
	d := yaml.NewDecoder(bytes.NewReader(raw))
	d.KnownFields(true)
	if err := d.Decode(&h); err != nil {
		return Household{}, errors.New("invalid household fields")
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return Household{}, errors.New("expected one YAML document")
	}
	if !present {
		h.Language = "en-US"
	}
	if err := h.Validate(); err != nil {
		return Household{}, err
	}
	return h, nil
}
func Encode(h Household) ([]byte, error) {
	if err := h.Validate(); err != nil {
		return nil, err
	}
	raw, err := yaml.Marshal(h)
	if err != nil {
		return nil, errors.New("cannot encode household")
	}
	if _, err = Decode(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// @impl CON-004-AC3 WLS-007-AC2 WLS-008-AC2 WLS-010-AC2 ACC-007-AC2
func (h Household) Validate() error {
	if h.SchemaVersion != 2 {
		return errors.New("unsupported household schema; use explicit migration")
	}
	if !validName(h.Name, 3, 50, false) {
		return errors.New("invalid household name")
	}
	if h.Language != "en-US" && h.Language != "de-CH" {
		return errors.New("invalid household language")
	}
	if !h.Appearance.valid() {
		return errors.New("invalid appearance")
	}
	if err := ValidatePasswordHash(h.AdminPasswordHash); err != nil {
		return err
	}
	ids := map[string]bool{}
	addID := func(id string) bool {
		if !safeID.MatchString(id) || len(id) > 128 || ids[id] {
			return false
		}
		ids[id] = true
		return true
	}
	people := func(ps []Person) (map[string]bool, error) {
		out := map[string]bool{}
		for i, p := range ps {
			if !addID(p.ID) || !validName(p.Name, 3, 50, true) {
				return nil, errors.New("invalid roster identity/name")
			}
			for _, prior := range ps[:i] {
				if strings.EqualFold(strings.TrimSpace(p.Name), strings.TrimSpace(prior.Name)) {
					return nil, errors.New("duplicate roster name")
				}
			}
			out[p.ID] = true
		}
		return out, nil
	}
	users, err := people(h.Users)
	if err != nil {
		return err
	}
	recipients, err := people(h.Recipients)
	if err != nil {
		return err
	}
	lists := map[string]map[string]bool{}
	codes := map[string]bool{}
	for i, list := range h.Lists {
		name := strings.TrimSpace(list.Name)
		if !addID(list.ID) || !utf8.ValidString(name) || utf8.RuneCountInString(name) < 3 || utf8.RuneCountInString(name) > 100 {
			return errors.New("invalid list identity/name")
		}
		for _, prior := range h.Lists[:i] {
			if strings.EqualFold(name, strings.TrimSpace(prior.Name)) {
				return errors.New("duplicate list name")
			}
		}
		selected := map[string]bool{}
		for _, id := range list.RecipientIDs {
			if !recipients[id] || selected[id] {
				return errors.New("invalid/duplicate selected recipient")
			}
			selected[id] = true
		}
		lists[list.ID] = selected
		linked := map[string]bool{}
		for _, link := range list.UserLinks {
			if !users[link.UserID] || linked[link.UserID] || !hexDigest.MatchString(link.Code) || codes[link.Code] {
				return errors.New("invalid/duplicate user link")
			}
			linked[link.UserID] = true
			codes[link.Code] = true
		}
		if len(linked) != len(users) {
			return errors.New("each user requires a link for every list")
		}
	}
	pictures := map[string]bool{}
	for _, wish := range h.Wishes {
		if !addID(wish.ID) {
			return errors.New("invalid/duplicate wish identity")
		}
		if err := ValidateWish(wish); err != nil {
			return err
		}
		if wish.Assignment != nil {
			a := wish.Assignment
			if !lists[a.ListID][a.RecipientID] {
				return errors.New("assignment references missing list/selected recipient")
			}
		}
		seen := map[string]bool{}
		for _, c := range wish.Claims {
			if !users[c.UserID] || seen[c.UserID] {
				return errors.New("invalid/duplicate claimant identity")
			}
			seen[c.UserID] = true
		}
		if wish.Picture != "" {
			if pictures[wish.Picture] {
				return errors.New("picture filename shared by multiple wishes")
			}
			pictures[wish.Picture] = true
		}
	}
	seenSessions := map[string]bool{}
	for _, s := range h.Sessions {
		if !hexDigest.MatchString(s.TokenHash) || !hexDigest.MatchString(s.PasswordHashFingerprint) || seenSessions[s.TokenHash] {
			return errors.New("invalid/duplicate session record")
		}
		seenSessions[s.TokenHash] = true
	}
	return nil
}

// @impl WSH-001-AC6 WSH-001-AC7 WSH-001-AC8 WSH-001-AC9 CON-006
func ValidateWish(w Wish) error {
	for key := range w.Extra {
		switch key {
		case "id", "title", "price", "links", "note", "picture_source", "picture", "assignment", "claims":
			return errors.New("future wish field conflicts with schema")
		}
	}
	for _, claim := range w.Claims {
		if _, exists := claim.Extra["user_id"]; exists {
			return errors.New("future claim field conflicts with schema")
		}
	}

	title := strings.TrimSpace(w.Title)
	if !utf8.ValidString(title) || utf8.RuneCountInString(title) < 5 || utf8.RuneCountInString(title) > 150 {
		return errors.New("invalid wish title")
	}
	if w.Price != "" && !decimal.MatchString(w.Price) {
		return errors.New("invalid wish reference price")
	}
	for _, link := range w.Links {
		if !HTTPURL(link) {
			return errors.New("invalid shop URL")
		}
	}
	if w.PictureSource != "" && !HTTPURL(w.PictureSource) {
		return errors.New("invalid image source URL")
	}
	if w.Picture != "" && !SafePicture(w.Picture) {
		return errors.New("unsafe picture filename")
	}
	return nil
}
func validationError(context string, err error) error { return fmt.Errorf("%s: %w", context, err) }
