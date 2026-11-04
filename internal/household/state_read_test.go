package household

import (
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
)

// @test CON-009-AC18
func TestReadRegularRejectsSymlinkReplacement(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("outside state"), 0600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "state")
	const expected = "trusted state"
	if err := os.WriteFile(path, []byte(expected), 0600); err != nil {
		t.Fatal(err)
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
			if err := os.WriteFile(regular, []byte(expected), 0600); err != nil {
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
	for range 5000 {
		raw, err := readRegular(path)
		if err == nil && string(raw) != expected {
			t.Fatalf("read replaced state outside the root: %q", raw)
		}
	}
}

// @test CON-009-AC18
func TestReadRegularFileTypes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state")
	if _, err := readRegular(path); !os.IsNotExist(err) {
		t.Fatalf("missing file: %v", err)
	}
	if _, err := readRegular(dir); err == nil {
		t.Fatal("directory accepted")
	}
	if err := os.WriteFile(path, []byte("state"), 0600); err != nil {
		t.Fatal(err)
	}
	if raw, err := readRegular(path); err != nil || string(raw) != "state" {
		t.Fatalf("regular read: %q %v", raw, err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink("state", link); err != nil {
		t.Fatal(err)
	}
	if _, err := readRegular(link); err == nil {
		t.Fatal("internal symlink accepted")
	}
	fifo := filepath.Join(dir, "fifo")
	if err := syscall.Mkfifo(fifo, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readRegular(fifo); err == nil {
		t.Fatal("FIFO accepted")
	}
}
