package toss

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"habitcashback/internal/payment"
)

func TestTossPayClient_Mode(t *testing.T) {
	// Create a mock server to avoid needing real certs
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client with mock server's TLS config
	client := &TossPayClient{
		baseURL: server.URL,
		apiKey:  "test-api-key",
		hc:      server.Client(),
	}

	if client.Mode() != "live" {
		t.Errorf("expected mode 'live', got '%s'", client.Mode())
	}
}

func TestNewTossPayClient_Validation(t *testing.T) {
	t.Run("Missing Cert File", func(t *testing.T) {
		_, err := NewTossPayClient("", "/path/to/key.pem", "api-key", "")
		if err == nil {
			t.Error("expected error for missing cert file")
		}
	})

	t.Run("Missing Key File", func(t *testing.T) {
		_, err := NewTossPayClient("/path/to/cert.pem", "", "api-key", "")
		if err == nil {
			t.Error("expected error for missing key file")
		}
	})

	t.Run("Missing API Key", func(t *testing.T) {
		_, err := NewTossPayClient("/path/to/cert.pem", "/path/to/key.pem", "", "")
		if err == nil {
			t.Error("expected error for missing API key")
		}
	})

	t.Run("Non-existent Cert File", func(t *testing.T) {
		_, err := NewTossPayClient("/nonexistent/cert.pem", "/nonexistent/key.pem", "api-key", "")
		if err == nil {
			t.Error("expected error for non-existent cert file")
		}
	})
}

func TestNewTossPayClientFromEnv(t *testing.T) {
	t.Run("Missing Environment Variables", func(t *testing.T) {
		// Clear relevant env vars
		originalCert := os.Getenv("AIT_MTLS_CERT_FILE")
		originalKey := os.Getenv("AIT_MTLS_KEY_FILE")
		originalAPIKey := os.Getenv("TOSSPAY_API_KEY")
		defer func() {
			os.Setenv("AIT_MTLS_CERT_FILE", originalCert)
			os.Setenv("AIT_MTLS_KEY_FILE", originalKey)
			os.Setenv("TOSSPAY_API_KEY", originalAPIKey)
		}()

		os.Setenv("AIT_MTLS_CERT_FILE", "")
		os.Setenv("AIT_MTLS_KEY_FILE", "")
		os.Setenv("TOSSPAY_API_KEY", "")

		_, err := NewTossPayClientFromEnv()
		if err == nil {
			t.Error("expected error when env vars are missing")
		}
	})
}

func TestTossPayClient_CreatePayment(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v3/payments" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Error("expected Content-Type: application/json")
			}
			if r.Header.Get("Authorization") == "" {
				t.Error("expected Authorization header")
			}

			resp := map[string]interface{}{
				"code":         0,
				"payToken":     "test_pay_token_123",
				"checkoutPage": "https://pay.toss.im/checkout/test",
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := &TossPayClient{
			baseURL: server.URL,
			apiKey:  "test-api-key",
			hc:      server.Client(),
		}

		ctx := context.Background()
		resp, err := client.CreatePayment(ctx, payment.CreateRequest{
			OrderNo:     "ORDER-001",
			Amount:      10000,
			ProductDesc: "Test Product",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.PayToken != "test_pay_token_123" {
			t.Errorf("expected payToken 'test_pay_token_123', got '%s'", resp.PayToken)
		}
		if resp.Mode != "live" {
			t.Errorf("expected mode 'live', got '%s'", resp.Mode)
		}
	})

	t.Run("Empty OrderNo", func(t *testing.T) {
		client := &TossPayClient{
			baseURL: "https://pay.toss.im",
			apiKey:  "test-api-key",
			hc:      http.DefaultClient,
		}

		ctx := context.Background()
		_, err := client.CreatePayment(ctx, payment.CreateRequest{
			Amount: 10000,
		})

		if err == nil {
			t.Fatal("expected error for empty orderNo")
		}

		var payErr *payment.PaymentError
		if !containsPaymentError(err, &payErr) {
			t.Fatalf("expected PaymentError, got %T", err)
		}
		if payErr.Code != payment.ErrCodeInvalidRequest {
			t.Errorf("expected code %s, got %s", payment.ErrCodeInvalidRequest, payErr.Code)
		}
	})

	t.Run("Invalid Amount", func(t *testing.T) {
		client := &TossPayClient{
			baseURL: "https://pay.toss.im",
			apiKey:  "test-api-key",
			hc:      http.DefaultClient,
		}

		ctx := context.Background()
		_, err := client.CreatePayment(ctx, payment.CreateRequest{
			OrderNo: "ORDER-001",
			Amount:  0,
		})

		if err == nil {
			t.Fatal("expected error for zero amount")
		}
	})

	t.Run("API Error Response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]interface{}{
				"code": 999,
				"msg":  "Invalid request",
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := &TossPayClient{
			baseURL: server.URL,
			apiKey:  "test-api-key",
			hc:      server.Client(),
		}

		ctx := context.Background()
		_, err := client.CreatePayment(ctx, payment.CreateRequest{
			OrderNo: "ORDER-001",
			Amount:  10000,
		})

		if err == nil {
			t.Fatal("expected error for API error response")
		}
	})

	t.Run("HTTP Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		client := &TossPayClient{
			baseURL: server.URL,
			apiKey:  "test-api-key",
			hc:      server.Client(),
		}

		ctx := context.Background()
		_, err := client.CreatePayment(ctx, payment.CreateRequest{
			OrderNo: "ORDER-001",
			Amount:  10000,
		})

		if err == nil {
			t.Fatal("expected error for HTTP error")
		}
	})
}

