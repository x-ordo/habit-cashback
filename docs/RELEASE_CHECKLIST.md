# 앱인토스 출시 체크리스트 (Release Checklist)

> **습관환급 (Habit Cashback) 토스 심사/출시 가이드**
> 공식 문서: https://developers-apps-in-toss.toss.im/

---

## 출시 플로우

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│   1. appName 확정  →  2. CORS 설정  →  3. .ait 빌드           │
│         │                  │                 │                  │
│         ▼                  ▼                 ▼                  │
│   4. 업로드 (콘솔/CLI)  →  5. QR 테스트 (1회 이상)            │
│         │                                    │                  │
│         ▼                                    ▼                  │
│   6. 검토 요청  →  7. 승인  →  8. 출시                        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 0. 사전 준비

### 콘솔 appName 확정

| 항목 | 값 | 파일 |
|------|-----|------|
| 콘솔 App ID | `habitcashback` | 토스 개발자센터 |
| granite.config.ts | `habitcashback` | `frontend/granite.config.ts` |

**주의**: 두 값은 **반드시 동일**해야 함

```typescript
// frontend/granite.config.ts
export default defineConfig({
  appName: "habitcashback",  // ← 콘솔과 동일
  // ...
});
```

### 도메인 확인

| 환경 | URL |
|------|-----|
| 라이브 | `https://habitcashback.apps.tossmini.com` |
| 테스트 | `https://habitcashback.private-apps.tossmini.com` |

---

## 1. 백엔드 설정

### CORS 허용 Origin 설정

**환경변수** (`ALLOW_ORIGIN`):
```bash
# 쉼표로 구분하여 2개 모두 등록
ALLOW_ORIGIN="https://habitcashback.apps.tossmini.com,https://habitcashback.private-apps.tossmini.com"
```

**Docker Compose 예시**:
```yaml
environment:
  - ALLOW_ORIGIN=https://habitcashback.apps.tossmini.com,https://habitcashback.private-apps.tossmini.com
```

### HTTPS 필수

- 라이브 환경은 **HTTPS만 허용**
- Caddy, Nginx, AWS ALB 등으로 TLS 종단 처리

### iOS 쿠키 이슈 회피

- iOS 13.4+ 서드파티 쿠키 차단
- **토큰 기반 인증 사용** (현재 구현 완료)

### mTLS 설정

```bash
# 환경변수
AIT_MTLS_CERT_FILE=/path/to/cert.pem
AIT_MTLS_KEY_FILE=/path/to/key.pem
AIT_TOSS_BASE_URL=https://apps-in-toss-api.toss.im
```

---

## 2. .ait 번들 생성

### 빌드 스크립트 실행

```bash
# 프로젝트 루트에서
./scripts/ait_build.sh
```

### 수동 빌드

```bash
cd frontend
pnpm install
pnpm build
npx ait build
```

### 빌드 결과 확인

```
frontend/
├── dist/           # Vite 빌드 결과
└── *.ait           # 앱인토스 번들
```

---

## 3. 업로드

### 옵션 A: 콘솔 업로드

1. 토스 개발자센터 접속
2. 앱 관리 > 앱 출시 메뉴
3. `.ait` 파일 업로드
4. QR 코드 생성

### 옵션 B: CLI 업로드 (권장)

```bash
# API Key 직접 전달
./scripts/ait_deploy.sh "<YOUR_AIT_API_KEY>"

# 또는 환경변수 사용
AIT_API_KEY="<YOUR_AIT_API_KEY>" ./scripts/ait_deploy.sh
```

**API Key 발급**: 토스 개발자센터 콘솔에서 발급

---

## 4. 토스앱 테스트

### 테스트 조건 (필수)

| 조건 | 설명 |
|------|------|
| 로그인 상태 | 토스앱 로그인 필수 |
| 워크스페이스 멤버 | 개발자센터 워크스페이스에 등록된 계정 |
| 연령 제한 | **만 19세 이상**만 테스트 가능 |

### 테스트 방법

1. 콘솔에서 QR 코드 생성
2. 토스앱 > QR 스캔
3. **최소 1회 이상** 앱 실행 완료

### 테스트 항목

- [ ] 로그인 (토스 로그인 SDK)
- [ ] 챌린지 목록 조회
- [ ] 챌린지 상세 진입
- [ ] 결제 플로우 (테스트 환경)
- [ ] 인증 제출 (사진/걸음수)
- [ ] 정산 내역 조회
- [ ] 이용약관/개인정보처리방침 접근

---

## 5. 검토 요청

### 요청 전 체크리스트

