package media

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"strconv"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

// EncodeWebP shells out to cwebp to encode an image.Image to WebP bytes.
func EncodeWebP(img image.Image, quality int) ([]byte, error) {
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
		"-q", strconv.Itoa(quality),
		inFile.Name(),
		"-o", outFile.Name(),
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("cwebp encode: %w: %s", err, out)
	}

	return os.ReadFile(outFile.Name())
}

// FitImage resizes img to fit within maxW x maxH, preserving aspect ratio.
func FitImage(img image.Image, maxW, maxH int) image.Image {
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

// MaxWebPDimension is the maximum width/height allowed by the WebP specification.
const MaxWebPDimension = 16383

// ProcessImage converts an image to WebP (unless too large) and generates a thumbnail.
// Returns (contentBytes, mimeType, thumbnailBytes, error).
func ProcessImage(data []byte, detectedMIME string) ([]byte, string, []byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", nil, fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()
	w := bounds.Max.X - bounds.Min.X
	h := bounds.Max.Y - bounds.Min.Y

	var content []byte
	var mime string

	if w > MaxWebPDimension || h > MaxWebPDimension {
		// Too large to convert — store original.
		content = data
		mime = detectedMIME
	} else {
		encoded, err := EncodeWebP(img, 85)
		if err != nil {
			return nil, "", nil, fmt.Errorf("encode webp: %w", err)
		}
		content = encoded
		mime = "image/webp"
	}

	// Generate thumbnail (512x512 max, preserve aspect ratio).
	thumb := FitImage(img, 512, 512)
	thumbBytes, err := EncodeWebP(thumb, 80)
	if err != nil {
		return nil, "", nil, fmt.Errorf("encode thumbnail: %w", err)
	}

	return content, mime, thumbBytes, nil
}
