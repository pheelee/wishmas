package app

import (
	data "github.com/pheelee/wishmas/internal/household"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// @test CLM-005-AC1 CLM-006-AC1 CLM-006-AC3 CLM-008-AC1
func TestHouseholdClaimToggleUsesLinkIdentity(t *testing.T) {
	dir, b, h := managementFixture(t)
	user := newAuthBrowser(b.handler)
	path := "/claims/toggle?code=" + h.Lists[0].UserLinks[0].Code
	page := user.get(t, "/?code="+h.Lists[0].UserLinks[0].Code)
	if !strings.Contains(page.Body.String(), `action="/claims/toggle?code=`) {
		t.Fatal("user cards have no claim toggle")
	}
	if code := managementPost(t, dir, user, path, url.Values{"wish": {"w4"}, "mode": {"add"}, "name": {"Forged Name"}}); code != 303 {
		t.Fatal("claim", code)
	}
	after, _, err := authStore(dir).Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Wishes[3].Claims) != 2 || after.Wishes[3].Claims[1].UserID != "u1" {
		t.Fatal("claim did not use link user ID")
	}
	if code := managementPost(t, dir, user, path, url.Values{"wish": {"w4"}, "mode": {"remove"}}); code != 303 {
		t.Fatal("remove", code)
	}
	after, _, _ = authStore(dir).Load()
	if len(after.Wishes[3].Claims) != 1 || after.Wishes[3].Claims[0].UserID != "u2" {
		t.Fatal("removal changed other user")
	}
	// Switching links in this same browser switches identity, without a name/cookie change.
	bobPath := "/claims/toggle?code=" + h.Lists[0].UserLinks[1].Code
	user.get(t, "/?code="+h.Lists[0].UserLinks[1].Code)
	if status := managementPost(t, dir, user, bobPath, url.Values{"wish": {"w1"}, "mode": {"remove"}}); status != 303 {
		t.Fatal("same-browser identity switch", status)
	}
	after, _, _ = authStore(dir).Load()
	if len(after.Wishes[0].Claims) != 1 || after.Wishes[0].Claims[0].UserID != "u1" {
		t.Fatal("previous link identity carried over")
	}
}

