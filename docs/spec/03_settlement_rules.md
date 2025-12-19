# 03. Settlement Rules (포인트 기반)

## 기본 컨셉
- 유저가 낸 돈은 “참가비(결제)”로 처리.
- 성공한 만큼 “환급(포인트)”이 지급됨.
- 실패분은 **보너스 풀(pool)**로 적립되고, 일부는 플랫폼 수수료로 전환.

## 변수 (Template)
- `feeKRW` : 참가비
- `days` : 기간
- `baseRefundRate` : 기본 환급률(예: 0.7 = 참가비의 70%를 ‘성공 시’ 돌려주는 느낌)
- `poolTakeRate` : 풀에서 플랫폼이 가져가는 비율(예: 0.5)
- `savePassPriceKRW` : 세이브권 결제 금액(예: 2,000)

## 일일 정산(추천)
- `basePointPerDay = floor(feeKRW * baseRefundRate / days)` 를 포인트(원 단위)로 지급
- PASS면 당일 지급(또는 익일 지급)
- FAIL이면 그 날의 `basePointPerDay`가 pool로 이동

## 종료 정산(보너스)
- `poolPoints = sum(FAIL day basePointPerDay)`
- `platformTake = floor(poolPoints * poolTakeRate)`
- `winnerPool = poolPoints - platformTake`
- 분배 방식(단순):
  - 완주자(FAILED 아닌 참가자)에게 `winnerPool / numFinishers` 균등 지급

## 세이브권(복구)
- FAIL 확정 전 “유예시간” 동안만 구매 가능(예: 마감 후 60분)
- 세이브권 구매 시:
  - 해당 day를 PASS로 변경
  - 당일 basePoint 지급 대상 복구
  - pool에서 해당 basePoint 차감(이미 적립됐으면 되돌림)

## 왜 이 규칙이 좋은가
- 유저에게는 “원금 느낌 환급”이 생김
- 플랫폼은 실패/풀 수수료 + 세이브권 매출로 수익이 꾸준히 발생
- 실제 현금 환불 없이 운영 가능(포인트 지급)
