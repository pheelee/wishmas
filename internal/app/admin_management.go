package app

import (
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	data "github.com/pheelee/wishmas/internal/household"
)

type managementRow struct {
	ID, Name, RenameKey, SaveKey, DeleteKey, QuestionKey, WarningKey string
	Wishes, Claims                                                   int
}
type listLinkView struct{ UserID, Name, Code string }
type recipientReadView struct {
	ID, Name, URL, TotalPrice      string
	Cards                          []wishCard
	ClaimedWishes, UnclaimedWishes int
	Wishes                         []data.Wish
}

func rosterRows(h data.Household, kind string) []managementRow {
	var out []managementRow
	appendRow := func(id, name string) {
		row := managementRow{ID: id, Name: name}
		switch kind {
		case "list":
			row.RenameKey = "HR33"
			row.SaveKey = "HR34"
			row.DeleteKey = "HR35"
			row.QuestionKey = "HR36"
			row.WarningKey = "HR37"
		case "recipient":
			row.RenameKey = "HR45"
			row.SaveKey = "HR46"
			row.DeleteKey = "HR47"
			row.QuestionKey = "HR48"
			row.WarningKey = "HR49"
		case "user":
			row.RenameKey = "HR60"
			row.SaveKey = "HR61"
			row.DeleteKey = "HR62"
			row.QuestionKey = "HR63"
			row.WarningKey = "HR64"
		}
		for _, w := range h.Wishes {
			if kind == "user" {
				for _, c := range w.Claims {
					if c.UserID == id {
						row.Claims++
					}
				}
			} else if w.Assignment != nil && ((kind == "list" && w.Assignment.ListID == id) || (kind == "recipient" && w.Assignment.RecipientID == id)) {
				row.Wishes++
				row.Claims += len(w.Claims)
			}
		}
		out = append(out, row)
	}
	switch kind {
	case "list":
		for _, p := range h.Lists {
			appendRow(p.ID, p.Name)
		}
	case "user":
		for _, p := range h.Users {
			appendRow(p.ID, p.Name)
		}
	case "recipient":
		for _, p := range h.Recipients {
			appendRow(p.ID, p.Name)
		}
	}
	return out
}

// @impl WLS-007-AC8 WLS-008-AC3 ACC-007-AC9: render household added order, independent of selection order.
func populateManagement(v *adminView, h data.Household, revision, listID string) bool {
	v.Revision = revision
	v.ListRows = rosterRows(h, "list")
	v.RecipientRows = rosterRows(h, "recipient")
	v.UserRows = rosterRows(h, "user")
	v.CurrentList = nil
	v.SelectedRecipients = nil
	v.AvailableRecipients = nil
	v.ListLinks = nil
	v.ReadRecipients = nil
	if listID == "" {
		return true
	}
	for _, list := range h.Lists {
		if list.ID != listID {
			continue
		}
		selected := list
		v.CurrentList = &selected
		for _, p := range h.Recipients {
			if slices.Contains(list.RecipientIDs, p.ID) {
				row := managementRow{ID: p.ID, Name: p.Name}
				read := recipientReadView{ID: p.ID, Name: p.Name}
				for _, w := range h.Wishes {
					if w.Assignment != nil && w.Assignment.ListID == list.ID && w.Assignment.RecipientID == p.ID {
						row.Wishes++
						row.Claims += len(w.Claims)
						read.Wishes = append(read.Wishes, w)
					}
				}
				v.SelectedRecipients = append(v.SelectedRecipients, row)
				v.ReadRecipients = append(v.ReadRecipients, read)
			} else {
				v.AvailableRecipients = append(v.AvailableRecipients, p)
			}
		}
		for _, u := range h.Users {
			for _, link := range list.UserLinks {
				if link.UserID == u.ID {
					v.ListLinks = append(v.ListLinks, listLinkView{u.ID, u.Name, link.Code})
				}
			}
		}
		return true
	}
	return false
}

// @impl WLS-007-AC1 WLS-007-AC2 WLS-008-AC2 ACC-007-AC2
func managementName(h data.Household, kind, id, name string) string {
	min, max := 3, 50
	bad, duplicate := "HR55", "HR56"
	if kind == "user" {
		bad, duplicate = "HR66", "HR67"
	}
	if kind == "list" {
		max = 100
		bad, duplicate = "HR40", "HR41"
	}
	if !utf8.ValidString(name) || utf8.RuneCountInString(name) < min || utf8.RuneCountInString(name) > max {
		return bad
	}
	if kind != "list" {
		for _, r := range name {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != ' ' {
				return bad
			}
		}
	}
	for _, row := range rosterRows(h, kind) {
		if row.ID != id && strings.EqualFold(strings.TrimSpace(row.Name), name) {
			return duplicate
		}
	}
	return ""
}

