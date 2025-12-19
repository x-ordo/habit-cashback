#!/usr/bin/env bash
set -euo pipefail

OWNER="${1:-}"
REPO="${2:-}"
ENVIRONMENT="${3:-}"
CERT_PEM="${4:-}"
KEY_PEM="${5:-}"
CA_PEM="${6:-}"

if [[ -z "$OWNER" || -z "$REPO" || -z "$ENVIRONMENT" || -z "$CERT_PEM" || -z "$KEY_PEM" ]]; then
  echo "Usage: $0 <OWNER> <REPO> <environment> <cert_pem_path> <key_pem_path> [ca_pem_path]"
  exit 1
fi

FULL="$OWNER/$REPO"

if [[ ! -f "$CERT_PEM" ]]; then
  echo "Missing cert pem: $CERT_PEM"
  exit 1
fi
if [[ ! -f "$KEY_PEM" ]]; then
  echo "Missing key pem: $KEY_PEM"
  exit 1
fi

echo "Setting environment secrets for $FULL::$ENVIRONMENT ..."

# PEMs: set as multi-line secrets by reading from file (stdin)
gh secret set TOSS_MTLS_CERT_PEM -R "$FULL" --env "$ENVIRONMENT" < "$CERT_PEM"
gh secret set TOSS_MTLS_KEY_PEM  -R "$FULL" --env "$ENVIRONMENT" < "$KEY_PEM"

if [[ -n "${CA_PEM:-}" && -f "$CA_PEM" ]]; then
  gh secret set TOSS_MTLS_CA_PEM -R "$FULL" --env "$ENVIRONMENT" < "$CA_PEM"
fi

# Decryption materials: prefer env vars for non-interactive run
if [[ -z "${TOSS_DECRYPTION_KEY_B64:-}" ]]; then
  read -r -p "TOSS_DECRYPTION_KEY_B64 (paste, then Enter): " TOSS_DECRYPTION_KEY_B64
fi
if [[ -z "${TOSS_DECRYPTION_AAD:-}" ]]; then
  read -r -p "TOSS_DECRYPTION_AAD (paste, then Enter): " TOSS_DECRYPTION_AAD
fi

gh secret set TOSS_DECRYPTION_KEY_B64 -R "$FULL" --env "$ENVIRONMENT" --body "$TOSS_DECRYPTION_KEY_B64"
gh secret set TOSS_DECRYPTION_AAD     -R "$FULL" --env "$ENVIRONMENT" --body "$TOSS_DECRYPTION_AAD"

echo "OK: secrets set for $FULL::$ENVIRONMENT"
