// Package ffwasm decodes a single video frame using FFmpeg compiled to
// WebAssembly and executed by wazero.
//
// This replaces shelling out to an ffmpeg binary. The embedded module is a
// decode-only, LGPL-only build of FFmpeg (no --enable-gpl, no libx264, no
// external libraries) linked with build/ffmpeg-wasm/frame.c, which decodes one
// frame and writes it as raw RGBA. Everything runs in-process with no cgo, so
// the API server needs no system packages.
//
// Rebuild the embedded module with build/ffmpeg-wasm/build.sh.
package ffwasm

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
)

//go:embed frame.wasm
var frameModule []byte

// mountPoint is where the directory holding the input file is mounted inside
// the guest. The guest sees nothing else of the host filesystem.
const mountPoint = "/in"

// maxFrameBytes caps the decoded RGBA frame at roughly 8K resolution, which
// bounds the memory a hostile file can make the host allocate.
const maxFrameBytes = 7680 * 4320 * 4

// headerSize is the size of the frame header written by frame.c: a 4-byte magic
// followed by two little-endian uint32 dimensions.
const headerSize = 12

var frameMagic = []byte("FRM1")

//nolint:gochecknoglobals // the compiled module is a process-wide cache; compiling it costs seconds.
var (
	compileOnce   sync.Once
	wazeroRuntime wazero.Runtime
	compiled      wazero.CompiledModule
	compileErr    error
)

// CompilerSupported reports whether wazero will machine-compile the module on
// this host. It falls back to an interpreter everywhere else, which decodes
// roughly 60x slower — a 1920x1080 frame takes about 25 seconds instead of 300
// milliseconds. Everything still works; it is a performance cliff, not a
// portability limit, and the embedded module itself is architecture-neutral.
func CompilerSupported() bool {
	switch runtime.GOARCH {
	case "amd64", "arm64":
		return true
	default:
		return false
	}
}

// Warm compiles the embedded module. Compilation takes on the order of a
// second, so servers should call this at startup rather than paying for it on
// the first upload. It is safe to call concurrently and more than once.
func Warm(ctx context.Context) error {
	compileOnce.Do(func() {
		r := wazero.NewRuntime(ctx)
		if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
			compileErr = fmt.Errorf("instantiate wasi: %w", err)
			_ = r.Close(ctx)
			return
		}
		module, err := r.CompileModule(ctx, frameModule)
		if err != nil {
			compileErr = fmt.Errorf("compile frame module: %w", err)
			_ = r.Close(ctx)
			return
		}
		wazeroRuntime = r
		compiled = module
	})
	return compileErr
}

// ExtractFrame decodes the frame at offsetSeconds from the video file at path
// and returns it as an image. The frame is returned at the video's native
// resolution.
func ExtractFrame(ctx context.Context, path string, offsetSeconds float64) (image.Image, error) {
	if err := Warm(ctx); err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", path, err)
	}

	// Mount only the directory containing the input, read-only.
	fsConfig := wazero.NewFSConfig().WithReadOnlyDirMount(filepath.Dir(absPath), mountPoint)
	guestPath := mountPoint + "/" + filepath.Base(absPath)

	var stdout, stderr bytes.Buffer
	config := wazero.NewModuleConfig().
		// An empty name lets several frames be extracted concurrently, each in
		// its own instance with its own linear memory.
		WithName("").
		WithFSConfig(fsConfig).
		WithArgs("frame", guestPath, strconv.FormatFloat(offsetSeconds, 'f', 3, 64)).
		WithStdout(&stdout).
		WithStderr(&stderr).
		WithSysNanotime().
		WithSysWalltime()

	instance, err := wazeroRuntime.InstantiateModule(ctx, compiled, config)
	if instance != nil {
		defer func() { _ = instance.Close(ctx) }()
	}
	if err != nil {
		var exitErr *sys.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() != 0 {
			return nil, fmt.Errorf("decode frame: exit %d: %s",
				exitErr.ExitCode(), trimDiagnostics(stderr.String()))
		}
		return nil, fmt.Errorf("run frame module: %w: %s", err, trimDiagnostics(stderr.String()))
	}

	img, err := decodeFrame(stdout.Bytes())
	if err != nil {
		if diag := trimDiagnostics(stderr.String()); diag != "" {
			return nil, fmt.Errorf("%w: %s", err, diag)
		}
		return nil, err
	}
	return img, nil
}

// decodeFrame parses the raw RGBA output of frame.c.
func decodeFrame(data []byte) (image.Image, error) {
	if len(data) < headerSize || !bytes.Equal(data[:4], frameMagic) {
		return nil, errors.New("decode frame: no frame written")
	}

	width := int(binary.LittleEndian.Uint32(data[4:8]))
	height := int(binary.LittleEndian.Uint32(data[8:12]))
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("decode frame: bad dimensions %dx%d", width, height)
	}
	if width*height*4 > maxFrameBytes {
		return nil, fmt.Errorf("decode frame: %dx%d exceeds the maximum frame size", width, height)
	}

	pixels := data[headerSize:]
	if len(pixels) != width*height*4 {
		return nil, fmt.Errorf("decode frame: got %d pixel bytes, want %d for %dx%d",
			len(pixels), width*height*4, width, height)
	}

	// swscale writes straight (non-premultiplied) alpha.
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	copy(img.Pix, pixels)
	return img, nil
}

// trimDiagnostics keeps guest stderr from flooding logs with decoder warnings.
func trimDiagnostics(s string) string {
	const limit = 512
	s = strings.TrimSpace(s)
	if len(s) > limit {
		return s[len(s)-limit:]
	}
	return s
}

// Close releases the compiled module and its runtime. It is intended for tests
// and for graceful shutdown; extraction after Close will fail.
func Close(ctx context.Context) error {
	if wazeroRuntime == nil {
		return nil
	}
	if err := wazeroRuntime.Close(ctx); err != nil {
		return fmt.Errorf("close wazero runtime: %w", err)
	}
	return nil
}
