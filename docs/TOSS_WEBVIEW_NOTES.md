# Apps in Toss(WebView) 체크리스트

## granite.config.ts 필수 포인트
- appName: 콘솔에 등록한 앱 ID와 **완전히 동일**
- brand.icon: 필수 필드 (빈 문자열 허용)
- web.commands: dev/build 명령 정확히
- outdir: dist
- webViewProps.type: partner (일반 서비스)

## 권한
- camera/photos 권한을 요청해야 인증 사진 플로우가 깔끔합니다.

## 심사 관점 (이 MVP 기준)
- 명칭/카피는 '사행성'으로 보이지 않게: **보증금/환급/정산** 톤 유지
- 환급/차감 로직은 '리워드/포인트'로 먼저 설계 (현금 환불 최소화)
- 어뷰징 방지(중복 사진/시간/EXIF)는 최소한 설계 문서로 제시
