# API 정의서 (API Definition)

> **앱인토스(Apps in Toss) 가이드라인 반영**
> 백엔드: Go 1.21+ / 프론트엔드: TypeScript + Fetch API

---

## 개요

| 항목 | 내용 |
|------|------|
| Base URL | `VITE_API_BASE_URL` 또는 same-origin |
| 인증 방식 | Bearer Token (HMAC-signed session) |
| Content-Type | `application/json; charset=utf-8` |
| 에러 형식 | `{ "error": "message" }` |

---

## 공통 헤더

### 요청 헤더

| 헤더 | 필수 | 설명 |
|------|------|------|
| `Authorization` | 인증 필요 API | `Bearer {sessionToken}` |
| `Content-Type` | POST | `application/json` |
| `Idempotency-Key` | 멱등성 필요 API | UUID (중복 요청 방지) |
| `X-Request-Id` | 선택 | 요청 추적 ID (미제공 시 서버 생성) |

### 응답 헤더

| 헤더 | 설명 |
|------|------|
| `X-Request-Id` | 요청 추적 ID |
| `Access-Control-Allow-Origin` | CORS origin |
| `X-Content-Type-Options` | `nosniff` |
| `X-Frame-Options` | `DENY` |

---

## 인증 체계

### 세션 토큰 형식

```
sv1.{base64url_payload}.{hmac_signature}
```

**Payload (Claims)**:
```json
{
  "sub": "toss:12345678",  // 사용자 식별자
  "iat": 1703145600,       // 발급 시간 (Unix)
  "exp": 1703232000        // 만료 시간 (Unix, 24시간)
}
```

### 인증 흐름

```
┌─────────────┐     appLogin()     ┌─────────────┐
│  Frontend   │ ─────────────────→ │  Toss SDK   │
│  (WebView)  │ ←───────────────── │             │
└──────┬──────┘  authorizationCode └─────────────┘
       │
       │ POST /v1/auth/toss/exchange
       ▼
┌─────────────┐     mTLS      ┌─────────────────────┐
│   Backend   │ ────────────→ │ apps-in-toss-api    │
│   (Go)      │ ←──────────── │ .toss.im            │
└──────┬──────┘  accessToken  └─────────────────────┘
       │
       │ sessionToken (자체 서명)
       ▼
┌─────────────┐
│  Frontend   │  → localStorage 저장
└─────────────┘
```

---

## API 엔드포인트

### 1. 시스템

#### GET /health

헬스 체크

**인증**: 불필요

**응답**:
```json
{
  "ok": true,
  "env": "local",
  "ts": "2025-12-21T13:00:00Z"
}
```

---

#### GET /meta

서비스 메타 정보

**인증**: 불필요

**응답**:
```json
{
  "name": "habitcashback-api",
  "version": "1.0.0",
  "commit": "abc1234",
  "links": {
    "terms": "/terms",
    "privacy": "/privacy",
    "support": "/support"
  },
  "toss": {
    "unlinkCallback": "/v1/auth/toss/unlink-callback"
  }
}
```

---

### 2. 인증 (Auth)

#### POST /v1/auth/exchange

스텁 로그인 (개발/테스트용)

**인증**: 불필요

**요청 Body**: (빈 객체 허용)
```json
{}
```

**응답** (200 OK):
```json
{
  "sessionToken": "sv1.eyJzdWIiOiJzdHViLXVzZXIi....",
  "accessToken": "sv1.eyJzdWIiOiJzdHViLXVzZXIi....",
  "mode": "stub",
  "expiresIn": 86400
}
```

| 필드 | 타입 | 설명 |
|------|------|------|
| sessionToken | string | 세션 토큰 (우선 사용) |
| accessToken | string | 호환성 별칭 |
| mode | string | `"stub"` - 스텁 모드 |
| expiresIn | number | 만료 시간 (초) |

---

#### POST /v1/auth/toss/exchange

