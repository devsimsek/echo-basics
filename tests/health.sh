#!/usr/bin/env sh
# tests/health.sh
# Simple health check for the Echo service
# Defaults: HOST=http://localhost:6070 (or override HOST and/or PORT env vars)

HOST=${HOST:-http://localhost:${PORT:-6070}}

echo "GET $HOST/api/health"
curl -sS -w "\nHTTP_STATUS:%{http_code}\n" "$HOST/api/health" | tee /dev/stderr