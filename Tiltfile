load('ext://restart_process', 'docker_build_with_restart')

api='hyperboard-api'
web='hyperboard-web'
cli='hyperboardctl'

def go_build(name, os, arch):
    local_resource(
        name + "-bin",
        ['go', 'build', '-o', 'bin/{}'.format(name), './cmd/{}'.format(name)],
        env={
            'GOOS': os,
            'GOARCH': arch,
        },
        deps=['go.mod', 'go.sum', './cmd/{}'.format(name), './pkg'],
    )

def container_build_with_restart(name):
    entrypoint = "/{}".format(name)
    docker_build_with_restart(
        ref=name,
        context='.',
        dockerfile='build/Containerfile',
        target=name,
        entrypoint=[entrypoint],
        live_update=[sync("bin/{}".format(name), entrypoint)]
    )

go_build(cli, 'darwin', 'arm64')
for name in [api, web]:
    go_build(name, 'linux', 'amd64')
    container_build_with_restart(name)
    k8s_yaml("deploy/tilt/{}.yaml".format(name))
k8s_resource(workload=web, port_forwards=8080)
k8s_resource(workload=api, port_forwards=8081)
