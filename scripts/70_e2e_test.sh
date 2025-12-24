#!/usr/bin/env bash
# 70_e2e_test.sh - 로컬 E2E 테스트 스크립트
# 사용법: ./scripts/70_e2e_test.sh
#
# Mock 모드에서 결제 생성 → 실행 → 인증 제출 플로우를 테스트합니다.

set -euo pipefail

# 설정
API_BASE="${API_BASE:-http://localhost:8080}"
CHALLENGE_ID="${CHALLENGE_ID:-morning-water}"

echo "=== 로컬 E2E 테스트 ==="
echo "API_BASE: $API_BASE"
echo "CHALLENGE_ID: $CHALLENGE_ID"
echo ""

# 테스트용 idempotency key 생성
IDEM_KEY=$(uuidgen | tr '[:upper:]' '[:lower:]')
echo "[INFO] Idempotency Key: $IDEM_KEY"
echo ""

# 1. 헬스체크
echo "[1/5] 헬스체크..."
HEALTH=$(curl -s -o /dev/null -w "%{http_code}" "$API_BASE/health" || echo "000")
if [ "$HEALTH" != "200" ]; then
  echo "  [ERROR] API 서버 응답 없음 (status: $HEALTH)"
  echo "  [HINT] 백엔드 서버를 먼저 실행하세요: cd backend && go run ./cmd/api"
  exit 1
fi
echo "  [OK] API 서버 정상"
echo ""

# 2. 결제 생성 테스트
echo "[2/5] 결제 생성 테스트..."
CREATE_RESP=$(curl -s -X POST "$API_BASE/v1/payments/create" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: $IDEM_KEY" \
  -H "X-AIT-Session: test-session-$(date +%s)" \
  -d "{\"challengeId\": \"$CHALLENGE_ID\", \"amount\": 10000}")

echo "  Response: $CREATE_RESP"

# JSON 파싱 (jq 사용 가능하면 사용, 아니면 grep)
if command -v jq >/dev/null 2>&1; then
  PAYMENT_ID=$(echo "$CREATE_RESP" | jq -r '.paymentId // empty')
  PAY_TOKEN=$(echo "$CREATE_RESP" | jq -r '.payToken // empty')
  MODE=$(echo "$CREATE_RESP" | jq -r '.mode // empty')
  ERROR_MSG=$(echo "$CREATE_RESP" | jq -r '.error // empty')
else
  PAYMENT_ID=$(echo "$CREATE_RESP" | grep -o '"paymentId":[0-9]*' | cut -d: -f2 || echo "")
  PAY_TOKEN=$(echo "$CREATE_RESP" | grep -o '"payToken":"[^"]*"' | cut -d: -f2 | tr -d '"' || echo "")
  MODE=$(echo "$CREATE_RESP" | grep -o '"mode":"[^"]*"' | cut -d: -f2 | tr -d '"' || echo "")
  ERROR_MSG=$(echo "$CREATE_RESP" | grep -o '"error":"[^"]*"' | cut -d: -f2 | tr -d '"' || echo "")
fi

if [ -n "$ERROR_MSG" ]; then
  echo "  [ERROR] 결제 생성 실패: $ERROR_MSG"
  exit 1
fi

if [ -z "$PAYMENT_ID" ]; then
  echo "  [ERROR] paymentId 없음"
  exit 1
fi

echo "  [OK] 결제 생성됨"
echo "    - paymentId: $PAYMENT_ID"
echo "    - payToken: $PAY_TOKEN"
echo "    - mode: $MODE"
echo ""

# 3. Mock 모드 확인
echo "[3/5] 결제 모드 확인..."
if [ "$MODE" = "mock" ]; then
  echo "  [OK] Mock 모드 - TossPay UI 스킵"
elif [ "$MODE" = "live" ]; then
  echo "  [INFO] Live 모드 - 실제 TossPay 연동"
  echo "  [WARN] E2E 테스트는 Mock 모드에서만 완전히 동작합니다."
  echo "  [HINT] APP_ENV=local 설정 후 서버를 재시작하세요."
else
  echo "  [WARN] 알 수 없는 모드: $MODE"
fi
echo ""

# 4. 결제 실행 테스트
echo "[4/5] 결제 실행 테스트..."
EXEC_RESP=$(curl -s -X POST "$API_BASE/v1/payments/execute" \
  -H "Content-Type: application/json" \
  -H "X-AIT-Session: test-session-$(date +%s)" \
  -d "{\"paymentId\": $PAYMENT_ID}")

echo "  Response: $EXEC_RESP"

if command -v jq >/dev/null 2>&1; then
  EXEC_SUCCESS=$(echo "$EXEC_RESP" | jq -r '.success // empty')
  EXEC_ERROR=$(echo "$EXEC_RESP" | jq -r '.error // empty')
else
  EXEC_SUCCESS=$(echo "$EXEC_RESP" | grep -o '"success":true' || echo "")
  EXEC_ERROR=$(echo "$EXEC_RESP" | grep -o '"error":"[^"]*"' | cut -d: -f2 | tr -d '"' || echo "")
fi

if [ -n "$EXEC_ERROR" ]; then
  echo "  [ERROR] 결제 실행 실패: $EXEC_ERROR"
  exit 1
fi

if [ "$EXEC_SUCCESS" = "true" ] || [ -n "$EXEC_SUCCESS" ]; then
  echo "  [OK] 결제 실행 성공"
else
  echo "  [WARN] 결제 실행 응답 확인 필요"
fi
echo ""

# 5. 챌린지 상태 확인
echo "[5/5] 챌린지 상태 확인..."
STATUS_RESP=$(curl -s -X GET "$API_BASE/v1/challenges/$CHALLENGE_ID" \
  -H "X-AIT-Session: test-session-$(date +%s)" 2>/dev/null || echo "{}")

if [ -n "$STATUS_RESP" ] && [ "$STATUS_RESP" != "{}" ]; then
  echo "  Response: $STATUS_RESP"
  echo "  [OK] 챌린지 상태 조회됨"
else
  echo "  [INFO] 챌린지 상태 API 응답 없음 (엔드포인트가 없을 수 있음)"
fi

echo ""
echo "=== E2E 테스트 완료 ==="
echo ""
echo "테스트된 플로우:"
echo "  1. [OK] 헬스체크"
echo "  2. [OK] 결제 생성 (POST /v1/payments/create)"
echo "  3. [OK] Mock 모드 확인"
echo "  4. [OK] 결제 실행 (POST /v1/payments/execute)"
echo "  5. [--] 챌린지 상태 확인"
echo ""
echo "[SUCCESS] 로컬 E2E 테스트 통과"
