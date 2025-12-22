# Toss Apps in Toss 로그인(토스 로그인) - 서버 mTLS 교환

## 흐름
1) 프론트에서 `appLogin()` 호출 → `{ authorizationCode, referrer }` 획득
2) 프론트가 백엔드 `POST /v1/auth/toss/exchange` 호출
3) 백엔드가 Apps in Toss API에 mTLS로 요청:
   - Base URL: `https://apps-in-toss-api.toss.im`
   - Endpoint: `POST /api-partner/v1/apps-in-toss/user/oauth2/generate-token`
   - Body: `{ "authorizationCode": "...", "referrer": "DEFAULT" | "SANDBOX" }`
4) 백엔드가 `sessionToken` 발급 후 반환 (현재는 in-memory)

## 서버 환경변수
- `AIT_MTLS_CERT_FILE` : PEM 인증서 경로
- `AIT_MTLS_KEY_FILE` : PEM 개인키 경로
- `AIT_BASE_URL` : 기본값 `https://apps-in-toss-api.toss.im`

## 주의
- Apps in Toss API는 mTLS 미설정 시 `ERR_NETWORK` 류의 네트워크 오류가 발생할 수 있습니다.
- Alpine 런타임 이미지에는 서버 인증서 검증을 위해 `ca-certificates`가 필요합니다. (Dockerfile 반영됨)

## TODO (프로덕션)
- sessionToken을 DB/Redis에 저장 (서버 재시작 시 유지)
- Refresh token 회전 및 만료 처리
- 필요 시 사용자 정보 조회 API 연결
