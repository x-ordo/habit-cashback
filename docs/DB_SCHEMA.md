# DB 스키마 정의서 (Database Schema Definition)

> **습관환급 (Habit Cashback) MVP 데이터베이스 설계**
> PostgreSQL 14+

---

## 개요

| 항목 | 내용 |
|------|------|
| DBMS | PostgreSQL 14+ |
| 스키마 버전 | v1.0 |
| 문자셋 | UTF-8 |
| 타임존 | UTC (TIMESTAMPTZ) |

---

## ERD (Entity Relationship Diagram)

```
┌─────────────┐       ┌─────────────────┐       ┌─────────────┐
│  app_user   │───1:N─│  participation  │───N:1─│  challenge  │
└──────┬──────┘       └────────┬────────┘       └─────────────┘
       │                       │
       │ 1:N                   │ 1:N
       ▼                       ▼
┌─────────────┐       ┌─────────────────┐
│   payment   │       │      proof      │
└─────────────┘       └─────────────────┘
       │
       │ 1:1
       ▼
┌─────────────┐
│  settlement │
└─────────────┘
       │
       │ 1:1
       ▼
┌─────────────┐
│    payout   │
└─────────────┘

┌─────────────┐       ┌─────────────────┐
│ idempotency │       │   revoked_session│
└─────────────┘       └─────────────────┘
```

---

## 테이블 정의

### 1. app_user (사용자)

토스 연동 사용자 정보

```sql
CREATE TABLE IF NOT EXISTS app_user (
  id            BIGSERIAL PRIMARY KEY,
  toss_user_key TEXT NOT NULL UNIQUE,        -- 토스 userKey (toss:12345678)
  status        TEXT NOT NULL DEFAULT 'active', -- active | suspended | deleted
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_app_user_status ON app_user(status);
```

| 컬럼 | 타입 | 필수 | 기본값 | 설명 |
|------|------|------|--------|------|
| id | BIGSERIAL | O | auto | PK |
| toss_user_key | TEXT | O | - | 토스 사용자 식별자 (unique) |
| status | TEXT | O | 'active' | 사용자 상태 |
| created_at | TIMESTAMPTZ | O | NOW() | 생성 시간 |
| updated_at | TIMESTAMPTZ | O | NOW() | 수정 시간 |

**status 값**:
| 값 | 설명 |
|----|------|
| active | 정상 |
| suspended | 정지 (부정 사용 등) |
| deleted | 탈퇴/삭제 |

---

### 2. challenge (챌린지)

챌린지 마스터 데이터

```sql
CREATE TABLE IF NOT EXISTS challenge (
  id          TEXT PRIMARY KEY,              -- walk-7000, bed-0700, lunch-proof
  title       TEXT NOT NULL,
  description TEXT,
  days        INT NOT NULL DEFAULT 3,        -- 챌린지 기간 (일)
  deposit     BIGINT NOT NULL DEFAULT 10000, -- 참가비 (원)
  proof_type  TEXT NOT NULL DEFAULT 'photo', -- photo | steps
  is_active   BOOLEAN NOT NULL DEFAULT true,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_challenge_active ON challenge(is_active);
```

| 컬럼 | 타입 | 필수 | 기본값 | 설명 |
|------|------|------|--------|------|
| id | TEXT | O | - | PK (챌린지 ID) |
| title | TEXT | O | - | 챌린지 제목 |
| description | TEXT | X | - | 상세 설명 |
| days | INT | O | 3 | 챌린지 기간 |
| deposit | BIGINT | O | 10000 | 참가비 (원) |
| proof_type | TEXT | O | 'photo' | 인증 방식 |
| is_active | BOOLEAN | O | true | 활성화 여부 |
| created_at | TIMESTAMPTZ | O | NOW() | 생성 시간 |
| updated_at | TIMESTAMPTZ | O | NOW() | 수정 시간 |

**proof_type 값**:
| 값 | 설명 |
|----|------|
| photo | 사진 인증 |
| steps | 걸음수 인증 |

