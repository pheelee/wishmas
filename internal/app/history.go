package app

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Numeric year directories are discovered from the household's saved data.
func (a *application) yearDirectories(h household) ([]int, error) {
	entries, err := os.ReadDir(a.householdDir(h))
	if err != nil {
		return nil, err
	}
	years := []int{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		year, err := strconv.Atoi(entry.Name())
		if err == nil && year > 0 && strconv.Itoa(year) == entry.Name() {
			years = append(years, year)
		}
	}
	sort.Ints(years)
	return years, nil
}

// @impl WLS-002-AC3 WLS-004-AC2 ACC-002-AC2
func (a *application) availableYears(h household, people []person) ([]int, error) {
	saved, err := a.yearDirectories(h)
	if err != nil {
		return nil, err
	}
	found := map[int]bool{a.currentYear(h): true}
	for _, year := range saved {
		for _, p := range people {
			if p.Archived {
				continue
			}
			_, err := os.Stat(a.wishlistPath(h, year, p.Name))
			if err == nil {
				found[year] = true
				break
			}
			if !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
		}
	}
	years := make([]int, 0, len(found))
	for year := range found {
		years = append(years, year)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(years)))
	return years, nil
}

// @impl WLS-001-AC6 WLS-002-AC4
// A stale form rendered for another year cannot mutate a historical context.
func (a *application) parsePeopleForm(w http.ResponseWriter, r *http.Request, h household) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	if err := r.ParseForm(); err != nil {
		a.renderError(w, 400, "Invalid person form")
		return false
	}
	current := strconv.Itoa(a.currentYear(h))
	for _, raw := range []string{r.URL.Query().Get("year"), r.PostForm.Get("year")} {
		if raw != "" && raw != current {
			a.renderError(w, 404, "404 page not found")
			return false
		}
	}
	return true
}

func activePersonIndex(people []person, name string) int {
	name = strings.TrimSpace(name)
	for i, p := range people {
		if !p.Archived && strings.EqualFold(p.Name, name) {
			return i
		}
	}
	return -1
}
func personURL(code, name string, year int) string {
	return "/wishlist?" + url.Values{"code": {code}, "person": {name}, "year": {strconv.Itoa(year)}}.Encode()
}

// @impl WLS-003-AC1 WLS-003-AC2 WLS-003-AC3 WLS-003-AC4 WLS-003-AC5 WLS-003-AC6
func (a *application) renamePerson(w http.ResponseWriter, r *http.Request, h household) {
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
	oldName := state.People[index].Name
	name := strings.TrimSpace(r.PostForm.Get("name"))
	problem := ""
	if !validName(name) {
		problem = "Use 3–50 Unicode letters and digits for the person's name."
	}
	for i, p := range state.People {
		if i != index && strings.EqualFold(name, p.Name) {
			problem = "That name is already in use, including archived people."
			break
		}
	}
	if problem != "" {
		data := a.peopleView(h, state, "Parent", h.ParentCode)
		data.PersonName = oldName
		data.ClaimPerson = oldName
		data.RenameName = name
		data.Error = problem
		data.Titles, _ = wishTitles(state.Lists[oldName])
		data.Wishes, _ = wishCards(state.Lists[oldName])
		personalizeClaims(&data, browserClaimName(r))
		a.render(w, 400, "wishlist.html", data)
		return
	}
	if name != oldName {
		moves, err := a.planRename(h, oldName, name)
		if err != nil {
			a.renderError(w, 500, "Unable to rename the person. Check the household files; no files were changed.")
			return
		}
		state.People[index].Name = name
		if err := a.applyRename(h, moves, state.People); err != nil {
			a.renderError(w, 500, "Unable to rename the person. Check the household files.")
			return
		}
	}
	http.Redirect(w, r, personURL(h.ParentCode, name, state.Year), http.StatusSeeOther)
}

type directoryMove struct{ from, to string }

// @impl WLS-003-AC3 CON-004
// Plan and validate every affected directory before moving any. Raw list files
// (including claims and comments) are moved, never re-encoded or rewritten.
func (a *application) planRename(h household, oldName, newName string) ([]directoryMove, error) {
	years, err := a.yearDirectories(h)
	if err != nil {
		return nil, err
	}
	moves := []directoryMove{}
	for _, year := range years {
		root := filepath.Join(a.householdDir(h), strconv.Itoa(year))
		events, err := os.ReadDir(root)
		if err != nil {
			return nil, err
		}
		for _, event := range events {
			if !event.IsDir() {
				continue
			}
			from := filepath.Join(root, event.Name(), oldName)
			info, err := os.Stat(from)
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return nil, err
			}
			if !info.IsDir() {
				return nil, fmt.Errorf("invalid person directory")
			}
			if event.Name() == "christmas" {
				if _, err := a.readWishlist(h, year, oldName); err != nil && !errors.Is(err, os.ErrNotExist) {
					return nil, err
				}
			}
			to := filepath.Join(root, event.Name(), newName)
			if _, err := os.Lstat(to); err == nil {
				return nil, fmt.Errorf("destination already exists")
			} else if !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
			moves = append(moves, directoryMove{from: from, to: to})
		}
	}
	return moves, nil
}

func rollbackMoves(moves []directoryMove) error {
	var result error
	for i := len(moves) - 1; i >= 0; i-- {
		if err := os.Rename(moves[i].to, moves[i].from); err != nil {
			result = errors.Join(result, err)
		}
	}
	return result
}

func (a *application) applyRename(h household, moves []directoryMove, people []person) error {
	for i, move := range moves {
		if err := os.Rename(move.from, move.to); err != nil {
			return errors.Join(err, rollbackMoves(moves[:i]))
		}
	}
	if err := writeYAML(a.peoplePath(h), roster{People: people}, false); err != nil {
		return errors.Join(err, rollbackMoves(moves))
	}
	return nil
}

// @impl WLS-004-AC1 WLS-004-AC2 WLS-004-AC3 WLS-004-AC4 WLS-003-AC4
func (a *application) archivePerson(w http.ResponseWriter, r *http.Request, h household) {
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
	state.People[index].Archived = true
	if err := writeYAML(a.peoplePath(h), roster{People: state.People}, false); err != nil {
		a.renderError(w, 500, "Unable to archive the person.")
		return
	}
	http.Redirect(w, r, accessLink(h.ParentCode), http.StatusSeeOther)
}
