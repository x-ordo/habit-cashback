# 적용 순서 (한 번에 끝내기)

## 0) 준비물
- GitHub Repo (public/private 상관없음)
- 스테이징 서버 1대 (Ubuntu 권장, Docker 설치)
- 도메인 (예: staging.example.com) → 서버 IP로 A 레코드 연결
- GitHub Actions Secrets (SSH + 도메인)

## 1) 로컬에서 레포 올리기 (Mac)
```bash
# 1) 압축 해제 후 이동
unzip habitcashback_repo_final_v09.zip -d ~/dev
cd ~/dev/habitcashback_repo_final_v09

# 2) git init
git init
git add .
git commit -m "init: habitcashback MVP"

# 3) GitHub 리포 연결 (이미 생성했다면 remote add만)
# gh cli가 있으면:
gh repo create x-ordo/habitcashback --private --source=. --remote=origin --push

# gh가 없으면:
# git remote add origin git@github.com:x-ordo/habitcashback.git
# git push -u origin main
```

## 2) 스테이징 서버 부트스트랩 (Mac)
```bash
bash scripts/40_bootstrap_staging_server_mac.sh <SERVER_IP> <SSH_USER>
```

서버에 `/opt/habitcashback/infra/staging` 폴더가 생성되고, docker compose가 올라갑니다.

## 3) GitHub Secrets 세팅 (Mac)
```bash
bash scripts/30_set_staging_github_secrets_mac.sh \
  x-ordo/habitcashback \
  <STAGING_SSH_HOST> <STAGING_SSH_USER> <PATH_TO_PRIVATE_KEY> \
  <STAGING_DOMAIN>
```

## 4) dev 이미지 자동 배포
- `dev` 브랜치에 push하면:
  - web/api 이미지가 GHCR에 푸시되고
  - 스테이징 서버에서 `docker compose pull && up -d` 자동 실행

## 5) Apps in Toss 심사/배포
```bash
cd frontend
npx ait init
# granite.config.ts 확인
npx ait deploy --api-key <YOUR_AIT_API_KEY>
```

## 3-1) (중요) GHCR Pull 권한
- 레포/패키지가 private이면 스테이징 서버가 이미지를 pull하려면 `docker login ghcr.io`가 필요합니다.
- 방법: GitHub PAT(패키지 read 권한) 생성 → Actions Secrets에 넣기
  - `GHCR_USERNAME`
  - `GHCR_TOKEN`

스크립트로 같이 넣으려면:
```bash
bash scripts/30_set_staging_github_secrets_mac.sh \
  x-ordo/habitcashback \
  <STAGING_SSH_HOST> <STAGING_SSH_USER> <PATH_TO_PRIVATE_KEY> \
  <STAGING_DOMAIN> \
  <GHCR_USERNAME> <PATH_TO_GHCR_TOKEN_FILE>
```


## v15 (Production)
- infra/prod 추가
- .github/workflows/prod-release.yml 추가
- docs/PROD_GO_LIVE.md 추가
