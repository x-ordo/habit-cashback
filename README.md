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
├── frontend/                # React WebView 앱
│   ├── src/
│   │   ├── app/            # App entry, Auth wrapper
│   │   ├── components/     # 공통 컴포넌트 (TopBar, BottomCTA 등)
│   │   ├── pages/          # 페이지 컴포넌트 (Login, Home, Proof 등)
│   │   ├── lib/            # API, 스토리지, 유틸리티
│   │   └── styles/         # 스타일시트
│   └── granite.config.ts   # Apps in Toss 배포 설정
├── backend/                 # Go API 서버
│   ├── cmd/
│   │   ├── api/            # HTTP API 서버
│   │   └── worker/         # 백그라운드 작업 워커
│   └── internal/
│       ├── proof/          # 인증 검증 로직 (EXIF 등)
│       ├── store/          # PostgreSQL 데이터 접근
│       └── toss/           # Apps in Toss mTLS 클라이언트
├── infra/                   # 인프라 설정
│   ├── staging/            # 스테이징 Docker Compose
│   └── prod/               # 프로덕션 Docker Compose
├── scripts/                 # 유틸리티 스크립트
├── docs/                    # 프로젝트 문서
├── db/                      # 데이터베이스 마이그레이션
│   └── migrations/         # SQL 마이그레이션 파일
└── .github/                 # GitHub 설정
    ├── workflows/          # CI/CD 워크플로우
    └── ISSUE_TEMPLATE/     # 이슈 템플릿
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