**초기 데이터**:
```sql
INSERT INTO challenge (id, title, days, deposit, proof_type) VALUES
  ('walk-7000', '매일 7,000보 걷기', 3, 10000, 'steps'),
  ('bed-0700', '아침 7시 이불 개기', 3, 10000, 'photo'),
  ('lunch-proof', '점심 도시락/샐러드 인증', 3, 10000, 'photo')
ON CONFLICT (id) DO NOTHING;
```

---

### 3. participation (챌린지 참여)

사용자의 챌린지 참여 정보

```sql
CREATE TABLE IF NOT EXISTS participation (
  id            BIGSERIAL PRIMARY KEY,
  user_id       BIGINT NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  challenge_id  TEXT NOT NULL REFERENCES challenge(id) ON DELETE RESTRICT,
  payment_id    BIGINT REFERENCES payment(id),  -- 결제 연결
  status        TEXT NOT NULL DEFAULT 'pending', -- pending | active | completed | failed | cancelled
  start_date    DATE,                            -- 챌린지 시작일
  end_date      DATE,                            -- 챌린지 종료일
  proof_count   INT NOT NULL DEFAULT 0,          -- 인증 완료 횟수
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  UNIQUE(user_id, challenge_id, start_date)      -- 같은 날 같은 챌린지 중복 참여 방지
);

CREATE INDEX IF NOT EXISTS idx_participation_user ON participation(user_id);
CREATE INDEX IF NOT EXISTS idx_participation_status ON participation(status);
CREATE INDEX IF NOT EXISTS idx_participation_end_date ON participation(end_date) WHERE status = 'active';
```

| 컬럼 | 타입 | 필수 | 기본값 | 설명 |
|------|------|------|--------|------|
| id | BIGSERIAL | O | auto | PK |
| user_id | BIGINT | O | - | FK → app_user |
| challenge_id | TEXT | O | - | FK → challenge |
| payment_id | BIGINT | X | - | FK → payment |
| status | TEXT | O | 'pending' | 참여 상태 |
| start_date | DATE | X | - | 시작일 |
| end_date | DATE | X | - | 종료일 |
| proof_count | INT | O | 0 | 인증 완료 횟수 |
| created_at | TIMESTAMPTZ | O | NOW() | 생성 시간 |
| updated_at | TIMESTAMPTZ | O | NOW() | 수정 시간 |

**status 값**:
| 값 | 설명 |
|----|------|
| pending | 결제 대기 |
| active | 진행 중 |
| completed | 성공 완료 |
| failed | 실패 |
| cancelled | 취소 |

---

### 4. payment (결제)

결제 정보

```sql
CREATE TABLE IF NOT EXISTS payment (
  id           BIGSERIAL PRIMARY KEY,
  user_id      BIGINT NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  challenge_id TEXT NOT NULL REFERENCES challenge(id) ON DELETE RESTRICT,
  order_no     TEXT NOT NULL UNIQUE,           -- 주문 번호 (pay_xxxxxxxx)
  pay_token    TEXT,                           -- 토스페이 토큰
  amount       BIGINT NOT NULL DEFAULT 0,      -- 결제 금액 (원)
  status       TEXT NOT NULL DEFAULT 'created', -- created | pending | done | failed | refunded
  pg_tx_id     TEXT,                           -- PG 거래 ID
  raw_json     JSONB,                          -- PG 응답 원본
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_user_created ON payment(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_order_no ON payment(order_no);
CREATE INDEX IF NOT EXISTS idx_payment_status ON payment(status);
```

| 컬럼 | 타입 | 필수 | 기본값 | 설명 |
|------|------|------|--------|------|
| id | BIGSERIAL | O | auto | PK |
| user_id | BIGINT | O | - | FK → app_user |
| challenge_id | TEXT | O | - | FK → challenge |
| order_no | TEXT | O | - | 주문 번호 (unique) |
| pay_token | TEXT | X | - | 토스페이 결제 토큰 |
| amount | BIGINT | O | 0 | 결제 금액 |
| status | TEXT | O | 'created' | 결제 상태 |
| pg_tx_id | TEXT | X | - | PG 거래 ID |
| raw_json | JSONB | X | - | PG 응답 원본 |
| created_at | TIMESTAMPTZ | O | NOW() | 생성 시간 |
| updated_at | TIMESTAMPTZ | O | NOW() | 수정 시간 |

