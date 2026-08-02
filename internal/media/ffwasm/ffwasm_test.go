package ffwasm

import (
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractFrameMissingFile(t *testing.T) {
	t.Parallel()
	_, err := ExtractFrame(t.Context(), filepath.Join(t.TempDir(), "absent.mp4"), 1)
	if err == nil {
		t.Fatal("expected an error for a missing file")
	}
	if !strings.Contains(err.Error(), "open input") {
		t.Errorf("error = %v, want it to mention the failed open", err)
	}
}

func TestExtractFrameNotAVideo(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "junk.mp4")
	if err := os.WriteFile(path, []byte(strings.Repeat("nonsense", 512)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ExtractFrame(t.Context(), path, 1); err == nil {
		t.Fatal("expected an error for a file that is not a video")
	}
}

func TestDecodeFrame(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		data    []byte
		wantErr string
	}{
		{name: "empty output", data: nil, wantErr: "no frame written"},
		{name: "bad magic", data: append([]byte("NOPE"), make([]byte, 8)...), wantErr: "no frame written"},
		{
			name:    "zero dimensions",
			data:    append([]byte("FRM1"), make([]byte, 8)...),
			wantErr: "bad dimensions",
		},
		{
			name:    "truncated pixels",
			data:    append([]byte("FRM1"), 2, 0, 0, 0, 2, 0, 0, 0, 0xff),
			wantErr: "pixel bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := decodeFrame(tt.data)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestDecodeFrameValid(t *testing.T) {
	t.Parallel()
	data := append([]byte("FRM1"), 2, 0, 0, 0, 1, 0, 0, 0)
	data = append(data, 1, 2, 3, 4, 5, 6, 7, 8)

	img, err := decodeFrame(data)
	if err != nil {
		t.Fatalf("decodeFrame error: %v", err)
	}
	if got := img.Bounds().Size(); got.X != 2 || got.Y != 1 {
		t.Fatalf("size = %v, want 2x1", got)
	}
	// The pixels are non-premultiplied, so compare them as stored.
	want := color.NRGBA{R: 5, G: 6, B: 7, A: 8}
	if got := img.At(1, 0); got != want {
		t.Errorf("pixel = %v, want %v", got, want)
	}
}
