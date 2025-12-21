# v12 Review Hardening

## 핵심
- Backend: `SESSION_SECRET` 기반 **서명 세션 토큰(sv1)** 발급/검증
- Frontend: 기존 `accessToken` 저장 로직 유지 (백엔드가 `sessionToken` + `accessToken` 둘 다 반환)

## 심사/운영 체크
- staging/prod에서 `ALLOW_ORIGIN`은 반드시 실제 도메인으로 고정
- staging/prod에서 `SESSION_SECRET` 필수
- Apps-in-Toss mTLS는 `infra/staging/secrets/`에 cert/key 마운트 후 `AIT_*` env로 지정
