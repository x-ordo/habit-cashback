package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"habitcashback/internal/store"
)

type Postgres struct {
	db *sql.DB
}

func Open(dsn string) (*Postgres, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	return &Postgres{db: db}, nil
}

func (p *Postgres) Ping(ctx context.Context) error { return p.db.PingContext(ctx) }
func (p *Postgres) Close() error                  { return p.db.Close() }

func (p *Postgres) UpsertUser(ctx context.Context, tossUserKey string) (int64, error) {
	var id int64
	err := p.db.QueryRowContext(ctx, `
		INSERT INTO app_user (toss_user_key) VALUES ($1)
		ON CONFLICT (toss_user_key) DO UPDATE SET toss_user_key = EXCLUDED.toss_user_key
		RETURNING id
	`, tossUserKey).Scan(&id)
	return id, err
}

func (p *Postgres) GetIdempotency(ctx context.Context, scope, key string) (bool, []byte, error) {
	var resp []byte
	err := p.db.QueryRowContext(ctx, `
		SELECT response_json
		FROM idempotency
		WHERE scope = $1 AND idem_key = $2
	`, scope, key).Scan(&resp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, resp, nil
}

func (p *Postgres) PutIdempotency(ctx context.Context, scope, key string, resp []byte) error {
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO idempotency (scope, idem_key, response_json) VALUES ($1, $2, $3)
		ON CONFLICT (scope, idem_key) DO NOTHING
	`, scope, key, resp)
	return err
}

func (p *Postgres) InsertPayment(ctx context.Context, pay store.Payment) error {
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO payment (user_id, order_no, pay_token, amount, status, raw_json)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, pay.UserID, pay.OrderNo, pay.PayToken, pay.Amount, pay.Status, pay.RawJSON)
	return err
}

func (p *Postgres) InsertPayout(ctx context.Context, po store.Payout) error {
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO payout (user_id, promotion_code, promotion_key, amount_points, status, raw_json)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, po.UserID, po.PromotionCode, po.PromotionKey, po.AmountPoints, po.Status, po.RawJSON)
	return err
}

func (p *Postgres) UpdatePayoutStatus(ctx context.Context, promotionKey string, status string, raw []byte) error {
	_, err := p.db.ExecContext(ctx, `
		UPDATE payout
		SET status=$2, raw_json=$3, updated_at=NOW()
		WHERE promotion_key=$1
	`, promotionKey, status, raw)
	return err
}

func (p *Postgres) ListPendingPayouts(ctx context.Context, limit int) ([]store.Payout, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT p.id, p.user_id, u.toss_user_key, p.promotion_code, p.promotion_key, p.amount_points, p.status, p.raw_json, p.updated_at, p.created_at
		FROM payout p
		JOIN app_user u ON u.id = p.user_id
		WHERE status IN ('REQUESTED','PENDING')
		ORDER BY updated_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []store.Payout
	for rows.Next() {
		var po store.Payout
		err = rows.Scan(&po.ID, &po.UserID, &po.TossUserKey, &po.PromotionCode, &po.PromotionKey, &po.AmountPoints, &po.Status, &po.RawJSON, &po.UpdatedAt, &po.CreatedAt)
		if err != nil {
			return nil, err
		}
		out = append(out, po)
	}
	return out, rows.Err()
}
