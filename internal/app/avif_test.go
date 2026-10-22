package app

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"image"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func avifFixture(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata/avif", name))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// @test WSH-006-AC5 WSH-006-AC6 CON-006
func TestAVIFConversionPreservesAlphaAndUsesFirstFrame(t *testing.T) {
	for _, tc := range []struct {
		name          string
		width, height int
	}{{"alpha.avif", 640, 320}, {"animated.avif", 16, 12}} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := optimizePicture(avifFixture(t, tc.name))
			if err != nil {
				t.Fatalf("AVIF source rejected: %v", err)
			}
			img, format, err := image.Decode(bytes.NewReader(raw))
			if err != nil || format != "webp" || img.Bounds().Size() != image.Pt(tc.width, tc.height) {
				t.Fatal("optimized output format/size", format, err)
			}
			red, _, blue, alpha := img.At(1, 1).RGBA()
			if red <= blue {
				t.Fatal("wrong AVIF first-frame/color")
			}
			if tc.name == "alpha.avif" && (alpha < 32000 || alpha > 34000) {
				t.Fatal("AVIF transparency lost", alpha)
			}
			if bytes.Contains(raw, []byte("ANIM")) {
				t.Fatal("optimized image animated")
			}
		})
	}
}

// @test WSH-006-AC4 WSH-006-AC5 WSH-006-AC12
func TestAVIFRejectsMalformedAndOversizedSources(t *testing.T) {
	source := avifFixture(t, "alpha.avif")
	for _, raw := range [][]byte{source[:24], []byte("not an AVIF image")} {
		if _, err := optimizePicture(raw); !errors.Is(err, errPictureFormat) {
			t.Fatal("malformed source accepted", err)
		}
	}
	// Change metadata only: the pixel limit must reject this before pixel decode.
	huge := bytes.Clone(source)
	changed := false
	for from := 0; from < len(huge); {
		i := bytes.Index(huge[from:], []byte("ispe"))
		if i < 0 {
			break
		}
		i += from
		if i+16 > len(huge) {
			t.Fatal("fixture ispe truncated")
		}
		binary.BigEndian.PutUint32(huge[i+8:i+12], 5000)
		binary.BigEndian.PutUint32(huge[i+12:i+16], 4001)
		changed = true
		from = i + 16
	}
	if !changed {
		t.Fatal("fixture has no image dimensions")
	}
	if _, err := optimizePicture(huge); !errors.Is(err, errPicturePixels) {
		t.Fatal("oversized AVIF did not hit pixel limit", err)
	}
}

// @test WSH-006-AC2 WSH-006-AC5 WSH-006-AC7 WSH-006-AC12 WSH-006-AC13 WSH-006-AC14
func TestAVIFHouseholdDownloadSaveAndRollback(t *testing.T) {
	dir, parent, _ := managementFixture(t)
	a := parent.handler.(*householdApp)
	s := authStore(dir)
	calls := 0
	source := avifFixture(t, "alpha.avif")
	a.pictureClient = &http.Client{Transport: pictureRoundTrip(func(r *http.Request) (*http.Response, error) {
		calls++
		if !strings.Contains(r.Header.Get("Accept"), "image/avif") {
			t.Error("download does not advertise AVIF")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"image/avif"}}, ContentLength: int64(len(source)), Body: io.NopCloser(bytes.NewReader(source)), Request: r}, nil
	})}
	form := url.Values{"title": {"AVIF source picture"}, "picture_url": {"https://images.example/source.avif"}}
	if code := managementPost(t, dir, parent, "/wishes/save", form); code != 303 {
		t.Fatal("AVIF wish save", code)
	}
	h, _, _ := s.Load()
	wish := h.Wishes[len(h.Wishes)-1]
	first := wish.Picture
	original, err := readHouseholdPicture(s, first)
	if err != nil || !strings.HasSuffix(first, ".webp") {
		t.Fatal("optimized file missing", err)
	}
	form.Set("wish", wish.ID)
	form.Set("note", "Retain the local copy")
	if code := managementPost(t, dir, parent, "/wishes/save", form); code != 303 || calls != 1 {
		t.Fatal("unchanged AVIF source downloaded", code, calls)
	}
	_, revision, _ := s.Load()
	source = []byte("invalid avif")
	form.Set("picture_url", "https://images.example/invalid.avif")
	if code := managementPost(t, dir, parent, "/wishes/save", form); code != 400 {
		t.Fatal("invalid replacement", code)
	}
	_, after, _ := s.Load()
	saved, _ := readHouseholdPicture(s, first)
	if revision != after || !bytes.Equal(saved, original) {
		t.Fatal("failed AVIF save modified existing data/image")
	}
	form.Set("picture_url", "")
	if code := managementPost(t, dir, parent, "/wishes/save", form); code != 303 {
		t.Fatal("remove AVIF picture", code)
	}
	if _, err := os.Stat(filepath.Join(s.PictureDirectory(), first)); !os.IsNotExist(err) {
		t.Fatal("optimized image retained after removal")
	}
}

// Temporary manual checks may provide the user's downloaded sample without
// making shop-owned image bytes or a live network request part of the suite.
// @test WSH-006-AC5 WSH-006-AC6
func TestAVIFOptionalRealSample(t *testing.T) {
	if remote := os.Getenv("WISHMAS_AVIF_SAMPLE_URL"); remote != "" {
		output, err := fetchPictureWithClient(context.Background(), newPictureClient(), remote)
		if err != nil {
			t.Fatal("real HTTP download failed", err)
		}
		cfg, format, err := image.DecodeConfig(bytes.NewReader(output))
		if err != nil || format != "webp" || cfg.Width > 640 || cfg.Height > 640 {
			t.Fatal("real HTTP conversion failed", err)
		}
		t.Logf("Public HTTPS AVIF -> WebP %d bytes (%dx%d)", len(output), cfg.Width, cfg.Height)
	}

	path := os.Getenv("WISHMAS_AVIF_SAMPLE")
	if path == "" {
		return
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	output, err := optimizePicture(raw)
	if err != nil {
		t.Fatal(err)
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(output))
	if err != nil || format != "webp" || cfg.Width > 640 || cfg.Height > 640 {
		t.Fatal("real AVIF sample conversion failed", err)
	}
	t.Logf("AVIF source %d bytes -> WebP %d bytes (%dx%d)", len(raw), len(output), cfg.Width, cfg.Height)
}
