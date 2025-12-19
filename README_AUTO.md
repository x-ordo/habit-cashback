# GitHub Pack v03 (습관환급)

이 폴더는 **GitHub 세팅 자동화**만 제공합니다.

포함:
- `.github/workflows/ci-go.yml` : Go backend CI (backend/ 기준)
- `scripts/05_detect_checks.sh` : 실제 생성된 Actions 체크 이름을 자동 추출
- `scripts/02_branch_protection.sh` : 추출된 체크 이름으로 브랜치 보호 자동 적용
- `.github/CODEOWNERS` : 코드오너 리뷰 강제용(핸들 교체 필요)

권장 순서:
1) 워크플로 파일 커밋/푸시 → Actions 체크 1회 생성
2) `scripts/05_detect_checks.sh`로 컨텍스트 확인
3) `scripts/02_branch_protection.sh` 실행
