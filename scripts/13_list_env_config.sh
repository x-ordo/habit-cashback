#!/usr/bin/env bash
set -euo pipefail

OWNER="${1:-}"
REPO="${2:-}"

if [[ -z "$OWNER" || -z "$REPO" ]]; then
  echo "Usage: $0 <OWNER> <REPO>"
  exit 1
fi

FULL="$OWNER/$REPO"

echo "== Repo variables =="
gh variable list -R "$FULL" || true

for ENV in staging production; do
  echo ""
  echo "== Environment: $ENV =="
  echo "-- secrets --"
  gh secret list -R "$FULL" --env "$ENV" || true
  echo "-- variables --"
  gh variable list -R "$FULL" --env "$ENV" || true
done
