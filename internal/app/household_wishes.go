package app

import (
	"errors"
	data "github.com/pheelee/wishmas/internal/household"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
)

type assignmentChoice struct{ Value, Label string }

func wishContextURL(list, recipient string) string {
	if list != "" {
		u := "/list?id=" + url.QueryEscape(list)
		if recipient != "" {
			u += "&recipient=" + url.QueryEscape(recipient)
		}
		return u
	}
	return "/"
}
func selectedAssignment(h data.Household, list, recipient string) bool {
	for _, l := range h.Lists {
		if l.ID == list {
			return slices.Contains(l.RecipientIDs, recipient)
		}
	}
	return false
}
func wishAssignmentLabel(h data.Household, w data.Wish, locale string) string {
	if w.Assignment == nil {
		return translate(locale, "HR76")
	}
	var list, recipient string
	for _, l := range h.Lists {
		if l.ID == w.Assignment.ListID {
			list = l.Name
		}
	}
	for _, p := range h.Recipients {
		if p.ID == w.Assignment.RecipientID {
			recipient = p.Name
		}
	}
	return translate(locale, "HR77", recipient, list)
}

// @impl CLM-005-AC3 CLM-007-AC1 CLM-007-AC2 CLM-007-AC3
// @impl WSH-001 WSH-005 WSH-006-AC8 WSH-006-AC9 WSH-006-AC11 WSH-006-AC15
func householdCard(h data.Household, w data.Wish, store *data.Store, code, user string, admin bool) wishCard {
	c := wishCard{Token: w.ID, Title: w.Title, Price: w.Price, Links: w.Links, LinksText: strings.Join(w.Links, "\n"), Note: w.Note, Picture: w.Picture, PictureSource: w.PictureSource, AssignmentLabel: wishAssignmentLabel(h, w, h.Language)}
	if w.Assignment != nil {
		c.AssignmentValue = w.Assignment.ListID + ":" + w.Assignment.RecipientID
	}
	for _, link := range w.Links {
		u, _ := url.Parse(link)
		c.Shops = append(c.Shops, shopLink{URL: link, Domain: u.Hostname()})
	}
	for _, cl := range w.Claims {
		name := ""
		if admin {
			for _, u := range h.Users {
				if u.ID == cl.UserID {
					name = u.Name
				}
			}
		}
		identity := ""
		if admin {
			identity = cl.UserID
		}
		c.Claims = append(c.Claims, claim{Name: name, UserID: identity})
		if user == cl.UserID {
			c.Mine = true
		}
	}
	c.Accent = "red"
	if completeWish(c) {
		c.Accent = "yellow"
		if len(c.Claims) > 0 {
			c.Accent = "green"
		}
	}
	if w.Picture != "" {
		c.PictureUnavailable = true
		if _, err := readHouseholdPicture(store, w.Picture); err == nil {
			q := url.Values{"wish": {w.ID}, "picture": {w.Picture}}
			if code != "" {
				q.Set("code", code)
			}
			c.PictureURL = "/wish-picture?" + q.Encode()
			c.PictureUnavailable = false
		}
	}
	return c
}

// @impl WLS-010-AC7 WLS-010-AC8 WLS-011 WSH-003
func prepareHouseholdWishes(v *adminView, h data.Household, store *data.Store, r *http.Request) bool {
	v.Admin = true
	v.ContextList = ""
	if v.CurrentList != nil {
		v.ContextList = v.CurrentList.ID
	}
	v.ContextRecipient = r.URL.Query().Get("recipient")
	if v.ContextRecipient != "" {
		if v.ContextList == "" || !selectedAssignment(h, v.ContextList, v.ContextRecipient) {
			return false
		}
		for _, p := range h.Recipients {
			if p.ID == v.ContextRecipient {
				person := p
				v.CurrentRecipient = &person
				v.PersonName = p.Name
			}
		}
	}
	v.ContextURL = wishContextURL(v.ContextList, v.ContextRecipient)
	for _, l := range h.Lists {
		for _, p := range h.Recipients {
			if slices.Contains(l.RecipientIDs, p.ID) {
				v.AssignmentChoices = append(v.AssignmentChoices, assignmentChoice{l.ID + ":" + p.ID, p.Name + " · " + l.Name})
			}
		}
	}
	v.FilterList = r.URL.Query().Get("filter_list")
	v.FilterRecipient = r.URL.Query().Get("filter_recipient")
	v.FilterUnassigned = r.URL.Query().Get("unassigned") == "yes"
	v.Wishes = nil
	for _, w := range h.Wishes {
		if v.ContextList != "" && (w.Assignment == nil || w.Assignment.ListID != v.ContextList) {
			continue
		}
		if v.ContextRecipient != "" && w.Assignment.RecipientID != v.ContextRecipient {
			continue
		}
		if v.FilterUnassigned && w.Assignment != nil {
			continue
		}
		if v.FilterList != "" && (w.Assignment == nil || w.Assignment.ListID != v.FilterList) {
			continue
		}
		if v.FilterRecipient != "" && (w.Assignment == nil || w.Assignment.RecipientID != v.FilterRecipient) {
			continue
		}
		v.Wishes = append(v.Wishes, householdCard(h, w, store, "", "", true))
	}
	for i := range v.ReadRecipients {
		row := &v.ReadRecipients[i]
		row.URL = wishContextURL(v.ContextList, row.ID)
		row.Cards = nil
		for _, w := range row.Wishes {
			row.Cards = append(row.Cards, householdCard(h, w, store, "", "", true))
		}
		row.TotalPrice, row.ClaimedWishes, row.UnclaimedWishes = summarizeWishes(row.Cards)
	}
	return true
}

