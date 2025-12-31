# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.1] - 2025-12-31

### Changed
- GitHub Organization 변경: `Prometheus-P` → `x-ordo`
- CLAUDE.md 개선: 프로젝트별 빌드 명령어, 아키텍처 다이어그램, 핵심 패턴 문서화

---

## [1.0.0] - 2025-12-22

### Added

#### Backend
- **DB 연동**: PostgreSQL 연결 및 CRUD 모듈 (`backend/internal/store`)
  - User, Challenge, Payment, Proof, Settlement 테이블 지원
  - 트랜잭션 기반 결제 실행
  - 멱등성 키 및 세션 폐기 관리
- **정산 API**: `GET /v1/settlements` 일괄 조회
- **인증 검증 강화**:
  - EXIF 메타데이터 촬영 시간 검증
  - SHA256 해시 기반 중복 사진 방지
- **배치 워커** (`backend/cmd/worker`):
  - 챌린지 종료 처리 (매일 00:05)
  - 정산 상태 갱신 (매일 00:10)
  - 멱등성 키 정리 (매시간)
  - 폐기 세션 정리 (매일 03:00)

#### Frontend
- 9개 화면 구현 완료:
  - LoginPage, HomePage, ChallengePage, ProofPage
  - HistoryPage, HelpPage, TermsPage, PrivacyPage, SupportPage
- Toss Design System 적용
- Apps in Toss SDK 연동

#### Infrastructure
- Docker 멀티스테이지 빌드 (API/Worker 분리)
- docker-compose: PostgreSQL, API, Worker, Web, Caddy
- GitHub Actions CI/CD:
  - Go 빌드 및 테스트
  - Frontend 빌드
  - Docker 이미지 빌드/푸시 (GHCR)

#### Documentation
- `docs/USER_FLOW.md`: 유저플로우 정의서
- `docs/API_DEFINITION.md`: API 명세서
- `docs/DB_SCHEMA.md`: DB 스키마 정의
- `docs/SCREEN_DEFINITION.md`: 화면 정의서

### Security
- HMAC-signed 세션 토큰
- mTLS 인증 (Apps in Toss)
- Rate limiting (120 req/min per IP)
- CORS origin 검증
- 세션 폐기 (unlink callback)

---

## [0.1.0] - 2025-12-21

### Added
- Initial commit: Habit Cashback MVP Kit
- 기본 프로젝트 구조
