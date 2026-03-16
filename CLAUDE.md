# Hyperboard

Hyperboard is an image and video hosting web application built with Go, following [Hypermedia Systems](https://hypermedia.systems/) principles. It consists of three binaries: `hyperboard-web` (web frontend), `hyperboard-api` (REST API), and `hyperboardctl` (CLI tool).

## Project Structure

- `cmd/` — Entry points for each binary (`hyperboard-api`, `hyperboard-web`, `hyperboardctl`)
- `internal/api/` — API server implementation, OpenAPI spec (`internal/api/spec/`), and generated server code
- `internal/web/` — Web frontend server (handlers, templates, static assets)
- `internal/cli/` — CLI tool implementation with subcommands (`posts/`, `notes/`, `tags/`, `tagcategories/`, `replace/`)
- `internal/db/` — Database layer (migrations in `internal/db/migrations/data/`, models in `internal/db/models/`, data access in `internal/db/store/`)
- `internal/search/` — Search query parsing, sorting, and tag-based search
- `internal/media/` — Media processing (images, video, perceptual hashing)
- `internal/storage/` — Storage abstraction (S3 and in-memory implementations)
- `internal/middleware/auth/` — HTTP Basic Auth middleware
- `internal/middleware/logging/` — HTTP request logging
- `internal/middleware/security/` — Security headers middleware
- `pkg/client/` — Generated API client (from OpenAPI spec)
- `pkg/types/` — Shared types (generated from OpenAPI spec)
- `build/Containerfile` — Multi-stage container build for all three binaries
- `deploy/tilt/` — Kubernetes manifests for local development
- `deploy/quadlet/` — Podman Quadlet container deployment files
- `docs/` — Project documentation

## Prerequisites

```
make install-deps
```

This installs Go, k3d, and Tilt via Homebrew.

## Code Generation

Run `make generate` to regenerate OpenAPI types, server stubs, and client code via `oapi-codegen` from specs in `internal/api/spec/`. This runs `go generate ./...` which processes `//go:generate` directives in source files.

Generated files (`gen.go`) should not be edited by hand.

If Tilt is running (`tilt get uiresources`), `go generate ./...` runs automatically when source files change. You can also manually trigger it with `tilt trigger generate`.

## Building

```
# Build container images for all binaries
make build-images

# Or build individual images
make build-image-hyperboard-api
make build-image-hyperboard-web
make build-image-hyperboardctl
```

The Tiltfile also builds binaries and container images automatically during local development.

## Local Development Environment

The project uses Tilt with a k3d (k3s-in-Docker) cluster for local development.

```
# Start the full stack (k3d cluster + Tilt with hot reload)
make start

# Stop and tear down
make stop
```

Default ports on the host (configurable in `Tiltfile`):
- Web: 8080
- API: 8081
- PostgreSQL: 5432
- S3 (RustFS): 9000 (API), 9001 (console)

## Running Tests

```
make test
```

## Linting

```
make lint
```

Always use `make lint` rather than running individual linter commands. This runs golangci-lint, go vet, go fix, and gofmt.

## Formatting

```
make format
```

## CI

```
make ci
```

## Cleaning Up

```
make clean
```

Removes generated files (`gen.go`) and built binaries.
