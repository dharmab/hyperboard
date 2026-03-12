package media

import (
	"bytes"
	"fmt"
	"image"
	"math/rand/v2"
	"os"
	"os/exec"
	"strconv"
	"strings"
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

// probeDuration uses ffprobe to return the duration of a video in seconds.
func probeDuration(path string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		path,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe duration: %w", err)
	}
	s := strings.TrimSpace(string(out))
	d, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("parse duration %q: %w", s, err)
	}
	return d, nil
}

// extractThumbnail extracts a frame at the given offset (in seconds) from the
// video at path, scales it to fit within 512x512, and returns WebP-encoded bytes.
func extractThumbnail(path string, offsetSeconds float64) ([]byte, error) {
	cmd := exec.Command("ffmpeg",
		"-ss", strconv.FormatFloat(offsetSeconds, 'f', 3, 64),
		"-i", path,
		"-vframes", "1",
		"-f", "image2pipe",
		"-vcodec", "png",
		"pipe:1",
	)
	pngData, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg extract frame: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, fmt.Errorf("decode frame: %w", err)
	}

	thumb := FitImage(img, 512, 512)
	thumbBytes, err := EncodeWebP(thumb, 80)
	if err != nil {
		return nil, fmt.Errorf("encode thumbnail: %w", err)
	}

	return thumbBytes, nil
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

	// Extract a frame at Wadsworth's constant (30%) into the video to hopefully
	// get a visually interesting thumbnail
	const wadsworthConstant = 0.30
	offset := 1.0
	if duration, err := probeDuration(tmpFile.Name()); err == nil && duration > 0 {
		offset = duration * wadsworthConstant
	}
	thumbBytes, err := extractThumbnail(tmpFile.Name(), offset)
	if err != nil {
		return nil, false, err
	}

	return thumbBytes, hasAudio, nil
}

// RegenerateVideoThumbnail extracts a thumbnail from a random frame between
// 25% and 75% of the video duration.
func RegenerateVideoThumbnail(data []byte) ([]byte, error) {
	tmpFile, err := os.CreateTemp("", "hyperboard-video-*")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	_ = tmpFile.Close()

	offset := 1.0
	if duration, err := probeDuration(tmpFile.Name()); err == nil && duration > 0 {
		// Pick a random position between 25% and 75% of the video.
		offset = duration * (0.25 + rand.Float64()*0.50) //nolint:gosec // not security-sensitive
	}

	return extractThumbnail(tmpFile.Name(), offset)
}
