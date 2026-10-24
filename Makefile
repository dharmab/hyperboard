.PHONY: help install-deps install-javascript-deps build-images build-image-hyperboard-api build-image-hyperboard-async build-image-hyperboard-web build-image-hyperboardctl generate lint format test coverage govulncheck lint-go lint-javascript format-go format-javascript lint-tilt start stop ci clean

.DEFAULT_GOAL := help

GO_PACKAGES = $(shell go list ./... | grep -v '/node_modules/')

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  %-30s %s\n", $$1, $$2}'

install-deps: ## Install development dependencies via Homebrew
	brew install go k3d tilt oven-sh/bun/bun
	$(MAKE) install-javascript-deps

install-javascript-deps: ## Install locked JavaScript dependencies
	bun install --frozen-lockfile

build-images: build-image-hyperboard-web build-image-hyperboard-api build-image-hyperboard-async build-image-hyperboardctl ## Build all container images

build-image-hyperboard-web build-image-hyperboard-api build-image-hyperboard-async build-image-hyperboardctl:
	docker build -f build/Containerfile --target $(@:build-image-%=%) -t $(@:build-image-%=%):latest .

generate: ## Regenerate code from database schema and OpenAPI specs
	go generate ./...

lint: lint-go lint-javascript ## Lint source code

lint-go: ## Lint Go source code
	go tool golangci-lint run
	go vet $(GO_PACKAGES)
	go fix -diff $(GO_PACKAGES)
	find . -type d -name node_modules -prune -o -type f -name '*.go' -exec gofmt -s -d {} +

format: format-go format-javascript ## Format source code

format-go: ## Format Go source code
	go fix $(GO_PACKAGES)
	find . -type d -name node_modules -prune -o -type f -name '*.go' -exec gofmt -s -w {} +

lint-javascript: install-javascript-deps ## Lint JavaScript source code
	bun run lint

format-javascript: install-javascript-deps ## Format JavaScript source code
	bun run format

test: ## Run tests
	go test -race ./...

coverage: ## Run tests with coverage report
	go test -race -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

govulncheck: ## Run govulncheck
	go tool govulncheck ./...

lint-tilt: ## Validate the Tiltfile
	tilt alpha tiltfile-result >/dev/null

start: ## Start local development environment (k3d + Tilt)
	k3d registry create hyperboard
	k3d cluster create hyperboard --registry-use hyperboard --wait
	tilt up

stop: ## Stop and tear down local development environment
	tilt down
	k3d cluster delete hyperboard
	k3d registry delete hyperboard

ci: build-images lint test lint-tilt ## Run CI pipeline (build, lint, test, Tiltfile validation)

clean: ## Remove generated files and built binaries
	find . -name 'gen.go' -delete
	rm -f bin/hyperboard-api bin/hyperboard-async bin/hyperboard-web bin/hyperboardctl
