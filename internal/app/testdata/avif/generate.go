//go:build ignore

// Generate original synthetic regression images. Run from the repository root:
// go run -tags nodynamic ./internal/app/testdata/avif/generate.go
package main

import (
	"github.com/gen2brain/avif"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
)

func main() {
	root := "internal/app/testdata/avif"
	write := func(name string, encode func(*os.File) error) {
		f, err := os.Create(filepath.Join(root, name))
		if err != nil {
			panic(err)
		}
		if err = encode(f); err != nil {
			f.Close()
			panic(err)
		}
		if err = f.Close(); err != nil {
			panic(err)
		}
	}
	opts := avif.Options{Lossless: true, Speed: 10}
	alpha := image.NewNRGBA(image.Rect(0, 0, 800, 400))
	draw.Draw(alpha, alpha.Bounds(), &image.Uniform{C: color.NRGBA{R: 220, G: 40, B: 30, A: 128}}, image.Point{}, draw.Src)
	write("alpha.avif", func(f *os.File) error { return avif.Encode(f, alpha, opts) })
	red := image.NewNRGBA(image.Rect(0, 0, 16, 12))
	draw.Draw(red, red.Bounds(), &image.Uniform{C: color.NRGBA{R: 240, G: 15, B: 10, A: 255}}, image.Point{}, draw.Src)
	blue := image.NewNRGBA(red.Bounds())
	draw.Draw(blue, blue.Bounds(), &image.Uniform{C: color.NRGBA{R: 10, G: 15, B: 240, A: 255}}, image.Point{}, draw.Src)
	write("animated.avif", func(f *os.File) error {
		return avif.EncodeAll(f, &avif.AVIF{Image: []image.Image{red, blue}, Delay: []float64{.1, .1}}, opts)
	})
}