토스 로그인 토큰 교환 (mTLS 필수)

**인증**: 불필요

**요청 Body**:
```json
{
  "authorizationCode": "toss_auth_code_from_sdk",
  "referrer": "DEFAULT"
}
```

| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| authorizationCode | string | O | 토스 SDK에서 받은 인가 코드 |
| referrer | string | X | 리퍼러 (기본: "DEFAULT") |

**응답** (200 OK):
```json
{
  "sessionToken": "sv1.eyJzdWIiOiJ0b3NzOjEyMzQ1Njc4Ii....",
  "accessToken": "sv1.eyJzdWIiOiJ0b3NzOjEyMzQ1Njc4Ii....",
  "mode": "toss",
  "expiresIn": 86400
}
```

| 필드 | 타입 | 설명 |
|------|------|------|
| sessionToken | string | 세션 토큰 |
| accessToken | string | 호환성 별칭 |
| mode | string | `"toss"` - 토스 연동 모드 |
| expiresIn | number | 만료 시간 (초) |

**에러 응답**:

| 상태 | 에러 | 설명 |
|------|------|------|
| 400 | `invalid json body` | JSON 파싱 실패 |
| 400 | `authorizationCode is required` | 인가 코드 누락 |
| 502 | `toss mTLS not configured on server` | mTLS 미설정 |
| 502 | `toss exchange failed` | 토스 API 교환 실패 |

---

#### POST /v1/auth/toss/unlink-callback

토스 연결 해제 콜백 (토스 콘솔에서 호출)

**인증**: Basic Auth (선택, `AIT_UNLINK_BASIC_AUTH` 설정 시)

**요청 (POST)**:
```json
{
  "userKey": 12345678,
  "referrer": "DEFAULT"
}
```

**요청 (GET)**:
```
GET /v1/auth/toss/unlink-callback?userKey=12345678&referrer=DEFAULT
```

| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| userKey | number | O | 토스 사용자 키 |
| referrer | string | X | 리퍼러 |

**응답** (200 OK):
```json
{
  "ok": true,
  "userKey": 12345678,
  "referrer": "DEFAULT"
}
```

**동작**: 해당 `userKey`의 세션을 즉시 무효화 (revoke)

---

#### GET /v1/me

현재 사용자 정보

**인증**: 필요

**응답** (200 OK):
```json
{
  "userId": "toss:12345678",
  "exp": 1703232000
}
```

---

### 3. 챌린지 (Challenges)

#### GET /v1/challenges

챌린지 목록 조회

**인증**: 필요

**응답** (200 OK):
```json
{
  "items": [
    {
      "id": "walk-7000",
      "title": "매일 7,000보 걷기",
      "days": 3,
      "deposit": 10000,
      "proofType": "steps"
    },
    {
      "id": "bed-0700",
      "title": "아침 7시 이불 개기",
      "days": 3,
      "deposit": 10000,
      "proofType": "photo"
    },
    {
      "id": "lunch-proof",
      "title": "점심 도시락/샐러드 인증",
      "days": 3,
      "deposit": 10000,
      "proofType": "photo"
    }
  ]
}
```

**Challenge 객체**:

| 필드 | 타입 | 설명 |
|------|------|------|
| id | string | 챌린지 ID |
| title | string | 챌린지 제목 |
| days | number | 챌린지 기간 (일) |
| deposit | number | 참가비 (원) |
| proofType | string | 인증 방식 (`"photo"` \| `"steps"`) |

---

### 4. 결제 (Payments)

#### POST /v1/payments/create

결제 생성

**인증**: 필요

**멱등성**: `Idempotency-Key` 헤더 사용 (2분 TTL)

**요청 Body**:
```json
{
  "challengeId": "bed-0700",
  "amount": 10000
}
```

| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| challengeId | string | O | 챌린지 ID |
| amount | number | O | 결제 금액 (원) |

