package media

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/gen2brain/webp"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

// EncodeWebP encodes an image.Image to lossy WebP bytes. The encoder is libwebp
// transpiled to Go, so this runs in-process with no cgo and no subprocess.
//
// Builds must set the nodynamic build tag; without it the library dlopens a
// system libwebp when one is present, which makes output depend on the host.
func EncodeWebP(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, webp.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("encode webp: %w", err)
	}
	return buf.Bytes(), nil
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

// MIMEWebP is the MIME type for WebP images.
const MIMEWebP = "image/webp"

// MIMEGif is the MIME type for GIF images.
const MIMEGif = "image/gif"

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

	if detectedMIME == MIMEGif {
		// Store GIF as-is; skip WebP conversion.
		content = data
		mime = MIMEGif
	} else if w > MaxWebPDimension || h > MaxWebPDimension {
		// Too large to convert — store original.
		content = data
		mime = detectedMIME
	} else if detectedMIME == MIMEWebP {
		// Already WebP — skip re-encoding to avoid quality loss.
		content = data
		mime = MIMEWebP
	} else {
		encoded, err := EncodeWebP(img, 85)
		if err != nil {
			return nil, "", nil, fmt.Errorf("encode webp: %w", err)
		}
		content = encoded
		mime = MIMEWebP
	}

	// Generate thumbnail (512x512 max, preserve aspect ratio).
	thumb := FitImage(img, 512, 512)
	thumbBytes, err := EncodeWebP(thumb, 80)
	if err != nil {
		return nil, "", nil, fmt.Errorf("encode thumbnail: %w", err)
	}

	return content, mime, thumbBytes, nil
}