**status 값**:
| 값 | 설명 |
|----|------|
| created | 생성됨 |
| pending | 결제 진행 중 |
| done | 결제 완료 |
| failed | 결제 실패 |
| refunded | 환불 완료 |

---

### 5. proof (인증)

인증 제출 기록

```sql
CREATE TABLE IF NOT EXISTS proof (
  id               BIGSERIAL PRIMARY KEY,
  participation_id BIGINT NOT NULL REFERENCES participation(id) ON DELETE CASCADE,
  user_id          BIGINT NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  challenge_id     TEXT NOT NULL REFERENCES challenge(id) ON DELETE RESTRICT,
  proof_date       DATE NOT NULL,              -- 인증 날짜
  proof_type       TEXT NOT NULL,              -- photo | steps
  image_hash       TEXT,                       -- 이미지 해시 (중복 검증)
  image_url        TEXT,                       -- 저장된 이미지 URL
  exif_timestamp   TIMESTAMPTZ,                -- EXIF 촬영 시간
  steps_count      INT,                        -- 걸음수 (steps 타입)
  status           TEXT NOT NULL DEFAULT 'pending', -- pending | approved | rejected
  reject_reason    TEXT,                       -- 거부 사유
  verified_at      TIMESTAMPTZ,                -- 검증 완료 시간
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  UNIQUE(participation_id, proof_date)         -- 날짜당 1회 인증
);

CREATE INDEX IF NOT EXISTS idx_proof_user ON proof(user_id);
CREATE INDEX IF NOT EXISTS idx_proof_participation ON proof(participation_id);
CREATE INDEX IF NOT EXISTS idx_proof_status ON proof(status);
CREATE INDEX IF NOT EXISTS idx_proof_image_hash ON proof(image_hash) WHERE image_hash IS NOT NULL;
```

| 컬럼 | 타입 | 필수 | 기본값 | 설명 |
|------|------|------|--------|------|
| id | BIGSERIAL | O | auto | PK |
| participation_id | BIGINT | O | - | FK → participation |
| user_id | BIGINT | O | - | FK → app_user |
| challenge_id | TEXT | O | - | FK → challenge |
| proof_date | DATE | O | - | 인증 날짜 |
| proof_type | TEXT | O | - | 인증 타입 |
| image_hash | TEXT | X | - | 이미지 SHA256 해시 |
| image_url | TEXT | X | - | 이미지 저장 URL |
| exif_timestamp | TIMESTAMPTZ | X | - | EXIF 촬영 시간 |
| steps_count | INT | X | - | 걸음수 |
| status | TEXT | O | 'pending' | 인증 상태 |
| reject_reason | TEXT | X | - | 거부 사유 |
| verified_at | TIMESTAMPTZ | X | - | 검증 시간 |
| created_at | TIMESTAMPTZ | O | NOW() | 생성 시간 |

**status 값**:
| 값 | 설명 |
|----|------|
| pending | 검증 대기 |
| approved | 승인 |
| rejected | 거부 |

---

### 6. settlement (정산)

챌린지 정산 정보

```sql
CREATE TABLE IF NOT EXISTS settlement (
  id               BIGSERIAL PRIMARY KEY,
  participation_id BIGINT NOT NULL UNIQUE REFERENCES participation(id) ON DELETE CASCADE,
  user_id          BIGINT NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  challenge_id     TEXT NOT NULL REFERENCES challenge(id) ON DELETE RESTRICT,
  payment_id       BIGINT REFERENCES payment(id),
  status           TEXT NOT NULL DEFAULT 'running', -- running | success | failed
  refundable       BOOLEAN NOT NULL DEFAULT false,
  deposit_amount   BIGINT NOT NULL DEFAULT 0,       -- 참가비
  reward_amount    BIGINT NOT NULL DEFAULT 0,       -- 리워드 금액
  payout_id        BIGINT REFERENCES payout(id),    -- 지급 연결
  settled_at       TIMESTAMPTZ,                     -- 정산 완료 시간
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_settlement_user ON settlement(user_id);
CREATE INDEX IF NOT EXISTS idx_settlement_status ON settlement(status);
CREATE INDEX IF NOT EXISTS idx_settlement_participation ON settlement(participation_id);
```

