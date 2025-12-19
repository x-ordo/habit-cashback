# 습관환급 (Habit Refund) — Apps in Toss MVP Kit v0.7

이 레포 키트는 **Go 백엔드 + Toss Apps-in-Toss Partner API(mTLS)** 기준으로
`로그인 → 참가(결제) → 인증(업로드) → 환급/리워드(포인트 지급) → 알림` 흐름을 **“바로 붙일 수 있는 형태”**로 제공합니다.

## What’s included
- `docs/spec/` : 흐름/정책/정산/멱등성/토스 API 맵
- `db/migrations/` : Postgres 최소 스키마
- `backend/` : Go API 서버 + Toss mTLS client + (초기) 워커
- `frontend_example/` : `appLogin()` + `checkoutPayment()` 최소 샘플

## Quick start (Dev)
1) Postgres 준비 + 마이그레이션
```bash
psql "$DATABASE_URL" -f db/migrations/001_init.sql
```

2) Toss 콘솔에서 mTLS 인증서 발급 후 파일을 준비
- `TOSS_MTLS_CERT_FILE=/path/to/client.crt`
- `TOSS_MTLS_KEY_FILE=/path/to/client.key`

3) 백엔드 실행
```bash
cd backend
cp .env.example .env
go run ./cmd/api
```

> 참고: 외부 모듈(chi/jwt/pgx/uuid)을 사용합니다. 일반적인 Go 환경에서는 `go mod tidy` / `go mod download`가 필요합니다.

## Endpoints (MVP)
- `POST /v1/auth/exchange` : authorizationCode/referrer → 내부 JWT 발급
- `POST /v1/payments/create` : make-payment → payToken 반환
- `POST /v1/payments/execute` : execute-payment
- `POST /v1/payouts/issue` : promotion get-key → execute
- `POST /v1/payouts/result` : execution-result
- `POST /v1/messages/send` : messenger send-message