**응답** (200 OK):
```json
{
  "paymentId": "pay_a1b2c3d4",
  "status": "created",
  "challengeId": "bed-0700",
  "amount": 10000
}
```

| 필드 | 타입 | 설명 |
|------|------|------|
| paymentId | string | 결제 ID (pay_ prefix) |
| status | string | `"created"` |
| challengeId | string | 챌린지 ID |
| amount | number | 결제 금액 |

**에러 응답**:

| 상태 | 에러 | 설명 |
|------|------|------|
| 400 | `challengeId and amount are required` | 필수 필드 누락 |
| 409 | `duplicate request` | 중복 요청 (멱등성 키) |

---

#### POST /v1/payments/execute

결제 실행

**인증**: 필요

**요청 Body**:
```json
{
  "paymentId": "pay_a1b2c3d4"
}
```

| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| paymentId | string | O | 결제 ID |

**응답** (200 OK):
```json
{
  "ok": true,
  "status": "done",
  "paymentId": "pay_a1b2c3d4"
}
```

---

### 5. 인증 제출 (Proofs)

#### POST /v1/proofs/submit

인증 제출

**인증**: 필요

**멱등성**: `Idempotency-Key` 헤더 사용 (2분 TTL)

**검증 로직**:
- **EXIF 검증**: 사진의 EXIF 메타데이터에서 촬영 시간을 추출하여 챌린지 기간 내 촬영 여부 확인
- **중복 방지**: 이미지 SHA256 해시를 저장하여 동일 사진 재사용 차단
  - 다른 사용자가 제출한 사진 사용 불가
  - 본인이 이미 제출한 사진 재사용 불가

**요청 Body (사진 인증)**:
```json
{
  "challengeId": "bed-0700",
  "imageBase64": "/9j/4AAQSkZJRgABAQAAAQABAAD..."
}
```

**요청 Body (걸음수 인증)**:
```json
{
  "challengeId": "walk-7000",
  "imageHash": "steps-demo"
}
```

| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| challengeId | string | O | 챌린지 ID |
| imageBase64 | string | 조건부 | Base64 인코딩 이미지 (photo 타입) |
| imageHash | string | 조건부 | 이미지 해시 또는 steps 식별자 |

**응답** (200 OK):
```json
{
  "ok": true,
  "status": "accepted",
  "warnings": ["EXIF 데이터를 읽을 수 없습니다"]
}
```

> `warnings` 필드는 EXIF 검증을 통과했지만 경고가 있는 경우에만 포함됩니다.

**에러 응답**:

| 상태 | 에러 | 설명 |
|------|------|------|
| 400 | `challengeId is required` | 챌린지 ID 누락 |
| 400 | `imageBase64 or imageHash is required` | 인증 데이터 누락 |
| 400 | `활성화된 챌린지 참여가 없습니다` | 결제 완료된 참여 없음 |
| 400 | `인증 실패: 사진이 챌린지 시작 전에 촬영되었습니다` | EXIF 날짜 검증 실패 |
| 400 | `이미 다른 사용자가 제출한 이미지입니다` | 타인의 사진 사용 시도 |
| 400 | `동일한 사진으로 이미 인증하셨습니다` | 본인 사진 재사용 시도 |
| 409 | `duplicate request` | 중복 요청 |

---

### 6. 정산 (Settlements)

#### GET /v1/settlements

현재 사용자의 전체 정산 내역 조회 (일괄 조회)

**인증**: 필요

**응답** (200 OK):
```json
{
  "items": [
    {
      "challengeId": "bed-0700",
      "status": "running",
      "refundable": false,
      "message": "진행중 (1/3일 완료)"
    },
    {
      "challengeId": "walk-7000",
      "status": "success",
      "refundable": true,
      "message": "성공! 환급 예정"
    }
  ]
}
```

