package store

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provides database operations
type Store struct {
	pool *pgxpool.Pool
}

// New creates a new Store with connection pool
func New(ctx context.Context) (*Store, error) {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &Store{pool: pool}, nil
}

// Close closes the connection pool
func (s *Store) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// ============ User Operations ============

type User struct {
	ID          int64
	TossUserKey string
	Status      string
	CreatedAt   time.Time
}

// GetOrCreateUser finds or creates a user by toss_user_key
func (s *Store) GetOrCreateUser(ctx context.Context, tossUserKey string) (*User, error) {
	const q = `
		INSERT INTO app_user (toss_user_key)
		VALUES ($1)
		ON CONFLICT (toss_user_key) DO UPDATE SET updated_at = NOW()
		RETURNING id, toss_user_key, status, created_at
	`
	var u User
	err := s.pool.QueryRow(ctx, q, tossUserKey).Scan(&u.ID, &u.TossUserKey, &u.Status, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get or create user: %w", err)
	}
	return &u, nil
}

// GetUserByTossKey finds a user by toss_user_key
func (s *Store) GetUserByTossKey(ctx context.Context, tossUserKey string) (*User, error) {
	const q = `SELECT id, toss_user_key, status, created_at FROM app_user WHERE toss_user_key = $1`
	var u User
	err := s.pool.QueryRow(ctx, q, tossUserKey).Scan(&u.ID, &u.TossUserKey, &u.Status, &u.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &u, nil
}

// ============ Challenge Operations ============

type Challenge struct {
	ID        string
	Title     string
	Days      int
	Deposit   int64
	ProofType string
	IsActive  bool
}

// ListChallenges returns all active challenges
func (s *Store) ListChallenges(ctx context.Context) ([]Challenge, error) {
	const q = `SELECT id, title, days, deposit, proof_type, is_active FROM challenge WHERE is_active = true ORDER BY id`
	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list challenges: %w", err)
	}
	defer rows.Close()

	var list []Challenge
	for rows.Next() {
		var c Challenge
		if err := rows.Scan(&c.ID, &c.Title, &c.Days, &c.Deposit, &c.ProofType, &c.IsActive); err != nil {
			return nil, fmt.Errorf("scan challenge: %w", err)
		}
		list = append(list, c)
	}
	return list, nil
}

// GetChallenge returns a challenge by ID
func (s *Store) GetChallenge(ctx context.Context, id string) (*Challenge, error) {
	const q = `SELECT id, title, days, deposit, proof_type, is_active FROM challenge WHERE id = $1`
	var c Challenge
	err := s.pool.QueryRow(ctx, q, id).Scan(&c.ID, &c.Title, &c.Days, &c.Deposit, &c.ProofType, &c.IsActive)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get challenge: %w", err)
	}
	return &c, nil
}

// ============ Participation Operations ============

type Participation struct {
	ID          int64
	UserID      int64
	ChallengeID string
	PaymentID   int64
	Status      string
	StartDate   time.Time
	EndDate     time.Time
	ProofCount  int
	CreatedAt   time.Time
}

// GetActiveParticipation returns the active participation for a user and challenge
func (s *Store) GetActiveParticipation(ctx context.Context, userID int64, challengeID string) (*Participation, error) {
	today := time.Now().Truncate(24 * time.Hour)
	const q = `
		SELECT id, user_id, challenge_id, payment_id, status, start_date, end_date, proof_count, created_at
		FROM participation
		WHERE user_id = $1 AND challenge_id = $2 AND status = 'active'
		AND start_date <= $3 AND end_date >= $3
		LIMIT 1
	`
	var p Participation
	err := s.pool.QueryRow(ctx, q, userID, challengeID, today).
		Scan(&p.ID, &p.UserID, &p.ChallengeID, &p.PaymentID, &p.Status, &p.StartDate, &p.EndDate, &p.ProofCount, &p.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get active participation: %w", err)
	}
	return &p, nil
}

// ============ Payment Operations ============

