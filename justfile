default: start

docker_build := "docker build -f build/Containerfile"

install-dependencies:
    brew install go k3d tilt

start:
    k3d registry create hyperboard
    k3d cluster create hyperboard --registry-use hyperboard --wait
    tilt up

stop:
    tilt down
    k3d cluster delete hyperboard
    k3d registry delete hyperboard