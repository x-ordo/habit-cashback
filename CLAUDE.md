# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

HabitCashback (습관환급) is an MVP habit-tracking app that runs in the Toss Apps in Toss (WebView) environment. Users deposit money as collateral for habit challenges, complete daily photo/step proofs, and receive cashback if successful.

## Build & Development Commands

### Frontend (`cd frontend`)
```bash
corepack enable && pnpm install   # Install dependencies
pnpm dev                          # Vite dev server (port 5173)
pnpm build                        # Production build
pnpm lint                         # ESLint check
```

### Backend (`cd backend`)
```bash
go mod tidy                       # Sync dependencies
go run ./cmd/api                  # API server (port 8080)
go run ./cmd/worker               # Background worker
go build -v ./cmd/api             # Build API binary
go build -v ./cmd/worker          # Build worker binary
go test -v ./...                  # Run all tests
go test -v ./internal/...         # Unit tests only
```

### Staging Deployment
```bash
cd infra/staging
cp .env.example .env              # Configure: STAGING_DOMAIN, IMAGE_TAG, DB creds
docker compose up -d
curl http://localhost/health      # Health check
```

### Apps in Toss Deployment
```bash
cd frontend
npx ait deploy --api-key <KEY>    # Deploy to Toss CDN
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Toss WebView (Frontend)                  │
│         React 18 + Vite + Emotion + TDS Components          │
└────────────────────────────┬────────────────────────────────┘
                             │ REST API (Bearer token)
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                     Go HTTP Server                          │
│    ├── /v1/auth/*     Toss mTLS OAuth + Stub auth          │
│    ├── /v1/me         User info                            │
│    ├── /v1/challenges Challenge list                        │
│    ├── /v1/payments/* TossPay integration                   │
│    ├── /v1/proofs/*   Photo/step verification (EXIF)       │
│    └── /v1/settlements Cashback status                      │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                     PostgreSQL 16+                          │
│   app_user, challenge, payment, participation, proof,       │
│   settlement                                                │
└─────────────────────────────────────────────────────────────┘
```

### Key Directories
- `frontend/src/pages/` - Page components (Login, Home, Proof, History)
- `frontend/src/lib/` - API client, storage, env config
- `backend/cmd/api/` - HTTP server entry point with all route handlers
- `backend/internal/store/` - PostgreSQL data access (pgx)
- `backend/internal/toss/` - Toss mTLS client for Apps in Toss API
- `backend/internal/payment/` - Payment service (Mock + TossPay)
- `infra/staging/` - Docker Compose stack (Caddy + API + Worker + DB)

### Key Patterns

**Authentication**: HMAC-signed session tokens (`sv1.{payload}.{signature}`), stored in localStorage. Backend validates via auth middleware (`backend/cmd/api/main.go:747`).

**Idempotency**: Payment and proof endpoints require `Idempotency-Key` header (UUID) for deduplication.

**Graceful Fallbacks**:
- No DATABASE_URL → in-memory store with hardcoded challenges
- No mTLS certs → stub auth for development
- No TossPay config → mock payment service

**Rate Limiting**: 120 requests/minute per IP (fixed-window).

## Git Commit Rules

**CRITICAL - These rules must be followed:**
- NO "Generated with [Claude Code]" in commit messages
- NO "Co-Authored-By: Claude" in commit messages

## Human Verification Required

The following areas require human review before deployment:
- Security-related code (auth, mTLS, session tokens)
- Payment/financial transaction logic
- Personal data processing
- External API integration (Toss APIs)
