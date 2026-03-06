#!/bin/bash
set -e

if [ "$#" -eq 0 ]; then
  set -- prometheus-varnish-exporter
elif [[ "$1" == -* ]]; then
  # If the first argument starts with a dash, prepend the default command
  set -- prometheus-varnish-exporter "$@"
else
  # Otherwise, use the provided command as is
  set -- "$@"
fi

exec "$@"
