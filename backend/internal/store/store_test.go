package store

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestNew_MissingDatabaseURL tests that New returns an error when DATABASE_URL is not set
func TestNew_MissingDatabaseURL(t *testing.T) {
	// Save and clear DATABASE_URL
	original := os.Getenv("DATABASE_URL")
	defer os.Setenv("DATABASE_URL", original)
	os.Setenv("DATABASE_URL", "")

	ctx := context.Background()
	_, err := New(ctx)

	if err == nil {
		t.Fatal("expected error when DATABASE_URL is not set")
	}

	expectedMsg := "DATABASE_URL not set"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestNew_InvalidDSN tests that New returns an error for invalid DSN
func TestNew_InvalidDSN(t *testing.T) {
	// Save original value
	original := os.Getenv("DATABASE_URL")
	defer os.Setenv("DATABASE_URL", original)

	// Set an invalid DSN
	os.Setenv("DATABASE_URL", "not-a-valid-dsn")

	ctx := context.Background()
	_, err := New(ctx)

	if err == nil {
		t.Fatal("expected error for invalid DSN")
	}
}

// TestUser struct validation
func TestUser_Fields(t *testing.T) {
	u := User{
		ID:          1,
		TossUserKey: "test-key",
		Status:      "active",
		CreatedAt:   time.Now(),
	}

	if u.ID != 1 {
		t.Errorf("expected ID 1, got %d", u.ID)
	}
	if u.TossUserKey != "test-key" {
		t.Errorf("expected TossUserKey 'test-key', got '%s'", u.TossUserKey)
	}
	if u.Status != "active" {
		t.Errorf("expected Status 'active', got '%s'", u.Status)
	}
}

// TestChallenge struct validation
func TestChallenge_Fields(t *testing.T) {
	c := Challenge{
		ID:        "morning-water",
		Title:     "아침 물 마시기",
		Days:      7,
		Deposit:   10000,
		ProofType: "photo",
		IsActive:  true,
	}

	if c.ID != "morning-water" {
		t.Errorf("expected ID 'morning-water', got '%s'", c.ID)
	}
	if c.Days != 7 {
		t.Errorf("expected Days 7, got %d", c.Days)
	}
	if c.Deposit != 10000 {
		t.Errorf("expected Deposit 10000, got %d", c.Deposit)
	}
}

// TestPayment struct validation
func TestPayment_Fields(t *testing.T) {
	p := Payment{
		ID:          1,
		UserID:      100,
		ChallengeID: "morning-water",
		OrderNo:     "ORD-001",
		PayToken:    "mock_pt_12345",
		Amount:      10000,
		Status:      "created",
		CreatedAt:   time.Now(),
	}

	if p.ID != 1 {
		t.Errorf("expected ID 1, got %d", p.ID)
	}
	if p.UserID != 100 {
		t.Errorf("expected UserID 100, got %d", p.UserID)
	}
	if p.OrderNo != "ORD-001" {
		t.Errorf("expected OrderNo 'ORD-001', got '%s'", p.OrderNo)
	}
	if p.PayToken != "mock_pt_12345" {
		t.Errorf("expected PayToken 'mock_pt_12345', got '%s'", p.PayToken)
	}
}

// TestParticipation struct validation
func TestParticipation_Fields(t *testing.T) {
	now := time.Now()
	p := Participation{
		ID:          1,
		UserID:      100,
		ChallengeID: "morning-water",
		PaymentID:   200,
		Status:      "active",
		StartDate:   now,
		EndDate:     now.AddDate(0, 0, 6),
		ProofCount:  3,
		CreatedAt:   now,
	}

	if p.Status != "active" {
		t.Errorf("expected Status 'active', got '%s'", p.Status)
	}
	if p.ProofCount != 3 {
		t.Errorf("expected ProofCount 3, got %d", p.ProofCount)
	}
}

// TestProof struct validation
func TestProof_Fields(t *testing.T) {
	now := time.Now()
	p := Proof{
		ID:              1,
		ParticipationID: 100,
		UserID:          50,
		ChallengeID:     "morning-water",
		ProofDate:       now,
		ProofType:       "photo",
		ImageHash:       "abc123def456",
		Status:          "accepted",
		CreatedAt:       now,
	}

	if p.ImageHash != "abc123def456" {
		t.Errorf("expected ImageHash 'abc123def456', got '%s'", p.ImageHash)
	}
	if p.Status != "accepted" {
		t.Errorf("expected Status 'accepted', got '%s'", p.Status)
	}
}

// TestSettlement struct validation
func TestSettlement_Fields(t *testing.T) {
	s := Settlement{
		ID:            1,
		UserID:        100,
		ChallengeID:   "morning-water",
		Status:        "success",
		Refundable:    true,
		DepositAmount: 10000,
		RewardAmount:  1000,
		Message:       "성공! 환급 예정",
		CreatedAt:     time.Now(),
	}

	if s.Status != "success" {
		t.Errorf("expected Status 'success', got '%s'", s.Status)
	}
	if !s.Refundable {
		t.Error("expected Refundable to be true")
	}
	if s.DepositAmount != 10000 {
		t.Errorf("expected DepositAmount 10000, got %d", s.DepositAmount)
	}
}

// TestBatchResult struct validation
func TestBatchResult_Fields(t *testing.T) {
	r := &BatchResult{
		Processed: 10,
		Failed:    2,
		Errors:    []string{"error1", "error2"},
	}

	if r.Processed != 10 {
		t.Errorf("expected Processed 10, got %d", r.Processed)
	}
	if r.Failed != 2 {
		t.Errorf("expected Failed 2, got %d", r.Failed)
	}
	if len(r.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(r.Errors))
	}
}