| 컬럼 | 타입 | 필수 | 기본값 | 설명 |
|------|------|------|--------|------|
| id | BIGSERIAL | O | auto | PK |
| participation_id | BIGINT | O | - | FK → participation (unique) |
| user_id | BIGINT | O | - | FK → app_user |
| challenge_id | TEXT | O | - | FK → challenge |
| payment_id | BIGINT | X | - | FK → payment |
| status | TEXT | O | 'running' | 정산 상태 |
| refundable | BOOLEAN | O | false | 환급 가능 여부 |
| deposit_amount | BIGINT | O | 0 | 참가비 |
| reward_amount | BIGINT | O | 0 | 리워드 금액 |
| payout_id | BIGINT | X | - | FK → payout |
| settled_at | TIMESTAMPTZ | X | - | 정산 완료 시간 |
| created_at | TIMESTAMPTZ | O | NOW() | 생성 시간 |
| updated_at | TIMESTAMPTZ | O | NOW() | 수정 시간 |

**status 값**:
| 값 | 설명 | refundable |
|----|------|------------|
| running | 진행 중 | false |
| success | 성공 | true |
| failed | 실패 | false |

---

### 7. payout (지급)

토스 포인트 지급 정보

```sql
CREATE TABLE IF NOT EXISTS payout (
  id              BIGSERIAL PRIMARY KEY,
  user_id         BIGINT NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  promotion_code  TEXT NOT NULL,               -- 프로모션 코드
  promotion_key   TEXT NOT NULL UNIQUE,        -- 프로모션 키 (멱등성)
  amount_points   BIGINT NOT NULL,             -- 지급 포인트
  status          TEXT NOT NULL DEFAULT 'requested', -- requested | success | failed | pending
  error_code      TEXT,                        -- 실패 시 에러 코드
  error_message   TEXT,                        -- 실패 시 에러 메시지
  raw_json        JSONB,                       -- 토스 API 응답 원본
  executed_at     TIMESTAMPTZ,                 -- 실행 시간
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payout_user ON payout(user_id);
CREATE INDEX IF NOT EXISTS idx_payout_status ON payout(status);
CREATE INDEX IF NOT EXISTS idx_payout_status_updated ON payout(status, updated_at ASC);
```

| 컬럼 | 타입 | 필수 | 기본값 | 설명 |
|------|------|------|--------|------|
| id | BIGSERIAL | O | auto | PK |
| user_id | BIGINT | O | - | FK → app_user |
| promotion_code | TEXT | O | - | 프로모션 코드 |
| promotion_key | TEXT | O | - | 프로모션 키 (unique) |
| amount_points | BIGINT | O | - | 지급 포인트 |
| status | TEXT | O | 'requested' | 지급 상태 |
| error_code | TEXT | X | - | 에러 코드 |
| error_message | TEXT | X | - | 에러 메시지 |
| raw_json | JSONB | X | - | API 응답 원본 |
| executed_at | TIMESTAMPTZ | X | - | 실행 시간 |
| created_at | TIMESTAMPTZ | O | NOW() | 생성 시간 |
| updated_at | TIMESTAMPTZ | O | NOW() | 수정 시간 |

**status 값**:
| 값 | 설명 |
|----|------|
| requested | 요청됨 |
| pending | 대기 중 |
| success | 성공 |
| failed | 실패 |

---

### 8. idempotency (멱등성)

API 멱등성 키 저장

