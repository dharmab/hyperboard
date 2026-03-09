package media

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"os/exec"
)

// ProbeHasAudio uses ffprobe to check if a video file contains an audio stream.
func ProbeHasAudio(data []byte) (bool, error) {
	tmpFile, err := os.CreateTemp("", "hyperboard-probe-*")
	if err != nil {
		return false, fmt.Errorf("create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return false, fmt.Errorf("write temp file: %w", err)
	}
	_ = tmpFile.Close()

	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-select_streams", "a",
		"-show_entries", "stream=codec_type",
		"-of", "csv=p=0",
		tmpFile.Name(),
	)
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("ffprobe: %w", err)
	}
	return len(bytes.TrimSpace(out)) > 0, nil
}

// ProcessVideo extracts a thumbnail from a video file using ffmpeg.
// Returns WebP thumbnail bytes.
func ProcessVideo(data []byte) ([]byte, error) {
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