type Payment struct {
	ID          int64
	UserID      int64
	ChallengeID string
	OrderNo     string
	PayToken    string
	Amount      int64
	Status      string
	CreatedAt   time.Time
}

// CreatePayment creates a new payment record
func (s *Store) CreatePayment(ctx context.Context, userID int64, challengeID, orderNo string, amount int64) (*Payment, error) {
	const q = `
		INSERT INTO payment (user_id, challenge_id, order_no, amount, status)
		VALUES ($1, $2, $3, $4, 'created')
		RETURNING id, user_id, challenge_id, order_no, amount, status, created_at
	`
	var p Payment
	err := s.pool.QueryRow(ctx, q, userID, challengeID, orderNo, amount).
		Scan(&p.ID, &p.UserID, &p.ChallengeID, &p.OrderNo, &p.Amount, &p.Status, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create payment: %w", err)
	}
	return &p, nil
}

// ExecutePayment updates payment status to 'done' and creates participation
func (s *Store) ExecutePayment(ctx context.Context, orderNo string) (*Payment, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update payment status
	const updateQ = `
		UPDATE payment SET status = 'done', updated_at = NOW()
		WHERE order_no = $1 AND status = 'created'
		RETURNING id, user_id, challenge_id, order_no, amount, status, created_at
	`
	var p Payment
	err = tx.QueryRow(ctx, updateQ, orderNo).
		Scan(&p.ID, &p.UserID, &p.ChallengeID, &p.OrderNo, &p.Amount, &p.Status, &p.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("payment not found or already executed")
	}
	if err != nil {
		return nil, fmt.Errorf("update payment: %w", err)
	}

	// Get challenge days
	var days int
	err = tx.QueryRow(ctx, `SELECT days FROM challenge WHERE id = $1`, p.ChallengeID).Scan(&days)
	if err != nil {
		return nil, fmt.Errorf("get challenge days: %w", err)
	}

	// Create participation
	startDate := time.Now().Truncate(24 * time.Hour)
	endDate := startDate.AddDate(0, 0, days-1)
	const partQ = `
		INSERT INTO participation (user_id, challenge_id, payment_id, status, start_date, end_date)
		VALUES ($1, $2, $3, 'active', $4, $5)
		ON CONFLICT (user_id, challenge_id, start_date) DO NOTHING
		RETURNING id
	`
	var partID int64
	err = tx.QueryRow(ctx, partQ, p.UserID, p.ChallengeID, p.ID, startDate, endDate).Scan(&partID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("create participation: %w", err)
	}

	// Create settlement record
	if partID > 0 {
		const settQ = `
			INSERT INTO settlement (participation_id, user_id, challenge_id, payment_id, status, deposit_amount)
			VALUES ($1, $2, $3, $4, 'running', $5)
			ON CONFLICT (participation_id) DO NOTHING
		`
		_, err = tx.Exec(ctx, settQ, partID, p.UserID, p.ChallengeID, p.ID, p.Amount)
		if err != nil {
			return nil, fmt.Errorf("create settlement: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &p, nil
}

// GetPaymentByOrderNo returns a payment by order number
func (s *Store) GetPaymentByOrderNo(ctx context.Context, orderNo string) (*Payment, error) {
	const q = `SELECT id, user_id, challenge_id, order_no, amount, status, created_at FROM payment WHERE order_no = $1`
	var p Payment
	err := s.pool.QueryRow(ctx, q, orderNo).
		Scan(&p.ID, &p.UserID, &p.ChallengeID, &p.OrderNo, &p.Amount, &p.Status, &p.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get payment: %w", err)
	}
	return &p, nil
}

// ============ Proof Operations ============

type Proof struct {
	ID              int64
	ParticipationID int64
	UserID          int64
	ChallengeID     string
	ProofDate       time.Time
	ProofType       string
	ImageHash       string
	Status          string
	CreatedAt       time.Time
}

// CheckDuplicateProofHash checks if an image hash has been used before by any user
// Returns the userID and challengeID of the existing proof if found
func (s *Store) CheckDuplicateProofHash(ctx context.Context, imageHash string, excludeUserID int64) (*Proof, error) {
	const q = `
		SELECT id, participation_id, user_id, challenge_id, proof_date, proof_type, image_hash, status, created_at
		FROM proof
		WHERE image_hash = $1 AND user_id != $2 AND status = 'accepted'
		LIMIT 1
	`
	var p Proof
	err := s.pool.QueryRow(ctx, q, imageHash, excludeUserID).
		Scan(&p.ID, &p.ParticipationID, &p.UserID, &p.ChallengeID, &p.ProofDate, &p.ProofType, &p.ImageHash, &p.Status, &p.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil // No duplicate found
	}
	if err != nil {
		return nil, fmt.Errorf("check duplicate hash: %w", err)
	}
	return &p, nil
}

// CheckSameUserDuplicateHash checks if the same user has already used this image hash
func (s *Store) CheckSameUserDuplicateHash(ctx context.Context, imageHash string, userID int64) (*Proof, error) {
	const q = `
		SELECT id, participation_id, user_id, challenge_id, proof_date, proof_type, image_hash, status, created_at
		FROM proof
		WHERE image_hash = $1 AND user_id = $2 AND status = 'accepted'
		LIMIT 1
	`
	var p Proof
	err := s.pool.QueryRow(ctx, q, imageHash, userID).
		Scan(&p.ID, &p.ParticipationID, &p.UserID, &p.ChallengeID, &p.ProofDate, &p.ProofType, &p.ImageHash, &p.Status, &p.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil // No duplicate found
	}
	if err != nil {
		return nil, fmt.Errorf("check same user duplicate: %w", err)
	}
	return &p, nil
}

// SubmitProof creates a new proof record
func (s *Store) SubmitProof(ctx context.Context, userID int64, challengeID, proofType, imageHash string) (*Proof, error) {
	// Find active participation
	today := time.Now().Truncate(24 * time.Hour)
	const partQ = `
		SELECT id FROM participation
		WHERE user_id = $1 AND challenge_id = $2 AND status = 'active'
		AND start_date <= $3 AND end_date >= $3
		LIMIT 1
	`
	var partID int64
	err := s.pool.QueryRow(ctx, partQ, userID, challengeID, today).Scan(&partID)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("no active participation found")
	}
	if err != nil {
		return nil, fmt.Errorf("find participation: %w", err)
	}

	// Create proof
	const proofQ = `
		INSERT INTO proof (participation_id, user_id, challenge_id, proof_date, proof_type, image_hash, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'accepted')
		ON CONFLICT (participation_id, proof_date) DO UPDATE SET
			image_hash = EXCLUDED.image_hash,
			status = 'accepted',
			created_at = NOW()
		RETURNING id, participation_id, user_id, challenge_id, proof_date, proof_type, image_hash, status, created_at
	`
	var p Proof
	err = s.pool.QueryRow(ctx, proofQ, partID, userID, challengeID, today, proofType, imageHash).
		Scan(&p.ID, &p.ParticipationID, &p.UserID, &p.ChallengeID, &p.ProofDate, &p.ProofType, &p.ImageHash, &p.Status, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create proof: %w", err)
	}

	// Update participation proof count
	const updateQ = `
		UPDATE participation SET
			proof_count = (SELECT COUNT(*) FROM proof WHERE participation_id = $1 AND status = 'accepted'),
			updated_at = NOW()
		WHERE id = $1
	`
	_, err = s.pool.Exec(ctx, updateQ, partID)
	if err != nil {
		return nil, fmt.Errorf("update proof count: %w", err)
	}

	return &p, nil
}

// ============ Settlement Operations ============

type Settlement struct {
	ID            int64
	UserID        int64
	ChallengeID   string
	Status        string
	Refundable    bool
	DepositAmount int64
	RewardAmount  int64
	Message       string
	CreatedAt     time.Time
}

// ListSettlementsByUser returns all settlements for a user
func (s *Store) ListSettlementsByUser(ctx context.Context, userID int64) ([]Settlement, error) {
	const q = `
		SELECT s.id, s.user_id, s.challenge_id, s.status, s.refundable, s.deposit_amount, s.reward_amount, s.created_at,
		       p.proof_count, c.days
		FROM settlement s
		JOIN participation p ON s.participation_id = p.id
		JOIN challenge c ON s.challenge_id = c.id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
	`
	rows, err := s.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list settlements: %w", err)
	}
	defer rows.Close()

	var list []Settlement
	for rows.Next() {
		var sett Settlement
		var proofCount, days int
		if err := rows.Scan(&sett.ID, &sett.UserID, &sett.ChallengeID, &sett.Status, &sett.Refundable,
			&sett.DepositAmount, &sett.RewardAmount, &sett.CreatedAt, &proofCount, &days); err != nil {
			return nil, fmt.Errorf("scan settlement: %w", err)
		}

		// Generate message based on status
		switch sett.Status {
		case "running":
			sett.Message = fmt.Sprintf("진행중 (%d/%d일 완료)", proofCount, days)
		case "success":
			sett.Message = "성공! 환급 예정"
		case "failed":
			sett.Message = "미완료"
		}

		list = append(list, sett)
	}
	return list, nil
}

// ============ Idempotency Operations ============

// CheckIdempotency checks if an idempotency key exists
func (s *Store) CheckIdempotency(ctx context.Context, scope, key string) (bool, error) {
	const q = `SELECT 1 FROM idempotency WHERE scope = $1 AND idem_key = $2 AND expires_at > NOW()`
	var exists int
	err := s.pool.QueryRow(ctx, q, scope, key).Scan(&exists)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check idempotency: %w", err)
	}
	return true, nil
}

