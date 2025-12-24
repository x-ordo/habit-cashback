package toss

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"habitcashback/internal/payment"
)

const DefaultTossPayBaseURL = "https://pay.toss.im"

// TossPayClient is a mTLS client for TossPay API.
// Implements payment.Service interface.
type TossPayClient struct {
	baseURL string
	apiKey  string
	hc      *http.Client
}

// NewTossPayClient creates a TossPay client with mTLS.
func NewTossPayClient(certFile, keyFile, apiKey, baseURL string) (*TossPayClient, error) {
	certFile = strings.TrimSpace(certFile)
	keyFile = strings.TrimSpace(keyFile)
	apiKey = strings.TrimSpace(apiKey)

	if certFile == "" || keyFile == "" {
		return nil, errors.New("mtls cert/key file path is required")
	}
	if apiKey == "" {
		return nil, errors.New("TossPay API key is required")
	}
	if baseURL == "" {
		baseURL = DefaultTossPayBaseURL
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load mtls cert/key: %w", err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{cert},
		},
	}

	hc := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	return &TossPayClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		hc:      hc,
	}, nil
}

// NewTossPayClientFromEnv creates a TossPay client from environment variables.
// Required:
//   - TOSSPAY_API_KEY
//   - AIT_MTLS_CERT_FILE
//   - AIT_MTLS_KEY_FILE
//
// Optional:
//   - TOSSPAY_BASE_URL (default: https://pay.toss.im)
func NewTossPayClientFromEnv() (*TossPayClient, error) {
	certFile := os.Getenv("AIT_MTLS_CERT_FILE")
	keyFile := os.Getenv("AIT_MTLS_KEY_FILE")
	apiKey := os.Getenv("TOSSPAY_API_KEY")
	baseURL := os.Getenv("TOSSPAY_BASE_URL")

	return NewTossPayClient(certFile, keyFile, apiKey, baseURL)
}

// Mode returns "live" to indicate this is the real TossPay implementation.
func (c *TossPayClient) Mode() string {
	return "live"
}

// TossPay API request/response types

type tossPayCreateRequest struct {
	OrderNo      string `json:"orderNo"`
	Amount       int64  `json:"amount"`
	ProductDesc  string `json:"productDesc"`
	RetURL       string `json:"retUrl,omitempty"`
	ResultURL    string `json:"resultUrl,omitempty"`
	AmountTaxFree int64  `json:"amountTaxFree,omitempty"`
}

type tossPayCreateResponse struct {
	Code        int    `json:"code"`
	Msg         string `json:"msg,omitempty"`
	PayToken    string `json:"payToken,omitempty"`
	CheckoutPage string `json:"checkoutPage,omitempty"`
}

type tossPayExecuteRequest struct {
	PayToken string `json:"payToken"`
}

type tossPayExecuteResponse struct {
	Code         int    `json:"code"`
	Msg          string `json:"msg,omitempty"`
	Mode         string `json:"mode,omitempty"`
	OrderNo      string `json:"orderNo,omitempty"`
	Amount       int64  `json:"amount,omitempty"`
	ApprovalTime string `json:"approvalTime,omitempty"`
	StateMsg     string `json:"stateMsg,omitempty"`
	TransactionId string `json:"transactionId,omitempty"`
}

type tossPayStatusResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg,omitempty"`
	PayToken  string `json:"payToken,omitempty"`
	OrderNo   string `json:"orderNo,omitempty"`
	PayStatus string `json:"payStatus,omitempty"`
	Amount    int64  `json:"amount,omitempty"`
}

