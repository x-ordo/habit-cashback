package payment

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
	"time"
)

// MockService implements Service for local development and testing.
// It simulates TossPay behavior without requiring mTLS or network calls.
type MockService struct {
	delay    time.Duration
	mu       sync.RWMutex
	payments map[string]*mockPayment // payToken -> payment data
}

type mockPayment struct {
	OrderNo   string
	Amount    int64
	Status    string
	CreatedAt time.Time
}

// NewMockService creates a new mock payment service.
func NewMockService() *MockService {
	return &MockService{
		delay:    150 * time.Millisecond,
		payments: make(map[string]*mockPayment),
	}
}

// NewMockServiceWithDelay creates a mock service with custom delay.
func NewMockServiceWithDelay(delay time.Duration) *MockService {
	return &MockService{
		delay:    delay,
		payments: make(map[string]*mockPayment),
	}
}

// Mode returns "mock" to indicate this is the mock implementation.
func (m *MockService) Mode() string {
	return "mock"
}

// CreatePayment simulates payment creation and returns a mock payToken.
func (m *MockService) CreatePayment(ctx context.Context, req CreateRequest) (*CreateResponse, error) {
	// Simulate network delay
	select {
	case <-time.After(m.delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Validate request
	if req.OrderNo == "" {
		return nil, NewPaymentError(ErrCodeInvalidRequest, "orderNo is required", nil)
	}
	if req.Amount <= 0 {
		return nil, NewPaymentError(ErrCodeInvalidRequest, "amount must be positive", nil)
	}

	// Generate mock payToken
	payToken := "mock_pt_" + generateRandomHex(16)

	// Store payment data
	m.mu.Lock()
	m.payments[payToken] = &mockPayment{
		OrderNo:   req.OrderNo,
		Amount:    req.Amount,
		Status:    "CREATED",
		CreatedAt: time.Now(),
	}
	m.mu.Unlock()

	return &CreateResponse{
		PayToken:    payToken,
		CheckoutURL: "mock://checkout/" + payToken,
		Mode:        "mock",
	}, nil
}

// ExecutePayment simulates payment execution.
func (m *MockService) ExecutePayment(ctx context.Context, payToken string) (*ExecuteResponse, error) {
	// Simulate network delay
	select {
	case <-time.After(m.delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Validate payToken
	if payToken == "" {
		return nil, NewPaymentError(ErrCodeInvalidRequest, "payToken is required", nil)
	}
	if !strings.HasPrefix(payToken, "mock_pt_") {
		return nil, NewPaymentError(ErrCodeInvalidRequest, "invalid mock payToken format", nil)
	}

	// Look up payment
	m.mu.Lock()
	payment, exists := m.payments[payToken]
	if !exists {
		m.mu.Unlock()
		return nil, NewPaymentError(ErrCodePaymentNotFound, "payment not found", nil)
	}

	// Check if already executed
	if payment.Status == "SUCCESS" {
		m.mu.Unlock()
		return nil, NewPaymentError(ErrCodeInvalidRequest, "payment already executed", nil)
	}

	// Update status
	payment.Status = "SUCCESS"
	m.mu.Unlock()

	return &ExecuteResponse{
		OrderNo:      payment.OrderNo,
		Amount:       payment.Amount,
		ApprovalTime: time.Now(),
		Status:       "SUCCESS",
		TxID:         "mock_tx_" + generateRandomHex(8),
	}, nil
}

// GetStatus retrieves the status of a mock payment.
func (m *MockService) GetStatus(ctx context.Context, payToken string) (*StatusResponse, error) {
	// Simulate network delay
	select {
	case <-time.After(m.delay / 2):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	if payToken == "" {
		return nil, NewPaymentError(ErrCodeInvalidRequest, "payToken is required", nil)
	}

	m.mu.RLock()
	payment, exists := m.payments[payToken]
	m.mu.RUnlock()

	if !exists {
		return nil, NewPaymentError(ErrCodePaymentNotFound, "payment not found", nil)
	}

	return &StatusResponse{
		PayToken: payToken,
		OrderNo:  payment.OrderNo,
		Status:   payment.Status,
		Amount:   payment.Amount,
	}, nil
}

// generateRandomHex generates a random hex string of specified length.
func generateRandomHex(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based if crypto/rand fails
		return hex.EncodeToString([]byte(time.Now().String()))[:length]
	}
	return hex.EncodeToString(bytes)
}
