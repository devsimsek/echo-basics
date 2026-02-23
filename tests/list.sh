#!/usr/bin/env sh
# tests/list.sh
# List all logs
# Defaults:
#   HOST -> http://localhost:${PORT:-6070}
#
HOST=${HOST:-http://localhost:${PORT:-6070}}

echo "GET $HOST/api/list"
if command -v jq >/dev/null 2>&1; then
  curl -sS "$HOST/api/list" | jq .
else
  curl -sS "$HOST/api/list"
fi