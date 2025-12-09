# prometheus-varnish-exporter Docker image

This is the Docker Image for [prometheus_varnish_exporter](https://github.com/jonnenauha/prometheus_varnish_exporter) sidecar container to expose Prometheus metrics for Varnish.

## Configuration

You can set the following environment variable to customize the behavior:

- `VARNISH_VSM` location of the Varnish VSM (defaults to `/var/lib/varnish`)
- `VARNISH_ADMIN_PORT` the varnish admin port (defaults to `6082`)
- `WAIT_TIMEOUT_ATTEMPTS` the number of attempts to wait for Varnish to be listening (defaults to 20)
- `WAIT_TIMEOUT_INTERVAL_SECONDS` the interval to wait for Varnish to be available (defaults to 1 second)
- `WAIT_TIMEOUT_INITIAL_SECONDS` the duration to wait before starting to wait for Varnish to be available (defaults to 10 seconds)

## Building

Navigate to the root of this repository and run:

```shell
goreleaser release --snapshot --clean
```

## Running

```shell
docker image ls ghcr.io/otto-de/prometheus-varnish-exporter
# select the desired image from the list and run it
docker run --rm -it <IMAGE from above> /bin/bash
# or in debug mode
docker run -e START_VARNISH=1 --rm -it <IMAGE from above>
```
