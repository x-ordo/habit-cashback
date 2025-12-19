# 01. End-to-End Flow (습관환급)

## 핵심 목표
- 유저가 “돈을 맡겼다”로 느끼게 하되, 구현/정산/CS는 **포인트 지급(프로모션)** 중심으로 단순화.
- 토스 심사/리스크: **사행성/도박**으로 보이지 않게, 문구/정책은 “환급(리워드) + 미션”으로 고정.

---

## A) 로그인 (Toss Login)
1. 프론트(앱인토스)에서 SDK `appLogin` 호출 → `{authorizationCode, referrer}`
2. 백엔드가 `generate-token` 호출 → `{accessToken, refreshToken, expiresIn}`
3. 백엔드가 `login-me` 호출 → `userKey` 획득 (이게 이후 모든 API의 핵심 식별자)
4. 백엔드가 우리 서비스 세션 토큰 발급(JWT) → 프론트에 전달

### 이유
- iOS 환경에서 쿠키 의존하면 깨질 수 있으니, **토큰 기반 인증**으로 간다(세션 JWT).

---

## B) 참가(결제) — Toss Pay
1. 프론트: “참가하기” 버튼
2. 백엔드: `make-payment` 호출 (주문번호 `orderNo`는 **유니크**)
3. 응답 `payToken` 저장
4. 프론트: SDK `checkoutPayment(payToken)`로 사용자 인증
5. 인증 성공 시 프론트 → 백엔드에 `payToken` 전달
6. 백엔드: `execute-payment(payToken, orderNo, isTestPayment)` 호출 → 승인 완료
7. 참가 상태: `ENROLLED → ACTIVE`

---

## C) 인증(제출)
- 하루 1회 제출 (사진/만보기 등)
- 제출은 `PENDING` → 자동검증 → `PASS|FAIL`
- 마감 이후에는 “세이브권” 처리만 허용(정책/문구로 고정)

---

## D) 정산(포인트 지급) — Promotion
**원칙:** 현금 환불이 아니라, “성공한 만큼 포인트 지급” (CS/PG 꼬임 최소화)

1. (지급 직전) `get-key` 호출 → key 발급(유효 1시간)
2. `execute-promotion(promotionCode, key, amount)` 호출
3. `execution-result`로 최종 `SUCCESS` 확인
4. 우리 DB에서 `PAYOUT SUCCESS`로 확정

---

## E) 알림 (Messenger)
- 마감 30분 전 / 실패 직후 / 성공 직후 3종만.
- 템플릿 변수는 백엔드에서 치환 후 발송.

---

## 장애/재시도 원칙 (중요)
- Toss 문서도 “중복 지급/중복 발송 방지”를 강하게 요구 → 우리 쪽에서 **멱등성 키**로 막는다.
