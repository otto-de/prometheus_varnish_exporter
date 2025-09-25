FROM golang:latest AS gobuilder

RUN apt update && apt install -y --no-install-recommends jq curl git

WORKDIR /prometheus_varnish_exporter

RUN git clone https://github.com/otto-de/prometheus_varnish_exporter /prometheus_varnish_exporter

# uncomment once we have releases
#RUN curl -s https://api.github.com/repos/otto-de/prometheus_varnish_exporter/releases/latest | jq -r .tag_name > /prometheus_varnish_exporter/Version
#RUN git checkout tags/$(cat /prometheus_varnish_exporter/Version)

RUN go build -o /go/bin/prometheus_varnish_exporter

FROM varnish:7.4.2

EXPOSE 9131
COPY --from=gobuilder /go/bin/prometheus_varnish_exporter /usr/local/bin
USER root
ENTRYPOINT ["/usr/local/bin/prometheus_varnish_exporter"]
