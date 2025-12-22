# Toss 심사/출시 제출 플레이북 (v13)

> 목적: “습관환급”을 **앱인토스(WebView/Granite)**로 **반려 없이** 통과시키기 위한 체크리스트 + 제출 순서.
> 핵심: **샌드박스 테스트 완료**, **iframe 금지**, **mTLS 필수**, **토큰 기반 인증**.

---

## 0) 반려 포인트(필수 회피)
- iframe 사용 금지 (유튜브 삽입 목적만 예외)
- 앱인토스 API는 **mTLS 미적용 시 ERR_NETWORK** 발생
- **샌드박스 테스트 1회 이상 필수** (미완료 시 반려 가능)
- **QR 테스트 조건**: 워크스페이스 멤버 + 토스 로그인 완료 + (정책상) 성인 조건

---

## 1) 콘솔 준비물

- **복붙용 값:** `docs/TOSS_CONSOLE_COPYPASTE.md`
- **필수 링크:** ` /privacy`, `/terms` (도메인별)
- **(권장) 연결 끊기 콜백:** `/v1/auth/toss/unlink-callback`
- AppName(App ID) 확정 → `frontend/granite.config.ts` 의 `appName`과 동일
- 아이콘 URL (선택) → `AIT_ICON_URL`
- 런타임 권한: camera/photos (필요 최소만)

---

## 2) 로컬에서 .ait 생성
```bash
cd frontend
pnpm i
npx ait build
```

---

## 3) 콘솔 업로드 → QR 테스트
1) 콘솔 업로드 → QR 발급  
2) 토스 앱에서 QR 스캔 → 앱 실행 확인  
3) 최소 1회 이상 “챌린지 진입 → 인증 화면”까지 동작 확인  
4) 테스트 완료 표기 확인 후 검토 요청

---

## 4) 심사 제출 전 “카피” 점검(사행성 회피)
권장 용어:
- 참가비 / 리워드 / 혜택 / 성공 조건 / 미지급  
리스크 단어(회피):
- 배팅 / 도박 / 투자 / 수익 보장 / 현금화

---

## 5) 운영/보안 점검
- `ALLOW_ORIGIN`은 스테이징/프로덕션에서 반드시 도메인으로 고정
- `SESSION_SECRET`은 스테이징/프로덕션에서 필수
- iOS 쿠키 제약 가능성 대비: **세션은 토큰 기반**으로 유지

---

## 6) 제출 전 자동 점검 커맨드
```bash
# iframe 금지 점검
bash scripts/10_check_iframe.sh

# 프론트 빌드
bash scripts/20_build_frontend.sh

# 백엔드 빌드
bash scripts/30_build_backend.sh

# 스테이징 스모크
bash scripts/40_smoke_staging.sh https://YOUR_STAGING_API
```
