-- habitcashback - minimal schema (v0.7)
-- NOTE: This is a pragmatic MVP schema; evolve as needed.

CREATE TABLE IF NOT EXISTS app_user (
  id            BIGSERIAL PRIMARY KEY,
  toss_user_key TEXT NOT NULL UNIQUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS idempotency (
  scope         TEXT NOT NULL,
  idem_key      TEXT NOT NULL,
  response_json JSONB NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (scope, idem_key)
);

CREATE TABLE IF NOT EXISTS payment (
  id         BIGSERIAL PRIMARY KEY,
  user_id    BIGINT NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  order_no   TEXT NOT NULL,
  pay_token  TEXT NOT NULL,
  amount     BIGINT NOT NULL DEFAULT 0,
  status     TEXT NOT NULL,
  raw_json   JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_user_created ON payment(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_order_no ON payment(order_no);

CREATE TABLE IF NOT EXISTS payout (
  id             BIGSERIAL PRIMARY KEY,
  user_id         BIGINT NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  promotion_code  TEXT NOT NULL,
  promotion_key   TEXT NOT NULL UNIQUE,
  amount_points   BIGINT NOT NULL,
  status          TEXT NOT NULL, -- REQUESTED | SUCCESS | FAIL | PENDING
  raw_json        JSONB,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payout_status_updated ON payout(status, updated_at ASC);
