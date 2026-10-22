package app

import (
	"bytes"
	"go.yaml.in/yaml/v3"
	"image"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	data "wishmas/internal/household"
)

// @test WSH-001-AC1 WSH-007-AC1 WSH-007-AC2 WLS-010-AC1
func TestHouseholdWishCreation(t *testing.T) {
	dir, b, _ := managementFixture(t)
	if page := b.get(t, "/"); !strings.Contains(page.Body.String(), `action="/wishes/save"`) {
		t.Fatal("household has no wish creation form")
	}
	if status := managementPost(t, dir, b, "/wishes/save", url.Values{"title": {"  Future train  "}}); status != 303 {
		t.Fatal("create", status)
	}
	h, _, err := authStore(dir).Load()
	if err != nil {
		t.Fatal(err)
	}
	w := h.Wishes[len(h.Wishes)-1]
	if w.Title != "Future train" || w.Assignment != nil {
		t.Fatal("new household wish incorrect")
	}
}

// @test WLS-010-AC2 WLS-010-AC3 WLS-010-AC4 WLS-010-AC6 WLS-010-AC7 WLS-010-AC8 WSH-007
func TestHouseholdWishAssignmentsEditingAndDeletion(t *testing.T) {
	dir, b, before := managementFixture(t)
	s := authStore(dir)
	for _, assignment := range []string{"l2:r1", "", "l1:r2"} {
		if status := managementPost(t, dir, b, "/wishes/assign", url.Values{"wish": {"w1"}, "assignment": {assignment}}); status != 303 {
			t.Fatal("assignment", status)
		}
		h, _, err := s.Load()
		if err != nil {
			t.Fatal(err)
		}
		expected := before.Wishes[0]
		expected.Assignment = h.Wishes[0].Assignment
		if !reflect.DeepEqual(h.Wishes[0], expected) {
			t.Fatal("assignment changed details or claims")
		}
		if assignment == "" && h.Wishes[0].Assignment != nil {
			t.Fatal("unassign failed")
		}
		if assignment != "" && h.Wishes[0].Assignment.ListID+":"+h.Wishes[0].Assignment.RecipientID != assignment {
			t.Fatal("move failed")
		}
	}
	_, rev, _ := s.Load()
	for _, invalid := range []string{"l2:r2", "l1:u1", "l1:r1:bad", "absent:r1"} {
		if code := managementPost(t, dir, b, "/wishes/assign", url.Values{"wish": {"w1"}, "assignment": {invalid}}); code != 400 {
			t.Fatal("invalid assignment", code)
		}
	}
	_, same, _ := s.Load()
	if rev != same {
		t.Fatal("invalid assignment saved")
	}
	if code := managementPost(t, dir, b, "/wishes/save", url.Values{"title": {"Recipient page wish"}, "context": {"l2"}, "context_recipient": {"r1"}}); code != 303 {
		t.Fatal("recipient create", code)
	}
	h, _, _ := s.Load()
	created := h.Wishes[len(h.Wishes)-1]
	if created.Assignment == nil || created.Assignment.ListID != "l2" || created.Assignment.RecipientID != "r1" {
		t.Fatal("creation context lost")
	}
	if code := managementPost(t, dir, b, "/wishes/save", url.Values{"wish": {"w1"}, "title": {"  Edited claimed train  "}, "price": {"2.00"}, "links": {"https://shop.example/a\nhttps://second.example/b"}, "note": {"New note"}}); code != 303 {
		t.Fatal("edit", code)
	}
	h, _, _ = s.Load()
	if h.Wishes[0].Title != "Edited claimed train" || len(h.Wishes[0].Claims) != 2 || h.Wishes[0].ID != "w1" || h.Wishes[0].Assignment.RecipientID != "r2" {
		t.Fatal("edit lost identity/claims/assignment")
	}
	for _, q := range []struct {
		query              string
		included, excluded string
	}{{"/?unassigned=yes", "Unassigned private wish", "Edited claimed train"}, {"/?filter_list=l2", "Recipient page wish", "Edited claimed train"}, {"/?filter_recipient=r2", "Edited claimed train", "Recipient page wish"}} {
		w := b.get(t, q.query)
		if !strings.Contains(w.Body.String(), `data-wish-title="`+q.included+`"`) || strings.Contains(w.Body.String(), `data-wish-title="`+q.excluded+`"`) {
			t.Fatal("filter scope", q.query)
		}
	}
	for _, path := range []string{"/wishes/save", "/wishes/delete", "/wishes/assign"} {
		if w := b.post(t, path+"?code="+before.Lists[0].UserLinks[0].Code, url.Values{"wish": {"w1"}, "revision": {same}}); w.Code != 403 {
			t.Fatal("user mutation permitted", w.Code)
		}
	}
	// Delete requires explicit confirmation and stale snapshots cannot apply.
	_, oldRev, _ := s.Load()
	if code := managementPost(t, dir, b, "/wishes/delete", url.Values{"wish": {"w1"}}); code != 400 {
		t.Fatal("unconfirmed delete")
	}
	if code := managementPost(t, dir, b, "/wishes/delete", url.Values{"wish": {"w1"}, "confirm": {"yes"}}); code != 303 {
		t.Fatal("delete", code)
	}
	h, _, _ = s.Load()
	if slices.ContainsFunc(h.Wishes, func(w data.Wish) bool { return w.ID == "w1" }) || len(h.Wishes) != 4 {
		t.Fatal("delete wrong wish")
	}
	if w := b.post(t, "/wishes/save", url.Values{"wish": {"w2"}, "title": {"Overwrite attempt"}, "revision": {oldRev}}); w.Code != 409 {
		t.Fatal("stale edit", w.Code)
	}
	if w := b.get(t, "/?recipient=r1"); w.Code != 404 {
		t.Fatal("recipient without list accepted", w.Code)
	}
	if w := b.get(t, "/list?id=l2&recipient=r2"); w.Code != 404 {
		t.Fatal("unselected recipient page")
	}
}

