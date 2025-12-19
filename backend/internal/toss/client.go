package toss

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	http *http.Client

	APIBaseURL string // https://apps-in-toss-api.toss.im
	PayBaseURL string // https://pay-apps-in-toss-api.toss.im
}

type ResultEnvelope[T any] struct {
	ResultType string `json:"resultType"`
	Success    *T     `json:"success,omitempty"`
	Result     *T     `json:"result,omitempty"` // some endpoints use `result`
	Error      any    `json:"error,omitempty"`
}

func NewClient(apiBaseURL, payBaseURL string, timeout time.Duration) (*Client, error) {
	hc, err := newMTLSHTTPClient(timeout)
	if err != nil {
		return nil, err
	}
	return &Client{
		http:       hc,
		APIBaseURL: apiBaseURL,
		PayBaseURL: payBaseURL,
	}, nil
}

// newMTLSHTTPClient builds an HTTP client with optional partner mTLS cert.
// If cert/key env vars are not provided, it returns an error (because Toss APIs require mTLS).
func newMTLSHTTPClient(timeout time.Duration) (*http.Client, error) {
	certFile := os.Getenv("TOSS_MTLS_CERT_FILE")
	keyFile := os.Getenv("TOSS_MTLS_KEY_FILE")
	caFile := os.Getenv("TOSS_MTLS_CA_FILE")

	if certFile == "" || keyFile == "" {
		return nil, errors.New("missing TOSS_MTLS_CERT_FILE or TOSS_MTLS_KEY_FILE")
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load keypair: %w", err)
	}

	var rootCAs *x509.CertPool
	if caFile != "" {
		b, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("read ca file: %w", err)
		}
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM(b); !ok {
			return nil, errors.New("append ca pem failed")
		}
		rootCAs = pool
	}

	tlsCfg := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
		RootCAs:      rootCAs, // nil -> system roots
	}

	tr := &http.Transport{
		TLSClientConfig:   tlsCfg,
		ForceAttemptHTTP2: true,
	}

	return &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}, nil
}

func doJSON[T any](ctx context.Context, hc *http.Client, method, url string, headers map[string]string, reqBody any) (ResultEnvelope[T], []byte, error) {
	var buf io.Reader
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			return ResultEnvelope[T]{}, nil, err
		}
		buf = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, buf)
	if err != nil {
		return ResultEnvelope[T]{}, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := hc.Do(req)
	if err != nil {
		return ResultEnvelope[T]{}, nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ResultEnvelope[T]{}, raw, fmt.Errorf("http %d: %s", resp.StatusCode, string(raw))
	}

	var env ResultEnvelope[T]
	if err := json.Unmarshal(raw, &env); err != nil {
		// some endpoints may not envelope - but Toss docs show envelopes
		return ResultEnvelope[T]{}, raw, err
	}
	return env, raw, nil
}
