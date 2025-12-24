package payment

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPaymentError(t *testing.T) {
	t.Run("Error without cause", func(t *testing.T) {
		err := NewPaymentError(ErrCodeInvalidRequest, "test message", nil)

		if err.Code != ErrCodeInvalidRequest {
			t.Errorf("expected code %s, got %s", ErrCodeInvalidRequest, err.Code)
		}
		if err.Message != "test message" {
			t.Errorf("expected message 'test message', got '%s'", err.Message)
		}
		if err.Error() != "test message" {
			t.Errorf("expected Error() 'test message', got '%s'", err.Error())
		}
		if err.Unwrap() != nil {
			t.Error("expected Unwrap() to return nil")
		}
	})

	t.Run("Error with cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := NewPaymentError(ErrCodeNetworkError, "network failed", cause)

		if err.Code != ErrCodeNetworkError {
			t.Errorf("expected code %s, got %s", ErrCodeNetworkError, err.Code)
		}
		expectedMsg := "network failed: underlying error"
		if err.Error() != expectedMsg {
			t.Errorf("expected Error() '%s', got '%s'", expectedMsg, err.Error())
		}
		if err.Unwrap() != cause {
			t.Error("expected Unwrap() to return the cause")
		}
	})
}

func TestMockService_Mode(t *testing.T) {
	svc := NewMockService()
	if svc.Mode() != "mock" {
		t.Errorf("expected mode 'mock', got '%s'", svc.Mode())
	}
}

func TestMockService_CreatePayment(t *testing.T) {
	svc := NewMockServiceWithDelay(0) // No delay for faster tests

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		req := CreateRequest{
			OrderNo:     "ORDER-001",
			ProductDesc: "Test Product",
			Amount:      10000,
		}

		resp, err := svc.CreatePayment(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.PayToken == "" {
			t.Error("expected non-empty payToken")
		}
		if resp.Mode != "mock" {
			t.Errorf("expected mode 'mock', got '%s'", resp.Mode)
		}
		if resp.CheckoutURL == "" {
			t.Error("expected non-empty checkoutURL")
		}
	})

	t.Run("Empty OrderNo", func(t *testing.T) {
		ctx := context.Background()
		req := CreateRequest{
			Amount: 10000,
		}

		_, err := svc.CreatePayment(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty orderNo")
		}

		var payErr *PaymentError
		if !errors.As(err, &payErr) {
			t.Fatalf("expected PaymentError, got %T", err)
		}
		if payErr.Code != ErrCodeInvalidRequest {
			t.Errorf("expected code %s, got %s", ErrCodeInvalidRequest, payErr.Code)
		}
	})

	t.Run("Invalid Amount", func(t *testing.T) {
		ctx := context.Background()
		req := CreateRequest{
			OrderNo: "ORDER-002",
			Amount:  0,
		}

		_, err := svc.CreatePayment(ctx, req)
		if err == nil {
			t.Fatal("expected error for zero amount")
		}

		var payErr *PaymentError
		if !errors.As(err, &payErr) {
			t.Fatalf("expected PaymentError, got %T", err)
		}
		if payErr.Code != ErrCodeInvalidRequest {
			t.Errorf("expected code %s, got %s", ErrCodeInvalidRequest, payErr.Code)
		}
	})

	t.Run("Negative Amount", func(t *testing.T) {
		ctx := context.Background()
		req := CreateRequest{
			OrderNo: "ORDER-003",
			Amount:  -1000,
		}

		_, err := svc.CreatePayment(ctx, req)
		if err == nil {
			t.Fatal("expected error for negative amount")
		}
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		svc := NewMockServiceWithDelay(1 * time.Second)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		req := CreateRequest{
			OrderNo: "ORDER-004",
			Amount:  10000,
		}

		_, err := svc.CreatePayment(ctx, req)
		if err == nil {
			t.Fatal("expected context cancellation error")
		}
	})
}