// @test CLM-005 CLM-006-AC2 CLM-006-AC4 CLM-006-AC5 CLM-007 CLM-008-AC1 CLM-008-AC2 CLM-008-AC3
func TestHouseholdClaimsAdminSelectionPrivacyAndRename(t *testing.T) {
	dir, parent, h := managementFixture(t)
	s := authStore(dir)
	user := newAuthBrowser(parent.handler)
	path := "/claims/toggle?code=" + h.Lists[0].UserLinks[0].Code
	user.get(t, "/?code="+h.Lists[0].UserLinks[0].Code)
	for _, tc := range []struct {
		values url.Values
		want   int
	}{{url.Values{"wish": {"w4"}, "mode": {"add"}}, 400}, {url.Values{"wish": {"w3"}, "mode": {"add"}, "user": {"u1"}}, 400}, {url.Values{"wish": {"w4"}, "mode": {"add"}, "user": {"r1"}}, 400}, {url.Values{"wish": {"w4"}, "mode": {"add"}, "user": {"u2"}}, 409}} {
		if code := managementPost(t, dir, parent, "/claims/toggle", tc.values); code != tc.want {
			t.Fatal("admin validation", code, tc.want)
		}
	}
	if code := managementPost(t, dir, parent, "/claims/toggle", url.Values{"wish": {"w4"}, "mode": {"add"}, "user": {"u1"}}); code != 303 {
		t.Fatal("admin selected claim", code)
	}
	body := user.get(t, "/?code="+h.Lists[0].UserLinks[0].Code).Body.String()
	if !strings.Contains(body, `aria-pressed="true"`) || strings.Contains(body, "Other User") || strings.Contains(body, `value="u2"`) || strings.Contains(body, "Unassigned private wish") {
		t.Fatal("user identity/privacy")
	}
	if code := managementPost(t, dir, user, path, url.Values{"wish": {"w4"}, "mode": {"remove"}, "user": {"u2"}}); code != 403 {
		t.Fatal("forged removal", code)
	}
	if code := managementPost(t, dir, user, path, url.Values{"wish": {"w4"}, "mode": {"remove-other"}, "user": {"u2"}}); code != 403 {
		t.Fatal("privileged mode", code)
	}
	if code := managementPost(t, dir, parent, "/manage/user/rename", url.Values{"id": {"u1"}, "name": {"Renamed Giver"}}); code != 303 {
		t.Fatal("rename")
	}
	after, _, _ := s.Load()
	if len(after.Wishes[3].Claims) != 2 || after.Wishes[3].Claims[1].UserID != "u1" {
		t.Fatal("rename cleared claims")
	}
	body = user.get(t, "/?code="+h.Lists[0].UserLinks[0].Code).Body.String()
	if !strings.Contains(body, "Renamed Giver") || strings.Contains(body, "Other User") {
		t.Fatal("renamed ownership/privacy")
	}
	body = parent.get(t, "/").Body.String()
	if !strings.Contains(body, `class="claimant-name">Renamed Giver`) {
		t.Fatal("admin existing claim name not updated")
	}
	// Each new admin form starts with an empty required user option, never a browser acting user.
	if strings.Contains(body, `value="u1" selected`) || !strings.Contains(body, `name="user" required`) {
		t.Fatal("admin acting identity persisted")
	}
	// Opening a link in the admin browser still restricts removal to that link user.
	if code := managementPost(t, dir, parent, path, url.Values{"wish": {"w4"}, "mode": {"remove"}}); code != 303 {
		t.Fatal("admin cookie user view")
	}
	after, _, _ = s.Load()
	if len(after.Wishes[3].Claims) != 1 || after.Wishes[3].Claims[0].UserID != "u2" {
		t.Fatal("link identity ignored")
	}
	if code := managementPost(t, dir, parent, "/claims/toggle", url.Values{"wish": {"w4"}, "mode": {"remove"}, "user": {"u2"}}); code != 303 {
		t.Fatal("admin remove another")
	}
	// Admin removal of preserved unassigned claims needs no acting user selection.
	if code := managementPost(t, dir, parent, "/claims/toggle", url.Values{"wish": {"w3"}, "mode": {"remove"}, "user": {"u2"}}); code != 303 {
		t.Fatal("unassigned removal")
	}
	after, _, _ = s.Load()
	if len(after.Wishes[3].Claims) != 0 || len(after.Wishes[2].Claims) != 1 || after.Wishes[2].Claims[0].UserID != "u1" {
		t.Fatal("last/other claim cleanup")
	}
}

// @test CLM-006-AC6 CLM-007-AC2 CLM-008-AC1 ACC-007-AC6 CON-007-AC4
func TestHouseholdClaimsStaleScopeRevocationAndGuards(t *testing.T) {
	dir, parent, h := managementFixture(t)
	s := authStore(dir)
	user := newAuthBrowser(parent.handler)
	code := h.Lists[0].UserLinks[0].Code
	path := "/claims/toggle?code=" + code
	user.get(t, "/?code="+code)
	_, stale, _ := s.Load()
	if status := managementPost(t, dir, user, path, url.Values{"wish": {"w4"}, "mode": {"add"}}); status != 303 {
		t.Fatal("claim")
	}
	_, current, _ := s.Load()
	w := user.post(t, path, url.Values{"wish": {"w4"}, "mode": {"add"}, "revision": {stale}})
	if w.Code != 409 || !strings.Contains(w.Body.String(), `role="alert"`) || !strings.Contains(w.Body.String(), `aria-pressed="true"`) || strings.Contains(w.Body.String(), "Other User") {
		t.Fatal("stale result/privacy", w.Code)
	}
	_, after, _ := s.Load()
	if current != after {
		t.Fatal("stale action inverted claim")
	}
	// Duplicate at a fresh revision remains an error; removal replay remains an error as well.
	if status := managementPost(t, dir, user, path, url.Values{"wish": {"w4"}, "mode": {"add"}}); status != 409 {
		t.Fatal("duplicate accepted")
	}
	for _, id := range []string{"w2", "w3", "absent"} {
		_, before, _ := s.Load()
		response := user.post(t, path, url.Values{"wish": {id}, "mode": {"add"}, "revision": {before}, "context": {"l2"}})
		if response.Code != 404 || strings.Contains(response.Body.String(), "Other list private wish") || strings.Contains(response.Body.String(), "Unassigned private wish") || strings.Contains(response.Body.String(), "Other User") {
			t.Fatal("scope/error privacy", response.Code)
		}
		_, after, _ := s.Load()
		if before != after {
			t.Fatal("scope failure wrote")
		}
	}
	if status := managementPost(t, dir, user, path, url.Values{"wish": {"w4"}, "mode": {"remove"}}); status != 303 {
		t.Fatal("remove")
	}
	if status := managementPost(t, dir, user, path, url.Values{"wish": {"w4"}, "mode": {"remove"}}); status != 409 {
		t.Fatal("removal replay inverted")
	}
	_, rev, _ := s.Load()
	bad := url.Values{"wish": {"w1"}, "mode": {"remove"}, "revision": {rev}, "csrf": {"bad"}}
	if w := user.send(t, "POST", path, bad); w.Code != 403 {
		t.Fatal("csrf")
	}
	user.headers.Set("Origin", "https://attacker.test")
	if w := user.post(t, path, url.Values{"wish": {"w1"}, "mode": {"remove"}, "revision": {rev}}); w.Code != 403 {
		t.Fatal("origin")
	}
	user.headers.Del("Origin")
	if w := user.get(t, path+"&wish=w1&mode=remove"); w.Code != 403 {
		t.Fatal("GET mutation")
	}
	if status := managementPost(t, dir, parent, "/manage/link/regenerate", url.Values{"id": {"l1"}, "user": {"u1"}}); status != 303 {
		t.Fatal("regenerate")
	}
	_, before, _ := s.Load()
	if w := parent.post(t, path, url.Values{"wish": {"w1"}, "mode": {"remove"}, "revision": {before}}); !strings.Contains(w.Body.String(), `name="password"`) {
		t.Fatal("revoked link acquired parent permissions")
	}
	_, after, _ = s.Load()
	if before != after {
		t.Fatal("revoked request wrote")
	}
}

