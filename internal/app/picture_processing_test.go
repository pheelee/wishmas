package app

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"testing"

	"github.com/gen2brain/webp"
)

type pictureRoundTrip func(*http.Request) (*http.Response, error)

func (f pictureRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func pictureFixture(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, color.NRGBA{200, 80, 30, 128})
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func mockPictureClient(a *householdApp, raw []byte, calls *int) {
	a.pictureClient = &http.Client{CheckRedirect: pictureRedirect, Transport: pictureRoundTrip(func(r *http.Request) (*http.Response, error) {
		if calls != nil {
			*calls++
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(raw)), Header: make(http.Header), ContentLength: int64(len(raw)), Request: r}, nil
	})}
}

// @test WSH-006-AC3: literal/private, mapped IPv4, special IPv6 and actual dial checks.
func TestPicturePublicAddresses(t *testing.T) {
	for _, raw := range []string{"127.0.0.1", "10.1.2.3", "172.16.1.1", "192.168.2.1", "169.254.169.254", "100.64.0.1", "0.0.0.0", "192.0.2.1", "198.18.1.1", "224.0.0.1", "::1", "::ffff:127.0.0.1", "fc00::1", "fe80::1", "64:ff9b::a00:1", "2001:db8::1", "2002:7f00:1::"} {
		if publicPictureIP(netip.MustParseAddr(raw)) {
			t.Errorf("allowed private/special %s", raw)
		}
	}
	for _, raw := range []string{"8.8.8.8", "1.1.1.1", "2606:4700:4700::1111"} {
		if !publicPictureIP(netip.MustParseAddr(raw)) {
			t.Errorf("denied public %s", raw)
		}
	}
	for _, raw := range []string{"file:///etc/passwd", "data:image/png,x", "https:///x", "ftp://images.example/x"} {
		if _, err := pictureURL(raw); !errors.Is(err, errPictureURL) {
			t.Errorf("URL accepted: %s", raw)
		}
	}
	if conn, err := publicPictureDial(context.Background(), "tcp", "localhost:80"); !errors.Is(err, errPicturePrivate) {
		if conn != nil {
			conn.Close()
		}
		t.Fatalf("private DNS destination: %v", err)
	}
}

// @test WSH-006-AC4 WSH-006-AC12
func TestPictureDownloadLimits(t *testing.T) {
	a := New(t.TempDir()).(*householdApp)
	source := pictureFixture(t, 2, 2)
	for _, tc := range []struct {
		name string
		size int64
		body []byte
		want error
	}{
		{"content-length", maxPictureBytes + 1, nil, errPictureBytes},
		{"streamed", -1, bytes.Repeat([]byte("x"), int(maxPictureBytes)+1), errPictureBytes},
		{"invalid", 3, []byte("svg"), errPictureFormat},
		{"valid", int64(len(source)), source, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a.pictureClient = &http.Client{Transport: pictureRoundTrip(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, ContentLength: tc.size, Body: io.NopCloser(bytes.NewReader(tc.body)), Header: make(http.Header)}, nil
			})}
			_, err := fetchPictureWithClient(context.Background(), a.pictureClient, "https://images.example/picture")
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
	for _, redirects := range []int{5, 6} {
		a.pictureClient = &http.Client{CheckRedirect: pictureRedirect, Transport: pictureRoundTrip(func(r *http.Request) (*http.Response, error) {
			n, _ := strconv.Atoi(r.URL.Query().Get("n"))
			h := make(http.Header)
			status := 200
			body := source
			if n < redirects {
				status = 302
				h.Set("Location", "https://images.example/?n="+strconv.Itoa(n+1))
				body = nil
			}
			return &http.Response{StatusCode: status, Header: h, Body: io.NopCloser(bytes.NewReader(body)), Request: r}, nil
		})}
		_, err := fetchPictureWithClient(context.Background(), a.pictureClient, "https://images.example/?n=0")
		if redirects == 5 && err != nil || redirects == 6 && !errors.Is(err, errPictureRedirect) {
			t.Fatalf("redirect count %d: %v", redirects, err)
		}
	}
	a.pictureClient = &http.Client{CheckRedirect: pictureRedirect, Transport: pictureRoundTrip(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 302, Header: http.Header{"Location": []string{"http://127.0.0.1/secret"}}, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
	})}
	if _, err := fetchPictureWithClient(context.Background(), a.pictureClient, "https://images.example/x"); !errors.Is(err, errPicturePrivate) {
		t.Fatalf("private redirect %v", err)
	}
	a.pictureClient = &http.Client{Transport: pictureRoundTrip(func(r *http.Request) (*http.Response, error) {
		deadline, ok := r.Context().Deadline()
		if !ok {
			t.Error("no total deadline")
		}
		_ = deadline
		return nil, context.DeadlineExceeded
	})}
	if _, err := fetchPictureWithClient(context.Background(), a.pictureClient, "https://images.example/x"); !errors.Is(err, errPictureTimeout) {
		t.Fatalf("timeout %v", err)
	}
}

