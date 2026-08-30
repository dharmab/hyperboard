package media

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDhash(t *testing.T) {
	t.Parallel()
	t.Run("identical images produce same hash", func(t *testing.T) {
		t.Parallel()
		img := syntheticImage(100, 100, color.White)
		h1 := Dhash(img)
		h2 := Dhash(img)
		assert.Equal(t, h1, h2)
	})

	t.Run("different images produce different hashes", func(t *testing.T) {
		t.Parallel()
		img1 := checkerboardImage(100, 100, 10)
		img2 := checkerboardImage(100, 100, 25)
		h1 := Dhash(img1)
		h2 := Dhash(img2)
		assert.NotEqual(t, h1, h2)
	})

	t.Run("patterned image produces non-zero hash", func(t *testing.T) {
		t.Parallel()
		img := checkerboardImage(100, 100, 10)
		h := Dhash(img)
		assert.NotZero(t, h)
	})
}

func syntheticImage(w, h int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, c)
		}
	}
	return img
}

func checkerboardImage(w, h, blockSize int) image.Image {
	img := image.NewGray(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			if ((x/blockSize)+(y/blockSize))%2 == 0 {
				img.SetGray(x, y, color.Gray{Y: 255})
			} else {
				img.SetGray(x, y, color.Gray{Y: 0})
			}
		}
	}
	return img
}
