.PHONY: install-deps build-images generate lint format test start stop ci clean

install-deps:
	brew install go k3d tilt

build-images: build-image-hyperboard-web build-image-hyperboard-api build-image-hyperboardctl

build-image-hyperboard-web build-image-hyperboard-api build-image-hyperboardctl:
	docker build -f build/Containerfile --target $(@:build-image-%=%) -t $(@:build-image-%=%):latest .

generate:
	go generate ./...

lint:
	go tool golangci-lint run
	go vet ./...
	go fix -diff ./...
	gofmt -s -d .

format:
	go fix ./...
	gofmt -s -w .

test:
	go test ./...

start:
	k3d registry create hyperboard
	k3d cluster create hyperboard --registry-use hyperboard --wait
	tilt up

stop:
	tilt down
	k3d cluster delete hyperboard
	k3d registry delete hyperboard

ci: build-images lint test

clean:
	find . -name 'gen.go' -delete
	rm -f internal/db/models/*.bob*.go
	rm -f bin/hyperboard-api bin/hyperboard-web bin/hyperboardctl