| 필드 | 타입 | 설명 |
|------|------|------|
| items | array | 정산 목록 |
| items[].challengeId | string | 챌린지 ID |
| items[].status | string | `"running"` \| `"success"` \| `"failed"` |
| items[].refundable | boolean | 환급 가능 여부 |
| items[].message | string? | 진행 상태 메시지 |

**정산 상태**:

| status | 설명 | refundable |
|--------|------|------------|
| running | 챌린지 진행 중 | false |
| success | 성공 (리워드 지급 대기/완료) | true |
| failed | 실패 (미지급) | false |

**참고**: 개별 조회 API (`GET /v1/settlements/:challengeId`)는 일괄 조회 API로 대체되었습니다.

---

## 에러 응답 형식

### 표준 에러

```json
{
  "error": "에러 메시지"
}
```

### HTTP 상태 코드

| 코드 | 설명 |
|------|------|
| 200 | 성공 |
| 204 | 성공 (본문 없음, OPTIONS) |
| 400 | 잘못된 요청 |
| 401 | 인증 필요 / 토큰 만료 / 세션 취소 |
| 405 | 허용되지 않는 메서드 |
| 409 | 중복 요청 (멱등성 키) |
| 429 | 요청 횟수 초과 (Rate Limit) |
| 502 | 외부 서비스 오류 (토스 API 등) |

---

## Rate Limiting

| 제한 | 값 |
|------|-----|
| 요청 수 | 120 req/min |
| 단위 | IP 기준 |
| 윈도우 | 1분 (Fixed Window) |

**초과 시 응답** (429):
```json
{
  "error": "rate limit exceeded"
}
```

---

## 멱등성 (Idempotency)

### 지원 엔드포인트

| 엔드포인트 | TTL |
|------------|-----|
| POST /v1/payments/create | 2분 |
| POST /v1/proofs/submit | 2분 |

### 사용 방법

```typescript
const idempotencyKey = crypto.randomUUID();

await apiPost("/v1/payments/create", {
  challengeId: "bed-0700",
  amount: 10000
}, idempotencyKey);
```

### 중복 요청 시

```json
{
  "error": "duplicate request"
}
```

---

## 프론트엔드 API 클라이언트

### api.ts

```typescript
import { API_BASE_URL } from "./env";
import { clearAccessToken, getAccessToken } from "./storage";

export async function apiGet<T>(path: string): Promise<T> {
  return apiRequest<T>(path, { method: "GET" });
}

export async function apiPost<T>(
  path: string,
  body?: unknown,
  idempotencyKey?: string
): Promise<T> {
  return apiRequest<T>(path, {
    method: "POST",
    body: body ? JSON.stringify(body) : undefined,
    headers: {
      "Content-Type": "application/json",
      ...(idempotencyKey ? { "Idempotency-Key": idempotencyKey } : {}),
    },
  });
}

async function apiRequest<T>(path: string, init: RequestInit): Promise<T> {
  const token = getAccessToken();
  const url = API_BASE_URL ? `${API_BASE_URL}${path}` : path;

  const res = await fetch(url, {
    ...init,
    headers: {
      ...(init.headers || {}),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  });

  const data = await res.json();

  if (!res.ok) {
    // 401: 세션 만료 또는 연결 해제 → 로그인 페이지로
    if (res.status === 401) {
      clearAccessToken();
      window.location.href = "/login?reason=unlinked";
    }
    throw new Error(data?.error || `HTTP ${res.status}`);
  }

  return data as T;
}
```

---

## 토스 API (mTLS)

### 환경 변수

