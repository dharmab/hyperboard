default: start

install-dependencies:
    brew install go k3d tilt

build-images: (build-image "hyperboard-web") (build-image "hyperboard-api") (build-image "hyperboardctl")

build-image target:
    docker build -f build/Containerfile --target {{target}} -t {{target}}:latest .

start:
    k3d registry create hyperboard
    k3d cluster create hyperboard --registry-use hyperboard --wait
    tilt up

stop:
    tilt down
    k3d cluster delete hyperboard
    k3d registry delete hyperboard
