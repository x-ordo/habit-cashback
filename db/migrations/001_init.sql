-- 습관환급 (Habit Cashback) DB 스키마 v1.0
-- PostgreSQL 14+

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