| 변수 | 필수 | 설명 |
|------|------|------|
| AIT_MTLS_CERT_FILE | O | mTLS 인증서 파일 경로 |
| AIT_MTLS_KEY_FILE | O | mTLS 키 파일 경로 |
| AIT_TOSS_BASE_URL | X | 토스 API URL (기본: https://apps-in-toss-api.toss.im) |

### 토스 API 엔드포인트

#### POST /api-partner/v1/apps-in-toss/user/oauth2/generate-token

인가 코드 → 액세스 토큰 교환

**요청**:
```json
{
  "authorizationCode": "...",
  "referrer": "DEFAULT"
}
```

**응답**:
```json
{
  "resultType": "SUCCESS",
  "success": {
    "accessToken": "...",
    "refreshToken": "...",
    "scope": "...",
    "tokenType": "Bearer",
    "expiresIn": 3600
  }
}
```

---

#### GET /api-partner/v1/apps-in-toss/user/oauth2/login-me

사용자 정보 조회

**헤더**: `Authorization: Bearer {accessToken}`

**응답**:
```json
{
  "resultType": "SUCCESS",
  "success": {
    "userKey": 12345678,
    "scope": "..."
  }
}
```

---

## 환경 변수 요약

### 백엔드

| 변수 | 필수 | 기본값 | 설명 |
|------|------|--------|------|
| PORT | X | 8080 | 서버 포트 |
| APP_ENV | X | local | 환경 (local/staging/prod) |
| APP_VERSION | X | dev | 앱 버전 |
| GIT_SHA | X | local | Git 커밋 해시 |
| ALLOW_ORIGIN | X | * | CORS origin |
| SESSION_SECRET | O* | - | 세션 서명 시크릿 (*local에서는 자동 생성) |
| AIT_MTLS_CERT_FILE | O* | - | mTLS 인증서 (*staging/prod 필수) |
| AIT_MTLS_KEY_FILE | O* | - | mTLS 키 (*staging/prod 필수) |
| AIT_TOSS_BASE_URL | X | https://apps-in-toss-api.toss.im | 토스 API URL |
| AIT_UNLINK_BASIC_AUTH | X | - | 연결 해제 콜백 Basic Auth (username:password) |

### 프론트엔드

| 변수 | 필수 | 기본값 | 설명 |
|------|------|--------|------|
| VITE_API_BASE_URL | X | "" (same-origin) | 백엔드 API URL |

---

## API 사용 예시

### 1. 로그인 흐름

```typescript
// 1. 토스 SDK 로그인
import { appLogin } from "@apps-in-toss/web-framework";
const { authorizationCode, referrer } = await appLogin();

// 2. 토큰 교환
const resp = await apiPost<{
  sessionToken: string;
  mode: "toss" | "stub";
}>("/v1/auth/toss/exchange", {
  authorizationCode,
  referrer,
});

// 3. 토큰 저장
setAccessToken(resp.sessionToken);
```

### 2. 챌린지 참여 흐름

```typescript
// 1. 결제 생성
const idempotencyKey = crypto.randomUUID();
const payment = await apiPost<{ paymentId: string }>(
  "/v1/payments/create",
  { challengeId: "bed-0700", amount: 10000 },
  idempotencyKey
);

// 2. 결제 실행 (향후 토스페이 연동)
await apiPost("/v1/payments/execute", {
  paymentId: payment.paymentId
});

// 3. 인증 페이지로 이동
navigate(`/proof/bed-0700`);
```

### 3. 인증 제출 흐름

```typescript
// 1. 이미지 Base64 변환
const base64 = await fileToBase64(imageFile);

// 2. 인증 제출
const idempotencyKey = crypto.randomUUID();
await apiPost("/v1/proofs/submit", {
  challengeId: "bed-0700",
  imageBase64: base64
}, idempotencyKey);

// 3. 정산 내역 페이지로 이동
navigate("/history");
```

---

## TODO (향후 구현)

- [x] GET /v1/settlements 일괄 조회 API 구현 (완료)
- [x] PostgreSQL DB 연동 (완료)
- [ ] 토스페이 결제 SDK 연동
- [ ] 토스 포인트 프로모션 API 연동
- [ ] 이미지 EXIF 검증 로직
- [ ] 이미지 중복 해시 검증
- [ ] Vision AI 자동 분류 (로드맵)