// @test WSH-001 WLS-011 WSH-005-AC6 WSH-005-AC7
func TestHouseholdWishValidationSummariesAndPrivacy(t *testing.T) {
	dir, b, _ := managementFixture(t)
	s := authStore(dir)
	for _, bad := range []url.Values{{"title": {"    "}}, {"title": {"abcd"}}, {"title": {strings.Repeat("a", 151)}}, {"title": {"Valid title"}, "price": {"-1"}}, {"title": {"Valid title"}, "price": {"1,25"}}, {"title": {"Valid title"}, "price": {"1.234"}}, {"title": {"Valid title"}, "links": {"example.org"}}} {
		_, rev, _ := s.Load()
		if code := managementPost(t, dir, b, "/wishes/save", bad); code != 400 {
			t.Fatal("invalid saved", code)
		}
		_, after, _ := s.Load()
		if rev != after {
			t.Fatal("validation wrote")
		}
	}
	for _, price := range []string{"0", "2", "10.5", strings.Repeat("9", 60) + ".99"} {
		if code := managementPost(t, dir, b, "/wishes/save", url.Values{"title": {"Duplicate allowed"}, "price": {price}, "context": {"l1"}, "context_recipient": {"r1"}}); code != 303 {
			t.Fatal("valid numeric", code)
		}
	}
	h, rev, _ := s.Load()
	if err := s.Update(rev, func(h *data.Household) error {
		h.Wishes = append(h.Wishes, data.Wish{ID: "complete", Title: "Unclaimed complete", Price: "2", Links: []string{"https://shop.example/"}, Assignment: &data.Assignment{ListID: "l1", RecipientID: "r1"}})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	h, _, _ = s.Load()
	_, store, _, _ := b.handler.(*householdApp).load()
	v := b.handler.(*householdApp).householdView(h)
	populateManagement(&v, h, "rev", "l1")
	req := httptest.NewRequest("GET", "https://wishmas.test/list?id=l1", nil)
	prepareHouseholdWishes(&v, h, store, req)
	if v.ReadRecipients[0].ClaimedWishes != 1 || v.ReadRecipients[0].UnclaimedWishes != 1 || v.ReadRecipients[0].TotalPrice != "1000000000000000000000000000000000000000000000000000000000014.49" {
		t.Fatalf("summary incorrect %s claimed=%d unclaimed=%d", v.ReadRecipients[0].TotalPrice, v.ReadRecipients[0].ClaimedWishes, v.ReadRecipients[0].UnclaimedWishes)
	}
	user := newAuthBrowser(b.handler)
	page := user.get(t, "/?code="+h.Lists[0].UserLinks[0].Code)
	body := page.Body.String()
	if strings.Contains(body, "Other User") || strings.Contains(body, "Unassigned private wish") || strings.Contains(body, "Other list private wish") || !strings.Contains(body, "Unclaimed complete") {
		t.Fatal("card privacy/scope")
	}
}

// @test WSH-006-AC1–AC15 CON-004-AC5 WLS-010-AC4 WLS-010-AC5
func TestHouseholdPicturesLifecycleAndScope(t *testing.T) {
	dir, b, h := managementFixture(t)
	a := b.handler.(*householdApp)
	s := authStore(dir)
	calls := 0
	mock := &application{}
	mockPictureClient(mock, pictureFixture(t, 800, 400), &calls)
	a.pictureClient = mock.pictureClient
	save := func(values url.Values) int { t.Helper(); return managementPost(t, dir, b, "/wishes/save", values) }
	values := url.Values{"title": {"Picture wish title"}, "price": {"12.50"}, "links": {"https://shop.example/item"}, "note": {"Preserved note"}, "picture_url": {"https://images.example/one.png"}, "context": {"l1"}, "context_recipient": {"r1"}}
	if code := save(values); code != 303 {
		t.Fatal("image create", code)
	}
	h, _, _ = s.Load()
	wish := h.Wishes[len(h.Wishes)-1]
	firstPicture := wish.Picture
	if !strings.HasPrefix(firstPicture, wish.ID+"-") || calls != 1 {
		t.Fatal("picture naming/download")
	}
	raw, err := readHouseholdPicture(s, firstPicture)
	if err != nil {
		t.Fatal(err)
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || format != "webp" || config.Width != 640 || config.Height != 320 {
		t.Fatal("optimized image")
	}
	imageURL := "/wish-picture?wish=" + wish.ID + "&picture=" + firstPicture
	if w := b.get(t, imageURL); w.Code != 200 || !bytes.Equal(raw, w.Body.Bytes()) {
		t.Fatal("admin image")
	}
	user := newAuthBrowser(a)
	code1, code2 := h.Lists[0].UserLinks[0].Code, h.Lists[1].UserLinks[0].Code
	if w := user.get(t, imageURL+"&code="+code1); w.Code != 200 || w.Header().Get("Content-Type") != "image/webp" {
		t.Fatal("own list image")
	}
	if w := b.get(t, imageURL+"&code="+code2); w.Code != 404 {
		t.Fatal("other list image allowed by admin cookie")
	}
	values.Set("wish", wish.ID)
	values.Set("title", "Edited picture wish")
	if code := save(values); code != 303 || calls != 1 {
		t.Fatal("unchanged source fetched", code, calls)
	}
	// Add a retained claim, then move and unassign without changing image bytes/source or claims.
	h, rev, _ := s.Load()
	if err := s.Update(rev, func(h *data.Household) error {
		h.Wishes[len(h.Wishes)-1].Claims = []data.Claim{{UserID: "u1"}}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if code := managementPost(t, dir, b, "/wishes/assign", url.Values{"wish": {wish.ID}, "assignment": {"l2:r1"}}); code != 303 {
		t.Fatal("image move", code)
	}
	if w := user.get(t, imageURL+"&code="+code1); w.Code != 404 {
		t.Fatal("old list picture after move")
	}
	if w := user.get(t, imageURL+"&code="+code2); w.Header().Get("Content-Type") != "image/webp" {
		t.Fatal("new list image")
	}
	if code := managementPost(t, dir, b, "/manage/link/regenerate", url.Values{"id": {"l2"}, "user": {"u1"}}); code != 303 {
		t.Fatal("regeneration")
	}
	if w := b.get(t, imageURL+"&code="+code2); w.Header().Get("Content-Type") == "image/webp" {
		t.Fatal("revoked link still serves image with admin cookie")
	}
	fresh, _, _ := s.Load()
	code2 = fresh.Lists[1].UserLinks[0].Code
	if w := user.get(t, imageURL+"&code="+code2); w.Header().Get("Content-Type") != "image/webp" {
		t.Fatal("replacement link cannot see assigned image")
	}
	if code := managementPost(t, dir, b, "/wishes/assign", url.Values{"wish": {wish.ID}, "assignment": {""}}); code != 303 {
		t.Fatal("unassign")
	}
	if w := user.get(t, imageURL+"&code="+code2); w.Code != 404 {
		t.Fatal("unassigned picture leaked")
	}
	h, _, _ = s.Load()
	current := h.Wishes[len(h.Wishes)-1]
	if current.Picture != firstPicture || len(current.Claims) != 1 || current.PictureSource != values.Get("picture_url") {
		t.Fatal("move changed image/claims")
	}
	// Entire form and original YAML/image survive failed replacement.
	_, revision, _ := s.Load()
	values.Set("picture_url", "http://127.0.0.1/private.png")
	values.Set("title", "  Alternate picture title  ")
	values.Set("links", "https://shop.example/next\n\nhttps://second.example/path")
	values.Set("note", "  draft note\nsecond line  ")
	values.Set("revision", revision)
	response := b.post(t, "/wishes/save", values)
	if response.Code != 400 || !strings.Contains(response.Body.String(), "Alternate picture title") || !strings.Contains(response.Body.String(), "draft note") || !strings.Contains(response.Body.String(), "127.0.0.1/private.png") {
		t.Fatal("failed draft not retained", response.Code)
	}
	_, after, _ := s.Load()
	if revision != after {
		t.Fatal("download failure modified YAML")
	}
	if got, _ := readHouseholdPicture(s, firstPicture); !bytes.Equal(got, raw) {
		t.Fatal("download failure modified picture")
	}
	values.Set("picture_url", "https://images.example/two.png")
	if code := save(values); code != 303 {
		t.Fatal("replace", code)
	}
	h, _, _ = s.Load()
	second := h.Wishes[len(h.Wishes)-1].Picture
	if second == firstPicture || calls != 2 {
		t.Fatal("replacement not fetched")
	}
	if _, err := os.Stat(filepath.Join(s.PictureDirectory(), firstPicture)); !os.IsNotExist(err) {
		t.Fatal("old file retained")
	}
	if w := b.get(t, imageURL); w.Code != 404 {
		t.Fatal("known old image URL works")
	}
	// Missing images are not re-downloaded and give a distinct placeholder.
	if err := os.Remove(filepath.Join(s.PictureDirectory(), second)); err != nil {
		t.Fatal(err)
	}
	if page := b.get(t, "/"); !strings.Contains(page.Body.String(), "picture-unavailable") {
		t.Fatal("missing image placeholder")
	}
	if calls != 2 {
		t.Fatal("view fetched missing image")
	}
	missing := householdCard(h, h.Wishes[len(h.Wishes)-1], s, "", "", true)
	if !missing.PictureUnavailable || missing.PictureURL != "" {
		t.Fatal("missing image must not expose available shelf")
	}
	values.Set("picture_url", "")
	if code := save(values); code != 303 {
		t.Fatal("clear missing image", code)
	}
	h, _, _ = s.Load()
	if h.Wishes[len(h.Wishes)-1].Picture != "" || h.Wishes[len(h.Wishes)-1].PictureSource != "" {
		t.Fatal("clear picture")
	}
	values.Set("picture_url", "https://images.example/three.png")
	if code := save(values); code != 303 {
		t.Fatal("add picture", code)
	}
	h, _, _ = s.Load()
	third := h.Wishes[len(h.Wishes)-1].Picture
	if code := managementPost(t, dir, b, "/wishes/delete", url.Values{"wish": {wish.ID}, "confirm": {"yes"}}); code != 303 {
		t.Fatal("picture delete", code)
	}
	if _, err := os.Stat(filepath.Join(s.PictureDirectory(), third)); !os.IsNotExist(err) {
		t.Fatal("delete retained image")
	}
}

// @test WSH-006-AC12 WSH-006-AC14 CON-004-AC4 CON-004-AC5
func TestHouseholdPictureWriteAndRemoveFailures(t *testing.T) {
	dir, b, _ := managementFixture(t)
	a := b.handler.(*householdApp)
	s := authStore(dir)
	mock := &application{}
	mockPictureClient(mock, pictureFixture(t, 20, 10), nil)
	a.pictureClient = mock.pictureClient
	picture := filepath.Join(s.PictureDirectory(), "w1-v1.webp")
	if err := os.MkdirAll(picture, 0700); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(picture, "block"), []byte("block"), 0600)
	_, rev, _ := s.Load()
	if code := managementPost(t, dir, b, "/wishes/delete", url.Values{"wish": {"w1"}, "confirm": {"yes"}}); code != 500 {
		t.Fatal("invalid old picture deletion", code)
	}
	_, same, _ := s.Load()
	if rev != same {
		t.Fatal("remove failure changed data")
	}
	if os.Geteuid() != 0 {
		parent := filepath.Join(dir, "households/Family")
		if err := os.Chmod(parent, 0500); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(parent, 0700)
		if code := managementPost(t, dir, b, "/wishes/save", url.Values{"title": {"Permission denied picture"}, "picture_url": {"https://images.example/image.png"}}); code != 500 {
			t.Fatal("permission failure", code)
		}
		_, same, _ := s.Load()
		if rev != same {
			t.Fatal("permission failure wrote data")
		}
	}
}

// @test WSH-006-AC2 WSH-006-AC11 CON-004-AC4 ACC-006-AC6
func TestPictureDownloadRevalidatesSessionAndRevision(t *testing.T) {
	for _, scenario := range []string{"changed", "revoked"} {
		t.Run(scenario, func(t *testing.T) {
			dir, b, _ := managementFixture(t)
			a := b.handler.(*householdApp)
			s := authStore(dir)
			started := make(chan struct{})
			release := make(chan struct{})
			png := pictureFixture(t, 20, 10)
			a.pictureClient = &http.Client{Transport: pictureRoundTrip(func(r *http.Request) (*http.Response, error) {
				close(started)
				<-release
				return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(png)), Header: http.Header{}, Request: r}, nil
			})}
			_, rev, _ := s.Load()
			values := url.Values{"title": {"Delayed picture wish"}, "picture_url": {"https://images.example/slow.png"}, "revision": {rev}}
			done := make(chan int, 1)
			go func() { done <- b.post(t, "/wishes/save", values).Code }()
			<-started
			if err := s.Update(rev, func(h *data.Household) error {
				if scenario == "revoked" {
					h.Sessions = nil
				} else {
					h.Lists[0].Name = "Changed during download"
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			_, stable, _ := s.Load()
			close(release)
			status := <-done
			if scenario == "revoked" && status != 401 || scenario == "changed" && status != 409 {
				t.Fatal("inflight authorization", status)
			}
			_, after, _ := s.Load()
			if stable != after {
				t.Fatal("inflight action saved")
			}
		})
	}
}

// @test WSH-007-AC3 CON-004-AC4
func TestHouseholdWishEditsPreserveFutureFields(t *testing.T) {
	dir, b, _ := managementFixture(t)
	s := authStore(dir)
	_, rev, _ := s.Load()
	var node yaml.Node
	if err := node.Encode(map[string]string{"nested": "retained"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Update(rev, func(h *data.Household) error {
		h.Wishes[0].Extra = map[string]yaml.Node{"future_wish": node}
		h.Wishes[0].Claims[0].Extra = map[string]yaml.Node{"future_claim": node}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	before, _, _ := s.Load()
	if code := managementPost(t, dir, b, "/wishes/save", url.Values{"wish": {"w1"}, "title": {"Edited future fields"}}); code != 303 {
		t.Fatal("edit", code)
	}
	after, _, _ := s.Load()
	// YAML source positions move when editable fields are rewritten; compare the persisted values.
	beforeExtra, _ := yaml.Marshal(before.Wishes[0].Extra)
	afterExtra, _ := yaml.Marshal(after.Wishes[0].Extra)
	beforeClaims, _ := yaml.Marshal(before.Wishes[0].Claims)
	afterClaims, _ := yaml.Marshal(after.Wishes[0].Claims)
	if !bytes.Equal(beforeExtra, afterExtra) || !bytes.Equal(beforeClaims, afterClaims) || !reflect.DeepEqual(before.Wishes[0].Assignment, after.Wishes[0].Assignment) {
		t.Fatalf("editing discarded data: extra %s -> %s; claims %s -> %s", beforeExtra, afterExtra, beforeClaims, afterClaims)
	}
}
