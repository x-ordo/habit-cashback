# 05. Toss API Mapping (Apps in Toss) — v0.7

> 목적: 백엔드(Go)가 호출해야 하는 Toss Partner API를 **엔드포인트/헤더/순서** 기준으로 고정.

## 1) 공통 전제

- Toss Partner API 호출은 **mTLS(클라이언트 인증서)**가 필요합니다.
- 대부분의 엔드포인트는 `X-Toss-User-Key` 헤더가 필요합니다. (`login-me`로 userKey 확보)

## 2) Base URLs

- Login / Promotion / Messenger: `https://apps-in-toss-api.toss.im`
- Toss Pay: `https://pay-apps-in-toss-api.toss.im`

## 3) Login (OAuth2)

1) 토스 앱 WebView에서 `appLogin()` 실행 → `{ authorizationCode, referrer }`
2) 서버에서 토큰 교환:

- `POST /api-partner/v1/apps-in-toss/user/oauth2/generate-token`
  - body: `{ authorizationCode, referrer }`

3) userKey 확보:

- `GET /api-partner/v1/apps-in-toss/user/oauth2/login-me`
  - header: `Authorization: Bearer <accessToken>`
  - success: `{ userKey }`

## 4) Toss Pay (보증금 결제)

- `POST /api-partner/v1/apps-in-toss/pay/make-payment`
  - header: `X-Toss-User-Key`, `Idempotency-Key`
  - body: `{ orderNo, productDesc, amount, amountTaxFree, isTestPayment }`
  - success: `{ payToken }`

- (프론트) `checkoutPayment({ payToken })`

- `POST /api-partner/v1/apps-in-toss/pay/execute-payment`
  - header: `X-Toss-User-Key`
  - body: `{ payToken, orderNo, isTestPayment }`

## 5) Promotion (포인트/리워드 지급)

권장 호출 순서:

- `POST /api-partner/v1/apps-in-toss/promotion/get-key`
  - header: `X-Toss-User-Key`
  - body: `{ promotionCode }`
  - success: `{ key }`

- `POST /api-partner/v1/apps-in-toss/promotion/execute-promotion`
  - header: `X-Toss-User-Key`, `Idempotency-Key`
  - body: `{ key, value }`

- `POST /api-partner/v1/apps-in-toss/promotion/execution-result`
  - header: `X-Toss-User-Key`
  - body: `{ key }`

## 6) Messenger (알림 발송)

- `POST /api-partner/v1/apps-in-toss/messenger/send-message`
  - header: `X-Toss-User-Key`
  - body: `{ templateSetCode, context }`

## 7) 심사/UX 규칙

- **iframe 불가**(일부 예외 존재) → WebView 화면 구성 시 iframe 의존 금지.
