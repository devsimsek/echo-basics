#!/usr/bin/env sh
# tests/fetch_by_id.sh
# Fetch a single log by ID.
#
# Usage:
#   ./tests/fetch_by_id.sh <id>
# or
#   ID=<id> ./tests/fetch_by_id.sh
#
# Defaults:
#   HOST -> http://localhost:${PORT:-6070}

HOST=${HOST:-http://localhost:${PORT:-6070}}
ID=${1:-${ID:-}}

if [ -z "$ID" ]; then
  echo "Usage: $0 <id>  OR set ID env var"
  exit 2
fi

echo "GET $HOST/api/fetch/i/$ID"
if command -v jq >/dev/null 2>&1; then
  curl -sS "$HOST/api/fetch/i/$ID" | jq .
else
  curl -sS "$HOST/api/fetch/i/$ID"
fi