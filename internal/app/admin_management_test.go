package app

import (
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	data "wishmas/internal/household"
)

// @test ACC-007-AC1 ACC-007-AC3 WLS-007-AC1 WLS-008-AC1 WLS-009-AC1
func TestHouseholdManagementCreatesRostersListsAndLinks(t *testing.T) {
	dir := t.TempDir()
	_, b := authSetup(t, dir)
	page := b.get(t, "/")
	if !strings.Contains(page.Body.String(), `action="/manage/list/create"`) {
		t.Fatal("admin landing has no list creation form")
	}
	store := authStore(dir)
	create := func(kind, name string) {
		t.Helper()
		_, rev, err := store.Load()
		if err != nil {
			t.Fatal(err)
		}
		w := b.post(t, "/manage/"+kind+"/create", url.Values{"name": {name}, "revision": {rev}})
		if w.Code != 303 {
			t.Fatalf("create %s status %d", kind, w.Code)
		}
	}
	create("user", "Grand Mum")
	create("list", "Christmas 2026!")
	create("recipient", "Grand Mum")
	create("list", "Birthday 2027")
	create("user", "Dad")
	h, rev, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(h.Lists) != 2 || len(h.Users) != 2 || len(h.Recipients) != 1 || len(h.Lists[0].UserLinks) != 2 || len(h.Lists[1].UserLinks) != 2 {
		t.Fatal("rosters/link cross product incorrect")
	}
	w := b.post(t, "/manage/selection/add", url.Values{"id": {h.Lists[0].ID}, "recipient": {h.Recipients[0].ID}, "revision": {rev}})
	if w.Code != 303 {
		t.Fatal("select recipient", w.Code)
	}
	h, _, _ = store.Load()
	if len(h.Lists[0].RecipientIDs) != 1 || len(h.Lists[1].RecipientIDs) != 0 {
		t.Fatal("selection isolation")
	}
	user := newAuthBrowser(b.handler)
	page = user.get(t, "/?code="+h.Lists[0].UserLinks[0].Code)
	if page.Code != 200 || !strings.Contains(page.Body.String(), h.Recipients[0].Name) || strings.Contains(page.Body.String(), h.Lists[1].Name) || strings.Contains(page.Body.String(), "/manage/") {
		t.Fatal("scoped list view incorrect")
	}
}

