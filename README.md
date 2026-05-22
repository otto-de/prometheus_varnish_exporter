# Varnish exporter for Prometheus

![Grafana example](dashboards/jonnenauha/dashboard.png)

Scrapes the `varnishstat -j` JSON output on each Prometheus collect and exposes all reported metrics. Metrics with
multiple backends or varnish defined identifiers (e.g. `VBE.*.happy SMA.*.c_bytes LCK.*.creat`) and other metrics with
similar structure (e.g. `MAIN.fetch_*`) are combined under a single metric name with distinguishable labels. Varnish
naming conventions are preserved as much as possible to be familiar to Varnish users when building queries, while trying to follow Prometheus conventions like lower casing and using `_` separators at the same time.

Handles runtime Varnish changes like adding new backends via vlc reload. Removed backends are reported by `varnishstat`
until Varnish is restarted.

Advanced users can use `-n -N`, they are passed to `varnishstat`.

Supported Varnish versions: **6.0 LTS** and newer (7.x, 8.x, 9.x). Older versions are not tested or supported.

Built queries can break on new versions if metric names or labels are refined. If you find bugs or have feature requests feel free to create issues or send PRs.

# Installing and running

You can find the latest binary releases for linux, darwin, windows, freebsd, openbsd and netbsd from
the [GitHub releases page](https://github.com/otto-de/prometheus_varnish_exporter/releases).

By default, the exporter listens on port `9131`. See `prometheus_varnish_exporter -h` for available options.

To test that `varnishstat` is found on the host machine and to preview all exported metrics run

    prometheus_varnish_exporter -test

# Troubleshooting

> Could not get hold of varnishd, is it running?
>
> 2020/12/18 20:22:33 [FATAL] Startup test: varnishstat scrape failed: exit status 1

The user running the exporter is missing permissions to access varnish services. To avoid using `sudo` see [#62](https://github.com/jonnenauha/prometheus_varnish_exporter/issues/62).

# Docker

Scraping metrics from Varnish running in a docker container is possible since 1.4.1. Resolve your Varnish container name
with `docker ps` and run the following. This will use `docker exec <container-name>` to execute varnishstat inside the
specified container.

    prometheus_varnish_exporter -docker-container-name <container_name>

## Images

Container images are built and published to our GHCR registry as part of the release process. Images are built for
multiple architectures (amd64, arm64) and for the latest major Varnish/Vinyl versions.

The image tags are in the format `<version>-varnish-<varnish-version>`. For example `1.7.0-varnish-8.0.0`.

[ghcr.io/otto-de/prometheus-varnish-exporter](https://github.com/orgs/otto-de/packages/container/package/prometheus-varnish-exporter)

# Grafana dashboards

The example dashboard is available [in JSON format](dashboards/jonnenauha/dashboard.json). Feel free to send PRs for additional dashboards or improvements to existing ones.

# Build

**One time setup**

This repository uses Golang with go modules. The minimum required version is defined in [go.mod](https://github.com/otto-de/prometheus_varnish_exporter/blob/master/go.mod#L3), later version are supported as well. To install go, [follow the upstream recommendations](https://golang.org/doc/install) or use the Golang package provided by your distribution.

**Development**

```bash
# clone
git clone git@github.com:otto-de/prometheus_varnish_exporter.git
cd prometheus_varnish_exporter

# build binary to current directory
go build

# draft local release with cross compilation
goreleaser release --snapshot --clean
```

## Release

Releases are done with [goreleaser](https://goreleaser.com/). Make sure you have it installed or use the GitHub Action
for releases.

To create a new release via the GitHub Action workflow you will have to create a new git tag with is a valid semantic
version.

1. Update `CHANGELOG.md` with the changes for the new version and commit the changes.
2. Create and push a new git tag:
   > [!NOTE]
   > Replace `v1.7.0` with your desired version number.
    ```bash
    git tag v1.7.0
    git push origin v1.7.0
    ```
3. Check the Actions tab on GitHub for the release workflow to complete.

This will trigger the release workflow that will build and publish
binaries to the GitHub releases and also build and publish docker images to the GHCR
registry.
