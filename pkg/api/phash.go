package api

import (
	"bytes"
	"fmt"
	"image"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

// dhash computes a 64-bit difference hash (dHash) of an image.
// The image is resized to 9x8 grayscale, then each pixel is compared
// to its right neighbor to produce a 64-bit hash. Similar images
// produce hashes with small Hamming distances.
func dhash(img image.Image) int64 {
	// Resize to 9x8 using high-quality interpolation.
	dst := image.NewGray(image.Rect(0, 0, 9, 8))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	var hash int64
	for y := range 8 {
		for x := range 8 {
			left := dst.GrayAt(x, y)
			right := dst.GrayAt(x+1, y)
			if left.Y > right.Y {
				hash |= 1 << (y*8 + x)
			}
		}
	}
	return hash
}

// dhashFromBytes decodes image bytes and computes the dHash.
func dhashFromBytes(data []byte) (int64, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("decode image for phash: %w", err)
	}
	return dhash(img), nil
}
