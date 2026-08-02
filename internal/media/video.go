package media

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"

	"github.com/dharmab/hyperboard/internal/media/ffwasm"
)

// extractThumbnail decodes the frame at the given offset (in seconds) from the
// video at path, scales it to fit within 512x512, and returns WebP-encoded bytes.
func extractThumbnail(ctx context.Context, path string, offsetSeconds float64) ([]byte, error) {
	img, err := ffwasm.ExtractFrame(ctx, path, offsetSeconds)
	if err != nil {
		return nil, fmt.Errorf("extract frame: %w", err)
	}

	thumb := FitImage(img, 512, 512)
	thumbBytes, err := EncodeWebP(thumb, 80)
	if err != nil {
		return nil, fmt.Errorf("encode thumbnail: %w", err)
	}

	return thumbBytes, nil
}

// writeTempVideo writes data to a temp file and returns its path along with a
// cleanup function. The decoder needs a seekable file; it does not read the
// whole video into memory.
func writeTempVideo(data []byte) (string, func(), error) {
	tmpFile, err := os.CreateTemp("", "hyperboard-video-*")
	if err != nil {
		return "", nil, fmt.Errorf("create temp file: %w", err)
	}
	cleanup := func() { _ = os.Remove(tmpFile.Name()) }

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		cleanup()
		return "", nil, fmt.Errorf("write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("close temp file: %w", err)
	}

	return tmpFile.Name(), cleanup, nil
}

// ProcessVideo extracts a thumbnail from a video file and probes for audio.
// Returns (thumbnailData, hasAudio, error).
func ProcessVideo(ctx context.Context, data []byte) ([]byte, bool, error) {
	info, err := probeVideo(data)
	if err != nil {
		return nil, false, fmt.Errorf("probe video: %w", err)
	}

	path, cleanup, err := writeTempVideo(data)
	if err != nil {
		return nil, false, err
	}
	defer cleanup()

	// Extract a frame at Wadsworth's constant (30%) into the video to hopefully
	// get a visually interesting thumbnail
	const wadsworthConstant = 0.30
	offset := 1.0
	if info.DurationSeconds > 0 {
		offset = info.DurationSeconds * wadsworthConstant
	}

	thumbBytes, err := extractThumbnail(ctx, path, offset)
	if err != nil {
		return nil, false, err
	}

	return thumbBytes, info.HasAudio, nil
}

// RegenerateVideoThumbnail extracts a thumbnail from a random frame between
// 25% and 75% of the video duration.
func RegenerateVideoThumbnail(ctx context.Context, data []byte) ([]byte, error) {
	info, err := probeVideo(data)
	if err != nil {
		return nil, fmt.Errorf("probe video: %w", err)
	}

	path, cleanup, err := writeTempVideo(data)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	offset := 1.0
	if info.DurationSeconds > 0 {
		// Pick a random position between 25% and 75% of the video.
		offset = info.DurationSeconds * (0.25 + rand.Float64()*0.50) //nolint:gosec // not security-sensitive
	}

	return extractThumbnail(ctx, path, offset)
}
