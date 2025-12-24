package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ===== Session Tests =====

func TestSignAndVerifySession(t *testing.T) {
	secret := "test-secret-key-12345"

	t.Run("Valid session", func(t *testing.T) {
		userID := "test-user-123"
		token := signSession(secret, userID, 1*time.Hour)

		if token == "" {
			t.Fatal("expected non-empty token")
		}
		if !strings.HasPrefix(token, "sv1.") {
			t.Errorf("expected token to start with 'sv1.', got: %s", token[:10])
		}

		claims, err := verifySession(secret, token)
		if err != nil {
			t.Fatalf("failed to verify session: %v", err)
		}

		if claims.Sub != userID {
			t.Errorf("expected Sub '%s', got '%s'", userID, claims.Sub)
		}
		if claims.Exp == 0 {
			t.Error("expected non-zero Exp")
		}
		if claims.Iat == 0 {
			t.Error("expected non-zero Iat")
		}
	})

	t.Run("Expired session", func(t *testing.T) {
		token := signSession(secret, "user", -1*time.Hour) // Already expired

		_, err := verifySession(secret, token)
		if err == nil {
			t.Fatal("expected error for expired session")
		}
		if !strings.Contains(err.Error(), "expired") {
			t.Errorf("expected 'expired' error, got: %v", err)
		}
	})

	t.Run("Wrong secret", func(t *testing.T) {
		token := signSession(secret, "user", 1*time.Hour)

		_, err := verifySession("wrong-secret", token)
		if err == nil {
			t.Fatal("expected error for wrong secret")
		}
	})

	t.Run("Invalid token format", func(t *testing.T) {
		_, err := verifySession(secret, "invalid-token")
		if err == nil {
			t.Fatal("expected error for invalid token format")
		}
	})

	t.Run("Empty token", func(t *testing.T) {
		_, err := verifySession(secret, "")
		if err == nil {
			t.Fatal("expected error for empty token")
		}
	})

	t.Run("Tampered payload", func(t *testing.T) {
		token := signSession(secret, "user", 1*time.Hour)
		parts := strings.Split(token, ".")
		// Tamper with payload
		parts[1] = "dGFtcGVyZWQ" // "tampered" in base64
		tamperedToken := strings.Join(parts, ".")

		_, err := verifySession(secret, tamperedToken)
		if err == nil {
			t.Fatal("expected error for tampered token")
		}
	})
}

// ===== Rate Limiter Tests =====

func TestRateLimiter(t *testing.T) {
	t.Run("Allow within limit", func(t *testing.T) {
		rl := newRateLimiter(5, time.Minute)

		for i := 0; i < 5; i++ {
			if !rl.Allow("test-ip") {
				t.Errorf("expected to allow request %d", i+1)
			}
		}
	})

	t.Run("Block after limit", func(t *testing.T) {
		rl := newRateLimiter(3, time.Minute)

		// Use up the limit
		for i := 0; i < 3; i++ {
			rl.Allow("test-ip")
		}

		// Should be blocked
		if rl.Allow("test-ip") {
			t.Error("expected to block request after limit exceeded")
		}
	})

	t.Run("Different IPs independent", func(t *testing.T) {
		rl := newRateLimiter(2, time.Minute)

		// Use up limit for IP1
		rl.Allow("ip1")
		rl.Allow("ip1")

		// IP2 should still be allowed
		if !rl.Allow("ip2") {
			t.Error("expected IP2 to be allowed independently")
		}
	})

	t.Run("Reset after window", func(t *testing.T) {
		rl := newRateLimiter(2, 10*time.Millisecond)

		// Use up the limit
		rl.Allow("test-ip")
		rl.Allow("test-ip")

		// Wait for window to reset
		time.Sleep(15 * time.Millisecond)

		// Should be allowed again
		if !rl.Allow("test-ip") {
			t.Error("expected to allow request after window reset")
		}
	})
}

// ===== Idempotency Store Tests =====

func TestIdemStore(t *testing.T) {
	t.Run("First use succeeds", func(t *testing.T) {
		store := newIdemStore()

		if !store.TryUse("key1", time.Minute) {
			t.Error("expected first use to succeed")
		}
	})

	t.Run("Duplicate blocked", func(t *testing.T) {
		store := newIdemStore()

		store.TryUse("key1", time.Minute)

		if store.TryUse("key1", time.Minute) {
			t.Error("expected duplicate to be blocked")
		}
	})

	t.Run("Different keys independent", func(t *testing.T) {
		store := newIdemStore()

		store.TryUse("key1", time.Minute)

		if !store.TryUse("key2", time.Minute) {
			t.Error("expected different key to succeed")
		}
	})

	t.Run("Expired key can be reused", func(t *testing.T) {
		store := newIdemStore()

		store.TryUse("key1", 10*time.Millisecond)

		time.Sleep(15 * time.Millisecond)

		if !store.TryUse("key1", time.Minute) {
			t.Error("expected expired key to be reusable")
		}
	})
}

// ===== Revoked Store Tests =====

