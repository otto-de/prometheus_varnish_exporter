#!/bin/bash
set -e

VARNISH_VSM="${VARNISH_VSM:-/var/lib/varnish}"

if [ "$#" -eq 0 ]; then
  set -- prometheus-varnish-exporter -n "${VARNISH_VSM}"
elif [[ "$1" == -* ]]; then
  # If the first argument starts with a dash, prepend the default command
  set -- prometheus-varnish-exporter "$@"
fi

exec "$@"
