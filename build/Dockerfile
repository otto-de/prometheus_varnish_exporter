ARG VARNISH_VERSION
FROM varnish:${VARNISH_VERSION}
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/prometheus-varnish-exporter /usr/bin/
COPY build/entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
