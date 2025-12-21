# Toss 콘솔 입력값 (복붙용)

> 목적: **Apps in Toss** 콘솔에 입력해야 하는 값을 한 번에 정리.
> 이 문서의 `{...}`는 당신 프로젝트 값으로 치환.

---

## 0) 기본 정보

- 앱 표시명(KR): **습관환급**
- 앱 표시명(EN): **Habit Refund** *(브랜딩 패치 기준)*
- 한 줄 설명: **보증금 걸고 습관 지키면 환급**
- 카테고리: **라이프/금융(리워드)** *(콘솔 옵션에 맞춰 선택)*

---

## 1) 도메인 / URL

### 1-1. Webview URL (앱 진입 주소)
- **Staging:** `https://{STAGING_DOMAIN}/`
- **Prod:** `https://{PROD_DOMAIN}/`

### 1-2. API Base URL (서버)
- **Staging:** `https://{STAGING_DOMAIN}/api` *(현재 Caddy 라우팅 기준)*
- **Prod:** `https://{PROD_DOMAIN}/api`

> 실제 경로는 repo의 `infra/*/Caddyfile` 기준으로 최종 확인.

---

## 2) Toss Login (Apps-in-Toss OAuth)

### 2-1. Redirect URI (인가코드 콜백)
- `https://{DOMAIN}/auth/callback`

> 프론트 라우트: `frontend/src/routes/auth/callback` (pack 기준)

### 2-2. Unlink Callback URL (로그인 끊김 콜백)
- `https://{DOMAIN}/v1/auth/toss/unlink-callback`

서버는 GET/POST 모두 지원합니다. `userKey`, `referrer`를 받습니다.

### 2-3. Basic Auth (선택)
- 콘솔에서 unlink callback에 Basic Auth를 걸었다면:
  - 서버 env에 아래를 넣어야 합니다.
  - `AIT_UNLINK_BASIC_AUTH="username:password"`

---

## 3) 권한 / 정책 링크 (심사에서 거의 필수)

### 3-1. 개인정보처리방침 URL
- `https://{DOMAIN}/privacy`

### 3-2. 이용약관 URL
- `https://{DOMAIN}/terms`

### 3-3. 고객센터 / 문의
- 이메일: `{SUPPORT_EMAIL}`
- (옵션) 카카오/전화: `{SUPPORT_CONTACT}`

---

## 4) 아이콘 / 스크린샷

- 아이콘: `assets/icon.png`
- 스크린샷(권장): `assets/screenshots/*`

---

## 5) 테스트 체크리스트 (Staging)

1. staging 도메인 접속
2. 로그인 → 세션 발급 (`/v1/auth/toss/exchange`)
3. 챌린지 목록 로딩 (`/v1/challenges`)
4. 인증 업로드 → 성공/실패 시나리오
5. 토스 콘솔에서 **연결 끊기** 실행 → 서버 unlink callback 200 확인
6. 앱 재진입 시 401 처리 → 로그인 화면/가이드 노출 확인

---

## 6) CI/배포 체크 (Staging)

- GH Actions: `staging-dev.yml` (dev 이미지 자동 배포)
- 서버 env: `SESSION_SECRET`, `ALLOW_ORIGIN`, `AIT_MTLS_CERT_FILE`, `AIT_MTLS_KEY_FILE`

