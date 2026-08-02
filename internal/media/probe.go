package media

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/abema/go-mp4"
	"github.com/at-wat/ebml-go"
	"github.com/at-wat/ebml-go/webm"
)

// videoInfo is the container metadata the thumbnail pipeline needs.
type videoInfo struct {
	// DurationSeconds is the duration of the movie, or 0 if the container did
	// not record one.
	DurationSeconds float64
	// HasAudio reports whether the container declares an audio track.
	HasAudio bool
}

// ErrUnknownContainer is returned when data is not an MP4/QuickTime or Matroska file.
var ErrUnknownContainer = errors.New("unrecognized video container")

// probeVideo reads duration and audio-track presence out of the container
// headers. It parses only boxes and EBML elements — no codec is involved — so
// it costs a few milliseconds and needs no external process.
func probeVideo(data []byte) (videoInfo, error) {
	switch {
	case isMP4(data):
		return probeMP4(bytes.NewReader(data))
	case isMatroska(data):
		return probeMatroska(bytes.NewReader(data))
	default:
		return videoInfo{}, ErrUnknownContainer
	}
}

// isMP4 reports whether data starts with an ISO base media file format box.
// MP4 and QuickTime both put a 4-byte size followed by a box type at offset 0;
// "ftyp" is required first in MP4 and conventional in QuickTime.
func isMP4(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	switch string(data[4:8]) {
	case "ftyp", "moov", "mdat", "free", "skip", "wide", "pnot":
		return true
	default:
		return false
	}
}

// isMatroska reports whether data starts with the EBML magic number shared by
// Matroska and WebM.
func isMatroska(data []byte) bool {
	return len(data) >= 4 && bytes.Equal(data[:4], []byte{0x1A, 0x45, 0xDF, 0xA3})
}

// probeMP4 reads mvhd for the duration and every trak's mdia/hdlr for audio.
func probeMP4(r io.ReadSeeker) (videoInfo, error) {
	var info videoInfo

	mvhds, err := mp4.ExtractBoxWithPayload(r, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeMvhd()})
	if err != nil {
		return info, fmt.Errorf("read mvhd: %w", err)
	}
	if len(mvhds) == 0 {
		return info, errors.New("mp4: no mvhd box")
	}
	mvhd, ok := mvhds[0].Payload.(*mp4.Mvhd)
	if !ok {
		return info, errors.New("mp4: unexpected mvhd payload")
	}
	if mvhd.Timescale != 0 {
		duration := uint64(mvhd.DurationV0)
		if mvhd.GetVersion() == 1 {
			duration = mvhd.DurationV1
		}
		info.DurationSeconds = float64(duration) / float64(mvhd.Timescale)
	}

	// The handler type identifies the track's media independently of the codec,
	// so this recognizes audio formats go-mp4's Probe does not classify.
	hdlrs, err := mp4.ExtractBoxWithPayload(r, nil, mp4.BoxPath{
		mp4.BoxTypeMoov(), mp4.BoxTypeTrak(), mp4.BoxTypeMdia(), mp4.BoxTypeHdlr(),
	})
	if err != nil {
		return info, fmt.Errorf("read hdlr: %w", err)
	}
	for _, box := range hdlrs {
		hdlr, ok := box.Payload.(*mp4.Hdlr)
		if !ok {
			continue
		}
		if string(hdlr.HandlerType[:]) == "soun" {
			info.HasAudio = true
			break
		}
	}

	return info, nil
}

// maxMatroskaHeaderBytes bounds how much of a Matroska file is parsed. Info and
// Tracks precede the first Cluster in any file a browser can play, so this
// avoids walking gigabytes of frame data to answer a header question. Hitting
// the limit is not an error as long as the header was complete.
const maxMatroskaHeaderBytes = 8 << 20

// matroskaHeader is the subset of the Matroska element tree that matters here.
// Clusters are deliberately absent so their contents are skipped as unknown
// elements.
type matroskaHeader struct {
	Segment struct {
		Info   webm.Info   `ebml:"Info"`
		Tracks webm.Tracks `ebml:"Tracks"`
	} `ebml:"Segment"`
}

// probeMatroska reads Segment/Info for the duration and Segment/Tracks for
// audio tracks.
func probeMatroska(r io.Reader) (videoInfo, error) {
	var info videoInfo

	var header matroskaHeader
	err := ebml.Unmarshal(io.LimitReader(r, maxMatroskaHeaderBytes), &header, ebml.WithIgnoreUnknown(true))
	// A truncated read is expected whenever the file is larger than the limit.
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return info, fmt.Errorf("read matroska header: %w", err)
	}

	// Duration is expressed in TimecodeScale units, which are nanoseconds.
	timecodeScale := header.Segment.Info.TimecodeScale
	if timecodeScale == 0 {
		timecodeScale = 1_000_000 // Matroska default: 1 ms.
	}
	const nanosecondsPerSecond = 1_000_000_000
	info.DurationSeconds = header.Segment.Info.Duration * float64(timecodeScale) / nanosecondsPerSecond

	// TrackType 2 is audio (1 is video).
	const audioTrackType = 2
	for _, track := range header.Segment.Tracks.TrackEntry {
		if track.TrackType == audioTrackType {
			info.HasAudio = true
			break
		}
	}

	if info.DurationSeconds == 0 && len(header.Segment.Tracks.TrackEntry) == 0 {
		return info, errors.New("matroska: no Info or Tracks element found")
	}

	return info, nil
}
