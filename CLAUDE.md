# Hyperboard

Hyperboard is an image and video hosting web application built with Go, following [Hypermedia Systems](https://hypermedia.systems/) principles. It consists of three binaries: `hyperboard-web` (web frontend), `hyperboard-api` (REST API), and `hyperboardctl` (CLI tool).

## Project Structure

- `cmd/` — Entry points for each binary (`hyperboard-api`, `hyperboard-web`, `hyperboardctl`)
- `internal/db/` — Database layer (migrations in `internal/db/migrations/data/`, generated models in `internal/db/models/`)
- `pkg/api/` — OpenAPI spec and generated API client/server code
- `pkg/types/` — Shared types (also generated from OpenAPI spec)
- `pkg/httplog/` — HTTP request logging
- `build/Containerfile` — Multi-stage container build for all three binaries
- `deploy/tilt/` — Kubernetes manifests for local development
- `scripts/generate/` — Code generation script (runs embedded Postgres, applies migrations, generates Bob ORM models)

## Prerequisites

```
make install-deps
```

This installs Go, k3d, and Tilt via Homebrew.

## Code Generation

Run `make generate` to regenerate:

- **Bob ORM models** from database schema — the generate script starts an embedded Postgres instance, runs migrations, and uses `bobgen-psql` to produce `internal/db/models/*.bob.go`
- **OpenAPI types and server stubs** via `oapi-codegen` from specs in `pkg/api/spec/`

Generated files should not be edited by hand. If the database schema changes (new migration in `internal/db/migrations/data/`), re-run generation.

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
go test ./...
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

Removes generated files (`gen.go`, `*.bob*.go`) and built binaries.