```sql
CREATE TABLE IF NOT EXISTS idempotency (
  scope         TEXT NOT NULL,               -- API 스코프 (paycreate, proof 등)
  idem_key      TEXT NOT NULL,               -- 멱등성 키
  response_json JSONB NOT NULL,              -- 저장된 응답
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  expires_at    TIMESTAMPTZ NOT NULL,        -- 만료 시간

  PRIMARY KEY (scope, idem_key)
);

CREATE INDEX IF NOT EXISTS idx_idempotency_expires ON idempotency(expires_at);
```

| 컬럼 | 타입 | 필수 | 기본값 | 설명 |
|------|------|------|--------|------|
| scope | TEXT | O | - | PK (API 스코프) |
| idem_key | TEXT | O | - | PK (멱등성 키) |
| response_json | JSONB | O | - | 캐시된 응답 |
| created_at | TIMESTAMPTZ | O | NOW() | 생성 시간 |
| expires_at | TIMESTAMPTZ | O | - | 만료 시간 |

**정리 쿼리**:
```sql
DELETE FROM idempotency WHERE expires_at < NOW();
```

---

### 9. revoked_session (취소된 세션)

토스 연결 해제 시 무효화된 세션

```sql
CREATE TABLE IF NOT EXISTS revoked_session (
  user_sub    TEXT PRIMARY KEY,              -- 사용자 식별자 (toss:12345678)
  revoked_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  reason      TEXT DEFAULT 'unlink'          -- unlink | logout | admin
);

CREATE INDEX IF NOT EXISTS idx_revoked_session_time ON revoked_session(revoked_at);
```

| 컬럼 | 타입 | 필수 | 기본값 | 설명 |
|------|------|------|--------|------|
| user_sub | TEXT | O | - | PK (사용자 식별자) |
| revoked_at | TIMESTAMPTZ | O | NOW() | 취소 시간 |
| reason | TEXT | X | 'unlink' | 취소 사유 |

---

## 전체 마이그레이션 SQL

