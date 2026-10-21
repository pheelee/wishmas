package app

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

type person struct {
	Name     string `yaml:"name"`
	Archived bool   `yaml:"archived,omitempty"`
}
type roster struct {
	People []person `yaml:"people"`
}

// Keep raw wish mappings so editing details preserves claims and future fields.
type wishlist struct {
	Wishes []yaml.Node `yaml:"wishes"`
}
type personCard struct {
	Name, URL, TotalPrice          string
	ClaimedWishes, UnclaimedWishes int
}
type peopleState struct {
	People  []person
	Lists   map[string]wishlist
	Missing []person
	Year    int
	Years   []int
}

func (a *application) householdDir(h household) string {
	if h.dir != "" {
		return h.dir
	}
	return filepath.Join(a.dataDir, "households", h.Name)
}
func (a *application) peoplePath(h household) string {
	return filepath.Join(a.householdDir(h), "people.yaml")
}
func (a *application) wishlistPath(h household, year int, name string) string {
	return filepath.Join(a.householdDir(h), strconv.Itoa(year), "christmas", name, "wishlist.yaml")
}

func readYAML(path string, target any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("expected one YAML document")
	}
	return nil
}

// @impl WLS-001-AC5 WLS-001-AC7 WLS-003-AC1 WLS-003-AC2 WLS-003-AC4 WLS-003-AC5 WLS-003-AC6
func (a *application) readPeople(h household) ([]person, error) {
	var data roster
	if err := readYAML(a.peoplePath(h), &data); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []person{}, nil
		}
		return nil, err
	}
	for i := range data.People {
		p := &data.People[i]
		p.Name = strings.TrimSpace(p.Name)
		if !validName(p.Name) {
			return nil, fmt.Errorf("invalid person name")
		}
		for _, prior := range data.People[:i] {
			if strings.EqualFold(p.Name, prior.Name) {
				return nil, fmt.Errorf("duplicate person name")
			}
		}
	}
	return data.People, nil
}

func (a *application) readWishlist(h household, year int, name string) (wishlist, error) {
	var list wishlist
	if err := readYAML(a.wishlistPath(h, year, name), &list); err != nil {
		return list, err
	}
	if _, err := wishTitles(list); err != nil {
		return list, err
	}
	return list, nil
}
func wishTitles(list wishlist) ([]string, error) {
	cards, err := wishCards(list)
	if err != nil {
		return nil, err
	}
	titles := []string{}
	for _, card := range cards {
		titles = append(titles, card.Title)
	}
	return titles, nil
}

// @impl WLS-001-AC5 WLS-001-AC7
// Validate all active current-year data before creating any missing files.
func (a *application) readPeopleState(h household) (peopleState, error) {
	return a.readPeopleYear(h, a.currentYear(h))
}

func (a *application) readPeopleYear(h household, year int) (peopleState, error) {
	state := peopleState{Year: year, Lists: map[string]wishlist{}}
	people, err := a.readPeople(h)
	if err != nil {
		return state, err
	}
	state.People = people
	state.Years, err = a.availableYears(h, people)
	if err != nil {
		return state, err
	}
	for _, p := range people {
		if p.Archived {
			continue
		}
		list, err := a.readWishlist(h, year, p.Name)
		if errors.Is(err, os.ErrNotExist) {
			if year != a.currentYear(h) {
				continue
			}
			state.Missing = append(state.Missing, p)
			list = wishlist{Wishes: []yaml.Node{}}
		} else if err != nil {
			return state, err
		}
		state.Lists[p.Name] = list
	}
	return state, nil
}

// @impl WLS-002-AC1 WLS-002-AC2 WLS-004-AC4 WLS-004-AC7
func (a *application) ensureCurrentLists(h household, state peopleState) error {
	for _, p := range state.Missing {
		if err := writeYAML(a.wishlistPath(h, state.Year, p.Name), wishlist{Wishes: []yaml.Node{}}, true); err != nil {
			return err
		}
	}
	return nil
}

