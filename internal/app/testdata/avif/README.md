# Synthetic AVIF regression fixtures

These images are generated entirely by generate.go; no shop/product image is redistributed.

- alpha.avif: 800×400 uniform translucent red, alpha 128/255. Conversion must become 640×320 WebP while preserving alpha.
- animated.avif: two 16×12 frames, red then blue. Conversion must retain only the red frame, without upscaling or animation.

Regenerate from the repository root with the pinned decoder/encoder version:

```sh
go run -tags nodynamic ./internal/app/testdata/avif/generate.go
```

AVIF tests mutate dimension metadata in memory to verify rejection above 20 megapixels before pixel decoding. Permanent tests use these local fixtures and mocked downloads, not live shop URLs.

An optional manual sample check is available without adding a shop image to the repository:

```sh
WISHMAS_AVIF_SAMPLE=/absolute/path/to/sample.avif CGO_ENABLED=0 go test -tags nodynamic ./internal/app -run '^TestAVIFOptionalRealSample$' -v
```

Set WISHMAS_AVIF_SAMPLE_URL instead to exercise the real public-address HTTPS downloader. Neither optional check runs network requests in the ordinary test suite.
