#!/usr/bin/env bash
# 60_validate_mtls.sh - mTLS 인증서 설정 검증
# 사용법: ./scripts/60_validate_mtls.sh
#
# 환경변수:
#   AIT_MTLS_CERT_FILE - mTLS 클라이언트 인증서 경로
#   AIT_MTLS_KEY_FILE  - mTLS 개인키 경로
#   TOSSPAY_API_KEY    - TossPay API 키

set -euo pipefail

echo "=== mTLS 인증서 설정 검증 ==="
echo ""

ERRORS=0

# 1. 환경변수 확인
echo "[1/5] 환경변수 확인..."

if [ -z "${AIT_MTLS_CERT_FILE:-}" ]; then
  echo "  [WARN] AIT_MTLS_CERT_FILE 미설정 (Mock 모드에서는 불필요)"
else
  echo "  [OK] AIT_MTLS_CERT_FILE: $AIT_MTLS_CERT_FILE"
fi

if [ -z "${AIT_MTLS_KEY_FILE:-}" ]; then
  echo "  [WARN] AIT_MTLS_KEY_FILE 미설정 (Mock 모드에서는 불필요)"
else
  echo "  [OK] AIT_MTLS_KEY_FILE: $AIT_MTLS_KEY_FILE"
fi

if [ -z "${TOSSPAY_API_KEY:-}" ]; then
  echo "  [WARN] TOSSPAY_API_KEY 미설정 (Mock 모드에서는 불필요)"
else
  echo "  [OK] TOSSPAY_API_KEY: (설정됨, 값 숨김)"
fi

echo ""

# 2. 인증서 파일 존재 확인
echo "[2/5] 인증서 파일 존재 확인..."

if [ -n "${AIT_MTLS_CERT_FILE:-}" ]; then
  if [ -f "$AIT_MTLS_CERT_FILE" ]; then
    echo "  [OK] 인증서 파일 존재: $AIT_MTLS_CERT_FILE"
  else
    echo "  [ERROR] 인증서 파일 없음: $AIT_MTLS_CERT_FILE"
    ERRORS=$((ERRORS + 1))
  fi
fi

if [ -n "${AIT_MTLS_KEY_FILE:-}" ]; then
  if [ -f "$AIT_MTLS_KEY_FILE" ]; then
    echo "  [OK] 개인키 파일 존재: $AIT_MTLS_KEY_FILE"
  else
    echo "  [ERROR] 개인키 파일 없음: $AIT_MTLS_KEY_FILE"
    ERRORS=$((ERRORS + 1))
  fi
fi

echo ""

# 3. 인증서 유효성 검사 (openssl 사용)
echo "[3/5] 인증서 유효성 검사..."

if [ -n "${AIT_MTLS_CERT_FILE:-}" ] && [ -f "${AIT_MTLS_CERT_FILE:-}" ]; then
  if command -v openssl >/dev/null 2>&1; then
    # 인증서 만료일 확인
    EXPIRY=$(openssl x509 -enddate -noout -in "$AIT_MTLS_CERT_FILE" 2>/dev/null | cut -d= -f2)
    if [ -n "$EXPIRY" ]; then
      echo "  [OK] 인증서 만료일: $EXPIRY"

      # 만료 7일 전 경고
      EXPIRY_EPOCH=$(date -j -f "%b %d %T %Y %Z" "$EXPIRY" +%s 2>/dev/null || date -d "$EXPIRY" +%s 2>/dev/null || echo "0")
      NOW_EPOCH=$(date +%s)
      DAYS_LEFT=$(( (EXPIRY_EPOCH - NOW_EPOCH) / 86400 ))

      if [ "$DAYS_LEFT" -lt 0 ]; then
        echo "  [ERROR] 인증서 만료됨!"
        ERRORS=$((ERRORS + 1))
      elif [ "$DAYS_LEFT" -lt 7 ]; then
        echo "  [WARN] 인증서 만료 임박: ${DAYS_LEFT}일 남음"
      else
        echo "  [OK] 인증서 유효: ${DAYS_LEFT}일 남음"
      fi
    fi

    # 인증서 Subject 확인
    SUBJECT=$(openssl x509 -subject -noout -in "$AIT_MTLS_CERT_FILE" 2>/dev/null)
    if [ -n "$SUBJECT" ]; then
      echo "  [INFO] $SUBJECT"
    fi
  else
    echo "  [SKIP] openssl 미설치"
  fi
else
  echo "  [SKIP] 인증서 파일 미설정"
fi

echo ""

# 4. 인증서-키 매칭 확인
echo "[4/5] 인증서-키 매칭 확인..."

if [ -n "${AIT_MTLS_CERT_FILE:-}" ] && [ -f "${AIT_MTLS_CERT_FILE:-}" ] && \
   [ -n "${AIT_MTLS_KEY_FILE:-}" ] && [ -f "${AIT_MTLS_KEY_FILE:-}" ]; then
  if command -v openssl >/dev/null 2>&1; then
    CERT_MODULUS=$(openssl x509 -modulus -noout -in "$AIT_MTLS_CERT_FILE" 2>/dev/null | md5)
    KEY_MODULUS=$(openssl rsa -modulus -noout -in "$AIT_MTLS_KEY_FILE" 2>/dev/null | md5)

    if [ "$CERT_MODULUS" = "$KEY_MODULUS" ]; then
      echo "  [OK] 인증서와 개인키 매칭 확인됨"
    else
      echo "  [ERROR] 인증서와 개인키가 매칭되지 않음!"
      ERRORS=$((ERRORS + 1))
    fi
  else
    echo "  [SKIP] openssl 미설치"
  fi
else
  echo "  [SKIP] 인증서 또는 키 파일 미설정"
fi

echo ""

# 5. APP_ENV 확인
echo "[5/5] APP_ENV 확인..."

APP_ENV="${APP_ENV:-}"
if [ -z "$APP_ENV" ]; then
  echo "  [INFO] APP_ENV 미설정 - Mock 모드로 동작"
elif [ "$APP_ENV" = "local" ] || [ "$APP_ENV" = "test" ]; then
  echo "  [INFO] APP_ENV=$APP_ENV - Mock 모드로 동작"
elif [ "$APP_ENV" = "staging" ] || [ "$APP_ENV" = "prod" ]; then
  echo "  [INFO] APP_ENV=$APP_ENV - 실제 TossPay 연동 모드"

  # Live 모드에서는 인증서 필수
  if [ -z "${AIT_MTLS_CERT_FILE:-}" ] || [ -z "${AIT_MTLS_KEY_FILE:-}" ] || [ -z "${TOSSPAY_API_KEY:-}" ]; then
    echo "  [ERROR] Live 모드에서는 mTLS 인증서와 API 키가 필수입니다!"
    ERRORS=$((ERRORS + 1))
  fi
else
  echo "  [WARN] 알 수 없는 APP_ENV: $APP_ENV"
fi

echo ""
echo "=== 검증 완료 ==="

if [ $ERRORS -gt 0 ]; then
  echo "[FAILED] $ERRORS개의 오류가 발견되었습니다."
  exit 1
else
  echo "[PASSED] 모든 검사 통과"
  exit 0
fi