// @test CLM-005-AC4 CLM-006-AC2 CLM-006-AC6 CON-004-AC4
func TestHouseholdClaimsConcurrentSameUserAndFailedWrites(t *testing.T) {
	dir, parent, h := managementFixture(t)
	s := authStore(dir)
	path := "/claims/toggle?code=" + h.Lists[0].UserLinks[0].Code
	first := newAuthBrowser(parent.handler)
	second := newAuthBrowser(New(dir))
	for _, b := range []*authBrowser{first, second} {
		b.get(t, "/?code="+h.Lists[0].UserLinks[0].Code)
	}
	_, rev, _ := s.Load()
	start := make(chan struct{})
	results := make(chan int, 2)
	for _, b := range []*authBrowser{first, second} {
		go func(b *authBrowser) {
			<-start
			results <- b.post(t, path, url.Values{"wish": {"w4"}, "mode": {"add"}, "revision": {rev}}).Code
		}(b)
	}
	close(start)
	one, two := <-results, <-results
	if !((one == 303 && two == 409) || (one == 409 && two == 303)) {
		t.Fatal("concurrent results", one, two)
	}
	after, _, _ := s.Load()
	if len(after.Wishes[3].Claims) != 2 {
		t.Fatal("concurrent duplicate")
	}
	if os.Geteuid() != 0 {
		root := filepath.Join(dir, "households/Family")
		if err := os.Chmod(root, 0500); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(root, 0700)
		_, before, _ := s.Load()
		w := first.post(t, path, url.Values{"wish": {"w1"}, "mode": {"remove"}, "revision": {before}})
		if w.Code != 500 || !strings.Contains(w.Body.String(), `role="alert"`) || strings.Contains(w.Body.String(), "Other User") {
			t.Fatal("write error/privacy", w.Code)
		}
		_, stable, _ := s.Load()
		if before != stable {
			t.Fatal("write failure changed claims")
		}
	}
}