// SetIdempotency stores an idempotency key
func (s *Store) SetIdempotency(ctx context.Context, scope, key string, ttl time.Duration) error {
	const q = `
		INSERT INTO idempotency (scope, idem_key, response_json, expires_at)
		VALUES ($1, $2, '{}', $3)
		ON CONFLICT (scope, idem_key) DO NOTHING
	`
	_, err := s.pool.Exec(ctx, q, scope, key, time.Now().Add(ttl))
	if err != nil {
		return fmt.Errorf("set idempotency: %w", err)
	}
	return nil
}

// ============ Revoked Session Operations ============

// RevokeSession marks a session as revoked
func (s *Store) RevokeSession(ctx context.Context, userSub, reason string) error {
	const q = `
		INSERT INTO revoked_session (user_sub, reason)
		VALUES ($1, $2)
		ON CONFLICT (user_sub) DO UPDATE SET revoked_at = NOW(), reason = EXCLUDED.reason
	`
	_, err := s.pool.Exec(ctx, q, userSub, reason)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

// IsSessionRevoked checks if a session is revoked
func (s *Store) IsSessionRevoked(ctx context.Context, userSub string) (bool, error) {
	const q = `SELECT 1 FROM revoked_session WHERE user_sub = $1`
	var exists int
	err := s.pool.QueryRow(ctx, q, userSub).Scan(&exists)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check revoked: %w", err)
	}
	return true, nil
}

