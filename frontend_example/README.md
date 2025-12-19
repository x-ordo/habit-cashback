# Apps-in-Toss Frontend Example (습관환급)

이 폴더는 “동작 예시”용입니다. 실제 서비스 프론트엔드에 그대로 복붙할 수 있게 최소 코드만 제공합니다.

## 핵심 SDK 호출

- `appLogin()` → `{ authorizationCode, referrer }`
- `checkoutPayment({ payToken })`

> 참고: Toss Apps-in-Toss WebView는 **iframe 사용이 제한**됩니다. 화면 구성은 React/DOM 기반으로만 가세요.

## 파일

- `src/toss.ts` : SDK wrapper
- `src/demo.ts` : 로그인 → 결제 생성 → 결제 UI → 결제 실행 순서 샘플
