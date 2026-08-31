package media

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeOrientedImage(t *testing.T) {
	t.Parallel()

	jpegData := encodeJPEG(t, syntheticColorImage(7, 5))
	baseline, _, err := image.Decode(bytes.NewReader(jpegData))
	require.NoError(t, err)
	width, height := baseline.Bounds().Dx(), baseline.Bounds().Dy()

	tests := []struct {
		name         string
		orientation  uint16
		expectWidth  int
		expectHeight int
		sourceAt     func(x, y int) (int, int)
	}{
		{"normal", 1, width, height, func(x, y int) (int, int) { return x, y }},
		{"mirror horizontal", 2, width, height, func(x, y int) (int, int) { return width - 1 - x, y }},
		{"rotate 180", 3, width, height, func(x, y int) (int, int) { return width - 1 - x, height - 1 - y }},
		{"mirror vertical", 4, width, height, func(x, y int) (int, int) { return x, height - 1 - y }},
		{"transpose", 5, height, width, func(x, y int) (int, int) { return y, x }},
		{"rotate 90 clockwise", 6, height, width, func(x, y int) (int, int) { return y, height - 1 - x }},
		{"transverse", 7, height, width, func(x, y int) (int, int) { return width - 1 - y, height - 1 - x }},
		{"rotate 90 counter-clockwise", 8, height, width, func(x, y int) (int, int) { return width - 1 - y, x }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			orientedData := addJPEGAPP1(t, jpegData, exifPayload(tt.orientation))
			got, format, decodeErr := decodeOrientedImage(orientedData)
			require.NoError(t, decodeErr)
			assert.Equal(t, "jpeg", format)
			require.Equal(t, image.Rect(0, 0, tt.expectWidth, tt.expectHeight), got.Bounds())

			for y := range tt.expectHeight {
				for x := range tt.expectWidth {
					sourceX, sourceY := tt.sourceAt(x, y)
					expected := color.RGBAModel.Convert(baseline.At(sourceX, sourceY))
					actual := color.RGBAModel.Convert(got.At(x, y))
					assert.Equal(t, expected, actual, "pixel at (%d, %d)", x, y)
				}
			}
		})
	}
}

func TestDecodeOrientedImageTreatsMissingOrMalformedEXIFAsNormal(t *testing.T) {
	t.Parallel()

	jpegData := encodeJPEG(t, syntheticColorImage(7, 5))
	baseline, _, err := image.Decode(bytes.NewReader(jpegData))
	require.NoError(t, err)

	tests := []struct {
		name string
		data []byte
	}{
		{name: "missing", data: jpegData},
		{name: "invalid orientation", data: addJPEGAPP1(t, jpegData, exifPayload(9))},
		{name: "truncated TIFF", data: addJPEGAPP1(t, jpegData, []byte("Exif\x00\x00MM"))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, _, decodeErr := decodeOrientedImage(tt.data)
			require.NoError(t, decodeErr)
			assertImagesEqual(t, baseline, got)
		})
	}
}

func encodeJPEG(t *testing.T, img image.Image) []byte {
	t.Helper()

	var output bytes.Buffer
	require.NoError(t, jpeg.Encode(&output, img, &jpeg.Options{Quality: 100}))
	return output.Bytes()
}

func exifPayload(orientation uint16) []byte {
	payload := make([]byte, 32)
	copy(payload, "Exif\x00\x00MM")
	binary.BigEndian.PutUint16(payload[8:10], 42)
	binary.BigEndian.PutUint32(payload[10:14], 8)
	binary.BigEndian.PutUint16(payload[14:16], 1)
	binary.BigEndian.PutUint16(payload[16:18], exifOrientationTag)
	binary.BigEndian.PutUint16(payload[18:20], exifTypeShort)
	binary.BigEndian.PutUint32(payload[20:24], 1)
	binary.BigEndian.PutUint16(payload[24:26], orientation)
	return payload
}

func addJPEGAPP1(t *testing.T, jpegData, payload []byte) []byte {
	t.Helper()
	require.GreaterOrEqual(t, len(jpegData), 2)
	require.LessOrEqual(t, len(payload)+2, int(^uint16(0)))

	output := make([]byte, 0, len(jpegData)+len(payload)+4)
	output = append(output, jpegData[:2]...)
	output = append(output, 0xff, 0xe1)
	length := make([]byte, 2)
	binary.BigEndian.PutUint16(length, uint16(len(payload)+2))
	output = append(output, length...)
	output = append(output, payload...)
	output = append(output, jpegData[2:]...)
	return output
}

func assertImagesEqual(t *testing.T, expected, actual image.Image) {
	t.Helper()
	require.Equal(t, expected.Bounds(), actual.Bounds())
	for y := expected.Bounds().Min.Y; y < expected.Bounds().Max.Y; y++ {
		for x := expected.Bounds().Min.X; x < expected.Bounds().Max.X; x++ {
			assert.Equal(t, color.RGBAModel.Convert(expected.At(x, y)), color.RGBAModel.Convert(actual.At(x, y)))
		}
	}
}
