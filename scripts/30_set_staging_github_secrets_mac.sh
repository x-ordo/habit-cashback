#!/usr/bin/env bash
set -euo pipefail

REPO="${1:-}"
HOST="${2:-}"
USER="${3:-}"
KEY_PATH="${4:-}"
DOMAIN="${5:-}"
GHCR_USERNAME="${6:-}"
GHCR_TOKEN_PATH="${7:-}"

if [[ -z "${REPO}" || -z "${HOST}" || -z "${USER}" || -z "${KEY_PATH}" || -z "${DOMAIN}" ]]; then
  echo "Usage: $0 <OWNER/REPO> <STAGING_SSH_HOST> <STAGING_SSH_USER> <PATH_TO_PRIVATE_KEY> <STAGING_DOMAIN> [GHCR_USERNAME] [PATH_TO_GHCR_TOKEN_FILE]"
  exit 1
fi

if ! command -v gh >/dev/null 2>&1; then
  echo "[!] gh (GitHub CLI) is required."
  exit 1
fi

KEY_CONTENT="$(cat "${KEY_PATH}")"

gh secret set STAGING_SSH_HOST -R "${REPO}" -b"${HOST}"
gh secret set STAGING_SSH_USER -R "${REPO}" -b"${USER}"
gh secret set STAGING_SSH_KEY  -R "${REPO}" -b"${KEY_CONTENT}"
gh secret set STAGING_DOMAIN   -R "${REPO}" -b"${DOMAIN}"

if [[ -n "${GHCR_USERNAME}" && -n "${GHCR_TOKEN_PATH}" ]]; then
  TOKEN_CONTENT="$(cat "${GHCR_TOKEN_PATH}")"
  gh secret set GHCR_USERNAME -R "${REPO}" -b"${GHCR_USERNAME}"
  gh secret set GHCR_TOKEN    -R "${REPO}" -b"${TOKEN_CONTENT}"
  echo "[ok] secrets set + GHCR credentials"
else
  echo "[ok] secrets set (SSH+DOMAIN). If repo/package is private, also set GHCR_USERNAME / GHCR_TOKEN."
fi