```sql
-- 001_init.sql (v1.0)
-- 습관환급 (Habit Cashback) DB 스키마

-- 1. 사용자
CREATE TABLE IF NOT EXISTS app_user (
  id            BIGSERIAL PRIMARY KEY,
  toss_user_key TEXT NOT NULL UNIQUE,
  status        TEXT NOT NULL DEFAULT 'active',
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_app_user_status ON app_user(status);

-- 2. 챌린지
CREATE TABLE IF NOT EXISTS challenge (
  id          TEXT PRIMARY KEY,
  title       TEXT NOT NULL,
  description TEXT,
  days        INT NOT NULL DEFAULT 3,
  deposit     BIGINT NOT NULL DEFAULT 10000,
  proof_type  TEXT NOT NULL DEFAULT 'photo',
  is_active   BOOLEAN NOT NULL DEFAULT true,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_challenge_active ON challenge(is_active);

-- 초기 챌린지 데이터
INSERT INTO challenge (id, title, days, deposit, proof_type) VALUES
  ('walk-7000', '매일 7,000보 걷기', 3, 10000, 'steps'),
  ('bed-0700', '아침 7시 이불 개기', 3, 10000, 'photo'),
  ('lunch-proof', '점심 도시락/샐러드 인증', 3, 10000, 'photo')
ON CONFLICT (id) DO NOTHING;

-- 3. 결제
CREATE TABLE IF NOT EXISTS payment (
  id           BIGSERIAL PRIMARY KEY,
  user_id      BIGINT NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  challenge_id TEXT NOT NULL REFERENCES challenge(id) ON DELETE RESTRICT,
  order_no     TEXT NOT NULL UNIQUE,
  pay_token    TEXT,
  amount       BIGINT NOT NULL DEFAULT 0,
  status       TEXT NOT NULL DEFAULT 'created',
  pg_tx_id     TEXT,
  raw_json     JSONB,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_payment_user_created ON payment(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_order_no ON payment(order_no);
CREATE INDEX IF NOT EXISTS idx_payment_status ON payment(status);

-- 4. 참여
CREATE TABLE IF NOT EXISTS participation (
  id            BIGSERIAL PRIMARY KEY,
  user_id       BIGINT NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  challenge_id  TEXT NOT NULL REFERENCES challenge(id) ON DELETE RESTRICT,
  payment_id    BIGINT REFERENCES payment(id),
  status        TEXT NOT NULL DEFAULT 'pending',
  start_date    DATE,
  end_date      DATE,
  proof_count   INT NOT NULL DEFAULT 0,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, challenge_id, start_date)
);
CREATE INDEX IF NOT EXISTS idx_participation_user ON participation(user_id);
CREATE INDEX IF NOT EXISTS idx_participation_status ON participation(status);
CREATE INDEX IF NOT EXISTS idx_participation_end_date ON participation(end_date) WHERE status = 'active';

-- 5. 인증
CREATE TABLE IF NOT EXISTS proof (
  id               BIGSERIAL PRIMARY KEY,
  participation_id BIGINT NOT NULL REFERENCES participation(id) ON DELETE CASCADE,
  user_id          BIGINT NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  challenge_id     TEXT NOT NULL REFERENCES challenge(id) ON DELETE RESTRICT,
  proof_date       DATE NOT NULL,
  proof_type       TEXT NOT NULL,
  image_hash       TEXT,
  image_url        TEXT,
  exif_timestamp   TIMESTAMPTZ,
  steps_count      INT,
  status           TEXT NOT NULL DEFAULT 'pending',
  reject_reason    TEXT,
  verified_at      TIMESTAMPTZ,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(participation_id, proof_date)
);
CREATE INDEX IF NOT EXISTS idx_proof_user ON proof(user_id);
CREATE INDEX IF NOT EXISTS idx_proof_participation ON proof(participation_id);
CREATE INDEX IF NOT EXISTS idx_proof_status ON proof(status);
CREATE INDEX IF NOT EXISTS idx_proof_image_hash ON proof(image_hash) WHERE image_hash IS NOT NULL;

-- 6. 지급
CREATE TABLE IF NOT EXISTS payout (
  id              BIGSERIAL PRIMARY KEY,
  user_id         BIGINT NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  promotion_code  TEXT NOT NULL,
  promotion_key   TEXT NOT NULL UNIQUE,
  amount_points   BIGINT NOT NULL,
  status          TEXT NOT NULL DEFAULT 'requested',
  error_code      TEXT,
  error_message   TEXT,
  raw_json        JSONB,
  executed_at     TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_payout_user ON payout(user_id);
CREATE INDEX IF NOT EXISTS idx_payout_status ON payout(status);
CREATE INDEX IF NOT EXISTS idx_payout_status_updated ON payout(status, updated_at ASC);

-- 7. 정산
CREATE TABLE IF NOT EXISTS settlement (
  id               BIGSERIAL PRIMARY KEY,
  participation_id BIGINT NOT NULL UNIQUE REFERENCES participation(id) ON DELETE CASCADE,
  user_id          BIGINT NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  challenge_id     TEXT NOT NULL REFERENCES challenge(id) ON DELETE RESTRICT,
  payment_id       BIGINT REFERENCES payment(id),
  status           TEXT NOT NULL DEFAULT 'running',
  refundable       BOOLEAN NOT NULL DEFAULT false,
  deposit_amount   BIGINT NOT NULL DEFAULT 0,
  reward_amount    BIGINT NOT NULL DEFAULT 0,
  payout_id        BIGINT REFERENCES payout(id),
  settled_at       TIMESTAMPTZ,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_settlement_user ON settlement(user_id);
CREATE INDEX IF NOT EXISTS idx_settlement_status ON settlement(status);
CREATE INDEX IF NOT EXISTS idx_settlement_participation ON settlement(participation_id);

-- 8. 멱등성
CREATE TABLE IF NOT EXISTS idempotency (
  scope         TEXT NOT NULL,
  idem_key      TEXT NOT NULL,
  response_json JSONB NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  expires_at    TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (scope, idem_key)
);
CREATE INDEX IF NOT EXISTS idx_idempotency_expires ON idempotency(expires_at);

-- 9. 취소된 세션
CREATE TABLE IF NOT EXISTS revoked_session (
  user_sub    TEXT PRIMARY KEY,
  revoked_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  reason      TEXT DEFAULT 'unlink'
);
CREATE INDEX IF NOT EXISTS idx_revoked_session_time ON revoked_session(revoked_at);
```

