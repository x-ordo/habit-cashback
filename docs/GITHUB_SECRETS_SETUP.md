# GitHub Secrets/Environments 세팅 (습관환급)

## 전제
- `gh` 로그인 (`gh auth login`)
- 대상 repo: `<OWNER>/<REPO>`
- **절대 PEM/키를 Git에 커밋하지 말 것** (secrets/ 폴더는 로컬 전용)

## 0) 로컬 파일 준비(예시)
로컬에 아래 파일을 둡니다 (경로는 원하는대로 가능).

- `./secrets/toss_mtls_cert.pem`
- `./secrets/toss_mtls_key.pem`
- (옵션) `./secrets/toss_mtls_ca.pem`

> PEM 파일은 Toss 개발자센터에서 발급받은 mTLS 인증서/키를 그대로 사용.

## 1) pack 반영
```bash
unzip ~/Downloads/habitcashback_github_pack_v04_secrets.zip -d .
git add .
git commit -m "chore: github env+secrets pack v04"
git push
```

## 2) Environment 생성
```bash
OWNER="<OWNER>"
REPO="<REPO>"
bash scripts/10_create_envs.sh "$OWNER" "$REPO"
```

## 3) Secrets 세팅 (production 예시)
복호화 키/ AAD는 환경변수로 넘기거나, 스크립트가 없으면 입력을 요구합니다.

```bash
export TOSS_DECRYPTION_KEY_B64="<YOUR_KEY_B64>"
export TOSS_DECRYPTION_AAD="<YOUR_AAD>"

bash scripts/11_set_toss_secrets.sh "$OWNER" "$REPO" production   ./secrets/toss_mtls_cert.pem   ./secrets/toss_mtls_key.pem   ./secrets/toss_mtls_ca.pem
```

staging도 동일:
```bash
bash scripts/11_set_toss_secrets.sh "$OWNER" "$REPO" staging   ./secrets/toss_mtls_cert.pem   ./secrets/toss_mtls_key.pem
```

## 4) Variables 세팅 (Secrets 아닌 값)
```bash
bash scripts/12_set_vars.sh "$OWNER" "$REPO"
```

## 5) 확인
```bash
bash scripts/13_list_env_config.sh "$OWNER" "$REPO"
```

## Naming 규칙(확정)
Environment Secrets (staging/production):
- `TOSS_MTLS_CERT_PEM`
- `TOSS_MTLS_KEY_PEM`
- (옵션) `TOSS_MTLS_CA_PEM`
- `TOSS_DECRYPTION_KEY_B64`
- `TOSS_DECRYPTION_AAD`

Repository Variables:
- `TOSS_DEVELOPER_CENTER_URL` (기본값: https://developers-apps-in-toss.toss.im/)
- `SERVICE_NAME` (habitcashback)
