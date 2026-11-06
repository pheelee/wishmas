package app

import (
	"bytes"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"

	"github.com/gen2brain/webp"
)

// @test WSH-006-AC16
func TestPictureReadRejectsNonRegularAndOversizedFiles(t *testing.T) {
	dir := t.TempDir()
	if _, err := readPicture(dir); err == nil {
		t.Fatal("directory accepted")
	}
	path := filepath.Join(dir, "oversized.webp")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate(maxPictureBytes + 1); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := readPicture(path); err != errPictureFormat {
		t.Fatalf("oversized picture: %v", err)
	}
	fifo := filepath.Join(dir, "fifo")
	if err := syscall.Mkfifo(fifo, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readPicture(fifo); err != errPictureFormat {
		t.Fatalf("FIFO: %v", err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink("oversized.webp", link); err != nil {
		t.Fatal(err)
	}
	if _, err := readPicture(link); err == nil {
		t.Fatal("internal symlink accepted")
	}
}

// @test WSH-006-AC16
func TestPictureReadRejectsSymlinkReplacement(t *testing.T) {
	encode := func(c color.RGBA) []byte {
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		img.Set(0, 0, c)
		var out bytes.Buffer
		if err := webp.Encode(&out, img, webp.Options{Quality: 80}); err != nil {
			t.Fatal(err)
		}
		return out.Bytes()
	}
	expected := encode(color.RGBA{G: 255, A: 255})
	outsideBytes := encode(color.RGBA{R: 255, A: 255})
	dir := t.TempDir()
	path := filepath.Join(dir, "picture.webp")
	outside := filepath.Join(t.TempDir(), "outside.webp")
	if err := os.WriteFile(outside, outsideBytes, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, expected, 0600); err != nil {
		t.Fatal(err)
	}
	if raw, err := readPicture(path); err != nil || !bytes.Equal(raw, expected) {
		t.Fatalf("fixture must be a readable picture: %v", err)
	}
	stop := make(chan struct{})
	var writer sync.WaitGroup
	writer.Add(1)
	go func() {
		defer writer.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			link := filepath.Join(dir, "link")
			if err := os.Symlink(outside, link); err != nil {
				t.Error(err)
				return
			}
			if err := os.Rename(link, path); err != nil {
				t.Error(err)
				return
			}
			regular := filepath.Join(dir, "regular")
			if err := os.WriteFile(regular, expected, 0600); err != nil {
				t.Error(err)
				return
			}
			if err := os.Rename(regular, path); err != nil {
				t.Error(err)
				return
			}
		}
	}()
	defer func() { close(stop); writer.Wait() }()
	for range 50000 {
		raw, err := readPicture(path)
		if err == nil && !bytes.Equal(raw, expected) {
			t.Fatal("read picture outside the root after symlink replacement")
		}
	}
}
