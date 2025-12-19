#!/usr/bin/env bash
set -euo pipefail

OWNER="${1:-}"
REPO="${2:-}"
BRANCH="${3:-main}"

if [[ -z "$OWNER" || -z "$REPO" ]]; then
  echo "Usage: $0 <OWNER> <REPO> [branch=main]"
  exit 1
fi

# 1) Determine required check contexts
# If REQUIRED_CONTEXTS_JSON is set, use it.
# Else try to auto-detect check runs for the branch HEAD.
# Else fallback to ["ci-go"].
REQUIRED_CONTEXTS_JSON="${REQUIRED_CONTEXTS_JSON:-}"

if [[ -z "$REQUIRED_CONTEXTS_JSON" ]]; then
  set +e
  REQUIRED_CONTEXTS_JSON="$(bash scripts/05_detect_checks.sh "$OWNER" "$REPO" "$BRANCH" 2>/dev/null)"
  set -e
fi

if [[ -z "$REQUIRED_CONTEXTS_JSON" || "$REQUIRED_CONTEXTS_JSON" == "[]" ]]; then
  REQUIRED_CONTEXTS_JSON='["ci-go"]'
fi

# 2) Apply protection via REST API: PUT /repos/{owner}/{repo}/branches/{branch}/protection
gh api -X PUT "repos/$OWNER/$REPO/branches/$BRANCH/protection" \
  -H "Accept: application/vnd.github+json" \
  -f required_status_checks.strict=true \
  -f required_status_checks.contexts:="$REQUIRED_CONTEXTS_JSON" \
  -f enforce_admins=true \
  -f required_pull_request_reviews.dismiss_stale_reviews=true \
  -f required_pull_request_reviews.require_code_owner_reviews=true \
  -f required_pull_request_reviews.required_approving_review_count=1 \
  -f required_linear_history=true \
  -f allow_force_pushes=false \
  -f allow_deletions=false \
  -f required_conversation_resolution=true >/dev/null

echo "OK: branch protection applied: $OWNER/$REPO@$BRANCH"
echo "Required checks: $REQUIRED_CONTEXTS_JSON"
