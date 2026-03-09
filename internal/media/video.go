package media

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"os/exec"
)

// probeHasAudio uses ffprobe to check if a video file contains an audio stream.
// The caller provides a path to the video file on disk.
func probeHasAudio(path string) (bool, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-select_streams", "a",
		"-show_entries", "stream=codec_type",
		"-of", "csv=p=0",
		path,
	)
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("ffprobe: %w", err)
	}
	return len(bytes.TrimSpace(out)) > 0, nil
}

// ProcessVideo extracts a thumbnail from a video file and probes for audio.
// Returns (thumbnailData, hasAudio, error). The video data is written to a
// single temp file shared by both ffmpeg and ffprobe.
func ProcessVideo(data []byte) ([]byte, bool, error) {
	tmpFile, err := os.CreateTemp("", "hyperboard-video-*")
	if err != nil {
		return nil, false, fmt.Errorf("create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return nil, false, fmt.Errorf("write temp file: %w", err)
	}
	_ = tmpFile.Close()

	// Probe for audio streams.
	hasAudio, err := probeHasAudio(tmpFile.Name())
	if err != nil {
		// Non-fatal: assume no audio on probe failure.
		hasAudio = false
	}

	// Extract a single frame as PNG.
	cmd := exec.Command("ffmpeg",
		"-i", tmpFile.Name(),
		"-ss", "00:00:01",
		"-vframes", "1",
		"-f", "image2pipe",
		"-vcodec", "png",
		"pipe:1",
	)
	pngData, err := cmd.Output()
	if err != nil {
		return nil, false, fmt.Errorf("ffmpeg extract frame: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, false, fmt.Errorf("decode frame: %w", err)
	}

	thumb := FitImage(img, 512, 512)
	thumbBytes, err := EncodeWebP(thumb, 80)
	if err != nil {
		return nil, false, fmt.Errorf("encode thumbnail: %w", err)
	}

	return thumbBytes, hasAudio, nil
}