func TestTossPayClient_ExecutePayment(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v3/execute" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			resp := map[string]interface{}{
				"code":          0,
				"orderNo":       "ORDER-001",
				"amount":        10000,
				"approvalTime":  time.Now().Format(time.RFC3339),
				"transactionId": "TX-12345",
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := &TossPayClient{
			baseURL: server.URL,
			apiKey:  "test-api-key",
			hc:      server.Client(),
		}

		ctx := context.Background()
		resp, err := client.ExecutePayment(ctx, "test_pay_token")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.OrderNo != "ORDER-001" {
			t.Errorf("expected orderNo 'ORDER-001', got '%s'", resp.OrderNo)
		}
		if resp.Amount != 10000 {
			t.Errorf("expected amount 10000, got %d", resp.Amount)
		}
		if resp.Status != "SUCCESS" {
			t.Errorf("expected status 'SUCCESS', got '%s'", resp.Status)
		}
	})

	t.Run("Empty PayToken", func(t *testing.T) {
		client := &TossPayClient{
			baseURL: "https://pay.toss.im",
			apiKey:  "test-api-key",
			hc:      http.DefaultClient,
		}

		ctx := context.Background()
		_, err := client.ExecutePayment(ctx, "")

		if err == nil {
			t.Fatal("expected error for empty payToken")
		}
	})
}

func TestTossPayClient_GetStatus(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v3/status" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			resp := map[string]interface{}{
				"code":      0,
				"payToken":  "test_pay_token",
				"orderNo":   "ORDER-001",
				"payStatus": "PAY_COMPLETE",
				"amount":    10000,
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := &TossPayClient{
			baseURL: server.URL,
			apiKey:  "test-api-key",
			hc:      server.Client(),
		}

		ctx := context.Background()
		resp, err := client.GetStatus(ctx, "test_pay_token")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.PayToken != "test_pay_token" {
			t.Errorf("expected payToken 'test_pay_token', got '%s'", resp.PayToken)
		}
		if resp.Status != "PAY_COMPLETE" {
			t.Errorf("expected status 'PAY_COMPLETE', got '%s'", resp.Status)
		}
	})

	t.Run("Empty PayToken", func(t *testing.T) {
		client := &TossPayClient{
			baseURL: "https://pay.toss.im",
			apiKey:  "test-api-key",
			hc:      http.DefaultClient,
		}

		ctx := context.Background()
		_, err := client.GetStatus(ctx, "")

		if err == nil {
			t.Fatal("expected error for empty payToken")
		}
	})
}

// Helper to check if error contains PaymentError
func containsPaymentError(err error, target **payment.PaymentError) bool {
	if pe, ok := err.(*payment.PaymentError); ok {
		*target = pe
		return true
	}
	return false
}
