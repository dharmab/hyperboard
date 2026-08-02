package media

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// Expected values match `ffprobe -show_entries format=duration` and
// `ffprobe -select_streams a` on the same fixtures.
func TestProbeVideo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		file         string
		wantDuration float64
		wantAudio    bool
	}{
		{name: "mp4 h264 without audio", file: "h264.mp4", wantDuration: 1.0, wantAudio: false},
		{name: "mp4 h264 with aac", file: "h264-audio.mp4", wantDuration: 1.0, wantAudio: true},
		{name: "quicktime h264 with aac", file: "h264-audio.mov", wantDuration: 1.0, wantAudio: true},
		{name: "mp4 hevc without audio", file: "hevc.mp4", wantDuration: 1.0, wantAudio: false},
		{name: "webm vp9 without audio", file: "vp9.webm", wantDuration: 1.0, wantAudio: false},
		{name: "webm vp8 with opus", file: "vp8-audio.webm", wantDuration: 1.008, wantAudio: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			info, err := probeVideo(readFixture(t, tt.file))
			if err != nil {
				t.Fatalf("probeVideo(%s) error: %v", tt.file, err)
			}
			if math.Abs(info.DurationSeconds-tt.wantDuration) > 0.01 {
				t.Errorf("duration = %v, want %v", info.DurationSeconds, tt.wantDuration)
			}
			if info.HasAudio != tt.wantAudio {
				t.Errorf("hasAudio = %v, want %v", info.HasAudio, tt.wantAudio)
			}
		})
	}
}

func TestProbeVideoUnknownContainer(t *testing.T) {
	t.Parallel()
	_, err := probeVideo([]byte("this is not a video file at all"))
	if !errors.Is(err, ErrUnknownContainer) {
		t.Errorf("error = %v, want ErrUnknownContainer", err)
	}
}

func TestProbeVideoTruncated(t *testing.T) {
	t.Parallel()
	data := readFixture(t, "h264.mp4")
	if _, err := probeVideo(data[:64]); err == nil {
		t.Error("expected an error for a truncated file")
	}
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return data
}
