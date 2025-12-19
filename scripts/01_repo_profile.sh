#!/usr/bin/env bash
set -euo pipefail

OWNER="${1:-}"
REPO="${2:-}"
if [[ -z "$OWNER" || -z "$REPO" ]]; then
  echo "Usage: $0 <OWNER> <REPO>"
  exit 1
fi

FULL="$OWNER/$REPO"
DESCRIPTION="습관환급 — 보증금을 걸고 습관을 지키면 환급(리워드) 받는 Apps in Toss 미션 플랫폼 (Go backend)"
HOMEPAGE="${HOMEPAGE_URL:-}"

if [[ -n "$HOMEPAGE" ]]; then
  gh repo edit "$FULL" --description "$DESCRIPTION" --homepage "$HOMEPAGE"
else
  gh repo edit "$FULL" --description "$DESCRIPTION"
fi

TOPICS=(
  "apps-in-toss" "toss" "fintech" "habit" "challenge" "reward"
  "go" "postgres" "oauth2" "mtls" "webview" "tds"
)

for t in "${TOPICS[@]}"; do
  gh repo edit "$FULL" --add-topic "$t" >/dev/null || true
done

echo "OK: repo profile updated: $FULL"
