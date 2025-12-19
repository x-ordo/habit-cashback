package store

import (
	"context"
	"time"
)

type IdempotencyRecord struct {
	Scope    string
	Key      string
	Response []byte
	CreatedAt time.Time
}

type Payment struct {
	ID        int64
	UserID    int64
	OrderNo   string
	PayToken  string
	Amount    int64
	Status    string
	RawJSON   []byte
	CreatedAt time.Time
}

type Payout struct {
	ID            int64
	UserID        int64
	TossUserKey   string
	PromotionCode string
	PromotionKey  string
	AmountPoints  int64
	Status        string // REQUESTED | SUCCESS | FAIL | PENDING
	RawJSON       []byte
	UpdatedAt     time.Time
	CreatedAt     time.Time
}

type Store interface {
	Ping(ctx context.Context) error
	Close() error

	UpsertUser(ctx context.Context, tossUserKey string) (userID int64, err error)

	// Idempotency: if already exists, returns (true, existingResponse)
	GetIdempotency(ctx context.Context, scope, key string) (found bool, resp []byte, err error)
	PutIdempotency(ctx context.Context, scope, key string, resp []byte) error

	InsertPayment(ctx context.Context, p Payment) error
	InsertPayout(ctx context.Context, p Payout) error
	UpdatePayoutStatus(ctx context.Context, promotionKey string, status string, raw []byte) error
	ListPendingPayouts(ctx context.Context, limit int) ([]Payout, error)
}