// @impl WLS-007-AC3 WLS-007-AC5 WLS-007-AC6 WLS-007-AC7 WLS-008-AC3 WLS-008-AC4 WLS-008-AC5 WLS-009-AC1 WLS-009-AC2 WLS-009-AC3 ACC-007-AC1 ACC-007-AC3 ACC-007-AC5 ACC-007-AC6 ACC-007-AC7 ACC-007-AC8
// @impl CLM-008-AC4 CLM-008-AC5 CLM-008-AC6
// All related changes are applied to the repository snapshot, validated, then saved atomically.
func mutateManagement(h *data.Household, kind, action, id, name, recipient, user string) error {
	missing := errors.New("HR111")
	if kind == "selection" {
		for i := range h.Lists {
			l := &h.Lists[i]
			if l.ID != id {
				continue
			}
			if !slices.ContainsFunc(h.Recipients, func(p data.Person) bool { return p.ID == recipient }) {
				return missing
			}
			if action == "add" {
				if slices.Contains(l.RecipientIDs, recipient) {
					return errors.New("HR57")
				}
				l.RecipientIDs = append(l.RecipientIDs, recipient)
				return nil
			}
			if action != "remove" || !slices.Contains(l.RecipientIDs, recipient) {
				return missing
			}
			l.RecipientIDs = slices.DeleteFunc(l.RecipientIDs, func(s string) bool { return s == recipient })
			clearAssignments(h, func(a data.Assignment) bool { return a.ListID == id && a.RecipientID == recipient })
			return nil
		}
		return missing
	}
	if kind == "link" && action == "regenerate" {
		for i := range h.Lists {
			if h.Lists[i].ID != id {
				continue
			}
			for j := range h.Lists[i].UserLinks {
				link := &h.Lists[i].UserLinks[j]
				if link.UserID == user {
					token, err := data.NewToken()
					if err != nil {
						return err
					}
					link.Code = token
					return nil
				}
			}
		}
		return missing
	}
	if kind != "list" && kind != "recipient" && kind != "user" {
		return missing
	}
	if action == "create" || action == "rename" {
		if key := managementName(*h, kind, id, name); key != "" {
			return errors.New(key)
		}
	}
	if action == "create" {
		generated, err := data.NewID()
		if err != nil {
			return err
		}
		switch kind {
		case "list":
			l := data.List{ID: generated, Name: name}
			for _, u := range h.Users {
				code, err := data.NewToken()
				if err != nil {
					return err
				}
				l.UserLinks = append(l.UserLinks, data.UserLink{UserID: u.ID, Code: code})
			}
			h.Lists = append(h.Lists, l)
		case "recipient":
			h.Recipients = append(h.Recipients, data.Person{ID: generated, Name: name})
		case "user":
			h.Users = append(h.Users, data.Person{ID: generated, Name: name})
			for i := range h.Lists {
				code, err := data.NewToken()
				if err != nil {
					return err
				}
				h.Lists[i].UserLinks = append(h.Lists[i].UserLinks, data.UserLink{UserID: generated, Code: code})
			}
		}
		return nil
	}
	if action != "rename" && action != "delete" {
		return missing
	}
	if kind == "list" {
		for i := range h.Lists {
			if h.Lists[i].ID != id {
				continue
			}
			if action == "rename" {
				h.Lists[i].Name = name
			} else {
				h.Lists = slices.Delete(h.Lists, i, i+1)
				clearAssignments(h, func(a data.Assignment) bool { return a.ListID == id })
			}
			return nil
		}
		return missing
	}
	people := &h.Recipients
	if kind == "user" {
		people = &h.Users
	}
	for i := range *people {
		if (*people)[i].ID != id {
			continue
		}
		if action == "rename" {
			(*people)[i].Name = name
			return nil
		}
		*people = slices.Delete(*people, i, i+1)
		if kind == "recipient" {
			for j := range h.Lists {
				h.Lists[j].RecipientIDs = slices.DeleteFunc(h.Lists[j].RecipientIDs, func(s string) bool { return s == id })
			}
			clearAssignments(h, func(a data.Assignment) bool { return a.RecipientID == id })
		} else {
			for j := range h.Lists {
				h.Lists[j].UserLinks = slices.DeleteFunc(h.Lists[j].UserLinks, func(l data.UserLink) bool { return l.UserID == id })
			}
			for j := range h.Wishes {
				h.Wishes[j].Claims = slices.DeleteFunc(h.Wishes[j].Claims, func(c data.Claim) bool { return c.UserID == id })
			}
		}
		return nil
	}
	return missing
}
func clearAssignments(h *data.Household, matches func(data.Assignment) bool) {
	for i := range h.Wishes {
		w := &h.Wishes[i]
		if w.Assignment != nil && matches(*w.Assignment) {
			w.Assignment = nil
			w.Claims = nil
		}
	}
}

