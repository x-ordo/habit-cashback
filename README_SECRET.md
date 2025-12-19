# GitHub Pack v04 — Secrets/Environments (습관환급)

목표: **Apps in Toss 연동에 필요한 민감값(mTLS/복호화 키 등)을 GitHub에 안전하게** 넣고,
`staging` / `production` 환경으로 분리한다.

포함:
- scripts/10_create_envs.sh : staging/production 환경 생성
- scripts/11_set_toss_secrets.sh : Toss 관련 secrets를 environment에 저장
- scripts/12_set_vars.sh : 변수(Secrets 아닌 값) 세팅
- scripts/13_list_env_config.sh : 세팅 결과 확인(list)
- docs/GITHUB_SECRETS_SETUP.md : 실행 순서/주의사항
- docs/ACTIONS_MTLS_SNIPPET.md : Actions에서 PEM 파일로 복원하는 스니펫