// @test CLM-007-AC5 CLM-008-AC4 CLM-008-AC5 CLM-008-AC6 CLM-008-AC7 WLS-010-AC5
func TestHouseholdClaimLifecycleIntegration(t *testing.T) {
	for _, operation := range []string{"user", "list", "selection", "recipient", "wish"} {
		t.Run(operation, func(t *testing.T) {
			dir, parent, h := managementFixture(t)
			s := authStore(dir)
			path := "/claims/toggle?code=" + h.Lists[0].UserLinks[0].Code
			user := newAuthBrowser(parent.handler)
			user.get(t, "/?code="+h.Lists[0].UserLinks[0].Code)
			// Create a real new claim, edit and move it without clearing either claimant.
			if status := managementPost(t, dir, user, path, url.Values{"wish": {"w4"}, "mode": {"add"}}); status != 303 {
				t.Fatal("real claim")
			}
			if status := managementPost(t, dir, parent, "/wishes/save", url.Values{"wish": {"w4"}, "title": {"Edited claim wish"}}); status != 303 {
				t.Fatal("edit")
			}
			for _, assignment := range []string{"l2:r1", "", "l1:r2"} {
				if status := managementPost(t, dir, parent, "/wishes/assign", url.Values{"wish": {"w4"}, "assignment": {assignment}}); status != 303 {
					t.Fatal("move", status)
				}
				current, _, _ := s.Load()
				if len(current.Wishes[3].Claims) != 2 {
					t.Fatal("move cleared real claims")
				}
			}
			values := url.Values{}
			route := ""
			switch operation {
			case "user":
				route = "/manage/user/delete"
				values.Set("id", "u1")
			case "list":
				route = "/manage/list/delete"
				values.Set("id", "l1")
			case "selection":
				route = "/manage/selection/remove"
				values.Set("id", "l1")
				values.Set("recipient", "r2")
			case "recipient":
				route = "/manage/recipient/delete"
				values.Set("id", "r2")
			case "wish":
				route = "/wishes/delete"
				values.Set("wish", "w4")
				values.Set("confirm", "yes")
			}
			if status := managementPost(t, dir, parent, route, values); status != 303 {
				t.Fatal("lifecycle delete", status)
			}
			after, _, _ := s.Load()
			if operation == "wish" {
				if slices.ContainsFunc(after.Wishes, func(w data.Wish) bool { return w.ID == "w4" }) {
					t.Fatal("deleted wish retained")
				}
			} else if operation == "user" {
				for _, w := range after.Wishes {
					if slices.ContainsFunc(w.Claims, func(c data.Claim) bool { return c.UserID == "u1" }) {
						t.Fatal("user claim retained")
					}
				}
			} else if after.Wishes[3].Assignment != nil || len(after.Wishes[3].Claims) != 0 {
				t.Fatal("destructive unassignment retained claim")
			}
			if len(after.Wishes[1].Claims) != 1 && operation != "user" {
				t.Fatal("other list claims lost")
			}
		})
	}
}

// @test CLM-005-AC1 CLM-006-AC6 CLM-007-AC4 CON-004-AC3
func TestHouseholdClaimsIgnoreBrowserNamesAndRejectInvalidManualData(t *testing.T) {
	dir, parent, h := managementFixture(t)
	s := authStore(dir)
	user := newAuthBrowser(parent.handler)
	user.cookies["wishmas_claim_name"] = url.QueryEscape("Other User")
	code := h.Lists[0].UserLinks[0].Code
	user.get(t, "/?code="+code)
	if status := managementPost(t, dir, user, "/claims/toggle?code="+code, url.Values{"wish": {"w4"}, "mode": {"add"}, "claimant": {"Other User"}, "name": {"Other User"}}); status != 303 {
		t.Fatal("legacy identity affected claim", status)
	}
	after, _, _ := s.Load()
	if after.Wishes[3].Claims[1].UserID != "u1" {
		t.Fatal("posted/browser name changed identity")
	}
	cfg := filepath.Join(dir, "households/Family/household.yaml")
	original, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	invalid := strings.Replace(string(original), "user_id: u1", "user_id: nonexistent", 1)
	if invalid == string(original) {
		t.Fatal("fixture mutation did not apply")
	}
	if err = os.WriteFile(cfg, []byte(invalid), 0600); err != nil {
		t.Fatal(err)
	}
	response := user.post(t, "/claims/toggle?code="+code, url.Values{"wish": {"w4"}, "mode": {"remove"}, "revision": {"old"}})
	if response.Code != 500 || response.Header().Get("Content-Language") != "en-US" {
		t.Fatal("invalid configuration guard", response.Code)
	}
	persisted, _ := os.ReadFile(cfg)
	if string(persisted) != invalid {
		t.Fatal("invalid manual data rewritten")
	}
}
