# 04. Idempotency & Anti-Abuse

Toss 문서 기준으로도 “중복 지급/중복 발송 방지”는 파트너 책임입니다.

## 1) Promotion 지급 멱등성
- `payout_idempotency_key = "promo:{challengeId}:{participantId}:{dayIndex}:{amount}"`
- DB에 UNIQUE 인덱스로 잡고, 이미 SUCCESS면 재호출 금지

### get-key 주의
- key 유효 1시간
- 사용한 key 재사용 시 에러
- 따라서 `get-key`는 execute 직전에만 발급

## 2) Push 멱등성
- `push_idempotency_key = "push:{template}:{participantId}:{event}:{date}"`

## 3) Payment 주문번호(orderNo) 정책
- orderNo는 매회 유니크
- 인증 완료 이후 재사용 불가
- 테스트/라이브 충돌 방지 위해 prefix 분리(예: TEST- / LIVE-)

## 4) 사진 인증 어뷰징(P0)
- 카메라 촬영 강제 + EXIF 시간 검증
- 이미지 해시(sha256)로 재사용 차단
