# HabitCashback (습관환급)

[![CI](https://github.com/Prometheus-P/habit-cashback/actions/workflows/ci.yml/badge.svg)](https://github.com/Prometheus-P/habit-cashback/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-Proprietary-red.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22-00ADD8.svg)](https://golang.org/)
[![Node Version](https://img.shields.io/badge/Node-20-339933.svg)](https://nodejs.org/)

**목표:** 토스 Apps in Toss(WebView) 환경에서 돌아가는 **'습관 인증 → 정산(환급/차감)'** MVP.

## Tech Stack

| Layer | Technology |
|-------|------------|
| Frontend | Vite + React + TDS (WebView) |
| Backend | Go (stub API, 리뷰/데모용) |
| Staging | Docker Compose + Caddy + GHCR 자동배포 |

## Project Structure

```
habit-cashback/
├── frontend/          # React WebView 앱
├── backend/           # Go API 서버
├── infra/             # 인프라 설정 (Docker, Caddy)
├── scripts/           # 유틸리티 스크립트
├── docs/              # 문서
└── db/                # 데이터베이스 스키마
```

## Getting Started

### Prerequisites

- Go 1.22+
- Node.js 20+
- Docker & Docker Compose
- pnpm (recommended)

### Local Development (Frontend)

```bash
cd frontend
corepack enable
pnpm i
pnpm dev
```

### Local Development (Backend)

```bash
cd backend
go mod tidy
go run ./cmd/api
# http://localhost:8080/health
```

## Deployment

### Staging (Single Server)

서버 준비는 `scripts/40_bootstrap_staging_server_mac.sh` 참고.

```bash
cd infra/staging
cp .env.example .env
# STAGING_DOMAIN, IMAGE_TAG 설정 후
docker compose up -d
```

### Apps in Toss (WebView)

```bash
cd frontend
npx ait init
# granite.config.ts 생성/수정
npx ait deploy --api-key <YOUR_AIT_API_KEY>
```

> `granite.config.ts` 템플릿은 `frontend/granite.config.ts` 참고.

## Contributing

이 프로젝트에 기여하고 싶으시다면:

1. Issue를 먼저 생성해주세요
2. Feature branch를 생성하세요 (`feature/your-feature`)
3. Pull Request를 제출하세요

자세한 내용은 [Contributing Guide](.github/CONTRIBUTING.md)를 참고하세요.

## Security

보안 취약점 발견 시 [Security Policy](.github/SECURITY.md)를 참고하세요.

## License

이 프로젝트는 **Proprietary License** 하에 배포됩니다.
상업적 사용 및 재배포는 사전 허가가 필요합니다.

자세한 내용은 [LICENSE](LICENSE) 파일을 참고하세요.

## Contact

- **Email**: parkdavid31@gmail.com
- **Issues**: [GitHub Issues](https://github.com/Prometheus-P/habit-cashback/issues)
