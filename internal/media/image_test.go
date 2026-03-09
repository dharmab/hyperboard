package media

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os/exec"
	"testing"
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
			if gotW != tt.expectW || gotH != tt.expectH {
				t.Errorf("FitImage(%dx%d, %d, %d) = %dx%d, want %dx%d",
					tt.srcW, tt.srcH, tt.maxW, tt.maxH, gotW, gotH, tt.expectW, tt.expectH)
			}
		})
	}
}

func TestProcessImage(t *testing.T) { //nolint:paralleltest // requires cwebp binary
	if _, err := exec.LookPath("cwebp"); err != nil {
		t.Skip("cwebp not available")
	}

	t.Run("small png to webp", func(t *testing.T) {
		img := syntheticColorImage(64, 64)
		pngData := encodePNG(t, img)

		content, mime, thumbnail, err := ProcessImage(pngData, "image/png")
		if err != nil {
			t.Fatalf("ProcessImage error: %v", err)
		}
		if mime != "image/webp" {
			t.Errorf("mime = %q, want %q", mime, "image/webp")
		}
		if len(content) == 0 {
			t.Error("content should not be empty")
		}
		if len(thumbnail) == 0 {
			t.Error("thumbnail should not be empty")
		}
	})
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
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode PNG: %v", err)
	}
	return buf.Bytes()
}