---

## 데이터 흐름

### 1. 챌린지 참여 흐름

```
1. 결제 생성 (payment: created)
       │
       ▼
2. 결제 실행 (payment: done)
       │
       ▼
3. 참여 생성 (participation: active)
       │
       ▼
4. 정산 생성 (settlement: running)
```

### 2. 인증 제출 흐름

```
1. 인증 제출 (proof: pending)
       │
       ▼
2. 검증 (EXIF, 중복 해시)
       │
   ┌───┴───┐
   ▼       ▼
approved  rejected
   │
   ▼
3. proof_count 증가
```

### 3. 정산 완료 흐름

```
1. 챌린지 기간 종료
       │
       ▼
2. proof_count 검증
       │
   ┌───┴───┐
   ▼       ▼
성공      실패
   │       │
   ▼       ▼
settlement  settlement
: success   : failed
   │
   ▼
3. payout 생성 (리워드 지급)
```

---

## 인덱스 요약

| 테이블 | 인덱스 | 용도 |
|--------|--------|------|
| app_user | idx_app_user_status | 상태별 조회 |
| challenge | idx_challenge_active | 활성 챌린지 조회 |
| participation | idx_participation_user | 사용자별 참여 조회 |
| participation | idx_participation_status | 상태별 조회 |
| participation | idx_participation_end_date | 종료 예정 조회 (배치) |
| payment | idx_payment_user_created | 사용자별 결제 내역 |
| payment | idx_payment_order_no | 주문번호 조회 |
| payment | idx_payment_status | 상태별 조회 |
| proof | idx_proof_user | 사용자별 인증 조회 |
| proof | idx_proof_participation | 참여별 인증 조회 |
| proof | idx_proof_status | 상태별 조회 |
| proof | idx_proof_image_hash | 중복 이미지 검색 |
| settlement | idx_settlement_user | 사용자별 정산 조회 |
| settlement | idx_settlement_status | 상태별 조회 |
| payout | idx_payout_user | 사용자별 지급 조회 |
| payout | idx_payout_status | 상태별 조회 |
| payout | idx_payout_status_updated | 재시도 대상 조회 |
| idempotency | idx_idempotency_expires | 만료 데이터 정리 |
| revoked_session | idx_revoked_session_time | 시간순 조회 |

---

## 배치 작업

### 1. 멱등성 데이터 정리

```sql
-- 매 시간 실행
DELETE FROM idempotency WHERE expires_at < NOW();
```

### 2. 챌린지 종료 처리

```sql
-- 매일 자정 실행
UPDATE participation
SET status = 'completed', updated_at = NOW()
WHERE status = 'active'
  AND end_date < CURRENT_DATE
  AND proof_count >= (SELECT days FROM challenge WHERE id = participation.challenge_id);

UPDATE participation
SET status = 'failed', updated_at = NOW()
WHERE status = 'active'
  AND end_date < CURRENT_DATE
  AND proof_count < (SELECT days FROM challenge WHERE id = participation.challenge_id);
```

### 3. 정산 상태 업데이트

```sql
-- 매일 자정 실행
UPDATE settlement s
SET status = 'success', refundable = true, settled_at = NOW(), updated_at = NOW()
FROM participation p
WHERE s.participation_id = p.id
  AND s.status = 'running'
  AND p.status = 'completed';

UPDATE settlement s
SET status = 'failed', refundable = false, settled_at = NOW(), updated_at = NOW()
FROM participation p
WHERE s.participation_id = p.id
  AND s.status = 'running'
  AND p.status = 'failed';
```