// CreatePayment creates a new payment on TossPay and returns a payToken.
func (c *TossPayClient) CreatePayment(ctx context.Context, req payment.CreateRequest) (*payment.CreateResponse, error) {
	if req.OrderNo == "" {
		return nil, payment.NewPaymentError(payment.ErrCodeInvalidRequest, "orderNo is required", nil)
	}
	if req.Amount <= 0 {
		return nil, payment.NewPaymentError(payment.ErrCodeInvalidRequest, "amount must be positive", nil)
	}

	payload := tossPayCreateRequest{
		OrderNo:     req.OrderNo,
		Amount:      req.Amount,
		ProductDesc: req.ProductDesc,
		RetURL:      req.ReturnURL,
		ResultURL:   req.ResultURL,
	}

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v3/payments", bytes.NewReader(body))
	if err != nil {
		return nil, payment.NewPaymentError(payment.ErrCodeInternalError, "failed to create request", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+c.apiKey)

	resp, err := c.hc.Do(httpReq)
	if err != nil {
		return nil, payment.NewPaymentError(payment.ErrCodeNetworkError, "failed to call TossPay API", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, payment.NewPaymentError(payment.ErrCodePaymentFailed, fmt.Sprintf("TossPay API error: status=%d body=%s", resp.StatusCode, string(raw)), nil)
	}

	var out tossPayCreateResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, payment.NewPaymentError(payment.ErrCodeInternalError, "failed to parse TossPay response", err)
	}

	// TossPay returns code 0 for success
	if out.Code != 0 {
		return nil, payment.NewPaymentError(payment.ErrCodePaymentFailed, fmt.Sprintf("TossPay error: code=%d msg=%s", out.Code, out.Msg), nil)
	}

	return &payment.CreateResponse{
		PayToken:    out.PayToken,
		CheckoutURL: out.CheckoutPage,
		Mode:        "live",
	}, nil
}

// ExecutePayment confirms and executes a payment.
func (c *TossPayClient) ExecutePayment(ctx context.Context, payToken string) (*payment.ExecuteResponse, error) {
	if payToken == "" {
		return nil, payment.NewPaymentError(payment.ErrCodeInvalidRequest, "payToken is required", nil)
	}

	payload := tossPayExecuteRequest{PayToken: payToken}
	body, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v3/execute", bytes.NewReader(body))
	if err != nil {
		return nil, payment.NewPaymentError(payment.ErrCodeInternalError, "failed to create request", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+c.apiKey)

	resp, err := c.hc.Do(httpReq)
	if err != nil {
		return nil, payment.NewPaymentError(payment.ErrCodeNetworkError, "failed to call TossPay API", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, payment.NewPaymentError(payment.ErrCodePaymentFailed, fmt.Sprintf("TossPay API error: status=%d body=%s", resp.StatusCode, string(raw)), nil)
	}

	var out tossPayExecuteResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, payment.NewPaymentError(payment.ErrCodeInternalError, "failed to parse TossPay response", err)
	}

	if out.Code != 0 {
		return nil, payment.NewPaymentError(payment.ErrCodePaymentFailed, fmt.Sprintf("TossPay error: code=%d msg=%s", out.Code, out.Msg), nil)
	}

	// Parse approval time
	var approvalTime time.Time
	if out.ApprovalTime != "" {
		// TossPay format: "2024-01-15T14:30:00+09:00"
		approvalTime, _ = time.Parse(time.RFC3339, out.ApprovalTime)
	}
	if approvalTime.IsZero() {
		approvalTime = time.Now()
	}

	return &payment.ExecuteResponse{
		OrderNo:      out.OrderNo,
		Amount:       out.Amount,
		ApprovalTime: approvalTime,
		Status:       "SUCCESS",
		TxID:         out.TransactionId,
	}, nil
}

// GetStatus retrieves the current status of a payment.
func (c *TossPayClient) GetStatus(ctx context.Context, payToken string) (*payment.StatusResponse, error) {
	if payToken == "" {
		return nil, payment.NewPaymentError(payment.ErrCodeInvalidRequest, "payToken is required", nil)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v3/status", bytes.NewReader([]byte(`{"payToken":"`+payToken+`"}`)))
	if err != nil {
		return nil, payment.NewPaymentError(payment.ErrCodeInternalError, "failed to create request", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+c.apiKey)

	resp, err := c.hc.Do(httpReq)
	if err != nil {
		return nil, payment.NewPaymentError(payment.ErrCodeNetworkError, "failed to call TossPay API", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, payment.NewPaymentError(payment.ErrCodePaymentFailed, fmt.Sprintf("TossPay API error: status=%d body=%s", resp.StatusCode, string(raw)), nil)
	}

	var out tossPayStatusResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, payment.NewPaymentError(payment.ErrCodeInternalError, "failed to parse TossPay response", err)
	}

	if out.Code != 0 {
		return nil, payment.NewPaymentError(payment.ErrCodePaymentFailed, fmt.Sprintf("TossPay error: code=%d msg=%s", out.Code, out.Msg), nil)
	}

	return &payment.StatusResponse{
		PayToken: out.PayToken,
		OrderNo:  out.OrderNo,
		Status:   out.PayStatus,
		Amount:   out.Amount,
	}, nil
}
