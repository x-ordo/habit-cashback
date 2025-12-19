#!/usr/bin/env bash
set -euo pipefail

OWNER="${1:-}"
REPO="${2:-}"
REF="${3:-main}"

if [[ -z "$OWNER" || -z "$REPO" ]]; then
  echo "Usage: $0 <OWNER> <REPO> [ref=main]"
  exit 1
fi

SHA="$(gh api "repos/$OWNER/$REPO/commits/$REF" -q '.sha')"

# List check runs for a Git reference:
# GET /repos/{owner}/{repo}/commits/{ref}/check-runs  (GitHub Docs)
# We'll output JSON array of check run names for use as required_status_checks.contexts
gh api "repos/$OWNER/$REPO/commits/$SHA/check-runs" -q '[.check_runs[].name] | unique'
