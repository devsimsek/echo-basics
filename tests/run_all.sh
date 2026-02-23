#!/usr/bin/env sh
# tests/run_all.sh
# Run a quick smoke test suite against the API using the other test scripts in this directory.
# Usage:
#   ./tests/run_all.sh
# Environment:
#   HOST - override full host including scheme and port (default http://localhost:${PORT:-6070})
#   PORT - used by default when HOST not set
#
# Behavior:
#  - Runs health -> create -> list -> fetch_by_flag -> (fetch_by_id, delete) -> list
#  - Attempts to auto-detect the created resource ID using jq if available, otherwise falls back to parsing create.sh output.
#  - Exits non-zero on any command failure (set -e). If you want to continue on failures, remove `set -e`.

set -eu

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
# Ensure we run the helper scripts from the tests directory
cd "$ROOT_DIR"

HOST=${HOST:-http://localhost:${PORT:-6070}}

echo "Running tests against: $HOST"
echo

# Helper to run a test script and print separators
run_script() {
  script="$1"
  shift
  echo ">>> Running $script $*"
  sh "./$script" "$@"
  echo ">>> Finished $script"
  echo
}

# Health check
run_script health.sh

# Create: capture full output so we can extract the JSON and ID
echo "== create =="
CREATE_OUTPUT="$(sh ./create.sh || true)" || true
echo "$CREATE_OUTPUT"
echo

# Try to extract JSON body (everything before a line that starts with "HTTP ")
JSON_BODY="$(printf "%s\n" "$CREATE_OUTPUT" | sed '/^HTTP /,$d')"
# Trim leading/trailing whitespace
JSON_BODY="$(printf "%s" "$JSON_BODY")"

CREATED_ID=""

if [ -n "$JSON_BODY" ]; then
  if command -v jq >/dev/null 2>&1; then
    # Try multiple common key names
    CREATED_ID="$(printf "%s\n" "$JSON_BODY" | jq -r '.ID // .id // .Id // .Id' 2>/dev/null || true)"
    if [ -z "$CREATED_ID" ] || [ "$CREATED_ID" = "null" ]; then
      CREATED_ID=""
    fi
  else
    # Fallback: try to parse the "HTTP ... created ID: <id>" line from create output
    HTTP_LINE="$(printf "%s\n" "$CREATE_OUTPUT" | grep '^HTTP ' | head -n1 || true)"
    if [ -n "$HTTP_LINE" ]; then
      # expected format: HTTP <code>, created ID: <id>
      CREATED_ID="$(printf "%s\n" "$HTTP_LINE" | awk -F'created ID: ' '{print $2}' | tr -d '[:space:]' || true)"
    fi
  fi
fi

if [ -n "$CREATED_ID" ]; then
  echo "Detected created ID: $CREATED_ID"
else
  echo "Warning: could not detect created ID. Install 'jq' for robust detection or check create.sh output."
fi

echo
# List
echo "== list =="
run_script list.sh

echo "== fetch_by_flag (INFO) =="
run_script fetch_by_flag.sh "INFO"

# If we have an ID, run fetch_by_id and delete
if [ -n "$CREATED_ID" ]; then
  echo "== fetch_by_id ($CREATED_ID) =="
  run_script fetch_by_id.sh "$CREATED_ID"

  echo "== delete ($CREATED_ID) =="
  run_script delete.sh "$CREATED_ID"

  echo "== list (after delete) =="
  run_script list.sh
else
  echo "Skipping fetch_by_id/delete because created ID unknown."
fi

echo
echo "All tests completed."