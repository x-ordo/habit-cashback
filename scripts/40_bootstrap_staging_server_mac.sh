#!/usr/bin/env bash
set -euo pipefail

HOST="${1:-}"
USER="${2:-ubuntu}"

if [[ -z "${HOST}" ]]; then
  echo "Usage: $0 <SERVER_IP_OR_HOST> <SSH_USER>"
  exit 1
fi

# 1) 서버에 폴더 생성
ssh "${USER}@${HOST}" "sudo mkdir -p /opt/habitcashback && sudo chown -R ${USER}:${USER} /opt/habitcashback"

# 2) 레포 내용 업로드 (git 대신 rsync)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

rsync -av --delete   --exclude ".git"   --exclude "frontend/node_modules"   "${ROOT_DIR}/" "${USER}@${HOST}:/opt/habitcashback/"

# 3) docker 설치 여부 체크 + compose 실행
ssh "${USER}@${HOST}" <<'EOF'
set -e
if ! command -v docker >/dev/null 2>&1; then
  echo "[!] docker not found. Install Docker first."
  exit 1
fi

cd /opt/habitcashback/infra/staging
cp -n .env.example .env || true
echo "[i] bootstrap complete. Edit /opt/habitcashback/infra/staging/.env then run: docker compose up -d"
EOF
