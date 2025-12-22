package toss

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const DefaultBaseURL = "https://apps-in-toss-api.toss.im"

// Client is a minimal Apps-in-Toss mTLS client used for Toss Login.
// Docs: https://developers-apps-in-toss.toss.im/login/develop.html
type Client struct {
	baseURL string
	hc      *http.Client
}

// New creates a client with mTLS using a client certificate.
func New(certFile, keyFile, baseURL string) (*Client, error) {
	certFile = strings.TrimSpace(certFile)
	keyFile = strings.TrimSpace(keyFile)
	if certFile == "" || keyFile == "" {
		return nil, errors.New("mtls cert/key file path is required")
	}
	if baseURL == "" {
		baseURL = DefaultBaseURL
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
		Timeout:   15 * time.Second,
	}

	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		hc:      hc,
	}, nil
}

// NewFromEnv creates a client from env vars.
// Required (staging/prod):
//   - AIT_MTLS_CERT_FILE
//   - AIT_MTLS_KEY_FILE
// Optional:
//   - AIT_TOSS_BASE_URL (default: https://apps-in-toss-api.toss.im)
func NewFromEnv() (*Client, error) {
	certFile := os.Getenv("AIT_MTLS_CERT_FILE")
	keyFile := os.Getenv("AIT_MTLS_KEY_FILE")
	baseURL := os.Getenv("AIT_TOSS_BASE_URL")
	if strings.TrimSpace(baseURL) == "" {
		baseURL = os.Getenv("AIT_BASE_URL") // backward compat
	}
	if strings.TrimSpace(certFile) == "" || strings.TrimSpace(keyFile) == "" {
		return nil, errors.New("AIT_MTLS_CERT_FILE / AIT_MTLS_KEY_FILE not set")
	}
	return New(certFile, keyFile, baseURL)
}

type GenerateTokenRequest struct {
	AuthorizationCode string `json:"authorizationCode"`
	Referrer          string `json:"referrer"`
}

type GenerateTokenSuccess struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	Scope        string `json:"scope"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int64  `json:"expiresIn"`
}

type APIError struct {
	ErrorCode string `json:"errorCode"`
	Reason    string `json:"reason"`
}

type GenerateTokenResponse struct {
	ResultType string               `json:"resultType"`
	Success    *GenerateTokenSuccess `json:"success"`
	Error      *APIError            `json:"error"`
}

type LoginMeSuccess struct {
	UserKey int64  `json:"userKey"`
	Scope   string `json:"scope"`
}

type LoginMeResponse struct {
	ResultType string         `json:"resultType"`
	Success    *LoginMeSuccess `json:"success"`
	Error      *APIError      `json:"error"`
}

func normalizeReferrer(referrer string) string {
	u := strings.ToUpper(strings.TrimSpace(referrer))
	if u == "" {
		return "DEFAULT"
	}
	return u
}

// GenerateUserToken exchanges authorizationCode -> accessToken/refreshToken.
func (c *Client) GenerateUserToken(ctx context.Context, authorizationCode, referrer string) (*GenerateTokenSuccess, error) {
	payload := GenerateTokenRequest{
		AuthorizationCode: strings.TrimSpace(authorizationCode),
		Referrer:          normalizeReferrer(referrer),
	}
	if payload.AuthorizationCode == "" {
		return nil, errors.New("authorizationCode is required")
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api-partner/v1/apps-in-toss/user/oauth2/generate-token", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("apps-in-toss api status=%d body=%s", resp.StatusCode, string(raw))
	}

	var out GenerateTokenResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode response: %w body=%s", err, string(raw))
	}
	if strings.ToUpper(out.ResultType) != "SUCCESS" || out.Success == nil {
		// some error responses use {"error":"invalid_grant"} (string-only); keep body for debugging
		return nil, fmt.Errorf("apps-in-toss api not success: resultType=%s body=%s", out.ResultType, string(raw))
	}
	return out.Success, nil
}

// LoginMe fetches userKey using accessToken.
// Content-type: application/json
// Method: GET
// URL: /api-partner/v1/apps-in-toss/user/oauth2/login-me
func (c *Client) LoginMe(ctx context.Context, accessToken string) (*LoginMeSuccess, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return nil, errors.New("accessToken is required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api-partner/v1/apps-in-toss/user/oauth2/login-me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("apps-in-toss api status=%d body=%s", resp.StatusCode, string(raw))
	}
	var out LoginMeResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode response: %w body=%s", err, string(raw))
	}
	if strings.ToUpper(out.ResultType) != "SUCCESS" || out.Success == nil {
		return nil, fmt.Errorf("apps-in-toss api not success: resultType=%s body=%s", out.ResultType, string(raw))
	}
	return out.Success, nil
}

// DecodeBase64URL is a small helper used by higher layers (optional).
func DecodeBase64URL(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
