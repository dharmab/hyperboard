package api

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"os/exec"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

// encodeWebP shells out to cwebp to encode an image.Image to WebP bytes.
func encodeWebP(img image.Image, quality int) ([]byte, error) {
	// Write the image to a temp PNG file for cwebp to read.
	inFile, err := os.CreateTemp("", "hyperboard-img-in-*.png")
	if err != nil {
		return nil, fmt.Errorf("create temp input file: %w", err)
	}
	defer func() { _ = os.Remove(inFile.Name()) }()

	// Encode as PNG into the temp file using standard library.
	if err := png.Encode(inFile, img); err != nil {
		_ = inFile.Close()
		return nil, fmt.Errorf("encode png for cwebp: %w", err)
	}
	_ = inFile.Close()

	outFile, err := os.CreateTemp("", "hyperboard-img-out-*.webp")
	if err != nil {
		return nil, fmt.Errorf("create temp output file: %w", err)
	}
	_ = outFile.Close()
	defer func() { _ = os.Remove(outFile.Name()) }()

	cmd := exec.Command("cwebp",
		"-q", fmt.Sprintf("%d", quality),
		inFile.Name(),
		"-o", outFile.Name(),
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("cwebp encode: %w: %s", err, out)
	}

	return os.ReadFile(outFile.Name())
}

// fitImage resizes img to fit within maxW x maxH, preserving aspect ratio.
func fitImage(img image.Image, maxW, maxH int) image.Image {
	bounds := img.Bounds()
	srcW := bounds.Max.X - bounds.Min.X
	srcH := bounds.Max.Y - bounds.Min.Y

	if srcW == 0 || srcH == 0 {
		return img
	}

	scaleW := float64(maxW) / float64(srcW)
	scaleH := float64(maxH) / float64(srcH)
	scale := scaleW
	if scaleH < scale {
		scale = scaleH
	}

	// Don't upscale.
	if scale >= 1.0 {
		return img
	}

	dstW := int(float64(srcW) * scale)
	dstH := int(float64(srcH) * scale)
	if dstW < 1 {
		dstW = 1
	}
	if dstH < 1 {
		dstH = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
	return dst
}

// processImage converts an image to WebP (unless too large) and generates a thumbnail.
// Returns (contentBytes, mimeType, thumbnailBytes, error).
func processImage(data []byte, detectedMIME string) ([]byte, string, []byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", nil, fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()
	w := bounds.Max.X - bounds.Min.X
	h := bounds.Max.Y - bounds.Min.Y

	var content []byte
	var mime string

	if w > 16383 || h > 16383 {
		// Too large to convert — store original.
		content = data
		mime = detectedMIME
	} else {
		encoded, err := encodeWebP(img, 85)
		if err != nil {
			return nil, "", nil, fmt.Errorf("encode webp: %w", err)
		}
		content = encoded
		mime = "image/webp"
	}

	// Generate thumbnail (512x512 max, preserve aspect ratio).
	thumb := fitImage(img, 512, 512)
	thumbBytes, err := encodeWebP(thumb, 80)
	if err != nil {
		return nil, "", nil, fmt.Errorf("encode thumbnail: %w", err)
	}

	return content, mime, thumbBytes, nil
}

// processVideo extracts a thumbnail from a video file using ffmpeg.
// Returns WebP thumbnail bytes.
func processVideo(data []byte) ([]byte, error) {
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

	thumb := fitImage(img, 512, 512)
	thumbBytes, err := encodeWebP(thumb, 80)
	if err != nil {
		return nil, fmt.Errorf("encode thumbnail: %w", err)
	}

	return thumbBytes, nil
}
