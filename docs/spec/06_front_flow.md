# 06. Front Flow (Apps in Toss) — v0.7

## Goal
프론트는 “SDK 호출(로그인/결제)”만 담당하고, **결제 생성/정산/리워드/메시지**는 백엔드가 책임진다.

## A) 로그인

1. `appLogin()` 실행 → `{ authorizationCode, referrer }`
2. `POST /v1/auth/exchange`로 전달
3. 서버가 `token(JWT)`를 반환
4. 이후 API 요청은 `Authorization: Bearer <token>`

## B) 보증금 결제

1. `POST /v1/payments/create` → `{ orderNo, payToken }`
2. `checkoutPayment({ payToken })` 실행
3. 결제 UI에서 사용자가 인증/승인 완료
4. `POST /v1/payments/execute` → 결제 최종 실행

## C) 리워드(환급/포인트 지급)

- 조건 충족 시 서버가 `POST /v1/payouts/issue` 호출
- 서버는 key 발급 → execute → (필요 시) result로 최종 확인