// @impl CON-004-AC3 CON-004-AC4 CON-003-AC10
func (a *householdApp) manage(w http.ResponseWriter, r *http.Request, store *data.Store, revision string, v adminView) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/manage/"), "/")
	if len(parts) != 2 {
		v.Error = "HR111"
		a.render(w, 404, "error.html", v)
		return
	}
	kind, action := parts[0], parts[1]
	id, name := r.Form.Get("id"), strings.TrimSpace(r.Form.Get("name"))
	if action == "create" {
		id = ""
	}
	context := r.Form.Get("context")
	status, key := 400, ""
	if r.Form.Get("revision") != revision {
		status, key = 409, "HR104"
	} else {
		err := store.Update(revision, func(current *data.Household) error {
			return mutateManagement(current, kind, action, id, name, r.Form.Get("recipient"), r.Form.Get("user"))
		})
		if err == nil {
			target := "/"
			if context != "" && !(kind == "list" && action == "delete" && id == context) {
				target = "/list?id=" + url.QueryEscape(context)
			}
			http.Redirect(w, r, target, 303)
			return
		}
		switch err.Error() {
		case "HR40", "HR41", "HR55", "HR56", "HR57", "HR66", "HR67":
			key = err.Error()
		case "HR111":
			status, key = 404, "HR111"
		default:
			status, key = 500, "HR105"
			if errors.Is(err, data.ErrStale) {
				status, key = 409, "HR104"
			}
		}
		key = storageMessage(err, key)
	}
	// A failed write never reuses a stale confirmation snapshot or leaks a partial mutation.
	fresh, rev, err := store.Load()
	if err != nil {
		v.Error = "HR103"
		a.render(w, 500, "error.html", v)
		return
	}
	v = a.householdView(fresh)
	v.CSRF = r.Form.Get("csrf")
	v.Error = key
	populateManagement(&v, fresh, rev, context)
	contextRequest := r.Clone(r.Context())
	u, _ := url.Parse(wishContextURL(context, ""))
	contextRequest.URL = u
	prepareHouseholdWishes(&v, fresh, store, contextRequest)
	if status == 400 || key == "HR119" || key == "HR113" {
		v.DraftKind = kind
		v.DraftName = name
		if action == "rename" {
			v.OpenModal = "rename-" + id
		}
	}
	template := "admin-home.html"
	if v.CurrentList != nil {
		template = "admin-list.html"
	}
	a.render(w, status, template, v)
}

// @impl ACC-007-AC4 ACC-007-AC5 WLS-009-AC4 ACC-006-AC8
func (a *householdApp) userList(w http.ResponseWriter, r *http.Request, h data.Household, list data.List, userID, code string, store *data.Store, v adminView) {
	a.userListStatus(w, r, h, list, userID, code, store, v, 200)
}
func (a *householdApp) userListStatus(w http.ResponseWriter, r *http.Request, h data.Household, list data.List, userID, code string, store *data.Store, v adminView, status int) {
	user := adminView{view: view{Locale: v.Locale, Language: v.Language, Name: h.Name, Appearance: v.Appearance, PersonName: list.Name}, Code: code, CSRF: v.CSRF, Revision: v.Revision, ContextRecipient: r.URL.Query().Get("recipient")}
	user.ClaimError = v.ClaimError
	for _, p := range h.Users {
		if p.ID == userID {
			user.ViewerName = p.Name
		}
	}
	// Only construct presentation cards for the current list, without other claimant names.
	for _, p := range h.Recipients {
		if !slices.Contains(list.RecipientIDs, p.ID) {
			continue
		}
		row := recipientReadView{ID: p.ID, Name: p.Name, URL: "/?code=" + url.QueryEscape(code) + "&recipient=" + url.QueryEscape(p.ID)}
		for _, wish := range h.Wishes {
			if wish.Assignment != nil && wish.Assignment.ListID == list.ID && wish.Assignment.RecipientID == p.ID {
				row.Cards = append(row.Cards, householdCard(h, wish, store, code, userID, false))
			}
		}
		row.TotalPrice, row.ClaimedWishes, row.UnclaimedWishes = summarizeWishes(row.Cards)
		user.ReadRecipients = append(user.ReadRecipients, row)
	}

	// The combined list view preserves household wish added order across recipients.
	for _, wish := range h.Wishes {
		if wish.Assignment != nil && wish.Assignment.ListID == list.ID {
			user.Wishes = append(user.Wishes, householdCard(h, wish, store, code, userID, false))
		}
	}
	if recipient := r.URL.Query().Get("recipient"); recipient != "" {
		user.ReadRecipients = slices.DeleteFunc(user.ReadRecipients, func(p recipientReadView) bool { return p.ID != recipient })
		if len(user.ReadRecipients) == 0 {
			user.Error = "HR111"
			a.render(w, 404, "error.html", user)
			return
		}
	}
	if r.URL.Query().Get("recipient") != "" {
		user.Wishes = user.ReadRecipients[0].Cards
		user.PersonName = user.ReadRecipients[0].Name
		user.ReadRecipients = nil
	}
	a.render(w, status, "admin-user.html", user)
}
