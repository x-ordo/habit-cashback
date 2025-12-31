# Contributing to HabitCashback

HabitCashback에 기여해 주셔서 감사합니다.

## Code of Conduct

이 프로젝트에 참여하는 모든 분들은 상호 존중과 전문성을 유지해 주시기 바랍니다.

## How to Contribute

### Reporting Bugs

버그를 발견하셨다면:

1. [GitHub Issues](https://github.com/x-ordo/habit-cashback/issues)에서 이미 보고된 버그인지 확인해주세요.
2. 새로운 버그라면 Bug Report 템플릿을 사용하여 이슈를 생성해주세요.

### Suggesting Features

새로운 기능을 제안하고 싶으시다면:

1. Feature Request 템플릿을 사용하여 이슈를 생성해주세요.
2. 기능의 필요성과 예상 구현 방법을 명확히 설명해주세요.

### Pull Requests

1. **Issue 먼저 생성**: PR을 제출하기 전에 관련 이슈를 먼저 생성해주세요.
2. **Fork & Branch**: 리포지토리를 Fork하고 feature branch를 생성하세요.
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **코드 작성**: 변경 사항을 구현하세요.
4. **테스트**: 변경 사항이 기존 기능을 손상시키지 않는지 확인하세요.
5. **커밋**: 명확한 커밋 메시지를 작성하세요.
6. **Push & PR**: 브랜치를 Push하고 Pull Request를 생성하세요.

## Development Setup

### Prerequisites

- Go 1.22+
- Node.js 20+
- Docker & Docker Compose
- pnpm

### Frontend

```bash
cd frontend
corepack enable
pnpm i
pnpm dev
```

### Backend

```bash
cd backend
go mod tidy
go run ./cmd/api
```

## Commit Message Convention

커밋 메시지는 다음 형식을 따라주세요:

```
<type>: <subject>

[optional body]
```

**Types:**
- `feat`: 새로운 기능
- `fix`: 버그 수정
- `docs`: 문서 수정
- `style`: 코드 포맷팅, 세미콜론 누락 등
- `refactor`: 코드 리팩토링
- `test`: 테스트 추가/수정
- `chore`: 빌드 프로세스 또는 보조 도구 변경

## Questions?

질문이 있으시면 parkdavid31@gmail.com으로 연락주세요.