// @impl CLM-007-AC5 CLM-008-AC7
// @impl WSH-007 WSH-001 WLS-010 CON-004-AC3 CON-004-AC4 WSH-006-AC1 WSH-006-AC2 WSH-006-AC12 WSH-006-AC13
func (a *householdApp) householdWishAction(w http.ResponseWriter, r *http.Request, h data.Household, store *data.Store, revision string, v adminView) {
	id := r.PostForm.Get("wish")
	action := strings.TrimPrefix(r.URL.Path, "/wishes/")
	list, recipient := r.PostForm.Get("context"), r.PostForm.Get("context_recipient")
	fail := func(status int, key string, draft wishCard) {
		a.householdWishError(w, r, store, list, recipient, status, key, draft, v.CSRF)
	}
	draft := wishCard{Token: id, Title: r.PostForm.Get("title"), Price: r.PostForm.Get("price"), Note: r.PostForm.Get("note"), LinksText: r.PostForm.Get("links"), PictureSource: strings.TrimSpace(r.PostForm.Get("picture_url"))}
	if store.ReadOnly() {
		fail(503, "HR113", draft)
		return
	}
	if r.PostForm.Get("revision") != revision {
		fail(409, "HR104", draft)
		return
	}
	if list != "" && !slices.ContainsFunc(h.Lists, func(l data.List) bool { return l.ID == list }) {
		fail(404, "HR111", draft)
		return
	}
	if recipient != "" && !selectedAssignment(h, list, recipient) {
		fail(404, "HR111", draft)
		return
	}
	index := slices.IndexFunc(h.Wishes, func(w data.Wish) bool { return w.ID == id })
	if (id != "" && index < 0) || (action != "save" && index < 0) {
		fail(404, "HR111", draft)
		return
	}
	old := data.Wish{}
	if index >= 0 {
		old = h.Wishes[index]
	}
	updated := old
	oldPicture, newPicture := old.Picture, old.Picture
	var raw []byte
	if action == "save" {
		for _, line := range strings.Split(draft.LinksText, "\n") {
			if link := strings.TrimSpace(line); link != "" {
				draft.Links = append(draft.Links, link)
			}
		}
		validated := draft
		if err := validateWish(&validated); err != nil {
			fail(400, err.Error(), draft)
			return
		}
		if id == "" {
			generated, err := data.NewID()
			if err != nil {
				fail(500, "HR105", draft)
				return
			}
			updated.ID = generated
			if recipient != "" {
				updated.Assignment = &data.Assignment{ListID: list, RecipientID: recipient}
			}
		}
		if draft.PictureSource == "" {
			newPicture = ""
		} else if draft.PictureSource != old.PictureSource {
			// Release the app lock during bounded network work, then revalidate credentials and revision.
			a.mu.Unlock()
			download, err := fetchPictureWithClient(r.Context(), a.pictureClient, draft.PictureSource)
			a.mu.Lock()
			current, currentStore, currentRev, loadErr := a.load()
			if loadErr != nil || current == nil {
				fail(500, "HR103", draft)
				return
			}
			if _, ok := a.session(r, *current); !ok {
				login := a.householdView(*current)
				login.CSRF = v.CSRF
				login.Error = "HR12"
				a.render(w, 401, "admin-login.html", login)
				return
			}
			if currentRev != revision {
				fail(409, "HR104", draft)
				return
			}
			store = currentStore
			h = *current
			if err != nil {
				fail(400, err.Error(), draft)
				return
			}
			raw = download
			version, err := data.NewID()
			if err != nil {
				fail(500, "I16", draft)
				return
			}
			newPicture = updated.ID + "-" + version + ".webp"
		}
		updated.Title = validated.Title
		updated.Price = draft.Price
		updated.Note = draft.Note
		updated.Links = draft.Links
		updated.PictureSource = draft.PictureSource
		updated.Picture = newPicture
	} else if action == "assign" {
		value := r.PostForm.Get("assignment")
		if value == "" {
			updated.Assignment = nil
		} else {
			parts := strings.Split(value, ":")
			if len(parts) != 2 || !selectedAssignment(h, parts[0], parts[1]) {
				fail(400, "HR91", draft)
				return
			}
			updated.Assignment = &data.Assignment{ListID: parts[0], RecipientID: parts[1]}
		}
	} else if action == "delete" {
		if r.PostForm.Get("confirm") != "yes" {
			fail(400, "HR112", draft)
			return
		}
		newPicture = ""
	} else {
		fail(404, "HR111", draft)
		return
	}
	change := func(current *data.Household) error {
		if id == "" {
			current.Wishes = append(current.Wishes, updated)
		} else {
			at := slices.IndexFunc(current.Wishes, func(w data.Wish) bool { return w.ID == id })
			if at < 0 {
				return data.ErrStale
			}
			if action == "delete" {
				current.Wishes = slices.Delete(current.Wishes, at, at+1)
			} else {
				current.Wishes[at] = updated
			}
		}
		return nil
	}
	var err error
	if oldPicture != newPicture || len(raw) > 0 {
		err = store.UpdatePicture(revision, oldPicture, newPicture, raw, change)
	} else {
		err = store.Update(revision, change)
	}
	if err != nil {
		key := "HR105"
		status := 500
		if errors.Is(err, data.ErrStale) {
			status, key = 409, "HR104"
		} else if errors.Is(err, data.ErrPictureRemove) {
			key = "I17"
		} else if oldPicture != newPicture {
			key = "I16"
		}
		fail(status, storageMessage(err, key), draft)
		return
	}
	http.Redirect(w, r, wishContextURL(list, recipient), 303)
}
func (a *householdApp) householdWishError(w http.ResponseWriter, r *http.Request, store *data.Store, list, recipient string, status int, key string, draft wishCard, csrf string) {
	h, rev, err := store.Load()
	if err != nil {
		a.render(w, 500, "error.html", adminView{view: view{Locale: "en-US", Error: "HR103"}})
		return
	}
	v := a.householdView(h)
	v.CSRF = csrf
	populateManagement(&v, h, rev, list)
	contextRequest := r.Clone(r.Context())
	u, _ := url.Parse(wishContextURL(list, recipient))
	contextRequest.URL = u
	if !prepareHouseholdWishes(&v, h, store, contextRequest) {
		v.Error = "HR111"
		a.render(w, 404, "error.html", v)
		return
	}
	v.WishError = translate(h.Language, key)
	if r.URL.Path == "/wishes/save" {
		if draft.Token != "" {
			v.Draft = draft
		} else {
			v.CreateDraft = draft
		}
	} else {
		v.Error = key
	}
	template := "admin-home.html"
	if list != "" {
		template = "admin-list.html"
	}
	a.render(w, status, template, v)
}

func readHouseholdPicture(store *data.Store, name string) ([]byte, error) {
	dir := store.PictureDirectory()
	if err := data.RegularPictureDirectory(dir); err != nil {
		return nil, err
	}
	if !data.SafePicture(name) {
		return nil, errPictureFormat
	}
	return readPicture(filepath.Join(dir, name))
}

// @impl WSH-006-AC11: current assignment and exact current picture are checked on every request.
func (a *householdApp) householdPicture(w http.ResponseWriter, r *http.Request, h data.Household, store *data.Store, list string, v adminView) {
	for _, wish := range h.Wishes {
		if wish.ID != r.URL.Query().Get("wish") || wish.Picture == "" || wish.Picture != r.URL.Query().Get("picture") {
			continue
		}
		if list != "" && (wish.Assignment == nil || wish.Assignment.ListID != list) {
			break
		}
		raw, err := readHouseholdPicture(store, wish.Picture)
		if err == nil {
			w.Header().Set("Content-Type", "image/webp")
			w.Write(raw)
			return
		}
	}
	v.Error = "HR111"
	a.render(w, 404, "error.html", v)
}
