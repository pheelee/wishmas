package app

import (
	"errors"
	data "github.com/pheelee/wishmas/internal/household"
	"net/http"
	"net/url"
	"slices"
)

// @impl CLM-005-AC1 CLM-005-AC4 CLM-006-AC1 CLM-006-AC2 CLM-006-AC4 CLM-006-AC5 CLM-006-AC6 CLM-008-AC1 CLM-008-AC2 CLM-008-AC3 CON-004-AC4
// The caller supplies the verified link identity (or admin role); form fields
// cannot replace it. Explicit add/remove modes never invert a stale action.
func (a *householdApp) householdClaim(w http.ResponseWriter, r *http.Request, h data.Household, store *data.Store, revision string, v adminView, listID, userID, code string) {
	admin := code == ""
	wishID, mode := r.PostForm.Get("wish"), r.PostForm.Get("mode")
	selected := r.PostForm.Get("user")
	fail := func(status int, key string) { a.claimFailure(w, r, store, v, listID, userID, code, status, key) }
	index := slices.IndexFunc(h.Wishes, func(w data.Wish) bool { return w.ID == wishID })
	if index < 0 {
		fail(404, "HR111")
		return
	}
	wish := h.Wishes[index]
	if !admin && (wish.Assignment == nil || wish.Assignment.ListID != listID) {
		fail(404, "HR111")
		return
	}
	if mode != "add" && mode != "remove" {
		if !admin && mode == "remove-other" {
			fail(403, "HR102")
		} else {
			fail(400, "HR112")
		}
		return
	}
	if admin {
		if !slices.ContainsFunc(h.Users, func(p data.Person) bool { return p.ID == selected }) {
			if mode == "add" {
				fail(400, "HR98")
			} else {
				fail(404, "HR111")
			}
			return
		}
		userID = selected
	} else if selected != "" && selected != userID {
		fail(403, "HR101")
		return
	}
	recipient := r.PostForm.Get("context_recipient")
	context := listID
	if admin {
		context = r.PostForm.Get("context")
	}
	if context != "" && !slices.ContainsFunc(h.Lists, func(l data.List) bool { return l.ID == context }) {
		fail(404, "HR111")
		return
	}
	if recipient != "" && !selectedAssignment(h, context, recipient) {
		fail(404, "HR111")
		return
	}
	if r.PostForm.Get("revision") != revision {
		fail(409, "E20")
		return
	}
	if mode == "add" && wish.Assignment == nil {
		fail(400, "HR100")
		return
	}
	at := slices.IndexFunc(wish.Claims, func(c data.Claim) bool { return c.UserID == userID })
	if mode == "add" && at >= 0 {
		fail(409, "HR99")
		return
	}
	if mode == "remove" && at < 0 {
		fail(409, "E25")
		return
	}
	err := store.Update(revision, func(current *data.Household) error {
		target := &current.Wishes[index]
		if mode == "add" {
			target.Claims = append(target.Claims, data.Claim{UserID: userID})
		} else {
			target.Claims = slices.Delete(target.Claims, at, at+1)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, data.ErrStale) {
			fail(409, "E20")
		} else {
			fail(500, storageMessage(err, "E28"))
		}
		return
	}
	target := wishContextURL(context, recipient)
	if !admin {
		q := url.Values{"code": {code}}
		if recipient != "" {
			q.Set("recipient", recipient)
		}
		target = "/?" + q.Encode()
	}
	http.Redirect(w, r, target, 303)
}

// @impl CLM-006-AC6 CLM-007-AC2 CLM-007-AC4: errors render a fresh authorized
// snapshot inline; no household roster or other claimant names reach user views.
func (a *householdApp) claimFailure(w http.ResponseWriter, r *http.Request, store *data.Store, v adminView, listID, userID, code string, status int, key string) {
	h, revision, err := store.Load()
	if err != nil {
		a.render(w, 500, "error.html", adminView{view: view{Locale: "en-US", Error: "HR103"}})
		return
	}
	fresh := a.householdView(h)
	fresh.CSRF = v.CSRF
	fresh.Revision = revision
	fresh.ClaimError = translate(h.Language, key)
	recipient := r.PostForm.Get("context_recipient")
	if code != "" {
		// Recheck the link on this snapshot as well: manual edits do not share the app mutex.
		for _, list := range h.Lists {
			if list.ID != listID {
				continue
			}
			for _, link := range list.UserLinks {
				if link.UserID == userID && constantEqual(link.Code, code) {
					q := url.Values{"code": {code}}
					if selectedAssignment(h, listID, recipient) {
						q.Set("recipient", recipient)
					}
					request := r.Clone(r.Context())
					request.URL = &url.URL{Path: "/", RawQuery: q.Encode()}
					a.userListStatus(w, request, h, list, userID, code, store, fresh, status)
					return
				}
			}
		}
		a.render(w, 200, "admin-login.html", fresh)
		return
	}
	if _, ok := a.session(r, h); !ok {
		fresh.Error = "HR12"
		a.render(w, 401, "admin-login.html", fresh)
		return
	}
	context := r.PostForm.Get("context")
	if !populateManagement(&fresh, h, revision, context) {
		populateManagement(&fresh, h, revision, "")
		context = ""
	}
	if !selectedAssignment(h, context, recipient) {
		recipient = ""
	}
	request := r.Clone(r.Context())
	request.URL, _ = url.Parse(wishContextURL(context, recipient))
	prepareHouseholdWishes(&fresh, h, store, request)
	page := "admin-home.html"
	if fresh.CurrentList != nil {
		page = "admin-list.html"
	}
	a.render(w, status, page, fresh)
}
