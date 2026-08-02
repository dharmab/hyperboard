# Video test fixtures

Decoding video needs real encoded video, so these are checked in. They are kept as small
as possible — 64x64, one second, heavily quantized — because nothing in the tests depends
on their content beyond "it decodes to something that is not a flat color". Together they
are about 35 KB and they never change.

Between them they cover every codec and container the API accepts, with and without an
audio track.

| File | Video | Audio |
|---|---|---|
| `h264.mp4` | H.264 | — |
| `h264-audio.mp4` | H.264 | AAC |
| `h264-audio.mov` | H.264 | AAC |
| `hevc.mp4` | HEVC | — |
| `vp9.webm` | VP9 | — |
| `vp8-audio.webm` | VP8 | Opus |

Regenerate with a system ffmpeg (not needed to build or test Hyperboard):

```sh
src="-f lavfi -i testsrc2=size=64x64:rate=10:duration=1"
audio="-f lavfi -i sine=frequency=440:duration=1"
aopts="-c:a aac -b:a 8k -ac 1 -ar 8000 -shortest"

ffmpeg $src -c:v libx264 -pix_fmt yuv420p -g 5 -crf 35 h264.mp4
ffmpeg $src $audio -c:v libx264 -pix_fmt yuv420p -g 5 -crf 35 $aopts h264-audio.mp4
ffmpeg $src $audio -c:v libx264 -pix_fmt yuv420p -g 5 -crf 35 $aopts h264-audio.mov
ffmpeg $src -c:v libx265 -pix_fmt yuv420p -g 5 -crf 35 -tag:v hvc1 hevc.mp4
ffmpeg $src -c:v libvpx-vp9 -pix_fmt yuv420p -g 5 -b:v 30k vp9.webm
ffmpeg $src $audio -c:v libvpx -pix_fmt yuv420p -g 5 -b:v 30k \
  -c:a libopus -b:a 6k -ac 1 -shortest vp8-audio.webm
```

The expected durations and audio flags in `probe_test.go` come from `ffprobe` run against
these files.
