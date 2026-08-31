package media

import (
	"encoding/binary"
	"image"
)

const (
	exifOrientationTag = 0x0112
	exifTypeShort      = 3
)

// exifOrientation returns the JPEG EXIF orientation, defaulting to normal for
// missing, malformed, or unsupported metadata.
func exifOrientation(data []byte) uint16 {
	if len(data) < 2 || data[0] != 0xff || data[1] != 0xd8 {
		return 1
	}

	for offset := 2; offset < len(data); {
		if data[offset] != 0xff {
			return 1
		}
		for offset < len(data) && data[offset] == 0xff {
			offset++
		}
		if offset >= len(data) {
			return 1
		}

		marker := data[offset]
		offset++
		if marker == 0xd9 || marker == 0xda {
			return 1
		}
		if marker == 0x01 || marker >= 0xd0 && marker <= 0xd7 {
			continue
		}
		if offset+2 > len(data) {
			return 1
		}

		segmentLength := int(binary.BigEndian.Uint16(data[offset : offset+2]))
		if segmentLength < 2 || offset+segmentLength > len(data) {
			return 1
		}
		segment := data[offset+2 : offset+segmentLength]
		offset += segmentLength

		if marker == 0xe1 {
			if orientation, ok := orientationFromEXIFSegment(segment); ok {
				return orientation
			}
		}
	}

	return 1
}

func orientationFromEXIFSegment(segment []byte) (uint16, bool) {
	const exifHeaderLength = 6
	if len(segment) < exifHeaderLength || string(segment[:exifHeaderLength]) != "Exif\x00\x00" {
		return 0, false
	}

	tiff := segment[exifHeaderLength:]
	if len(tiff) < 8 {
		return 0, false
	}

	var byteOrder binary.ByteOrder
	switch string(tiff[:2]) {
	case "II":
		byteOrder = binary.LittleEndian
	case "MM":
		byteOrder = binary.BigEndian
	default:
		return 0, false
	}
	if byteOrder.Uint16(tiff[2:4]) != 42 {
		return 0, false
	}

	ifdOffset := uint64(byteOrder.Uint32(tiff[4:8]))
	if ifdOffset+2 > uint64(len(tiff)) {
		return 0, false
	}
	entryCount := uint64(byteOrder.Uint16(tiff[ifdOffset : ifdOffset+2]))
	entriesOffset := ifdOffset + 2
	if entryCount > (uint64(len(tiff))-entriesOffset)/12 {
		return 0, false
	}

	for index := range entryCount {
		entryOffset := entriesOffset + index*12
		entry := tiff[entryOffset : entryOffset+12]
		if byteOrder.Uint16(entry[:2]) != exifOrientationTag {
			continue
		}
		if byteOrder.Uint16(entry[2:4]) != exifTypeShort || byteOrder.Uint32(entry[4:8]) != 1 {
			return 0, false
		}

		orientation := byteOrder.Uint16(entry[8:10])
		if orientation < 1 || orientation > 8 {
			return 0, false
		}
		return orientation, true
	}

	return 0, false
}

func applyEXIFOrientation(src image.Image, orientation uint16) image.Image {
	if orientation <= 1 || orientation > 8 {
		return src
	}

	srcBounds := src.Bounds()
	width, height := srcBounds.Dx(), srcBounds.Dy()
	dstWidth, dstHeight := width, height
	if orientation >= 5 {
		dstWidth, dstHeight = height, width
	}
	dst := image.NewRGBA(image.Rect(0, 0, dstWidth, dstHeight))

	for y := range height {
		for x := range width {
			var dstX, dstY int
			switch orientation {
			case 2:
				dstX, dstY = width-1-x, y
			case 3:
				dstX, dstY = width-1-x, height-1-y
			case 4:
				dstX, dstY = x, height-1-y
			case 5:
				dstX, dstY = y, x
			case 6:
				dstX, dstY = height-1-y, x
			case 7:
				dstX, dstY = height-1-y, width-1-x
			case 8:
				dstX, dstY = y, width-1-x
			}
			dst.Set(dstX, dstY, src.At(srcBounds.Min.X+x, srcBounds.Min.Y+y))
		}
	}

	return dst
}
