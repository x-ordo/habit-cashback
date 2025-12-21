#!/bin/bash
# ait_deploy.sh - .ait 번들 업로드 스크립트
# 사용법: ./scripts/ait_deploy.sh <API_KEY>
#      또는: AIT_API_KEY="..." ./scripts/ait_deploy.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
FRONTEND_DIR="$ROOT_DIR/frontend"

echo "=== .ait 번들 업로드 ==="
echo ""

# 1. API Key 확인
API_KEY="${1:-$AIT_API_KEY}"
if [ -z "$API_KEY" ]; then
  echo "[ERROR] API Key가 필요합니다."
  echo ""
  echo "사용법:"
  echo "  ./scripts/ait_deploy.sh <API_KEY>"
  echo "  AIT_API_KEY='...' ./scripts/ait_deploy.sh"
  echo ""
  echo "API Key는 토스 개발자센터 콘솔에서 발급받을 수 있습니다."
  exit 1
fi

cd "$FRONTEND_DIR"

# 2. .ait 파일 확인
AIT_FILE=$(find . -maxdepth 1 -name "*.ait" -type f | head -1)
if [ -z "$AIT_FILE" ]; then
  echo "[ERROR] .ait 파일이 없습니다. 먼저 빌드를 실행하세요:"
  echo "  ./scripts/ait_build.sh"
  exit 1
fi

AIT_NAME=$(basename "$AIT_FILE")
echo "[INFO] 업로드 파일: $AIT_NAME"
echo ""

# 3. appName 확인
APP_NAME=$(grep -o 'appName:.*"[^"]*"' granite.config.ts | head -1 | sed 's/.*"\([^"]*\)".*/\1/')
echo "[INFO] appName: $APP_NAME"
echo ""

# 4. 업로드 실행
echo "[1/2] 업로드 중..."
npx ait deploy --api-key "$API_KEY"

echo ""
echo "=== 업로드 완료 ==="
echo ""
echo "다음 단계:"
echo "  1. 토스 개발자센터 콘솔에서 QR 코드 생성"
echo "  2. 토스앱에서 QR 스캔하여 테스트 (최소 1회)"
echo "  3. 테스트 완료 후 '검토 요청' 클릭"
echo ""
echo "테스트 URL:"
echo "  라이브:  https://${APP_NAME}.apps.tossmini.com"
echo "  테스트: https://${APP_NAME}.private-apps.tossmini.com"
echo ""
