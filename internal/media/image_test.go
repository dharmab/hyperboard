package media

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFitImage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		srcW, srcH       int
		maxW, maxH       int
		expectW, expectH int
	}{
		{
			name: "no resize needed",
			srcW: 100, srcH: 100,
			maxW: 512, maxH: 512,
			expectW: 100, expectH: 100,
		},
		{
			name: "scale down wide image",
			srcW: 1024, srcH: 512,
			maxW: 512, maxH: 512,
			expectW: 512, expectH: 256,
		},
		{
			name: "scale down tall image",
			srcW: 512, srcH: 1024,
			maxW: 512, maxH: 512,
			expectW: 256, expectH: 512,
		},
		{
			name: "scale down large square",
			srcW: 2048, srcH: 2048,
			maxW: 512, maxH: 512,
			expectW: 512, expectH: 512,
		},
		{
			name: "zero dimension returns original",
			srcW: 0, srcH: 0,
			maxW: 512, maxH: 512,
			expectW: 0, expectH: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			img := image.NewRGBA(image.Rect(0, 0, tt.srcW, tt.srcH))
			result := FitImage(img, tt.maxW, tt.maxH)
			bounds := result.Bounds()
			gotW := bounds.Max.X - bounds.Min.X
			gotH := bounds.Max.Y - bounds.Min.Y
			assert.Equal(t, tt.expectW, gotW)
			assert.Equal(t, tt.expectH, gotH)
		})
	}
}

func TestProcessImageRejectsInvalidMedia(t *testing.T) {
	t.Parallel()

	_, _, _, err := ProcessImage(t.Context(), []byte("not an image"), "image/png")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidMedia)
}

func TestProcessImage(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("cwebp"); err != nil {
		t.Skip("cwebp not available")
	}

	t.Run("small png to webp", func(t *testing.T) {
		t.Parallel()

		img := syntheticColorImage(64, 64)
		pngData := encodePNG(t, img)

		content, mime, thumbnail, err := ProcessImage(t.Context(), pngData, "image/png")
		require.NoError(t, err)
		assert.Equal(t, MIMEWebP, mime)
		assert.NotEmpty(t, content)
		assert.NotEmpty(t, thumbnail)
	})

	t.Run("webp passthrough", func(t *testing.T) {
		t.Parallel()

		img := syntheticColorImage(64, 64)
		webpData, err := EncodeWebP(t.Context(), img, 85)
		require.NoError(t, err)

		content, mime, thumbnail, err := ProcessImage(t.Context(), webpData, MIMEWebP)
		require.NoError(t, err)
		assert.Equal(t, MIMEWebP, mime)
		assert.Equal(t, webpData, content)
		assert.NotEmpty(t, thumbnail)
	})

	for _, orientation := range []uint16{6, 8} {
		t.Run(fmt.Sprintf("JPEG EXIF orientation %d", orientation), func(t *testing.T) {
			t.Parallel()

			jpegData := encodeJPEG(t, syntheticColorImage(40, 20))
			jpegData = addJPEGAPP1(t, jpegData, exifPayload(orientation))

			content, mime, thumbnail, err := ProcessImage(t.Context(), jpegData, "image/jpeg")
			require.NoError(t, err)
			assert.Equal(t, MIMEWebP, mime)

			contentImage, _, err := image.Decode(bytes.NewReader(content))
			require.NoError(t, err)
			assert.Equal(t, image.Rect(0, 0, 20, 40), contentImage.Bounds())

			thumbnailImage, _, err := image.Decode(bytes.NewReader(thumbnail))
			require.NoError(t, err)
			assert.Equal(t, image.Rect(0, 0, 20, 40), thumbnailImage.Bounds())
		})
	}
}

func syntheticColorImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	return img
}

func encodePNG(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	encodeErr := png.Encode(&buf, img)
	require.NoError(t, encodeErr)
	return buf.Bytes()
}