func TestMockService_ExecutePayment(t *testing.T) {
	svc := NewMockServiceWithDelay(0)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()

		// First create a payment
		createResp, err := svc.CreatePayment(ctx, CreateRequest{
			OrderNo: "ORDER-EXEC-001",
			Amount:  15000,
		})
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		// Execute the payment
		execResp, err := svc.ExecutePayment(ctx, createResp.PayToken)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if execResp.OrderNo != "ORDER-EXEC-001" {
			t.Errorf("expected orderNo 'ORDER-EXEC-001', got '%s'", execResp.OrderNo)
		}
		if execResp.Amount != 15000 {
			t.Errorf("expected amount 15000, got %d", execResp.Amount)
		}
		if execResp.Status != "SUCCESS" {
			t.Errorf("expected status 'SUCCESS', got '%s'", execResp.Status)
		}
		if execResp.TxID == "" {
			t.Error("expected non-empty txID")
		}
		if execResp.ApprovalTime.IsZero() {
			t.Error("expected non-zero approval time")
		}
	})

	t.Run("Empty PayToken", func(t *testing.T) {
		ctx := context.Background()

		_, err := svc.ExecutePayment(ctx, "")
		if err == nil {
			t.Fatal("expected error for empty payToken")
		}

		var payErr *PaymentError
		if !errors.As(err, &payErr) {
			t.Fatalf("expected PaymentError, got %T", err)
		}
		if payErr.Code != ErrCodeInvalidRequest {
			t.Errorf("expected code %s, got %s", ErrCodeInvalidRequest, payErr.Code)
		}
	})

	t.Run("Invalid PayToken Format", func(t *testing.T) {
		ctx := context.Background()

		_, err := svc.ExecutePayment(ctx, "invalid_token")
		if err == nil {
			t.Fatal("expected error for invalid payToken format")
		}

		var payErr *PaymentError
		if !errors.As(err, &payErr) {
			t.Fatalf("expected PaymentError, got %T", err)
		}
		if payErr.Code != ErrCodeInvalidRequest {
			t.Errorf("expected code %s, got %s", ErrCodeInvalidRequest, payErr.Code)
		}
	})

	t.Run("PayToken Not Found", func(t *testing.T) {
		ctx := context.Background()

		_, err := svc.ExecutePayment(ctx, "mock_pt_nonexistent12345678")
		if err == nil {
			t.Fatal("expected error for non-existent payToken")
		}

		var payErr *PaymentError
		if !errors.As(err, &payErr) {
			t.Fatalf("expected PaymentError, got %T", err)
		}
		if payErr.Code != ErrCodePaymentNotFound {
			t.Errorf("expected code %s, got %s", ErrCodePaymentNotFound, payErr.Code)
		}
	})

	t.Run("Double Execution", func(t *testing.T) {
		ctx := context.Background()

		// Create a payment
		createResp, _ := svc.CreatePayment(ctx, CreateRequest{
			OrderNo: "ORDER-DOUBLE-001",
			Amount:  20000,
		})

		// First execution should succeed
		_, err := svc.ExecutePayment(ctx, createResp.PayToken)
		if err != nil {
			t.Fatalf("first execution failed: %v", err)
		}

		// Second execution should fail
		_, err = svc.ExecutePayment(ctx, createResp.PayToken)
		if err == nil {
			t.Fatal("expected error for double execution")
		}

		var payErr *PaymentError
		if !errors.As(err, &payErr) {
			t.Fatalf("expected PaymentError, got %T", err)
		}
		if payErr.Code != ErrCodeInvalidRequest {
			t.Errorf("expected code %s, got %s", ErrCodeInvalidRequest, payErr.Code)
		}
	})
}

func TestMockService_GetStatus(t *testing.T) {
	svc := NewMockServiceWithDelay(0)

	t.Run("Created Payment", func(t *testing.T) {
		ctx := context.Background()

		// Create a payment
		createResp, _ := svc.CreatePayment(ctx, CreateRequest{
			OrderNo: "ORDER-STATUS-001",
			Amount:  25000,
		})

		// Get status
		status, err := svc.GetStatus(ctx, createResp.PayToken)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if status.PayToken != createResp.PayToken {
			t.Errorf("expected payToken '%s', got '%s'", createResp.PayToken, status.PayToken)
		}
		if status.OrderNo != "ORDER-STATUS-001" {
			t.Errorf("expected orderNo 'ORDER-STATUS-001', got '%s'", status.OrderNo)
		}
		if status.Amount != 25000 {
			t.Errorf("expected amount 25000, got %d", status.Amount)
		}
		if status.Status != "CREATED" {
			t.Errorf("expected status 'CREATED', got '%s'", status.Status)
		}
	})

	t.Run("Executed Payment", func(t *testing.T) {
		ctx := context.Background()

		// Create and execute a payment
		createResp, _ := svc.CreatePayment(ctx, CreateRequest{
			OrderNo: "ORDER-STATUS-002",
			Amount:  30000,
		})
		svc.ExecutePayment(ctx, createResp.PayToken)

		// Get status
		status, err := svc.GetStatus(ctx, createResp.PayToken)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if status.Status != "SUCCESS" {
			t.Errorf("expected status 'SUCCESS', got '%s'", status.Status)
		}
	})

	t.Run("Empty PayToken", func(t *testing.T) {
		ctx := context.Background()

		_, err := svc.GetStatus(ctx, "")
		if err == nil {
			t.Fatal("expected error for empty payToken")
		}

		var payErr *PaymentError
		if !errors.As(err, &payErr) {
			t.Fatalf("expected PaymentError, got %T", err)
		}
		if payErr.Code != ErrCodeInvalidRequest {
			t.Errorf("expected code %s, got %s", ErrCodeInvalidRequest, payErr.Code)
		}
	})

	t.Run("PayToken Not Found", func(t *testing.T) {
		ctx := context.Background()

		_, err := svc.GetStatus(ctx, "mock_pt_notfound123456")
		if err == nil {
			t.Fatal("expected error for non-existent payToken")
		}

		var payErr *PaymentError
		if !errors.As(err, &payErr) {
			t.Fatalf("expected PaymentError, got %T", err)
		}
		if payErr.Code != ErrCodePaymentNotFound {
			t.Errorf("expected code %s, got %s", ErrCodePaymentNotFound, payErr.Code)
		}
	})
}

func TestMockService_Concurrency(t *testing.T) {
	svc := NewMockServiceWithDelay(0)
	ctx := context.Background()

	// Create multiple payments concurrently
	const numPayments = 100
	results := make(chan *CreateResponse, numPayments)
	errors := make(chan error, numPayments)

	for i := 0; i < numPayments; i++ {
		go func(idx int) {
			resp, err := svc.CreatePayment(ctx, CreateRequest{
				OrderNo: "CONCURRENT-" + string(rune('A'+idx%26)),
				Amount:  int64(1000 + idx),
			})
			if err != nil {
				errors <- err
			} else {
				results <- resp
			}
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numPayments; i++ {
		select {
		case <-results:
			successCount++
		case err := <-errors:
			t.Errorf("concurrent payment failed: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for concurrent payments")
		}
	}

	if successCount != numPayments {
		t.Errorf("expected %d successful payments, got %d", numPayments, successCount)
	}
}
