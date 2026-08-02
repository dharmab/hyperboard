# Media pipeline

Hyperboard processes uploads entirely in-process. There are no `cwebp`, `ffprobe` or
`ffmpeg` subprocesses and no cgo, so `hyperboard-api` runs on a distroless image with no
shell, package manager or libc.

| Step | Implementation |
|---|---|
| Decode uploaded image | `image` stdlib + `golang.org/x/image/webp` |
| Encode WebP (q85 content, q80 thumbnails) | `github.com/gen2brain/webp` — libwebp transpiled to Go |
| Video duration and audio-track detection | `github.com/abema/go-mp4`, `github.com/at-wat/ebml-go` |
| Video frame extraction | FFmpeg compiled to WebAssembly, run by `github.com/tetratelabs/wazero` |

## WebP encoding

`gen2brain/webp` is libwebp compiled to WASM and transpiled to Go, so encoding is a normal
Go function call.

**Builds must set the `nodynamic` build tag.** Without it the library `dlopen`s a system
libwebp when one is installed and only falls back to the transpiled code, which makes
output depend on the host. The Makefile, Tiltfile, `build/Containerfile` and
`.golangci.yml` all set it; if you invoke `go build` or `go test` directly, pass
`-tags nodynamic`.

## Container probing

Neither piece of video metadata Hyperboard needs requires a decoder. Duration comes from
the MP4 `mvhd` box or the Matroska `Segment/Info` element; audio detection comes from a
track's `hdlr` handler type (`soun`) or `TrackEntry.TrackType == 2`. Both are container
parsing, take about a millisecond, and match `ffprobe` exactly on the fixtures in
`internal/media/testdata/`.

The handler type is used rather than go-mp4's `Probe` helper because `Probe` only
classifies a few codecs, while `hdlr` is codec-agnostic.

## Frame extraction

`internal/media/ffwasm` embeds `frame.wasm`: a decode-only build of FFmpeg targeting
`wasm32-wasip1`, linked with `build/ffmpeg-wasm/frame.c`. The C program opens the video,
seeks to the requested offset, decodes one frame, converts it to RGBA with swscale, and
writes it to stdout with a 12-byte header. Go scales the result and encodes the thumbnail.

wazero executes the module with only the temp directory holding the input mounted,
read-only. The guest has no network, no threads and no other filesystem access.

Compiling the module takes about half a second, so `hyperboard-api` calls `ffwasm.Warm`
at startup rather than paying for it on the first upload. After that, extraction costs
roughly 10 ms for a 320×240 clip and 310 ms for 1920×1080 — about 3–4× slower than a
native `ffmpeg` subprocess, which is not material on an upload path.

Supported: H.264, HEVC, VP8, VP9, AV1, MPEG-4 Part 2 and MJPEG video in MP4/QuickTime and
Matroska/WebM containers. That covers every MIME type the API accepts.

### CPU architectures

`frame.wasm` is architecture-neutral, so it needs no attention when cross-compiling.
`CGO_ENABLED=0 GOOS=... GOARCH=... go build` works for every target Go supports, 32-bit
included, with no wasi-sdk and no build tags beyond `nodynamic`.

What does depend on the host is *speed*. wazero machine-compiles WebAssembly on amd64 and
arm64 and falls back to an interpreter everywhere else. Measured on the same machine with
the same module:

| | 64x64 frame | 1920x1080 frame | Module compile |
|---|---|---|---|
| wazero compiler (amd64, arm64) | 13 ms | 313 ms | 500 ms |
| wazero interpreter (everything else) | 814 ms | **25.6 s** | 52 ms |

Roughly 60-80x. Uploads still work on, say, riscv64 or ppc64le, but a 1080p video would
tie up a request for half a minute. `hyperboard-api` logs a warning at startup when it
lands on such a host. Images are unaffected — the WebP encoder is transpiled Go, not
WebAssembly.

Rebuilding the module (as opposed to using it) needs an amd64 or arm64 Linux builder,
since that is what wasi-sdk publishes; the Containerfile fails with a clear message
otherwise. The resulting `frame.wasm` is the same either way.

### Rebuilding the module

```
make build-ffmpeg-wasm
```

This runs `build/ffmpeg-wasm/build.sh`, which builds `build/ffmpeg-wasm/Containerfile`
with Docker and writes the artifacts into `internal/media/ffwasm/`. Re-run it after
editing `frame.c` or bumping `FFMPEG_VERSION` / `WASI_SDK_VERSION`. Nothing else in the
repo needs the wasi-sdk toolchain — `go build` uses the checked-in module.

Because the build pins an FFmpeg release, security fixes arrive by bumping
`FFMPEG_VERSION` in the Containerfile and rebuilding, not through the base image's package
manager.

## Licensing

Hyperboard is MIT. The embedded FFmpeg build enables **no GPL components**: there is no
`--enable-gpl`, no `--enable-version3`, and no external libraries at all — no libx264, no
libvpx, no libaom. Hyperboard never encodes video, so the encoders that carry the GPL are
not needed. Everything linked in is LGPL 2.1 or BSD: FFmpeg's native H.264, HEVC, VP8/VP9
and AV1 decoders, the mov/mp4 and Matroska demuxers, and swscale.

Static linking against LGPL 2.1 code carries a relinking obligation (section 6). It is
satisfied by what is in this repository:

- `internal/media/ffwasm/LICENSE.ffmpeg` — the LGPL 2.1 text.
- `internal/media/ffwasm/BUILD-INFO.txt` — the exact upstream version and the complete
  configure line.
- `build/ffmpeg-wasm/` — the build recipe and the source of the program linked against
  FFmpeg, so anyone can rebuild `frame.wasm` against a modified FFmpeg and drop it in.

H.264 and HEVC *patent* licensing is unchanged by any of this; it applies the same way to
a system `ffmpeg`.

## Container image

Removing the `ffmpeg` and `webp` packages lets `hyperboard-api` move from `debian:13-slim`
to the same distroless base as `hyperboard-web`. Measured on linux/amd64:

| | Binary | Rootfs | Compressed |
|---|---|---|---|
| Subprocess pipeline on `debian:13-slim` | 24.8 MB | 532 MB | 201 MB |
| In-process pipeline on distroless | 36.9 MB | 41 MB | 20 MB |

The binary grows by 12 MB, of which 2.8 MB is the FFmpeg module and the rest is the WebP
encoder and container parsers. The image shrinks by 491 MB, along with the entire Debian
userland CVE surface.
