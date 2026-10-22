package app

import (
	"bytes"
	"image"
	"os"
)

func readPicture(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxPictureBytes {
		return nil, errPictureFormat
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || format != "webp" || config.Width > 640 || config.Height > 640 {
		return nil, errPictureFormat
	}
	return raw, nil
}
