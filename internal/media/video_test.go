package media

import (
	"bytes"
	"image"
	"testing"

	"golang.org/x/image/webp"
)

// TestProcessVideo decodes a frame out of each accepted container with the
// embedded FFmpeg module and checks the resulting thumbnail.
func TestProcessVideo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		file      string
		wantAudio bool
	}{
		{name: "mp4 h264", file: "h264.mp4", wantAudio: false},
		{name: "mp4 h264 with aac", file: "h264-audio.mp4", wantAudio: true},
		{name: "quicktime h264", file: "h264-audio.mov", wantAudio: true},
		{name: "mp4 hevc", file: "hevc.mp4", wantAudio: false},
		{name: "webm vp9", file: "vp9.webm", wantAudio: false},
		{name: "webm vp8 with opus", file: "vp8-audio.webm", wantAudio: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			thumb, hasAudio, err := ProcessVideo(t.Context(), readFixture(t, tt.file))
			if err != nil {
				t.Fatalf("ProcessVideo(%s) error: %v", tt.file, err)
			}
			if hasAudio != tt.wantAudio {
				t.Errorf("hasAudio = %v, want %v", hasAudio, tt.wantAudio)
			}

			img, err := webp.Decode(bytes.NewReader(thumb))
			if err != nil {
				t.Fatalf("thumbnail is not decodable WebP: %v", err)
			}
			// The fixtures are 64x64, which already fits in 512x512.
			bounds := img.Bounds()
			if bounds.Dx() != 64 || bounds.Dy() != 64 {
				t.Errorf("thumbnail is %dx%d, want 64x64", bounds.Dx(), bounds.Dy())
			}
			if isUniform(t, img) {
				t.Error("thumbnail is a single flat color, so no frame was really decoded")
			}
		})
	}
}

func TestRegenerateVideoThumbnail(t *testing.T) {
	t.Parallel()
	thumb, err := RegenerateVideoThumbnail(t.Context(), readFixture(t, "h264.mp4"))
	if err != nil {
		t.Fatalf("RegenerateVideoThumbnail error: %v", err)
	}
	if len(thumb) == 0 {
		t.Fatal("thumbnail is empty")
	}
	if _, err := webp.Decode(bytes.NewReader(thumb)); err != nil {
		t.Fatalf("thumbnail is not decodable WebP: %v", err)
	}
}

func TestProcessVideoRejectsNonVideo(t *testing.T) {
	t.Parallel()
	_, _, err := ProcessVideo(t.Context(), []byte("not a video"))
	if err == nil {
		t.Fatal("expected an error for a non-video payload")
	}
}

// isUniform reports whether every pixel in img is the same color, which is what
// a failed decode tends to produce.
func isUniform(t *testing.T, img image.Image) bool {
	t.Helper()
	bounds := img.Bounds()
	r0, g0, b0, _ := img.At(bounds.Min.X, bounds.Min.Y).RGBA()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r != r0 || g != g0 || b != b0 {
				return false
			}
		}
	}
	return true
}