// @impl WLS-001-AC1 WLS-001-AC2 WLS-001-AC6 WLS-003-AC1 WLS-003-AC2 WLS-003-AC4 WLS-003-AC5 WLS-003-AC6
func (a *application) addPerson(w http.ResponseWriter, r *http.Request, h household) {
	if !a.parsePeopleForm(w, r, h) {
		return
	}
	state, err := a.readPeopleState(h)
	if err != nil {
		a.renderError(w, 500, "Unable to read wishlist data. Check the YAML files; no files were changed.")
		return
	}
	name := strings.TrimSpace(r.PostForm.Get("name"))
	problem := ""
	if !validName(name) {
		problem = "Use 3–50 Unicode letters and digits for the person's name."
	}
	for _, p := range state.People {
		if strings.EqualFold(name, p.Name) {
			problem = "That name is already in use, including archived people."
			break
		}
	}
	if problem != "" {
		data := a.peopleView(h, state, "Parent", h.ParentCode)
		data.Error = problem
		data.PersonName = name
		personalizeClaims(&data, browserClaimName(r))
		a.render(w, 400, "home.html", data)
		return
	}
	file := a.wishlistPath(h, state.Year, name)
	// Exclusive creation protects manual data or files added after validation.
	if err := writeYAML(file, wishlist{Wishes: []yaml.Node{}}, true); err != nil {
		a.renderError(w, 500, "Unable to create the wishlist. Existing files were not overwritten.")
		return
	}
	state.People = append(state.People, person{Name: name})
	if err := writeYAML(a.peoplePath(h), roster{People: state.People}, false); err != nil {
		_ = os.Remove(file) // Roll back only the empty file created by this request.
		a.renderError(w, 500, "Unable to save the new person.")
		return
	}
	http.Redirect(w, r, accessLink(h.ParentCode), http.StatusSeeOther)
}

// @impl WLS-001-AC8 ACC-002-AC3
func (a *application) peopleView(h household, state peopleState, role, code string) view {
	data := view{Name: h.Name, Timezone: h.Timezone, Role: role, Year: state.Year, Code: code, HomeLink: accessLink(code), IsCurrent: state.Year == a.currentYear(h)}
	if role == "Parent" {
		data.Years = state.Years
	}
	if !data.IsCurrent {
		data.HomeLink += "&year=" + strconv.Itoa(state.Year)
	}
	if role == "Parent" {
		data.ParentLink = accessLink(h.ParentCode)
		data.FamilyLink = accessLink(h.FamilyCode)
	}
	for _, p := range state.People {
		if p.Archived {
			continue
		}
		if _, exists := state.Lists[p.Name]; !exists {
			continue
		}
		query := url.Values{"code": {code}, "person": {p.Name}, "year": {strconv.Itoa(state.Year)}}
		// @impl WLS-006-AC1 WLS-006-AC2 WLS-006-AC3 WLS-006-AC6
		// readPeopleYear has already validated these wish mappings.
		cards, _ := wishCards(state.Lists[p.Name])
		total, claimed, unclaimed := summarizeWishes(cards)
		data.People = append(data.People, personCard{Name: p.Name, URL: "/wishlist?" + query.Encode(), TotalPrice: total, ClaimedWishes: claimed, UnclaimedWishes: unclaimed})
	}
	return data
}

// @impl WLS-001-AC5 WLS-001-AC7 WLS-002-AC1 WLS-002-AC2 ACC-002-AC3 ACC-002-AC4
func (a *application) showPeople(w http.ResponseWriter, r *http.Request, h household, role, code string) {
	year := a.currentYear(h)
	if raw := r.URL.Query().Get("year"); raw != "" {
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
	available := false
	for _, saved := range state.Years {
		if saved == year {
			available = true
			break
		}
	}
	if !available {
		a.renderError(w, 404, "404 page not found")
		return
	}
	data := a.peopleView(h, state, role, code)
	page := "home.html"
	if r.URL.Path == "/wishlist" {
		name := strings.TrimSpace(r.URL.Query().Get("person"))
		for _, p := range state.People {
			if !p.Archived && strings.EqualFold(p.Name, name) {
				if _, exists := state.Lists[p.Name]; exists {
					data.PersonName = p.Name
				}
				break
			}
		}
		if data.PersonName == "" {
			a.renderError(w, 404, "404 page not found")
			return
		}
		data.ClaimPerson = data.PersonName
		data.RenameName = data.PersonName
		data.Titles, _ = wishTitles(state.Lists[data.PersonName])
		data.Wishes, _ = wishCards(state.Lists[data.PersonName])
		page = "wishlist.html"
	}
	current := state
	if !data.IsCurrent {
		current, err = a.readPeopleState(h)
		if err != nil {
			a.renderError(w, 500, "Unable to read current-year data. Check the YAML files; no files were changed.")
			return
		}
	}
	if err := a.ensureCurrentLists(h, current); err != nil {
		a.renderError(w, 500, "Unable to create current-year wishlists. Existing files were not overwritten.")
		return
	}
	personalizeClaims(&data, browserClaimName(r))
	a.render(w, 200, page, data)
}

// Write via a synced temporary file. New list files are linked exclusively so a
// concurrent manual edit cannot be overwritten between validation and writing.
func writeYAML(path string, value any, exclusive bool) error {
	raw, err := yaml.Marshal(value)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".wishmas-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(raw); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if exclusive {
		return os.Link(file.Name(), path)
	}
	return os.Rename(file.Name(), path)
}
