# 02. State Machine

## Challenge Instance
- `DRAFT` : 생성만 됨
- `ENROLLING` : 참가 모집 중(결제 가능)
- `RUNNING` : 인증 진행 중
- `SETTLING` : 지급/정산 처리 중
- `CLOSED` : 종료

## Participant
- `CREATED` : 아직 결제 전
- `ENROLLED` : 결제 완료(참가 확정)
- `ACTIVE` : 진행 중
- `FAILED` : 규칙 위반/미제출로 탈락(일부 챌린지는 일일 실패만 존재)
- `COMPLETED` : 종료 시점까지 유지

## Payment
- `CREATED` (make-payment 성공, payToken 확보)
- `AUTHED` (checkoutPayment 인증 성공)
- `PAID` (execute-payment 성공)
- `REFUNDED` (refund-payment 성공)
- `FAILED`

## Submission
- `PENDING`
- `PASS`
- `FAIL`

## Payout (Promotion)
- `PENDING` : 지급 대상 확정
- `KEY_ISSUED` : get-key 성공
- `REQUESTED` : execute-promotion 호출 완료
- `SUCCESS` : execution-result 최종 성공
- `FAILED` : 실패(재시도 큐로 이동)
