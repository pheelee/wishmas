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
	Picture            string     `yaml:"picture,omitempty"`
	PictureURL         string     `yaml:"-"`
	PictureSource      string     `yaml:"picture_source,omitempty"`
	PictureUnavailable bool       `yaml:"-"`
	Mine               bool       `yaml:"-"`
	Title              string     `yaml:"title"`
	Claims             []claim    `yaml:"claims,omitempty"`
	Price              string     `yaml:"price,omitempty"`
	Note               string     `yaml:"note,omitempty"`
	Links              []string   `yaml:"links,omitempty"`
	Accent             string     `yaml:"-"`
	Token              string     `yaml:"-"`
	Shops              []shopLink `yaml:"-"`
	LinksText          string     `yaml:"-"`
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
		// @impl WSH-005-AC6 WSH-005-AC7: status follows wish data,
		// independently of list position, browser sorting or claim ownership.
		card.Accent = "red"
		if completeWish(card) {
			card.Accent = "yellow"
			if len(card.Claims) > 0 {
				card.Accent = "green"
			}
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
	wishToken := r.PostForm.Get("wish")
	target := -1
	for i, c := range cards {
		if c.Token == wishToken {
			target = i
			break
		}
	}
	if wishToken != "" && target < 0 {
		a.renderError(w, 409, "This wish changed. Reload the wishlist before editing again.")
		return
	}
	draft := wishCard{Title: r.PostForm.Get("title"), Price: r.PostForm.Get("price"), Note: r.PostForm.Get("note"), Token: wishToken, LinksText: r.PostForm.Get("links"), PictureSource: strings.TrimSpace(r.PostForm.Get("picture_url"))}
	for _, line := range strings.Split(draft.LinksText, "\n") {
		if raw := strings.TrimSpace(line); raw != "" {
			draft.Links = append(draft.Links, raw)
		}
	}
	if err := validateWish(&draft); err != nil {
		a.wishSaveError(w, r, h, state, name, cards, draft, err)
		return
	}
	oldPicture, oldSource := "", ""
	if target >= 0 {
		oldPicture, oldSource = cards[target].Picture, cards[target].PictureSource
	}
	draft.Picture = oldPicture
	var pictureBytes []byte
	if draft.PictureSource == "" {
		draft.Picture = ""
	} else if draft.PictureSource != oldSource {
		// Network work must not block other visitors or code revocation.
		a.mu.Unlock()
		pictureBytes, err = a.fetchPicture(r.Context(), draft.PictureSource)
		a.mu.Lock()
		current, loadErr := a.load()
		if loadErr != nil || current == nil {
			a.renderError(w, 500, "Unable to read household data. Check the YAML files; no files were changed.")
			return
		}
		if current.ParentCode != r.URL.Query().Get("code") {
			a.language = "en-US"
			a.renderError(w, 404, "404 page not found")
			return
		}
		h = *current
		a.language = householdLocale(h.Language)
		if !a.parsePeopleForm(w, r, h) {
			return
		}
		state, loadErr = a.readPeopleState(h)
		if loadErr != nil {
			a.renderError(w, 500, "Unable to read wishlist data. Check the YAML files; no files were changed.")
			return
		}
		currentIndex := activePersonIndex(state.People, name)
		if currentIndex < 0 || state.People[currentIndex].Name != name {
			a.renderError(w, 404, "404 page not found")
			return
		}
		list = state.Lists[name]
		cards, _ = wishCards(list)
		if target >= 0 && (target >= len(cards) || cards[target].Token != wishToken) {
			a.renderError(w, 409, "This wish changed. Reload the wishlist before editing again.")
			return
		}
		if err != nil {
			a.wishSaveError(w, r, h, state, name, cards, draft, err)
			return
		}
		id, tokenErr := token()
		if tokenErr != nil {
			a.wishSaveError(w, r, h, state, name, cards, draft, errPictureSave)
			return
		}
		draft.Picture = id + ".webp"
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
			if key != "title" && key != "price" && key != "note" && key != "links" && key != "picture" && key != "picture_source" {
				kept = append(kept, node.Content[i], node.Content[i+1])
			}
		}
		node.Content = append(kept, fresh.Content...)
	}
	if err := savePictureWish(a.wishlistPath(h, state.Year, name), list, oldPicture, draft.Picture, pictureBytes); err != nil {
		a.wishSaveError(w, r, h, state, name, cards, draft, err)
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
		data.ClaimPerson = name
		data.Draft = cards[target]
		data.HomeLink = personURL(h.ParentCode, name, state.Year)
		a.render(w, 200, "delete-wish.html", data)
		return
	}
	// Remove the entire mapping, including attached claims. No notification side
	// effects are performed. Other wish mappings retain their contents and order.
	list.Wishes = append(list.Wishes[:target], list.Wishes[target+1:]...)
	if err := savePictureWish(a.wishlistPath(h, state.Year, name), list, cards[target].Picture, "", nil); err != nil {
		a.renderError(w, 500, pictureError(err))
		return
	}
	http.Redirect(w, r, personURL(h.ParentCode, name, state.Year), http.StatusSeeOther)
}

func (a *application) wishSaveError(w http.ResponseWriter, r *http.Request, h household, state peopleState, name string, cards []wishCard, draft wishCard, err error) {
	// Keep the submitted draft verbatim; normalization applies only to a successful save.
	draft.Title = r.PostForm.Get("title")
	draft.LinksText = r.PostForm.Get("links")
	draft.PictureSource = r.PostForm.Get("picture_url")
	data := a.peopleView(h, state, "Parent", h.ParentCode)
	data.PersonName, data.ClaimPerson, data.RenameName = name, name, name
	data.Wishes, data.WishError, data.Draft = cards, err.Error(), draft
	if draft.Token == "" {
		data.CreateDraft = draft
	}
	personalizeClaims(&data, browserClaimName(r))
	a.render(w, 400, "wishlist.html", data)
}
