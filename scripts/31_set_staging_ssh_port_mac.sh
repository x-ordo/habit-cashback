#!/usr/bin/env bash
set -euo pipefail
REPO="${1:-}"
PORT="${2:-22}"

if [[ -z "${REPO}" ]]; then
  echo "Usage: $0 <OWNER/REPO> [PORT]"
  exit 1
fi

gh secret set STAGING_SSH_PORT -R "${REPO}" -b"${PORT}"
echo "[ok] STAGING_SSH_PORT=${PORT}"
