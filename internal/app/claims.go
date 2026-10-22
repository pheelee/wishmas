package app

import (
	"fmt"
	"go.yaml.in/yaml/v3"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type claim struct {
	UserID string `yaml:"-"`
	Name   string `yaml:"name"`
}

// @impl CON-004 CLM-001-AC4 CLM-001-AC6 CLM-002-AC3
// Validate a presentation copy without normalizing or rewriting saved YAML.
func validateClaims(claims []claim) error {
	for i := range claims {
		claims[i].Name = strings.TrimSpace(claims[i].Name)
		if !validName(claims[i].Name) {
			return fmt.Errorf("invalid claimant name")
		}
		for _, prior := range claims[:i] {
			if strings.EqualFold(prior.Name, claims[i].Name) {
				return fmt.Errorf("duplicate claimant name")
			}
		}
	}
	return nil
}

const claimNameCookie = "wishmas_claim_name"

// @impl CLM-001-AC1 CLM-001-AC4 CLM-001-AC6 CLM-001-AC7
func browserClaimName(r *http.Request) string {
	cookie, err := r.Cookie(claimNameCookie)
	if err != nil {
		return ""
	}
	name, err := url.QueryUnescape(cookie.Value)
	name = strings.TrimSpace(name)
	if err != nil || !validName(name) {
		return ""
	}
	return name
}
func rememberClaimName(w http.ResponseWriter, r *http.Request, name string) {
	http.SetCookie(w, &http.Cookie{Name: claimNameCookie, Value: url.QueryEscape(name), Path: "/", MaxAge: 365 * 24 * 60 * 60, HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteLaxMode})
}

// @impl CLM-003-AC3 CLM-003-AC4
func personalizeClaims(data *view, name string) {
	data.ViewerName = name
	for i := range data.Wishes {
		data.Wishes[i].Mine = false
		for _, claim := range data.Wishes[i].Claims {
			if name != "" && strings.EqualFold(name, claim.Name) {
				data.Wishes[i].Mine = true
				break
			}
		}
	}
}

func (a *application) claimView(h household, state peopleState, role, code, person, viewer string) view {
	data := a.peopleView(h, state, role, code)
	if person != "" {
		data.PersonName = person
		data.ClaimPerson = person
		data.RenameName = person
		data.Wishes, _ = wishCards(state.Lists[person])
	}
	personalizeClaims(&data, viewer)
	return data
}
func (a *application) renderClaimPage(w http.ResponseWriter, status int, data view) {
	page := "home.html"
	if data.PersonName != "" {
		page = "wishlist.html"
	}
	a.render(w, status, page, data)
}

// @impl CLM-001-AC2 CLM-001-AC3 CLM-001-AC6
// Changing a browser name never rewrites any claim or household file.
func (a *application) changeClaimName(w http.ResponseWriter, r *http.Request, h household, role, code string) {
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	if err := r.ParseForm(); err != nil {
		a.renderError(w, 400, "Invalid name form")
		return
	}
	year := a.currentYear(h)
	if raw := r.PostForm.Get("year"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 || strconv.Itoa(value) != raw || (role != "Parent" && value != year) {
			a.renderError(w, 404, "404 page not found")
			return
		}
		year = value
	}
	state, err := a.readPeopleYear(h, year)
	if err != nil {
		a.renderError(w, 500, "Unable to read wishlist data. Check the YAML files; no files were changed.")
		return
	}
	person := ""
	if raw := r.PostForm.Get("person"); raw != "" {
		index := activePersonIndex(state.People, raw)
		if index < 0 {
			a.renderError(w, 404, "404 page not found")
			return
		}
		person = state.People[index].Name
		if _, exists := state.Lists[person]; !exists {
			a.renderError(w, 404, "404 page not found")
			return
		}
	}
	name := strings.TrimSpace(r.PostForm.Get("name"))
	if !validName(name) {
		data := a.claimView(h, state, role, code, person, browserClaimName(r))
		data.NameDraft = name
		data.NameError = "Use 3–50 Unicode letters and digits for your claim name."
		data.EditClaimName = true
		a.renderClaimPage(w, 400, data)
		return
	}
	rememberClaimName(w, r, name)
	destination := accessLink(code)
	if year != a.currentYear(h) {
		destination += "&year=" + strconv.Itoa(year)
	}
	if person != "" {
		destination = personURL(code, person, year)
	}
	http.Redirect(w, r, destination, http.StatusSeeOther)
}

// @impl CLM-001-AC1 CLM-001-AC5 CLM-001-AC6 CLM-002-AC1 CLM-002-AC2 CLM-002-AC3 CLM-002-AC4 CLM-002-AC5 CLM-002-AC6
// @impl CLM-003-AC6 CLM-004-AC1 CLM-004-AC2 CLM-004-AC3 CLM-004-AC4 CLM-004-AC5 CLM-004-AC6 CLM-004-AC7
func (a *application) toggleClaim(w http.ResponseWriter, r *http.Request, h household, role, code string) {
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
	person := state.People[index].Name
	viewer := browserClaimName(r)
	data := a.claimView(h, state, role, code, person, viewer)
	fail := func(status int, message string) { data.ClaimError = message; a.renderClaimPage(w, status, data) }
	target := -1
	for i, card := range data.Wishes {
		if card.Token == r.PostForm.Get("wish") {
			target = i
			break
		}
	}
	mode := r.PostForm.Get("mode")
	if viewer == "" && mode == "add" {
		if _, supplied := r.PostForm["name"]; supplied {
			name := strings.TrimSpace(r.PostForm.Get("name"))
			if !validName(name) {
				data.NamePrompt = true
				data.PendingWish = r.PostForm.Get("wish")
				data.NameDraft = name
				data.NameError = "Use 3–50 Unicode letters and digits for your claim name."
				a.renderClaimPage(w, 400, data)
				return
			}
			viewer = name
			rememberClaimName(w, r, viewer)
			personalizeClaims(&data, viewer)
		} else if target >= 0 {
			data.NamePrompt = true
			data.PendingWish = data.Wishes[target].Token
			a.renderClaimPage(w, 200, data)
			return
		}
	}
	if target < 0 {
		fail(409, "This wishlist changed. Claim status has been refreshed; please try again.")
		return
	}
	card := data.Wishes[target]
	remove := -1
	switch mode {
	case "add":
		if viewer == "" {
			fail(400, "Choose a name before claiming a wish.")
			return
		}
		for _, claim := range card.Claims {
			if strings.EqualFold(viewer, claim.Name) {
				fail(409, "This wish is already claimed by you. Claim status has been refreshed.")
				return
			}
		}
	case "remove":
		for i, claim := range card.Claims {
			if viewer != "" && strings.EqualFold(viewer, claim.Name) {
				remove = i
				break
			}
		}
		if remove < 0 {
			fail(403, "You can only remove a claim under your current name. Claim status has been refreshed.")
			return
		}
	case "remove-other":
		if role != "Parent" {
			fail(403, "Only parents can remove another person's claim. Claim status has been refreshed.")
			return
		}
		name := strings.TrimSpace(r.PostForm.Get("claimant"))
		for i, claim := range card.Claims {
			if strings.EqualFold(name, claim.Name) {
				remove = i
				break
			}
		}
		if remove < 0 {
			fail(409, "That claim changed. Claim status has been refreshed.")
			return
		}
	default:
		fail(400, "Invalid claim action. Claim status has been refreshed.")
		return
	}
	list := state.Lists[person]
	// Resolve aliases in the copy before changing a sequence; shared YAML aliases
	// must not mutate another wish or become dangling when a claim is removed.
	for i := range list.Wishes {
		resolved, err := copyClaimNode(&list.Wishes[i], map[*yaml.Node]bool{})
		if err != nil {
			fail(500, "Unable to save the claim. Check the YAML aliases; existing claims are unchanged.")
			return
		}
		list.Wishes[i] = resolved
	}
	node := &list.Wishes[target]
	var sequence *yaml.Node
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == "claims" {
			sequence = node.Content[i+1]
			break
		}
	}
	if sequence == nil {
		sequence = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "claims"}, sequence)
	}
	if mode == "add" {
		// A null/omitted claims field represents an empty sequence.
		sequence.Kind = yaml.SequenceNode
		sequence.Tag = "!!seq"
		sequence.Value = ""
		var newClaim yaml.Node
		if err := newClaim.Encode(claim{Name: viewer}); err != nil {
			fail(500, "Unable to save the claim. Existing claims are unchanged.")
			return
		}
		sequence.Content = append(sequence.Content, &newClaim)
	} else {
		sequence.Content = append(sequence.Content[:remove], sequence.Content[remove+1:]...)
	}
	if err := writeYAML(a.wishlistPath(h, state.Year, person), list, false); err != nil {
		fail(500, "Unable to save the claim. Existing claims are unchanged.")
		return
	}
	http.Redirect(w, r, personURL(code, person, state.Year), http.StatusSeeOther)
}

func copyClaimNode(node *yaml.Node, active map[*yaml.Node]bool) (yaml.Node, error) {
	if active[node] {
		return yaml.Node{}, fmt.Errorf("cyclic YAML alias")
	}
	active[node] = true
	defer delete(active, node)
	if node.Kind == yaml.AliasNode {
		return copyClaimNode(node.Alias, active)
	}
	result := *node
	result.Anchor = ""
	result.Alias = nil
	result.Content = nil
	for _, child := range node.Content {
		copy, err := copyClaimNode(child, active)
		if err != nil {
			return yaml.Node{}, err
		}
		result.Content = append(result.Content, &copy)
	}
	return result, nil
}
