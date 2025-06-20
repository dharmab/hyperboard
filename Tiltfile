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

load('ext://restart_process', 'docker_build_with_restart')
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

api='hyperboard-api'
web='hyperboard-web'
db='postgresql'
cli='hyperboardctl'

go_build(cli, 'darwin', 'arm64') # (re)build the CLI automatically
for name in [api, web]:
    go_build(name, 'linux', 'amd64') # (re)build the API and web server binaries automatically
    container_build_with_restart(name) # (re)build the API and web server container images automatically
for name in [api, web, db]:
    k8s_yaml("deploy/tilt/{}.yaml".format(name)) # continuously deploy the API, web server, and database manifests

# TCP ports that are bound on the host machine.
# Change these if you have port conflicts.
host_web_port=8080
host_api_port=8081
host_db_port=5432

k8s_resource(
    workload=db,
    port_forwards="{}:5432".format(host_db_port), # Make the database accessible on the host machine (connect with psql/DataGrip/DBeaver)
)
k8s_resource(
    workload=api,
    port_forwards="{}:8080".format(host_api_port), # Make the API accessible on the host machine (access with cli/curl)
    resource_deps=[db], # Wait to start the API until the database is ready
)
k8s_resource(
    workload=web,
    port_forwards="{}:8080".format(host_web_port), # Make the website accessible on the host machine (open in a web browser)
    resource_deps=[api], # Wait to start the website until the API is ready
)
