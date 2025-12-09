#!/bin/bash
set -e

# optionally start a test varnish process for debugging
if [ -n "$START_VARNISH" ]; then
  echo "[INFO] Starting test varnish process..."
  varnishd -b 0.0.0.0:8080 -n /var/lib/varnish -T 0.0.0.0:6082
fi

VARNISH_VSM="${VARNISH_VSM:-/var/lib/varnish}"

# waiting for varnish to be ready if wait attempts and interval is set greater than 0
WAIT_ATTEMPTS="${WAIT_ATTEMPTS:-20}"
WAIT_INTERVAL="${WAIT_INTERVAL:-1}"
WAIT_INITIAL="${WAIT_INITIAL:-2}"
if [ "$#" -eq 0 ]; then
  set -- prometheus-varnish-exporter -n "${VARNISH_VSM}"
  if [ "${WAIT_ATTEMPTS}" -gt 0 ] && [ "${WAIT_INTERVAL}" -gt 0 ]; then
    echo "[INFO] Waiting ${WAIT_INITIAL}s before checking VSM at ${VARNISH_VSM}..."
    sleep "${WAIT_INITIAL}"

    for i in $(seq 1 "${WAIT_ATTEMPTS}"); do
      echo "[INFO] Attempt ${i}/${WAIT_ATTEMPTS}: Checking VSM..."

      if varnishstat -n "${VARNISH_VSM}" -1 >/dev/null 2>&1; then
        echo "[INFO] VSM is ready and accessible. Continue start..."
        break
      fi

      if [ "$i" -eq "${WAIT_ATTEMPTS}" ]; then
        echo "[ERROR] VSM not ready after ${WAIT_ATTEMPTS} attempts" >&2
        exit 1
      fi

      echo "[INFO] VSM not ready, waiting ${WAIT_INTERVAL}s..."
      sleep "${WAIT_INTERVAL}"
    done
  fi
fi

exec "$@"
