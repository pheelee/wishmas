package app

import (
	"bytes"
	"image"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

// @impl WSH-006-AC16: validate and read the same no-follow, root-relative handle.
func readPicture(path string) ([]byte, error) {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	defer root.Close()
	dir, err := root.Open(".")
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	// openat enforces NOFOLLOW even for in-root symlinks; NONBLOCK prevents
	// FIFO replacements from hanging before the regular-file check.
	fd, err := syscall.Openat(int(dir.Fd()), filepath.Base(path), syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, &os.PathError{Op: "openat", Path: path, Err: err}
	}
	file := os.NewFile(uintptr(fd), path)
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxPictureBytes {
		return nil, errPictureFormat
	}
	raw, err := io.ReadAll(io.LimitReader(file, maxPictureBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > maxPictureBytes {
		return nil, errPictureFormat
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || format != "webp" || config.Width > 640 || config.Height > 640 {
		return nil, errPictureFormat
	}
	return raw, nil
}