func managementFixture(t *testing.T) (string, *authBrowser, data.Household) {
	t.Helper()
	dir := t.TempDir()
	_, b := authSetup(t, dir)
	s := authStore(dir)
	_, rev, _ := s.Load()
	err := s.Update(rev, func(h *data.Household) error {
		h.Users = []data.Person{{ID: "u1", Name: "Grand Mum"}, {ID: "u2", Name: "Other User"}}
		h.Recipients = []data.Person{{ID: "r1", Name: "Grand Mum"}, {ID: "r2", Name: "Other Recipient"}}
		h.Lists = []data.List{
			{ID: "l1", Name: "Christmas 2001", RecipientIDs: []string{"r2", "r1"}, UserLinks: []data.UserLink{{UserID: "u1", Code: strings.Repeat("a", 64)}, {UserID: "u2", Code: strings.Repeat("b", 64)}}},
			{ID: "l2", Name: "Birthday 2027", RecipientIDs: []string{"r1"}, UserLinks: []data.UserLink{{UserID: "u1", Code: strings.Repeat("c", 64)}, {UserID: "u2", Code: strings.Repeat("d", 64)}}},
		}
		h.Wishes = []data.Wish{
			{ID: "w1", Title: "Selected first wish", Price: "0", Links: []string{"https://example.test/product"}, Note: "Original note", Picture: "w1-v1.webp", Assignment: &data.Assignment{ListID: "l1", RecipientID: "r1"}, Claims: []data.Claim{{UserID: "u1"}, {UserID: "u2"}}},
			{ID: "w2", Title: "Other list private wish", Assignment: &data.Assignment{ListID: "l2", RecipientID: "r1"}, Claims: []data.Claim{{UserID: "u1"}}},
			{ID: "w3", Title: "Unassigned private wish", Claims: []data.Claim{{UserID: "u1"}, {UserID: "u2"}}},
			{ID: "w4", Title: "Selected other recipient", Assignment: &data.Assignment{ListID: "l1", RecipientID: "r2"}, Claims: []data.Claim{{UserID: "u2"}}},
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	h, _, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	return dir, b, h
}
func managementPost(t *testing.T, dir string, b *authBrowser, path string, values url.Values) int {
	t.Helper()
	if values == nil {
		values = url.Values{}
	}
	_, rev, err := authStore(dir).Load()
	if err != nil {
		t.Fatal(err)
	}
	values.Set("revision", rev)
	return b.post(t, path, values).Code
}

// @test ACC-007-AC4 ACC-007-AC5 ACC-007-AC6 ACC-006-AC8 WLS-009-AC4 CON-004-AC4
func TestManagementScopesLinksAndRenames(t *testing.T) {
	dir, b, before := managementFixture(t)
	s := authStore(dir)
	user := newAuthBrowser(b.handler)
	code := before.Lists[0].UserLinks[0].Code
	page := user.get(t, "/?code="+code)
	body := page.Body.String()
	for _, expected := range []string{"Grand Mum", "Other Recipient", "Selected first wish", "Selected other recipient", "Original note"} {
		if !strings.Contains(body, expected) {
			t.Fatal("scoped view missing", expected)
		}
	}
	for _, private := range []string{"Other User", "Other list private wish", "Unassigned private wish", "Birthday 2027", before.Lists[1].UserLinks[0].Code, `/manage/`, `name="name"`} {
		if strings.Contains(body, private) {
			t.Fatal("private scope leaked", private)
		}
	}
	filtered := user.get(t, "/?code="+code+"&recipient=r2")
	if !strings.Contains(filtered.Body.String(), "Selected other recipient") || strings.Contains(filtered.Body.String(), "Selected first wish") {
		t.Fatal("recipient view scope")
	}
	if w := user.get(t, "/?code="+code+"&recipient=absent"); w.Code != 404 {
		t.Fatal("unknown recipient", w.Code)
	}
	if w := b.get(t, "/?code="+code); strings.Contains(w.Body.String(), `/manage/`) {
		t.Fatal("admin cookie elevated user link")
	}
	_, rev, _ := s.Load()
	for _, actor := range []*authBrowser{b, user} {
		if w := actor.post(t, "/manage/list/delete?code="+code, url.Values{"id": {"l1"}, "revision": {rev}}); w.Code != 403 {
			t.Fatal("user admin mutation accepted", w.Code)
		}
	}
	unauth := newAuthBrowser(b.handler)
	unauth.get(t, "/")
	if w := unauth.post(t, "/manage/user/create", url.Values{"name": {"Intruder"}, "revision": {rev}}); w.Code != 401 {
		t.Fatal("unauth mutation", w.Code)
	}
	for _, row := range []struct{ kind, id, name string }{{"list", "l1", "Renamed old list"}, {"recipient", "r1", "Renamed Recipient"}, {"user", "u1", "Renamed User"}} {
		if status := managementPost(t, dir, b, "/manage/"+row.kind+"/rename", url.Values{"id": {row.id}, "name": {row.name}}); status != 303 {
			t.Fatal("rename", status)
		}
	}
	renamed, _, _ := s.Load()
	if !reflect.DeepEqual(renamed.Lists[0].UserLinks, before.Lists[0].UserLinks) || !reflect.DeepEqual(renamed.Wishes, before.Wishes) || !reflect.DeepEqual(renamed.Lists[0].RecipientIDs, before.Lists[0].RecipientIDs) {
		t.Fatal("rename changed identity references")
	}
	if w := user.get(t, "/?code="+code); !strings.Contains(w.Body.String(), "Renamed User") || !strings.Contains(w.Body.String(), "Renamed Recipient") {
		t.Fatal("rename not visible on page load")
	}
	if status := managementPost(t, dir, b, "/manage/link/regenerate", url.Values{"id": {"l1"}, "user": {"u1"}, "context": {"l1"}}); status != 303 {
		t.Fatal("regeneration", status)
	}
	after, _, _ := s.Load()
	if after.Lists[0].UserLinks[0].Code == code || !reflect.DeepEqual(after.Wishes, before.Wishes) || after.Lists[0].UserLinks[1] != before.Lists[0].UserLinks[1] || !reflect.DeepEqual(after.Lists[1], before.Lists[1]) {
		t.Fatal("regeneration changed unrelated data")
	}
	if w := b.get(t, "/?code="+code); !strings.Contains(w.Body.String(), `name="password"`) {
		t.Fatal("revoked link used admin cookie")
	}
	for _, link := range []data.UserLink{after.Lists[0].UserLinks[0], after.Lists[0].UserLinks[1], after.Lists[1].UserLinks[0]} {
		if w := user.get(t, "/?code="+link.Code); w.Code != 200 || strings.Contains(w.Body.String(), `name="password"`) {
			t.Fatal("other/new link failed")
		}
	}
	if w := b.get(t, "/"); !strings.Contains(w.Body.String(), `/manage/list/create`) {
		t.Fatal("admin cookie lost")
	}
}

// @test WLS-007-AC6 WLS-007-AC7 WLS-008-AC5 WLS-009-AC2 WLS-009-AC3 ACC-007-AC7 ACC-007-AC8
func TestManagementDestructiveCleanupIsScoped(t *testing.T) {
	for _, kind := range []string{"list", "recipient", "selection", "user"} {
		t.Run(kind, func(t *testing.T) {
			dir, b, before := managementFixture(t)
			s := authStore(dir)
			picture := filepath.Join(dir, "households/Family/pictures/w1-v1.webp")
			if err := os.MkdirAll(filepath.Dir(picture), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(picture, []byte("retained-picture"), 0600); err != nil {
				t.Fatal(err)
			}
			path := "/manage/" + kind + "/delete"
			id := "l1"
			if kind == "recipient" {
				id = "r1"
			}
			if kind == "user" {
				id = "u1"
			}
			if kind == "selection" {
				path = "/manage/selection/remove"
			}
			// Rendering/canceling a confirmation does not mutate anything.
			_, rev, _ := s.Load()
			if b.get(t, "/").Code != 200 || b.get(t, "/list?id=l1").Code != 200 {
				t.Fatal("confirmation pages")
			}
			_, still, _ := s.Load()
			if rev != still {
				t.Fatal("confirmation GET wrote data")
			}
			if code := managementPost(t, dir, b, path, url.Values{"id": {id}, "recipient": {"r1"}}); code != 303 {
				t.Fatal("delete status", code)
			}
			after, _, err := s.Load()
			if err != nil {
				t.Fatal(err)
			}
			if len(after.Wishes) != len(before.Wishes) {
				t.Fatal("wish deleted")
			}
			for i, wish := range after.Wishes {
				expected := before.Wishes[i]
				if kind == "user" {
					expected.Claims = slices.DeleteFunc(slices.Clone(expected.Claims), func(c data.Claim) bool { return c.UserID == "u1" })
				} else if expected.Assignment != nil && ((kind == "list" && expected.Assignment.ListID == "l1") || (kind == "recipient" && expected.Assignment.RecipientID == "r1") || (kind == "selection" && expected.Assignment.ListID == "l1" && expected.Assignment.RecipientID == "r1")) {
					expected.Assignment = nil
					expected.Claims = nil
				}
				if len(expected.Claims) == 0 {
					expected.Claims = nil
				}
				if !reflect.DeepEqual(wish, expected) {
					t.Fatalf("%s changed wrong wish fields: %s", kind, wish.ID)
				}
			}
			image, err := os.ReadFile(picture)
			if err != nil || string(image) != "retained-picture" {
				t.Fatal("picture modified")
			}
			switch kind {
			case "list":
				if len(after.Lists) != 1 || !reflect.DeepEqual(after.Lists[0], before.Lists[1]) || !reflect.DeepEqual(after.Users, before.Users) || !reflect.DeepEqual(after.Recipients, before.Recipients) {
					t.Fatal("list cleanup scope")
				}
			case "recipient":
				if len(after.Recipients) != 1 || !reflect.DeepEqual(after.Users, before.Users) || len(after.Lists[0].RecipientIDs) != 1 || len(after.Lists[1].RecipientIDs) != 0 {
					t.Fatal("recipient cleanup scope")
				}
			case "selection":
				if !reflect.DeepEqual(after.Recipients, before.Recipients) || !reflect.DeepEqual(after.Lists[1], before.Lists[1]) || len(after.Lists[0].RecipientIDs) != 1 {
					t.Fatal("selection cleanup scope")
				}
			case "user":
				if len(after.Users) != 1 || !reflect.DeepEqual(after.Recipients, before.Recipients) || len(after.Lists[0].UserLinks) != 1 || len(after.Lists[1].UserLinks) != 1 || after.Lists[0].UserLinks[0] != before.Lists[0].UserLinks[1] {
					t.Fatal("user cleanup scope")
				}
			}
			if kind == "list" || kind == "user" {
				visitor := newAuthBrowser(b.handler)
				if w := visitor.get(t, "/?code="+before.Lists[0].UserLinks[0].Code); !strings.Contains(w.Body.String(), `name="password"`) {
					t.Fatal("deleted identity link remained valid")
				}
			}
		})
	}
}

// @test WLS-007-AC2 WLS-007-AC8 WLS-008-AC2 WLS-008-AC3 ACC-007-AC2 ACC-007-AC9 CON-004-AC4
func TestManagementValidationStaleAndWriteFailures(t *testing.T) {
	dir, b, _ := managementFixture(t)
	s := authStore(dir)
	for _, tc := range []struct {
		kind, name string
		status     int
	}{{"list", " Christmas 2001 ", 400}, {"list", "ab", 400}, {"list", strings.Repeat("z", 101), 400}, {"recipient", " grand mum ", 400}, {"user", "grand MUM", 400}, {"user", "ab", 400}, {"recipient", "Name!", 400}, {"user", "Name\tName", 400}, {"recipient", strings.Repeat("𐐀", 51), 400}, {"list", "  Birthday: É 🎄  ", 303}, {"user", strings.Repeat("𐐀", 50), 303}, {"recipient", "  Ünicode Person 9  ", 303}} {
		_, before, _ := s.Load()
		status := managementPost(t, dir, b, "/manage/"+tc.kind+"/create", url.Values{"name": {tc.name}})
		if status != tc.status {
			t.Fatalf("%s validation got %d want %d", tc.kind, status, tc.status)
		}
		_, after, _ := s.Load()
		if tc.status == 400 && before != after {
			t.Fatal("invalid draft wrote data")
		}
	}
	h, stale, _ := s.Load()
	if h.Lists[2].Name != "Birthday: É 🎄" || h.Recipients[2].Name != "Ünicode Person 9" || h.Users[0].ID != "u1" {
		t.Fatal("trim/added order")
	}
	if status := managementPost(t, dir, b, "/manage/list/rename", url.Values{"id": {"l1"}, "name": {"Manually editable old list"}}); status != 303 {
		t.Fatal("old list immutable")
	}
	_, before, _ := s.Load()
	w := b.post(t, "/manage/list/delete", url.Values{"id": {"l1"}, "revision": {stale}})
	if w.Code != 409 {
		t.Fatal("stale confirmation accepted", w.Code)
	}
	_, after, _ := s.Load()
	if before != after {
		t.Fatal("stale wrote")
	}
	for _, values := range []url.Values{{"id": {"absent"}, "name": {"Valid New Name"}}, {"id": {"u1"}, "name": {"Valid New Name"}}} {
		if status := managementPost(t, dir, b, "/manage/list/rename", values); status != 404 {
			t.Fatal("tampered IDs", status)
		}
	}
	if status := managementPost(t, dir, b, "/manage/selection/add", url.Values{"id": {"l1"}, "recipient": {"r1"}}); status != 400 {
		t.Fatal("duplicate selection")
	}
	if status := managementPost(t, dir, b, "/manage/selection/add", url.Values{"id": {"l1"}, "recipient": {"u1"}}); status != 404 {
		t.Fatal("cross roster selection")
	}
	if status := managementPost(t, dir, b, "/manage/link/regenerate", url.Values{"id": {"l1"}, "user": {"r1"}}); status != 404 {
		t.Fatal("cross roster user")
	}
	if os.Geteuid() != 0 {
		parent := filepath.Join(dir, "households/Family")
		if err := os.Chmod(parent, 0500); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(parent, 0700)
		_, before, _ := s.Load()
		status := managementPost(t, dir, b, "/manage/user/delete", url.Values{"id": {"u1"}})
		if status != 500 {
			t.Fatal("write failure", status)
		}
		_, after, _ := s.Load()
		if before != after {
			t.Fatal("failed write modified data")
		}
	}
}
