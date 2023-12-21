# Shig SFU
Shig is a Fediverse service designed to distribute and clone live streams among Fediverse instances. 
Shig is constructed based on the WHIP/WHEP approach for both incoming and outgoing streams.

## Documentaion

- [Use Cases](docs/use-cases.md)
- [Entities](docs/entities.md)
- [Architecture](docs/architecture.md)
- [WHIP/WHEP: Lobby Session initialization](docs/whip-whep.md)
- [ShigCLT: Start stream via commandline tool](docs/cli.md)
- [Backpressure](docs/backpressure.md)

## Quick Start

### Run SFU

```shell
make run
```

### Run with some Data

Normally, the SFU (Shig instance) listens for data from other Fediverse instances.
However, to facilitate development, you could start the SFU with an included SQLite database.
Run inside the SFU project

```shell
move shig-dev.db shig.db
```

Now you can start the SFU with some data.

### Build

```shell
make build
``` 

#### If you have Mac M1 / M2 -> Build a linux arm64 with Docker

Once time build a docker image for your container

```shell
export DOCKER_DEFAULT_PLATFORM=linux/amd64
docker build -t shig-builder .
```

###### Build Shig Instance
```shell
export DOCKER_DEFAULT_PLATFORM=linux/amd64
docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp shig-builder make build-linux
```
###### Build Shig CLT
```shell
export DOCKER_DEFAULT_PLATFORM=linux/amd64
docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp shig-builder make build-clt-linux
```


## Monitoring

### Requirement

please install grafana loki docker plugin

```shell
docker plugin install grafana/loki-docker-driver:latest --alias loki --grant-all-permissions
```

### Start Develop Monitoring 
```shell
make monitor
```


