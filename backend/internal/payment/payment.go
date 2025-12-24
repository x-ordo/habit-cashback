// Package payment provides payment service interface and implementations.
package payment

import (
	"context"
	"time"
)

// Service defines the interface for payment operations.
// Implementations include MockService (local/test) and TossPayService (staging/prod).
type Service interface {
	// CreatePayment creates a new payment and returns a payToken for checkout.
	CreatePayment(ctx context.Context, req CreateRequest) (*CreateResponse, error)

	// ExecutePayment confirms and executes a payment after user approval.
	ExecutePayment(ctx context.Context, payToken string) (*ExecuteResponse, error)

	// GetStatus retrieves the current status of a payment.
	GetStatus(ctx context.Context, payToken string) (*StatusResponse, error)

	// Mode returns the service mode ("mock" or "live").
	Mode() string
}

// CreateRequest contains the data needed to create a payment.
type CreateRequest struct {
	OrderNo     string // Unique order number
	ProductDesc string // Product description for TossPay UI
	Amount      int64  // Payment amount in KRW
	ReturnURL   string // URL to return after payment (optional)
	ResultURL   string // URL for server callback (optional)
}

// CreateResponse contains the result of payment creation.
type CreateResponse struct {
	PayToken    string // Token to use for checkout
	CheckoutURL string // URL for payment UI (if applicable)
	Mode        string // "mock" or "live"
}

// ExecuteResponse contains the result of payment execution.
type ExecuteResponse struct {
	OrderNo      string    // Order number
	Amount       int64     // Payment amount
	ApprovalTime time.Time // Time of approval
	Status       string    // Payment status (e.g., "SUCCESS", "FAILED")
	TxID         string    // Transaction ID from payment provider
}

// StatusResponse contains the current payment status.
type StatusResponse struct {
	PayToken string // Payment token
	OrderNo  string // Order number
	Status   string // Current status
	Amount   int64  // Payment amount
}

// Error codes for payment operations.
const (
	ErrCodeInvalidRequest  = "INVALID_REQUEST"
	ErrCodePaymentNotFound = "PAYMENT_NOT_FOUND"
	ErrCodePaymentFailed   = "PAYMENT_FAILED"
	ErrCodeNetworkError    = "NETWORK_ERROR"
	ErrCodeInternalError   = "INTERNAL_ERROR"
)

// PaymentError represents a payment-specific error.
type PaymentError struct {
	Code    string
	Message string
	Cause   error
}

func (e *PaymentError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *PaymentError) Unwrap() error {
	return e.Cause
}

// NewPaymentError creates a new PaymentError.
func NewPaymentError(code, message string, cause error) *PaymentError {
	return &PaymentError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}
