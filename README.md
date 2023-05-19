# Shig SFU
Selected Forward Unit for distributed environments

## Quick Start
### Start SFU

```shell
make run
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
