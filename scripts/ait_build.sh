#!/bin/bash
# ait_build.sh - .ait 번들 생성 스크립트
# 사용법: ./scripts/ait_build.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
FRONTEND_DIR="$ROOT_DIR/frontend"

echo "=== .ait 번들 생성 ==="
echo ""

# 1. frontend 디렉토리 확인
if [ ! -d "$FRONTEND_DIR" ]; then
  echo "[ERROR] frontend 디렉토리가 없습니다: $FRONTEND_DIR"
  exit 1
fi

cd "$FRONTEND_DIR"

# 2. granite.config.ts 확인
if [ ! -f "granite.config.ts" ]; then
  echo "[ERROR] granite.config.ts 파일이 없습니다."
  exit 1
fi

# 3. appName 확인
APP_NAME=$(grep -o 'appName:.*"[^"]*"' granite.config.ts | head -1 | sed 's/.*"\([^"]*\)".*/\1/')
echo "[INFO] appName: $APP_NAME"
echo ""

# 4. 의존성 설치
echo "[1/3] 의존성 설치..."
if command -v pnpm &> /dev/null; then
  pnpm install
elif command -v npm &> /dev/null; then
  npm install
else
  echo "[ERROR] pnpm 또는 npm이 설치되어 있지 않습니다."
  exit 1
fi

# 5. 프론트엔드 빌드
echo ""
echo "[2/3] 프론트엔드 빌드..."
if command -v pnpm &> /dev/null; then
  pnpm build
else
  npm run build
fi

# 6. .ait 번들 생성
echo ""
echo "[3/3] .ait 번들 생성..."
npx ait build

# 7. 결과 확인
AIT_FILE=$(find . -maxdepth 1 -name "*.ait" -type f | head -1)
if [ -z "$AIT_FILE" ]; then
  echo "[ERROR] .ait 파일 생성 실패"
  exit 1
fi

AIT_PATH="$(cd "$(dirname "$AIT_FILE")" && pwd)/$(basename "$AIT_FILE")"
AIT_SIZE=$(du -h "$AIT_FILE" | cut -f1)

echo ""
echo "=== 빌드 완료 ==="
echo "파일: $AIT_PATH"
echo "크기: $AIT_SIZE"
echo ""
echo "다음 단계:"
echo "  1. 콘솔 업로드: 토스 개발자센터에서 직접 업로드"
echo "  2. CLI 업로드:  ./scripts/ait_deploy.sh <API_KEY>"
echo ""