// TestBatchStats struct validation
func TestBatchStats_Fields(t *testing.T) {
	s := &BatchStats{
		ActiveParticipations:   5,
		RunningSettlements:     3,
		PendingIdempotencyKeys: 100,
		RevokedSessions:        2,
	}

	if s.ActiveParticipations != 5 {
		t.Errorf("expected ActiveParticipations 5, got %d", s.ActiveParticipations)
	}
	if s.RunningSettlements != 3 {
		t.Errorf("expected RunningSettlements 3, got %d", s.RunningSettlements)
	}
}

// Integration tests - These require a running PostgreSQL database
// Skip if DATABASE_URL is not set or we're in CI without a database

func skipIfNoDatabase(t *testing.T) *Store {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store, err := New(ctx)
	if err != nil {
		t.Skipf("Could not connect to database: %v", err)
	}

	return store
}

func TestIntegration_GetOrCreateUser(t *testing.T) {
	store := skipIfNoDatabase(t)
	defer store.Close()

	ctx := context.Background()
	testKey := "test-user-" + time.Now().Format("20060102150405")

	// First call should create user
	user1, err := store.GetOrCreateUser(ctx, testKey)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	if user1.TossUserKey != testKey {
		t.Errorf("expected toss_user_key '%s', got '%s'", testKey, user1.TossUserKey)
	}

	// Second call should return same user
	user2, err := store.GetOrCreateUser(ctx, testKey)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if user1.ID != user2.ID {
		t.Errorf("expected same user ID, got %d and %d", user1.ID, user2.ID)
	}
}

func TestIntegration_ListChallenges(t *testing.T) {
	store := skipIfNoDatabase(t)
	defer store.Close()

	ctx := context.Background()

	challenges, err := store.ListChallenges(ctx)
	if err != nil {
		t.Fatalf("failed to list challenges: %v", err)
	}

	// Should have at least one challenge if database is seeded
	if len(challenges) == 0 {
		t.Log("No challenges found - database may not be seeded")
	}

	// Verify challenge structure if we have any
	for _, c := range challenges {
		if c.ID == "" {
			t.Error("challenge ID should not be empty")
		}
		if c.Days <= 0 {
			t.Errorf("challenge %s: days should be positive, got %d", c.ID, c.Days)
		}
		if c.Deposit <= 0 {
			t.Errorf("challenge %s: deposit should be positive, got %d", c.ID, c.Deposit)
		}
	}
}

func TestIntegration_Idempotency(t *testing.T) {
	store := skipIfNoDatabase(t)
	defer store.Close()

	ctx := context.Background()
	scope := "test"
	key := "idem-" + time.Now().Format("20060102150405.000")

	// First check should return false
	exists, err := store.CheckIdempotency(ctx, scope, key)
	if err != nil {
		t.Fatalf("failed to check idempotency: %v", err)
	}
	if exists {
		t.Error("expected idempotency key to not exist initially")
	}

	// Set the key
	err = store.SetIdempotency(ctx, scope, key, 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to set idempotency: %v", err)
	}

	// Second check should return true
	exists, err = store.CheckIdempotency(ctx, scope, key)
	if err != nil {
		t.Fatalf("failed to check idempotency: %v", err)
	}
	if !exists {
		t.Error("expected idempotency key to exist after setting")
	}
}

func TestIntegration_RevokedSessions(t *testing.T) {
	store := skipIfNoDatabase(t)
	defer store.Close()

	ctx := context.Background()
	userSub := "test-sub-" + time.Now().Format("20060102150405")

	// Check should return false initially
	revoked, err := store.IsSessionRevoked(ctx, userSub)
	if err != nil {
		t.Fatalf("failed to check revoked: %v", err)
	}
	if revoked {
		t.Error("expected session to not be revoked initially")
	}

	// Revoke the session
	err = store.RevokeSession(ctx, userSub, "test revocation")
	if err != nil {
		t.Fatalf("failed to revoke session: %v", err)
	}

	// Check should return true now
	revoked, err = store.IsSessionRevoked(ctx, userSub)
	if err != nil {
		t.Fatalf("failed to check revoked: %v", err)
	}
	if !revoked {
		t.Error("expected session to be revoked after revoking")
	}
}