// @test WSH-006-AC4 WSH-006-AC5 WSH-006-AC6
func TestPictureOptimization(t *testing.T) {
	for _, size := range []image.Point{{1280, 640}, {320, 100}, {640, 1280}} {
		raw, err := optimizePicture(pictureFixture(t, size.X, size.Y))
		if err != nil {
			t.Fatal(err)
		}
		img, format, err := image.Decode(bytes.NewReader(raw))
		if err != nil || format != "webp" {
			t.Fatalf("output: %s %v", format, err)
		}
		want := size
		if size.X > 640 {
			want = image.Pt(640, 320)
		} else if size.Y > 640 {
			want = image.Pt(320, 640)
		}
		if img.Bounds().Size() != want {
			t.Fatalf("dimensions %v want %v", img.Bounds(), want)
		}
		_, _, _, alpha := img.At(0, 0).RGBA()
		if alpha < 32000 || alpha > 34000 {
			t.Fatalf("alpha lost: %d", alpha)
		}
	}
	// Change only PNG IHDR and its CRC: rejection must precede decoding pixels.
	huge := pictureFixture(t, 1, 1)
	binary.BigEndian.PutUint32(huge[16:20], 5000)
	binary.BigEndian.PutUint32(huge[20:24], 4001)
	binary.BigEndian.PutUint32(huge[29:33], crc32.ChecksumIEEE(huge[12:29]))
	if _, err := optimizePicture(huge); !errors.Is(err, errPicturePixels) {
		t.Fatalf("resolution: %v", err)
	}
	if _, err := optimizePicture([]byte(`<svg xmlns="http://www.w3.org/2000/svg"/>`)); !errors.Is(err, errPictureFormat) {
		t.Fatalf("SVG: %v", err)
	}
	var jpegBytes bytes.Buffer
	jpeg.Encode(&jpegBytes, image.NewRGBA(image.Rect(0, 0, 2, 3)), nil)
	if _, err := optimizePicture(jpegBytes.Bytes()); err != nil {
		t.Fatalf("JPEG: %v", err)
	}
	red := image.NewPaletted(image.Rect(0, 0, 4, 4), color.Palette{color.RGBA{255, 0, 0, 255}})
	blue := image.NewPaletted(image.Rect(0, 0, 4, 4), color.Palette{color.RGBA{0, 0, 255, 255}})
	var animated bytes.Buffer
	gif.EncodeAll(&animated, &gif.GIF{Image: []*image.Paletted{red, blue}, Delay: []int{10, 10}})
	output, err := optimizePicture(animated.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	still, _, err := image.Decode(bytes.NewReader(output))
	if err != nil {
		t.Fatal(err)
	}
	r, _, b, _ := still.At(1, 1).RGBA()
	if r < b {
		t.Fatal("wrong GIF frame")
	}
	if bytes.Contains(output, []byte("ANIM")) {
		t.Fatal("output animated")
	}
	if _, err := optimizePicture(output); err != nil {
		t.Fatalf("WebP source: %v", err)
	}
	// Animated WebP uses the same first-frame-only decoder contract.
	var animation bytes.Buffer
	if err := webp.EncodeAll(&animation, &webp.WEBP{Image: []image.Image{red, blue}, Delay: []int{100, 100}}, webp.Options{Quality: 80}); err != nil {
		t.Fatal(err)
	}
	converted, err := optimizePicture(animation.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	first, _, err := image.Decode(bytes.NewReader(converted))
	if err != nil {
		t.Fatal(err)
	}
	r, _, b, _ = first.At(1, 1).RGBA()
	if r < b || bytes.Contains(converted, []byte("ANIM")) {
		t.Fatal("animated WebP first frame not retained")
	}
}
