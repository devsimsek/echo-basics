#!/usr/bin/env sh
# tests/fetch_by_flag.sh
# Fetch logs by flag (default "INFO").
# Usage:
#   ./tests/fetch_by_flag.sh [FLAG]
# or set env: FLAG=ERROR ./tests/fetch_by_flag.sh
#
HOST=${HOST:-http://localhost:${PORT:-6070}}
FLAG=${1:-${FLAG:-"INFO"}}

echo "GET $HOST/api/fetch/f/$FLAG"
if command -v jq >/dev/null 2>&1; then
  curl -sS "$HOST/api/fetch/f/$FLAG" | jq .
else
  curl -sS "$HOST/api/fetch/f/$FLAG"
fi