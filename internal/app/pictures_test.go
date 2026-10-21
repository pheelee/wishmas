package app

import (
	"net/url"
	"strings"
	"testing"
)

// @test WSH-006-AC1 WSH-006-AC3 WSH-006-AC12
func TestPictureInputAndPrivateSource(t *testing.T) {
	a := newLegacy(t.TempDir())
	parent, _ := setupHousehold(t, a, "Europe/Zurich")
	addTestPerson(t, a, parent, "Anna")
	page := request(t, a, "GET", wishLink(t, parent, "Anna", "2026"), nil)
	if !strings.Contains(page.Body.String(), `name="picture_url"`) {
		t.Fatal("image URL field missing")
	}
	response := request(t, a, "POST", personAction(t, parent, "/wishes/save"), url.Values{"person": {"Anna"}, "year": {"2026"}, "title": {"Picture train"}, "picture_url": {"http://127.0.0.1/private.png"}})
	if response.Code != 400 {
		t.Fatalf("private image source was not rejected: %d", response.Code)
	}
}
