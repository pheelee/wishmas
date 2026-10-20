package app

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"
)

type shopLink struct{ URL, Domain string }
type wishCard struct {
	Title     string     `yaml:"title"`
	Claims    []claim    `yaml:"claims,omitempty"`
	Price     string     `yaml:"price,omitempty"`
	Note      string     `yaml:"note,omitempty"`
	Links     []string   `yaml:"links,omitempty"`
	Token     string     `yaml:"-"`
	Shops     []shopLink `yaml:"-"`
	LinksText string     `yaml:"-"`
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

// @impl WSH-001-AC5 WSH-001-AC10 WSH-001-AC11
func wishCards(list wishlist) ([]wishCard, error) {
	cards := []wishCard{}
	raw, err := yaml.Marshal(list)
	if err != nil {
		return nil, err
	}
	revision := sha256.Sum256(raw)
	for i, node := range list.Wishes {
		if node.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("invalid wish")
		}
		var card wishCard
		if err := node.Decode(&card); err != nil {
			return nil, err
		}
		if err := validateWish(&card); err != nil {
			return nil, err
		}
		if err := validateClaims(card.Claims); err != nil {
			return nil, err
		}
		// Bind the position to the whole list snapshot so a repeated delete
		// cannot target an identical wish that shifted into the removed slot.
		card.Token = fmt.Sprintf("%d-%x", i, revision)
		cards = append(cards, card)
	}
	return cards, nil
}

// @impl WLS-001-AC3 WLS-001-AC4 WLS-001-AC6
// @impl WSH-001-AC1 WSH-001-AC3 WSH-001-AC4 WSH-001-AC6 WSH-002-AC1 WSH-002-AC2 WSH-002-AC3 WSH-002-AC5 WSH-002-AC7
func (a *application) saveWish(w http.ResponseWriter, r *http.Request, h household) {
	if !a.parsePeopleForm(w, r, h) {
		return
	}
	state, err := a.readPeopleState(h)
	if err != nil {
		a.renderError(w, 500, "Unable to read wishlist data. Check the YAML files; no files were changed.")
		return
	}
	index := activePersonIndex(state.People, strings.TrimSpace(r.PostForm.Get("person")))
	if index < 0 {
		a.renderError(w, 404, "404 page not found")
		return
	}
	name := state.People[index].Name
	list := state.Lists[name]
	cards, _ := wishCards(list)
	token := r.PostForm.Get("wish")
	target := -1
	for i, c := range cards {
		if c.Token == token {
			target = i
			break
		}
	}
	if token != "" && target < 0 {
		a.renderError(w, 409, "This wish changed. Reload the wishlist before editing again.")
		return
	}
	draft := wishCard{Title: r.PostForm.Get("title"), Price: r.PostForm.Get("price"), Note: r.PostForm.Get("note"), Token: token, LinksText: r.PostForm.Get("links")}
	for _, line := range strings.Split(draft.LinksText, "\n") {
		if raw := strings.TrimSpace(line); raw != "" {
			draft.Links = append(draft.Links, raw)
		}
	}
	if err := validateWish(&draft); err != nil {
		data := a.peopleView(h, state, "Parent", h.ParentCode)
		data.PersonName = name
		data.RenameName = name
		data.Wishes = cards
		data.WishError = err.Error()
		data.Draft = draft
		if draft.Token == "" {
			data.CreateDraft = draft
		}
		a.render(w, 400, "wishlist.html", data)
		return
	}
	var fresh yaml.Node
	if err := fresh.Encode(draft); err != nil {
		a.renderError(w, 500, "Unable to prepare wish.")
		return
	}
	if target < 0 {
		list.Wishes = append(list.Wishes, fresh)
	} else {
		// Replace only editable fields on the fresh disk node; keep claims, comments,
		// and any future fields. Omitting an optional field clears it.
		node := &list.Wishes[target]
		kept := []*yaml.Node{}
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i].Value
			if key != "title" && key != "price" && key != "note" && key != "links" {
				kept = append(kept, node.Content[i], node.Content[i+1])
			}
		}
		node.Content = append(kept, fresh.Content...)
	}
	if err := writeYAML(a.wishlistPath(h, state.Year, name), list, false); err != nil {
		a.renderError(w, 500, "Unable to save the wish.")
		return
	}
	http.Redirect(w, r, personURL(h.ParentCode, name, state.Year), http.StatusSeeOther)
}

// @impl WSH-002-AC1 WSH-002-AC2 WSH-002-AC4 WSH-002-AC5 WSH-002-AC6 WSH-002-AC7
func (a *application) deleteWish(w http.ResponseWriter, r *http.Request, h household) {
	if !a.parsePeopleForm(w, r, h) {
		return
	}
	state, err := a.readPeopleState(h)
	if err != nil {
		a.renderError(w, 500, "Unable to read wishlist data. Check the YAML files; no files were changed.")
		return
	}
	index := activePersonIndex(state.People, r.PostForm.Get("person"))
	if index < 0 {
		a.renderError(w, 404, "404 page not found")
		return
	}
	name := state.People[index].Name
	list := state.Lists[name]
	cards, _ := wishCards(list)
	target := -1
	for i, card := range cards {
		if card.Token == r.PostForm.Get("wish") {
			target = i
			break
		}
	}
	if target < 0 {
		a.renderError(w, 409, "This wish changed. Reload the wishlist before deleting it.")
		return
	}
	if r.PostForm.Get("confirm") != "yes" {
		data := a.peopleView(h, state, "Parent", h.ParentCode)
		data.PersonName = name
		data.Draft = cards[target]
		data.HomeLink = personURL(h.ParentCode, name, state.Year)
		a.render(w, 200, "delete-wish.html", data)
		return
	}
	// Remove the entire mapping, including attached claims. No notification side
	// effects are performed. Other wish mappings retain their contents and order.
	list.Wishes = append(list.Wishes[:target], list.Wishes[target+1:]...)
	if err := writeYAML(a.wishlistPath(h, state.Year, name), list, false); err != nil {
		a.renderError(w, 500, "Unable to delete the wish.")
		return
	}
	http.Redirect(w, r, personURL(h.ParentCode, name, state.Year), http.StatusSeeOther)
}
