.PHONY: help install-deps build-images build-ffmpeg-wasm generate lint format test coverage govulncheck start stop ci clean

.DEFAULT_GOAL := help

# github.com/gen2brain/webp dlopens a system libwebp unless the nodynamic build
# tag is set, which would make WebP output depend on the host. Exported so every
# go invocation below (including golangci-lint) picks it up.
export GOFLAGS := -tags=nodynamic

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  %-30s %s\n", $$1, $$2}'

install-deps: ## Install development dependencies via Homebrew
	brew install go k3d tilt

build-images: build-image-hyperboard-web build-image-hyperboard-api build-image-hyperboardctl ## Build all container images

build-image-hyperboard-web build-image-hyperboard-api build-image-hyperboardctl:
	docker build -f build/Containerfile --target $(@:build-image-%=%) -t $(@:build-image-%=%):latest .

generate: ## Regenerate code from database schema and OpenAPI specs
	go generate ./...

lint: ## Run linters
	go tool golangci-lint run
	go vet ./...
	go fix -diff ./...
	gofmt -s -d .

format: ## Format source code
	go fix ./...
	gofmt -s -w .

test: ## Run tests
	go test -race ./...

coverage: ## Run tests with coverage report
	go test -race -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

govulncheck: ## Run govulncheck
	go tool govulncheck ./...

start: ## Start local development environment (k3d + Tilt)
	k3d registry create hyperboard
	k3d cluster create hyperboard --registry-use hyperboard --wait
	tilt up

stop: ## Stop and tear down local development environment
	tilt down
	k3d cluster delete hyperboard
	k3d registry delete hyperboard

ci: build-images lint test ## Run CI pipeline (build, lint, test)

build-ffmpeg-wasm: ## Rebuild the embedded LGPL FFmpeg WASM module
	build/ffmpeg-wasm/build.sh

clean: ## Remove generated files and built binaries
	find . -name 'gen.go' -delete
	rm -f bin/hyperboard-api bin/hyperboard-web bin/hyperboardctl
