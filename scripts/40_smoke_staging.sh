#!/usr/bin/env bash
set -euo pipefail
BASE="${1:-}"
if [[ -z "${BASE}" ]]; then
  echo "usage: bash scripts/40_smoke_staging.sh https://YOUR_STAGING_API"
  exit 1
fi

echo "[smoke] GET ${BASE}/health"
curl -fsS "${BASE}/health" | head -c 300; echo
echo "[smoke] GET ${BASE}/meta"
curl -fsS "${BASE}/meta" | head -c 300; echo

echo "[smoke] POST auth/exchange"
TOKEN=$(curl -fsS -X POST "${BASE}/v1/auth/exchange" -H "Content-Type: application/json" -d '{}' | python3 -c "import sys,json; print(json.load(sys.stdin).get('accessToken',''))")
if [[ -z "${TOKEN}" ]]; then
  echo "❌ token empty"
  exit 1
fi
echo "✅ token ok"

echo "[smoke] GET challenges"
curl -fsS "${BASE}/v1/challenges" -H "Authorization: Bearer ${TOKEN}" | head -c 400; echo

echo "✅ smoke done"
