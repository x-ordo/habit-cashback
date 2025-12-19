#!/usr/bin/env bash
set -euo pipefail

OWNER="${1:-}"
REPO="${2:-}"

if [[ -z "$OWNER" || -z "$REPO" ]]; then
  echo "Usage: $0 <OWNER> <REPO>"
  exit 1
fi

# REST API: Create or update an environment
# PUT /repos/{owner}/{repo}/environments/{environment_name}
# We'll create 2 environments without additional protection rules (MVP default).
for ENV in staging production; do
  gh api -X PUT "repos/$OWNER/$REPO/environments/$ENV" \
    -H "Accept: application/vnd.github+json" >/dev/null
  echo "OK: environment ensured: $OWNER/$REPO::$ENV"
done
