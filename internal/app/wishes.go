package app

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

type shopLink struct{ URL, Domain string }
type wishCard struct {
	AssignmentLabel, AssignmentValue string     `yaml:"-"`
	Picture                          string     `yaml:"picture,omitempty"`
	PictureURL                       string     `yaml:"-"`
	PictureSource                    string     `yaml:"picture_source,omitempty"`
	PictureUnavailable               bool       `yaml:"-"`
	Mine                             bool       `yaml:"-"`
	Title                            string     `yaml:"title"`
	Claims                           []claim    `yaml:"claims,omitempty"`
	Price                            string     `yaml:"price,omitempty"`
	Note                             string     `yaml:"note,omitempty"`
	Links                            []string   `yaml:"links,omitempty"`
	Accent                           string     `yaml:"-"`
	Token                            string     `yaml:"-"`
	Shops                            []shopLink `yaml:"-"`
	LinksText                        string     `yaml:"-"`
}

var pricePattern = regexp.MustCompile(`^[0-9]+(\.[0-9]{1,2})?$`)

// @impl WSH-001-AC2 WSH-001-AC6 WSH-001-AC7 WSH-001-AC8 WSH-001-AC9
func validateWish(w *wishCard) error {
	w.Title = strings.TrimSpace(w.Title)
	if n := utf8.RuneCountInString(w.Title); n < 5 || n > 150 {
		return errors.New("E39")
	}
	if w.Price != "" && !pricePattern.MatchString(w.Price) {
		return errors.New("E40")
	}
	w.Shops = nil
	for _, raw := range w.Links {
		u, err := url.Parse(raw)
		if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return errors.New("E41")
		}
		w.Shops = append(w.Shops, shopLink{URL: raw, Domain: u.Hostname()})
	}
	w.LinksText = strings.Join(w.Links, "\n")
	return nil
}
