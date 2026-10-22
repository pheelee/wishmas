package app

import (
	"fmt"
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
		return fmt.Errorf("Use 5–150 characters for the title.")
	}
	if w.Price != "" && !pricePattern.MatchString(w.Price) {
		return fmt.Errorf("Use a nonnegative price with a decimal point and at most two decimal places.")
	}
	w.Shops = nil
	for _, raw := range w.Links {
		u, err := url.Parse(raw)
		if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return fmt.Errorf("Shop links must be complete HTTP or HTTPS URLs.")
		}
		w.Shops = append(w.Shops, shopLink{URL: raw, Domain: u.Hostname()})
	}
	w.LinksText = strings.Join(w.Links, "\n")
	return nil
}
