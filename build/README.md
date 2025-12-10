# prometheus-varnish-exporter Docker image

This is the Docker Image for [prometheus_varnish_exporter](https://github.com/jonnenauha/prometheus_varnish_exporter) sidecar container to expose Prometheus metrics for Varnish.

## Configuration

You can set the following environment variable to customize the behavior:

- `VARNISH_VSM` location of the Varnish VSM (defaults to `/var/lib/varnish`)

## Building

Navigate to the root of this repository and run:

```shell
goreleaser release --snapshot --clean
```

## Running

```shell
docker image ls ghcr.io/otto-de/prometheus-varnish-exporter
# select the desired image from the list and run it like this:
docker run --rm -it ghcr.io/otto-de/prometheus-varnish-exporter:latest-arm64 --help
# or in debug mode
docker run -e START_VARNISH=1 --rm -it <IMAGE from above>
```
