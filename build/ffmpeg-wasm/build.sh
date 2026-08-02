#!/usr/bin/env bash
# Rebuild internal/media/ffwasm/frame.wasm from build/ffmpeg-wasm/Containerfile.
#
# The result is checked into the repo so that `go build` needs no toolchain beyond Go.
# Re-run this when bumping FFMPEG_VERSION or WASI_SDK_VERSION, or when editing frame.c.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "${here}/../.." && pwd)"
out="${root}/internal/media/ffwasm"

mkdir -p "${out}"
docker build \
  -f "${here}/Containerfile" \
  --target artifact \
  --output "type=local,dest=${out}" \
  "${here}"

ls -l "${out}/frame.wasm"