// ============ Batch Job Operations ============

// BatchResult holds the result of a batch operation
type BatchResult struct {
	Processed int
	Failed    int
	Errors    []string
}

// CloseExpiredParticipations marks participations as completed or failed based on end_date
// Returns the number of participations processed
func (s *Store) CloseExpiredParticipations(ctx context.Context) (*BatchResult, error) {
	today := time.Now().Truncate(24 * time.Hour)
	result := &BatchResult{Errors: []string{}}

	// Find all active participations that have ended
	const findQ = `
		SELECT p.id, p.user_id, p.challenge_id, p.proof_count, c.days
		FROM participation p
		JOIN challenge c ON p.challenge_id = c.id
		WHERE p.status = 'active' AND p.end_date < $1
	`
	rows, err := s.pool.Query(ctx, findQ, today)
	if err != nil {
		return nil, fmt.Errorf("find expired participations: %w", err)
	}
	defer rows.Close()

	type expiredPart struct {
		ID          int64
		UserID      int64
		ChallengeID string
		ProofCount  int
		Days        int
	}
	var expired []expiredPart
	for rows.Next() {
		var ep expiredPart
		if err := rows.Scan(&ep.ID, &ep.UserID, &ep.ChallengeID, &ep.ProofCount, &ep.Days); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("scan: %v", err))
			continue
		}
		expired = append(expired, ep)
	}

	// Update each participation
	for _, ep := range expired {
		newStatus := "failed"
		if ep.ProofCount >= ep.Days {
			newStatus = "success"
		}

		const updateQ = `UPDATE participation SET status = $1, updated_at = NOW() WHERE id = $2`
		_, err := s.pool.Exec(ctx, updateQ, newStatus, ep.ID)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("update participation %d: %v", ep.ID, err))
			continue
		}
		result.Processed++
	}

	return result, nil
}

