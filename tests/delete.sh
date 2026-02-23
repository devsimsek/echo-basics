#!/usr/bin/env sh
# tests/delete.sh
# Delete a log by ID.
#
# Usage:
#   ./tests/delete.sh <id>
# or
#   ID=<id> ./tests/delete.sh
#
# Defaults:
#   HOST -> http://localhost:${PORT:-6070}

HOST=${HOST:-http://localhost:${PORT:-6070}}
ID=${1:-${ID:-}}

if [ -z "$ID" ]; then
  echo "Usage: $0 <id>  OR set ID env var"
  exit 2
fi

echo "DELETE $HOST/api/delete/$ID"

# Send DELETE request and append HTTP status on a new line.
RESP=$(curl -sS -X DELETE -w "\n%{http_code}" "$HOST/api/delete/$ID" || true)

HTTP_CODE=$(printf "%s\n" "$RESP" | tail -n1)
BODY=$(printf "%s\n" "$RESP" | sed '$d')

if [ -n "$BODY" ]; then
  if command -v jq >/dev/null 2>&1; then
    echo "$BODY" | jq .
  else
    echo "$BODY"
  fi
fi

echo "HTTP $HTTP_CODE"