```bash
# iframe 사용 확인 (유튜브 외 금지)
./scripts/10_check_iframe.sh

# 프론트엔드 빌드 확인
./scripts/20_build_frontend.sh

# 백엔드 빌드 확인
./scripts/30_build_backend.sh

# 스테이징 스모크 테스트
./scripts/40_smoke_staging.sh https://YOUR_STAGING_API
```

### 검토 요청 절차

1. 테스트 완료 확인 (콘솔에 표시)
2. **검토 요청하기** 버튼 클릭
3. 심사 대기 (영업일 기준 최대 3일)

### 반려 시

1. 반려 사유 확인
2. 코드 수정
3. 새 `.ait` 빌드 및 업로드
4. 재검토 요청

---

## 6. 출시

### 승인 후

1. 콘솔에서 **출시하기** 클릭
2. 즉시 전체 사용자에게 반영

### 출시 전 최종 확인

- [ ] 프로덕션 환경변수 설정 완료
- [ ] HTTPS 종단 확인
- [ ] CORS 설정 확인
- [ ] mTLS 인증서 유효성 확인
- [ ] 고객센터 연락처 정확성

---

## 환경변수 체크리스트

### 백엔드 (필수)

| 변수 | 값 예시 | 설명 |
|------|---------|------|
| `PORT` | `8080` | 서버 포트 |
| `APP_ENV` | `prod` | 환경 (local/staging/prod) |
| `SESSION_SECRET` | `(32+ 랜덤 문자열)` | 세션 서명 시크릿 |
| `ALLOW_ORIGIN` | `https://xxx.apps.tossmini.com,https://xxx.private-apps.tossmini.com` | CORS 허용 origin |
| `AIT_MTLS_CERT_FILE` | `/certs/cert.pem` | mTLS 인증서 |
| `AIT_MTLS_KEY_FILE` | `/certs/key.pem` | mTLS 키 |

### 프론트엔드 (빌드 시)

| 변수 | 값 예시 | 설명 |
|------|---------|------|
| `VITE_API_BASE_URL` | `https://api.example.com` | 백엔드 API URL |
| `VITE_APP_DISPLAY_NAME` | `습관환급` | 앱 표시명 |
| `VITE_SUPPORT_EMAIL` | `support@example.com` | 고객센터 이메일 |

### 앱인토스

| 변수 | 값 예시 | 설명 |
|------|---------|------|
| `AIT_APP_NAME` | `habitcashback` | 콘솔 App ID |
| `AIT_API_KEY` | `(콘솔에서 발급)` | 배포 API 키 |

---

## 반려 포인트 (회피 필수)

| 항목 | 설명 | 대응 |
|------|------|------|
| iframe 사용 | 유튜브 외 금지 | `scripts/10_check_iframe.sh` 실행 |
| HTTP 사용 | 라이브에서 HTTPS만 허용 | TLS 종단 설정 |
| mTLS 미설정 | 토스 API 호출 실패 | 인증서 설정 |
| 테스트 미완료 | 검토 요청 불가 | QR 테스트 1회 이상 |
| 사행성 표현 | 심사 반려 | 용어 가이드 준수 |

### 사행성 회피 용어

| 사용 | 회피 |
|------|------|
| 참가비 | ~~배팅금~~ |
| 리워드/환급 | ~~수익/투자금~~ |
| 성공 조건 | ~~배당률~~ |
| 미지급 | ~~손실~~ |

---

## 문제 해결

### QR 테스트 안 됨

1. 워크스페이스 멤버 여부 확인
2. 토스앱 로그인 상태 확인
3. 만 19세 이상 계정인지 확인

### 토스 API 호출 실패

1. mTLS 인증서 경로 확인
2. 인증서 유효기간 확인
3. `AIT_TOSS_BASE_URL` 확인

### CORS 에러

1. `ALLOW_ORIGIN`에 두 도메인 모두 포함 확인
2. 프로토콜(https) 포함 확인
3. 후행 슬래시 없이 설정

---

## 스크립트 목록

| 스크립트 | 설명 |
|----------|------|
| `scripts/ait_build.sh` | .ait 번들 생성 |
| `scripts/ait_deploy.sh` | .ait 업로드 |
| `scripts/10_check_iframe.sh` | iframe 사용 검사 |
| `scripts/20_build_frontend.sh` | 프론트엔드 빌드 |
| `scripts/30_build_backend.sh` | 백엔드 빌드 |
| `scripts/40_smoke_staging.sh` | 스테이징 스모크 테스트 |
| `scripts/50_preflight_review.sh` | 심사 전 점검 |

---

## 타임라인

| 단계 | 예상 소요 |
|------|----------|
| 빌드 및 업로드 | 10분 |
| QR 테스트 | 30분 |
| 검토 요청 ~ 승인 | 영업일 1~3일 |
| 출시 | 즉시 |
