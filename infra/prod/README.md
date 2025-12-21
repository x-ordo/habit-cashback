# Staging (dev) 배포

## 1) 전제
- 서버에 `/opt/habitcashback` 경로로 레포가 배치되어 있어야 합니다.
- GitHub Actions(`production-dev.yml`)가 SSH로 서버에 접속해 `docker compose up -d`를 수행합니다.

## 2) Apps in Toss mTLS 인증서 준비 (필수)
Apps in Toss API 호출은 서버 mTLS가 필수입니다.

서버에 아래 파일을 준비하세요:

- `infra/production/secrets/ait_mtls_cert.pem`
- `infra/production/secrets/ait_mtls_key.pem`

docker-compose가 위 파일을 컨테이너 `/run/secrets/`로 마운트합니다.

## 3) 런타임 환경변수
`infra/production/docker-compose.yml`의 `api.environment`에서 아래 변수를 사용합니다.

- `AIT_BASE_URL` (default: https://apps-in-toss-api.toss.im)
- `AIT_MTLS_CERT_FILE` (default: /run/secrets/ait_mtls_cert.pem)
- `AIT_MTLS_KEY_FILE`  (default: /run/secrets/ait_mtls_key.pem)

## 4) 토스 로그인 토큰 교환 엔드포인트
- `POST /v1/auth/toss/exchange`
  - body: `{ "authorizationCode": string, "referrer": "DEFAULT" | "SANDBOX" }`
  - response: `{ "sessionToken": string, "expiresIn": string }`

로컬 브라우저에서는 `appLogin()`이 동작하지 않으므로 프론트에서 자동으로 `/v1/auth/exchange`(stub)로 폴백합니다.


## Production note
- Open TCP 80/443 to allow Caddy automatic HTTPS.
- Prefer a separate server for prod (no shared staging). Use non-root user + firewall.