// UpdateSettlementStatuses updates settlement records based on participation status
func (s *Store) UpdateSettlementStatuses(ctx context.Context) (*BatchResult, error) {
	result := &BatchResult{Errors: []string{}}

	// Update settlements based on participation status
	const updateQ = `
		UPDATE settlement s
		SET
			status = p.status,
			refundable = (p.status = 'success'),
			updated_at = NOW()
		FROM participation p
		WHERE s.participation_id = p.id
		AND s.status = 'running'
		AND p.status IN ('success', 'failed')
	`
	tag, err := s.pool.Exec(ctx, updateQ)
	if err != nil {
		return nil, fmt.Errorf("update settlements: %w", err)
	}

	result.Processed = int(tag.RowsAffected())
	return result, nil
}

// CleanupExpiredIdempotencyKeys deletes expired idempotency keys
func (s *Store) CleanupExpiredIdempotencyKeys(ctx context.Context) (*BatchResult, error) {
	result := &BatchResult{Errors: []string{}}

	const deleteQ = `DELETE FROM idempotency WHERE expires_at < NOW()`
	tag, err := s.pool.Exec(ctx, deleteQ)
	if err != nil {
		return nil, fmt.Errorf("cleanup idempotency: %w", err)
	}

	result.Processed = int(tag.RowsAffected())
	return result, nil
}

// CleanupOldRevokedSessions deletes revoked sessions older than 30 days
func (s *Store) CleanupOldRevokedSessions(ctx context.Context) (*BatchResult, error) {
	result := &BatchResult{Errors: []string{}}

	cutoff := time.Now().Add(-30 * 24 * time.Hour)
	const deleteQ = `DELETE FROM revoked_session WHERE revoked_at < $1`
	tag, err := s.pool.Exec(ctx, deleteQ, cutoff)
	if err != nil {
		return nil, fmt.Errorf("cleanup revoked sessions: %w", err)
	}

	result.Processed = int(tag.RowsAffected())
	return result, nil
}

// GetBatchStats returns statistics for batch job monitoring
type BatchStats struct {
	ActiveParticipations   int
	RunningSettlements     int
	PendingIdempotencyKeys int
	RevokedSessions        int
}

func (s *Store) GetBatchStats(ctx context.Context) (*BatchStats, error) {
	stats := &BatchStats{}

	// Active participations
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM participation WHERE status = 'active'`).Scan(&stats.ActiveParticipations)
	if err != nil {
		return nil, fmt.Errorf("count active participations: %w", err)
	}

	// Running settlements
	err = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM settlement WHERE status = 'running'`).Scan(&stats.RunningSettlements)
	if err != nil {
		return nil, fmt.Errorf("count running settlements: %w", err)
	}

	// Pending idempotency keys
	err = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM idempotency WHERE expires_at > NOW()`).Scan(&stats.PendingIdempotencyKeys)
	if err != nil {
		return nil, fmt.Errorf("count idempotency keys: %w", err)
	}

	// Revoked sessions
	err = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM revoked_session`).Scan(&stats.RevokedSessions)
	if err != nil {
		return nil, fmt.Errorf("count revoked sessions: %w", err)
	}

	return stats, nil
}
