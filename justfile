default: start

install-dependencies:
    brew install go k3d tilt

build-images: (build-image "hyperboard-web") (build-image "hyperboard-api") (build-image "hyperboardctl")

build-image target:
    docker build -f build/Containerfile --target {{target}} -t {{target}}:latest .

lint: lint-go

lint-go:
    @echo "Running Go linter..."
    go tool golangci-lint run
    @echo "Checking Go code formatting..."
    gofmt -s -d .

format: format-go

format-go:
    @echo "Formatting Go code..."
    gofmt -s -w .

start:
    k3d registry create hyperboard
    k3d cluster create hyperboard --registry-use hyperboard --wait
    tilt up

stop:
    tilt down
    k3d cluster delete hyperboard
    k3d registry delete hyperboard
