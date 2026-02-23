#!/usr/bin/env sh
# tests/create.sh
# Create a single log record. You can override FLAG and MESSAGE env vars.
# Defaults:
#   HOST -> http://localhost:${PORT:-6070}
#   FLAG -> "INFO"
#   MESSAGE -> "test message from curl"

HOST=${HOST:-http://localhost:${PORT:-6070}}
FLAG=${FLAG:-"INFO"}
MESSAGE=${MESSAGE:-"test message from curl"}

echo "POST $HOST/api/create"

# Send request and append HTTP status on a new line, then split response/body and code safely.
RESP=$(curl -sS -H "Content-Type: application/json" -d "{\"flag\":\"${FLAG}\",\"message\":\"${MESSAGE}\"}" -w "\n%{http_code}" "$HOST/api/create" || true)

HTTP_CODE=$(printf "%s\n" "$RESP" | tail -n1)
BODY=$(printf "%s\n" "$RESP" | sed '$d')

if command -v jq >/dev/null 2>&1; then
  echo "$BODY" | jq .
  # Try several common JSON key casings to extract ID
  CREATED_ID=$(echo "$BODY" | jq -r '.ID // .id // .Id // .Id' 2>/dev/null || true)
  echo "HTTP $HTTP_CODE, created ID: $CREATED_ID"
else
  echo "$BODY"
  echo "HTTP $HTTP_CODE"
  echo "Install jq to pretty-print and auto-extract the created ID"
fi