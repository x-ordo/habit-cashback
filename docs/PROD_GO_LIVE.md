# Production Go-Live Checklist (습관환급)

## 0) 전제
- **Staging에서 Toss Login + 챌린지 흐름 + unlink callback**까지 통과한 상태에서만 진행.
- Prod는 staging과 **서버/도메인/시크릿 분리**.

## 1) 인프라
- DNS: `PROD_DOMAIN` A/AAAA 레코드 설정
- 방화벽: 80/443 오픈 (Caddy 자동 HTTPS)
- 서버: Docker + Compose 설치, `/opt/habitcashback` 경로 준비
- mTLS: `secrets/ait_mtls_cert.pem`, `secrets/ait_mtls_key.pem` 배치 또는 GH Secret로 주입

## 2) GitHub Secrets (prod-release.yml 기준)
- PROD_SSH_HOST
- PROD_SSH_USER
- PROD_SSH_KEY
- PROD_DOMAIN
- GHCR_USERNAME
- GHCR_TOKEN
- AIT_MTLS_CERT_PEM_B64 (optional if server has file)
- AIT_MTLS_KEY_PEM_B64 (optional if server has file)
- AIT_UNLINK_BASIC_AUTH (optional)

## 3) 배포
- Actions → **prod-release** → Run workflow
  - image_tag: `prod` (기본) 또는 `release-YYYYMMDD`

## 4) 검증 (필수)
- `https://{PROD_DOMAIN}/health` → 200
- `https://{PROD_DOMAIN}/meta` → version/commit 확인
- Toss 로그인 → 홈/챌린지 진입 → 인증 업로드 → 기록 페이지
- Toss 콘솔에서 연결 끊기 → unlink callback 200 → 재진입 시 login 유도

## 5) 심사 제출 전 링크
- `https://{PROD_DOMAIN}/terms`
- `https://{PROD_DOMAIN}/privacy`
- `https://{PROD_DOMAIN}/support`
