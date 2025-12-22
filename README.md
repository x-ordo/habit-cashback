# HabitRefund (습관환급)
**목표:** 토스 Apps in Toss(WebView) 환경에서 돌아가는 **'습관 인증 → 정산(환급/차감)'** MVP.

- Frontend: Vite + React + TDS (WebView)
- Backend: Go (stub API, 리뷰/데모용)
- Staging: Docker Compose + Caddy + GHCR 자동배포

## 로컬 실행 (웹)

```bash
cd frontend
corepack enable
pnpm i
pnpm dev
```

## 로컬 실행 (백엔드)

```bash
cd backend
go mod tidy
go run ./cmd/api
# http://localhost:8080/health
```

## 스테이징 배포 (서버 1대)

서버 준비는 `scripts/40_bootstrap_staging_server_mac.sh` 참고.

```bash
cd infra/staging
cp .env.example .env
# STAGING_DOMAIN, IMAGE_TAG 설정 후
docker compose up -d
```

## Apps in Toss(WebView) 배포

frontend 디렉토리에서:

```bash
cd frontend
npx ait init
# granite.config.ts 생성/수정
npx ait deploy --api-key <YOUR_AIT_API_KEY>
```

> `granite.config.ts` 템플릿은 `frontend/granite.config.ts` 참고.