func TestRevokedStore(t *testing.T) {
	t.Run("Not revoked initially", func(t *testing.T) {
		store := newRevokedStore()

		if store.IsRevoked("user1") {
			t.Error("expected user to not be revoked initially")
		}
	})

	t.Run("Revoked after calling Revoke", func(t *testing.T) {
		store := newRevokedStore()

		store.Revoke("user1")

		if !store.IsRevoked("user1") {
			t.Error("expected user to be revoked")
		}
	})

	t.Run("Different users independent", func(t *testing.T) {
		store := newRevokedStore()

		store.Revoke("user1")

		if store.IsRevoked("user2") {
			t.Error("expected user2 to not be revoked")
		}
	})

	t.Run("Empty string not revoked", func(t *testing.T) {
		store := newRevokedStore()

		store.Revoke("")

		if store.IsRevoked("") {
			t.Error("expected empty string to not be considered revoked")
		}
	})
}

// ===== CORS Tests =====

func TestParseAllowedOrigins(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"Empty string", "", []string{"*"}},
		{"Wildcard", "*", []string{"*"}},
		{"Single origin", "https://example.com", []string{"https://example.com"}},
		{"Multiple origins", "https://a.com,https://b.com", []string{"https://a.com", "https://b.com"}},
		{"With spaces", "  https://a.com , https://b.com  ", []string{"https://a.com", "https://b.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAllowedOrigins(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d origins, got %d", len(tt.expected), len(result))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("expected origin[%d] = '%s', got '%s'", i, tt.expected[i], v)
				}
			}
		})
	}
}

func TestMatchOrigin(t *testing.T) {
	tests := []struct {
		name            string
		allowedOrigins  []string
		requestOrigin   string
		expectedAllowed string
	}{
		{"Wildcard allows any", []string{"*"}, "https://example.com", "https://example.com"},
		{"Wildcard with no origin", []string{"*"}, "", "*"},
		{"Exact match", []string{"https://a.com", "https://b.com"}, "https://a.com", "https://a.com"},
		{"Second match", []string{"https://a.com", "https://b.com"}, "https://b.com", "https://b.com"},
		{"No match returns first", []string{"https://a.com", "https://b.com"}, "https://c.com", "https://a.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.requestOrigin != "" {
				req.Header.Set("Origin", tt.requestOrigin)
			}

			result := matchOrigin(req, tt.allowedOrigins)
			if result != tt.expectedAllowed {
				t.Errorf("expected '%s', got '%s'", tt.expectedAllowed, result)
			}
		})
	}
}

// ===== HTTP Handler Tests =====

func TestHealthEndpoint(t *testing.T) {
	// Create a minimal handler for testing
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeCORS(w, r, []string{"*"})
		writeJSON(w, http.StatusOK, jsonMap{"ok": true, "env": "test"})
	})

	t.Run("Returns OK", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["ok"] != true {
			t.Error("expected ok: true")
		}
	})

	t.Run("Sets CORS headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set("Origin", "https://example.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Header().Get("Access-Control-Allow-Origin") == "" {
			t.Error("expected Access-Control-Allow-Origin header")
		}
	})
}

func TestPreflightHandling(t *testing.T) {
	allowedOrigins := []string{"https://example.com"}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if preflight(w, r, allowedOrigins) {
			return
		}
		writeJSON(w, http.StatusOK, jsonMap{"ok": true})
	})

	t.Run("OPTIONS returns no content", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", rec.Code)
		}
	})

	t.Run("GET proceeds normally", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})
}

func TestAuthMiddleware(t *testing.T) {
	secret := "test-secret"
	revoked := newRevokedStore()

	protectedHandler := auth(secret, revoked)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := mustClaims(r.Context())
		writeJSON(w, http.StatusOK, jsonMap{"userId": claims.Sub})
	}))

	t.Run("Missing Authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()

		protectedHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("Invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()

		protectedHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("Valid token", func(t *testing.T) {
		token := signSession(secret, "test-user", time.Hour)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		protectedHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(rec.Body).Decode(&resp)
		if resp["userId"] != "test-user" {
			t.Errorf("expected userId 'test-user', got '%v'", resp["userId"])
		}
	})

	t.Run("Revoked session", func(t *testing.T) {
		revoked.Revoke("revoked-user")
		token := signSession(secret, "revoked-user", time.Hour)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		protectedHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401 for revoked session, got %d", rec.Code)
		}
	})
}

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusOK, jsonMap{"key": "value"})

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["key"] != "value" {
		t.Errorf("expected key='value', got '%s'", resp["key"])
	}
}

func TestWriteErr(t *testing.T) {
	rec := httptest.NewRecorder()
	writeErr(rec, http.StatusBadRequest, "test error")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp["error"] != "test error" {
		t.Errorf("expected error='test error', got '%s'", resp["error"])
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		expected   string
	}{
		{"With port", "192.168.1.1:12345", "192.168.1.1"},
		{"Without port", "192.168.1.1", "192.168.1.1"},
		{"IPv6 with port", "[::1]:12345", "::1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr

			result := clientIP(req)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestShortHash(t *testing.T) {
	hash1 := shortHash("test-input")
	hash2 := shortHash("test-input")
	hash3 := shortHash("different-input")

	if hash1 != hash2 {
		t.Error("expected same input to produce same hash")
	}

	if hash1 == hash3 {
		t.Error("expected different input to produce different hash")
	}

	if len(hash1) != 12 { // 6 bytes = 12 hex chars
		t.Errorf("expected hash length 12, got %d", len(hash1))
	}
}
