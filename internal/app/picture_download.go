package app

import (
	"bytes"
	"context"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	_ "github.com/gen2brain/avif"
	"github.com/gen2brain/webp"
	"golang.org/x/image/draw"
)

const maxPictureBytes int64 = 10 * 1024 * 1024
const maxPicturePixels = 20_000_000

var (
	errPictureURL      = errors.New("I08")
	errPicturePrivate  = errors.New("I09")
	errPictureDownload = errors.New("I10")
	errPictureTimeout  = errors.New("I11")
	errPictureRedirect = errors.New("I12")
	errPictureBytes    = errors.New("I13")
	errPicturePixels   = errors.New("I14")
	errPictureFormat   = errors.New("I15")
	errPictureSave     = errors.New("I16")
	errPictureRemove   = errors.New("I17")
)

// @impl WSH-006-AC1 WSH-006-AC3 WSH-006-AC4 CON-006
func pictureURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" || u.Opaque != "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, errPictureURL
	}
	if ip, err := netip.ParseAddr(u.Hostname()); err == nil && !publicPictureIP(ip) {
		return nil, errPicturePrivate
	}
	return u, nil
}

var nonPublicPictureNetworks = func() []netip.Prefix {
	var result []netip.Prefix
	for _, prefix := range []string{"0.0.0.0/8", "100.64.0.0/10", "192.0.0.0/24", "192.0.2.0/24", "198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24", "240.0.0.0/4", "2001::/23", "2001:db8::/32", "2002::/16", "3fff::/20"} {
		result = append(result, netip.MustParsePrefix(prefix))
	}
	return result
}()

func publicPictureIP(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	// Restrict IPv6 to allocated global unicast space; excludes translation,
	// IPv4-compatible, site-local and other special-purpose address families.
	if ip.Is6() && !netip.MustParsePrefix("2000::/3").Contains(ip) {
		return false
	}
	for _, prefix := range nonPublicPictureNetworks {
		if prefix.Contains(ip) {
			return false
		}
	}
	return true
}

// Pin the actual connection to a checked address. Never validate a DNS answer
// then let a second DNS lookup choose a different (potentially private) peer.
func publicPictureDial(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, errPictureURL
	}
	addresses, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	if len(addresses) == 0 {
		return nil, errPictureDownload
	}
	for _, ip := range addresses {
		if !publicPictureIP(ip) {
			return nil, errPicturePrivate
		}
	}
	var last error
	for _, ip := range addresses {
		conn, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if err == nil {
			return conn, nil
		}
		last = err
	}
	return nil, last
}
func pictureRedirect(req *http.Request, via []*http.Request) error {
	if len(via) > 5 {
		return errPictureRedirect
	}
	_, err := pictureURL(req.URL.String())
	return err
}
func newPictureClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second, CheckRedirect: pictureRedirect, Transport: &http.Transport{Proxy: nil, DialContext: publicPictureDial, TLSHandshakeTimeout: 10 * time.Second, ResponseHeaderTimeout: 30 * time.Second, DisableKeepAlives: true, DisableCompression: true}}
}

// @impl WSH-006-AC2 WSH-006-AC3 WSH-006-AC4 WSH-006-AC12
func fetchPictureWithClient(ctx context.Context, client *http.Client, raw string) ([]byte, error) {
	u, err := pictureURL(raw)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errPictureURL
	}
	req.Header.Set("Accept", "image/jpeg, image/png, image/webp, image/gif, image/avif")
	response, err := client.Do(req)
	if err != nil {
		if errors.Is(err, errPicturePrivate) {
			return nil, errPicturePrivate
		}
		if errors.Is(err, errPictureRedirect) {
			return nil, errPictureRedirect
		}
		if errors.Is(err, errPictureURL) {
			return nil, errPictureURL
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, errPictureTimeout
		}
		var timeout net.Error
		if errors.As(err, &timeout) && timeout.Timeout() {
			return nil, errPictureTimeout
		}
		return nil, errPictureDownload
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, errPictureDownload
	}
	if response.ContentLength > maxPictureBytes {
		return nil, errPictureBytes
	}
	rawBytes, err := io.ReadAll(io.LimitReader(response.Body, maxPictureBytes+1))
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, errPictureTimeout
		}
		return nil, errPictureDownload
	}
	if int64(len(rawBytes)) > maxPictureBytes {
		return nil, errPictureBytes
	}
	return optimizePicture(rawBytes)
}

// @impl WSH-006-AC4 WSH-006-AC5 WSH-006-AC6 CON-006: inspect dimensions before pixel allocation.
func optimizePicture(raw []byte) ([]byte, error) {
	config, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || (format != "jpeg" && format != "png" && format != "webp" && format != "gif" && format != "avif") {
		return nil, errPictureFormat
	}
	if config.Width < 1 || config.Height < 1 || int64(config.Width) > maxPicturePixels/int64(config.Height) {
		return nil, errPicturePixels
	}
	// Each decoder's Decode returns only the first frame, never DecodeAll.
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, errPictureFormat
	}
	bounds := image.Rect(0, 0, config.Width, config.Height)
	if img.Bounds() != bounds {
		canvas := image.NewNRGBA(bounds)
		draw.Draw(canvas, img.Bounds(), img, img.Bounds().Min, draw.Src)
		img = canvas
	}
	width, height := config.Width, config.Height
	if width > 640 || height > 640 {
		if width >= height {
			height = max(1, height*640/width)
			width = 640
		} else {
			width = max(1, width*640/height)
			height = 640
		}
		resized := image.NewNRGBA(image.Rect(0, 0, width, height))
		draw.CatmullRom.Scale(resized, resized.Bounds(), img, img.Bounds(), draw.Src, nil)
		img = resized
	}
	var out bytes.Buffer
	if err := webp.Encode(&out, img, webp.Options{Quality: 80, Method: 4, Exact: true}); err != nil {
		return nil, errPictureSave
	}
	return out.Bytes(), nil
}

// errors are catalog IDs so shared render paths localize them without leaking
// remote URLs, network internals or raw decoder errors into the interface.
func pictureError(err error) string {
	if err == nil {
		return ""
	}
	return strings.TrimSpace(err.Error())
}
