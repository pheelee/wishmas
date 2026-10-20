package app

import (
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func deletionFixture(t *testing.T) (*application, string, string, string, string) {
	t.Helper()
	a, dir, parent, family := newPeopleApp(t)
	addTestPerson(t, a, parent, "Anna")
	writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes:\n  - title: Wooden train\n    claims: [{name: Auntie}, {name: Uncle}]\n  - title: Another train\n    claims: [{name: Cousin}]\n")
	page := request(t, a, "GET", wishLink(t, parent, "Anna", "2026"), nil)
	if page.Code != 200 {
		t.Fatalf("list: %d", page.Code)
	}
	return a, dir, parent, family, field(t, page.Body.String(), "wish")
}

// @test WSH-002-AC1 WSH-002-AC4 WSH-002-AC6 WSH-002-AC7
func TestWishDeletionConfirmationCancellationAndRemoval(t *testing.T) {
	a, dir, parent, _, token := deletionFixture(t)
	form := url.Values{"person": {"Anna"}, "year": {"2026"}, "wish": {token}}
	before := snapshotFiles(t, dir)
	confirmation := request(t, a, "POST", personAction(t, parent, "/wishes/delete"), form)
	if confirmation.Code != 200 {
		t.Fatalf("confirmation: %d", confirmation.Code)
	}
	for _, want := range []string{"Wooden train", "Permanently delete wish", "Cancel", "attached claims"} {
		if !strings.Contains(confirmation.Body.String(), want) {
			t.Fatalf("missing %s", want)
		}
	}
	if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("request deleted before confirmation")
	}
	cancelled := request(t, a, "GET", wishLink(t, parent, "Anna", "2026"), nil)
	if cancelled.Code != 200 || !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("cancel changed data")
	}
	form.Set("confirm", "no")
	got := request(t, a, "POST", personAction(t, parent, "/wishes/delete"), form)
	if got.Code != 200 || !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("unconfirmed value deleted data")
	}
	form.Set("confirm", "yes")
	got = request(t, a, "POST", personAction(t, parent, "/wishes/delete"), form)
	if got.Code != 303 || got.Header().Get("Location") != personURL(strings.TrimPrefix(parent, "/?code="), "Anna", 2026) {
		t.Fatalf("delete: %d", got.Code)
	}
	raw, _ := os.ReadFile(listFile(dir, "2026", "Anna"))
	for _, removed := range []string{"Wooden train", "Auntie", "Uncle"} {
		if strings.Contains(string(raw), removed) {
			t.Fatalf("retained %s", removed)
		}
	}
	for _, kept := range []string{"Another train", "Cousin"} {
		if !strings.Contains(string(raw), kept) {
			t.Fatalf("lost %s", kept)
		}
	}
	// A repeated confirmation must not delete the next wish after indexes shift.
	before = snapshotFiles(t, dir)
	got = request(t, a, "POST", personAction(t, parent, "/wishes/delete"), form)
	if got.Code != 409 || !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatal("replayed delete changed data")
	}
	// Unclaimed wishes follow the same confirmation flow and leave an empty list.
	writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes: [{title: Another train}]\n")
	page := request(t, a, "GET", wishLink(t, parent, "Anna", "2026"), nil)
	form.Set("wish", field(t, page.Body.String(), "wish"))
	got = request(t, a, "POST", personAction(t, parent, "/wishes/delete"), form)
	if got.Code != 303 {
		t.Fatal(got.Code)
	}
	page = request(t, a, "GET", wishLink(t, parent, "Anna", "2026"), nil)
	if !strings.Contains(page.Body.String(), "No wishes yet.") {
		t.Fatal("final wish not removed")
	}
}

// @test WSH-002-AC2 WSH-002-AC5 WSH-002-AC6 CON-004
func TestWishDeletionGuards(t *testing.T) {
	for _, scenario := range []string{"family", "old-year", "query-year", "year-switch", "archived", "missing-person", "empty-token", "changed", "reordered", "new-claim", "invalid", "revoked", "get"} {
		t.Run(scenario, func(t *testing.T) {
			a, dir, parent, family, token := deletionFixture(t)
			form := url.Values{"person": {"Anna"}, "year": {"2026"}, "wish": {token}, "confirm": {"yes"}}
			link := parent
			method := "POST"
			status := 404
			switch scenario {
			case "family":
				link = family
			case "old-year":
				form.Set("year", "2025")
			case "query-year":
				link += "&year=2025"
			case "year-switch":
				a.now = func() time.Time { return time.Date(2027, 1, 2, 0, 0, 0, 0, time.UTC) }
			case "archived":
				writeTestYAML(t, peopleFile(dir), "people: [{name: Anna, archived: true}]\n")
			case "missing-person":
				form.Set("person", "Bobby")
			case "empty-token":
				form.Del("wish")
				status = 409
			case "changed":
				writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes: [{title: Updated train}]\n")
				status = 409
			case "reordered":
				writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes:\n  - title: Another train\n    claims: [{name: Cousin}]\n  - title: Wooden train\n    claims: [{name: Auntie}, {name: Uncle}]\n")
				status = 409
			case "new-claim":
				writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes:\n  - title: Wooden train\n    claims: [{name: Auntie}, {name: Uncle}, {name: Third}]\n")
				status = 409
			case "invalid":
				writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes: [broken]\n")
				status = 500
			case "revoked":
				got := request(t, a, "POST", personAction(t, parent, "/access/regenerate"), url.Values{"role": {"parent"}})
				if got.Code != 303 {
					t.Fatalf("regenerate %d", got.Code)
				}
			case "get":
				method = "GET"
			}
			before := snapshotFiles(t, dir)
			got := request(t, a, method, personAction(t, link, "/wishes/delete"), form)
			if got.Code != status {
				t.Fatalf("got %d, want %d", got.Code, status)
			}
			if !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
				t.Fatal("denied deletion changed data")
			}
			// Confirmation requests must obey the same guards as final submissions.
			if method == "POST" {
				form.Del("confirm")
				got = request(t, a, method, personAction(t, link, "/wishes/delete"), form)
				if got.Code != status || !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
					t.Fatalf("confirmation guard: %d", got.Code)
				}
			}
		})
	}
}

// @test WSH-002-AC4 WSH-002-AC6
func TestRepeatedDeletionDoesNotRemoveIdenticalNextWish(t *testing.T) {
	a, dir, parent, _, _ := deletionFixture(t)
	writeTestYAML(t, listFile(dir, "2026", "Anna"), "wishes: [{title: Same train}, {title: Same train}]\n")
	page := request(t, a, "GET", wishLink(t, parent, "Anna", "2026"), nil)
	form := url.Values{"person": {"Anna"}, "year": {"2026"}, "wish": {field(t, page.Body.String(), "wish")}, "confirm": {"yes"}}
	if got := request(t, a, "POST", personAction(t, parent, "/wishes/delete"), form); got.Code != 303 {
		t.Fatal(got.Code)
	}
	before := snapshotFiles(t, dir)
	got := request(t, a, "POST", personAction(t, parent, "/wishes/delete"), form)
	if got.Code != 409 || !reflect.DeepEqual(before, snapshotFiles(t, dir)) {
		t.Fatalf("replay removed identical next wish: %d", got.Code)
	}
}
