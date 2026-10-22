package app

import (
	"embed"
	"html/template"
	"unicode"
	"unicode/utf8"
)

//go:embed web/*
var web embed.FS
var pages = template.Must(template.New("").Funcs(template.FuncMap{"t": func(string, ...any) string { return "" }, "price": func(string) string { return "" }}).ParseFS(web, "web/*.html"))

// Shared presentation only; all household persistence uses internal/household.
type view struct {
	Locale, Language, ReferrerPolicy string
	ViewerName, ClaimError           string
	Appearance                       *appearance
	Bulbs                            []bulbDecoration
	LeftSnow, RightSnow              []snowDecoration
	Wishes                           []wishCard
	Draft, CreateDraft               wishCard
	WishError                        string
	Name, Error, PersonName          string
}

func validName(name string) bool {
	size := utf8.RuneCountInString(name)
	if size < 3 || size > 50 {
		return false
	}
	for _, char := range name {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}
