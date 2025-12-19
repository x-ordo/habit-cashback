#!/usr/bin/env bash
set -euo pipefail

OWNER="${1:-}"
REPO="${2:-}"

if [[ -z "$OWNER" || -z "$REPO" ]]; then
  echo "Usage: $0 <OWNER> <REPO>"
  exit 1
fi

FULL="$OWNER/$REPO"

# Variables are NOT secrets; keep them non-sensitive.
# You can safely commit these defaults in docs/config too, but keeping in GitHub variables helps centralize.
gh variable set SERVICE_NAME -R "$FULL" --body "habitcashback" >/dev/null
gh variable set TOSS_DEVELOPER_CENTER_URL -R "$FULL" --body "https://developers-apps-in-toss.toss.im/" >/dev/null

echo "OK: repository variables set: $FULL"
