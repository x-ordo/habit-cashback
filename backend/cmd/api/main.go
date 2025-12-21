package main

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"habitcashback/internal/toss"
)

type jsonMap map[string]any

func main() {
	port := getenv("PORT", "8080")
	allowOriginsRaw := getenv("ALLOW_ORIGIN", "*")
	appEnv := getenv("APP_ENV", "local")
	version := getenv("APP_VERSION", "dev")
	commit := getenv("GIT_SHA", "local")

	// Parse ALLOW_ORIGIN: comma-separated list or "*"
	// Example: "https://habitcashback.apps.tossmini.com,https://habitcashback.private-apps.tossmini.com"
	allowedOrigins := parseAllowedOrigins(allowOriginsRaw)

	secret := strings.TrimSpace(os.Getenv("SESSION_SECRET"))
	if secret == "" {
		if appEnv == "local" {
			secret = mustRandomHex(32) // dev convenience
			log.Printf("[warn] SESSION_SECRET not set. Using ephemeral local secret: %s...", secret[:8])
		} else {
			log.Fatal("SESSION_SECRET is required in non-local environments")
		}
	}

	// Idempotency and simple rate limiting (MVP hardening)
	idem := newIdemStore()
	rl := newRateLimiter(120, time.Minute) // 120 req/min per IP

	// Apps in Toss mTLS client (optional in local; required in staging/prod)
	var tossClient *toss.Client
	if c, err := toss.NewFromEnv(); err != nil {
		log.Printf("[warn] toss mTLS client disabled: %v", err)
		tossClient = nil
	} else {
		tossClient = c
	}

	mux := http.NewServeMux()

	// ---- Health/meta
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if preflight(w, r, allowedOrigins) {
			return
		}
		writeCORS(w, r, allowedOrigins)
		writeJSON(w, http.StatusOK, jsonMap{"ok": true, "env": appEnv, "ts": time.Now().UTC().Format(time.RFC3339)})
	})
	mux.HandleFunc("/meta", func(w http.ResponseWriter, r *http.Request) {
		if preflight(w, r, allowedOrigins) {
			return
		}
		writeCORS(w, r, allowedOrigins)
		writeJSON(w, http.StatusOK, jsonMap{"name": "habitcashback-api", "version": version, "commit": commit, "links": jsonMap{"terms": "/terms", "privacy": "/privacy", "support": "/support"}, "toss": jsonMap{"unlinkCallback": "/v1/auth/toss/unlink-callback"}})
	})

	// ---- Auth (stub)
	mux.HandleFunc("/v1/auth/exchange", func(w http.ResponseWriter, r *http.Request) {
		if preflight(w, r, allowedOrigins) {
			return
		}
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeCORS(w, r, allowedOrigins)

		// Accept any body (provider swap later)
		userID := "stub-user"
		sess := signSession(secret, userID, 24*time.Hour)

		writeJSON(w, http.StatusOK, jsonMap{
			"sessionToken": sess,
			"accessToken":  sess, // compatibility alias
			"mode":         "stub",
			"expiresIn":    24 * 3600,
		})
	})

	// ---- Auth (Apps in Toss mTLS)
	mux.HandleFunc("/v1/auth/toss/exchange", func(w http.ResponseWriter, r *http.Request) {
		if preflight(w, r, allowedOrigins) {
			return
		}
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeCORS(w, r, allowedOrigins)

		var body struct {
			AuthorizationCode string `json:"authorizationCode"`
			Referrer          string `json:"referrer"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid json body")
			return
		}
		if strings.TrimSpace(body.AuthorizationCode) == "" {
			writeErr(w, http.StatusBadRequest, "authorizationCode is required")
			return
		}

		// local fallback (no mTLS)
		if tossClient == nil && appEnv == "local" {
			userID := "stub-user"
			sess := signSession(secret, userID, 24*time.Hour)
			writeJSON(w, http.StatusOK, jsonMap{
				"sessionToken": sess,
				"accessToken":  sess,
				"mode":         "stub",
				"expiresIn":    24 * 3600,
			})
			return
		}
		if tossClient == nil {
			writeErr(w, http.StatusBadGateway, "toss mTLS not configured on server")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
		defer cancel()

		// Exchange authorizationCode -> Toss access token (we don't expose it to frontend; we mint our own session token)
		success, err := tossClient.GenerateUserToken(ctx, body.AuthorizationCode, body.Referrer)
		if err != nil {
			writeErr(w, http.StatusBadGateway, "toss exchange failed")
			return
		}

		uid := "toss:" + shortHash(success.AccessToken)
		if me, err := tossClient.LoginMe(ctx, success.AccessToken); err == nil && me != nil && me.UserKey > 0 {
			uid = fmt.Sprintf("toss:%d", me.UserKey)
		}
		sess := signSession(secret, uid, 24*time.Hour)

		writeJSON(w, http.StatusOK, jsonMap{
			"sessionToken": sess,
			"accessToken":  sess,
			"mode":         "toss",
			"expiresIn":    24 * 3600,
		})
	})



// ---- Auth (unlink callback configured in Toss console)
mux.HandleFunc("/v1/auth/toss/unlink-callback", func(w http.ResponseWriter, r *http.Request) {
	if preflight(w, r, allowedOrigins) {
		return
	}
	writeCORS(w, r, allowedOrigins)

	// Optional Basic Auth: set AIT_UNLINK_BASIC_AUTH="username:password"
	expect := strings.TrimSpace(os.Getenv("AIT_UNLINK_BASIC_AUTH"))
	if expect != "" {
		ah := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(strings.ToLower(ah), "basic ") {
			writeErr(w, http.StatusUnauthorized, "missing basic auth")
			return
		}
		raw := strings.TrimSpace(ah[len("Basic "):])
		dec, err := base64.StdEncoding.DecodeString(raw)
		if err != nil || string(dec) != expect {
			writeErr(w, http.StatusUnauthorized, "invalid basic auth")
			return
		}
	}

	var userKey int64
	var referrer string

	switch r.Method {
	case http.MethodGet:
		q := r.URL.Query()
		userKeyStr := strings.TrimSpace(q.Get("userKey"))
		referrer = strings.TrimSpace(q.Get("referrer"))
		if userKeyStr == "" {
			writeErr(w, http.StatusBadRequest, "userKey is required")
			return
		}
		uk, err := strconv.ParseInt(userKeyStr, 10, 64)
		if err != nil || uk <= 0 {
			writeErr(w, http.StatusBadRequest, "invalid userKey")
			return
		}
		userKey = uk
	case http.MethodPost:
		var body struct {
			UserKey   int64  `json:"userKey"`
			Referrer string `json:"referrer"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid json")
			return
		}
		if body.UserKey <= 0 {
			writeErr(w, http.StatusBadRequest, "userKey is required")
			return
		}
		userKey = body.UserKey
		referrer = strings.TrimSpace(body.Referrer)
	default:
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	sub := fmt.Sprintf("toss:%d", userKey)
	revoked.Revoke(sub)

	writeJSON(w, http.StatusOK, jsonMap{
		"ok":       true,
		"userKey":  userKey,
		"referrer": strings.ToUpper(referrer),
	})
})

	// ---- Protected endpoints
	mux.Handle("/v1/me", auth(secret, revoked)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeCORS(w, r, allowedOrigins)
		claims := mustClaims(r.Context())
		writeJSON(w, http.StatusOK, jsonMap{"userId": claims.Sub, "exp": claims.Exp})
	})))

	mux.Handle("/v1/challenges", auth(secret, revoked)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if preflight(w, r, allowedOrigins) {
			return
		}
		if r.Method != http.MethodGet {
			writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeCORS(w, r, allowedOrigins)

		writeJSON(w, http.StatusOK, jsonMap{
			"items": []jsonMap{
				{"id": "walk-7000", "title": "매일 7,000보 걷기", "days": 3, "deposit": 10000, "proofType": "steps"},
				{"id": "bed-0700", "title": "아침 7시 이불 개기", "days": 3, "deposit": 10000, "proofType": "photo"},
				{"id": "lunch-proof", "title": "점심 도시락/샐러드 인증", "days": 3, "deposit": 10000, "proofType": "photo"},
			},
		})
	})))

	mux.Handle("/v1/payments/create", auth(secret, revoked)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if preflight(w, r, allowedOrigins) {
			return
		}
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeCORS(w, r, allowedOrigins)

		idk := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if idk == "" {
			idk = "auto-" + mustRandomHex(8)
		}
		if !idem.TryUse("paycreate:"+idk, 2*time.Minute) {
			writeErr(w, http.StatusConflict, "duplicate request")
			return
		}

		var body struct {
			ChallengeID string `json:"challengeId"`
			Amount      int    `json:"amount"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid json body")
			return
		}
		if strings.TrimSpace(body.ChallengeID) == "" || body.Amount <= 0 {
			writeErr(w, http.StatusBadRequest, "challengeId and amount are required")
			return
		}

		paymentID := "pay_" + mustRandomHex(8)
		writeJSON(w, http.StatusOK, jsonMap{
			"paymentId":  paymentID,
			"status":     "created",
			"challengeId": body.ChallengeID,
			"amount":     body.Amount,
		})
	})))

	mux.Handle("/v1/payments/execute", auth(secret, revoked)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if preflight(w, r, allowedOrigins) {
			return
		}
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeCORS(w, r, allowedOrigins)

		var body struct {
			PaymentID string `json:"paymentId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid json body")
			return
		}
		if strings.TrimSpace(body.PaymentID) == "" {
			writeErr(w, http.StatusBadRequest, "paymentId is required")
			return
		}

		writeJSON(w, http.StatusOK, jsonMap{"ok": true, "status": "done", "paymentId": body.PaymentID})
	})))

	mux.Handle("/v1/proofs/submit", auth(secret, revoked)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if preflight(w, r, allowedOrigins) {
			return
		}
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeCORS(w, r, allowedOrigins)

		idk := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if idk == "" {
			idk = "auto-" + mustRandomHex(8)
		}
		if !idem.TryUse("proof:"+idk, 2*time.Minute) {
			writeErr(w, http.StatusConflict, "duplicate request")
			return
		}

		var body struct {
			ChallengeID string `json:"challengeId"`
			ImageBase64 string `json:"imageBase64"`
			ImageHash   string `json:"imageHash"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid json body")
			return
		}
		if strings.TrimSpace(body.ChallengeID) == "" {
			writeErr(w, http.StatusBadRequest, "challengeId is required")
			return
		}
		// MVP: accept base64 or hash (steps)
		if body.ImageBase64 == "" && body.ImageHash == "" {
			writeErr(w, http.StatusBadRequest, "imageBase64 or imageHash is required")
			return
		}

		writeJSON(w, http.StatusOK, jsonMap{"ok": true, "status": "accepted"})
	})))

	// Global wrapper (security headers + req id + rate limit + log)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// preflight short-circuit
		if r.Method == http.MethodOptions {
			writeCORS(w, r, allowedOrigins)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")

		// request id
		reqID := strings.TrimSpace(r.Header.Get("X-Request-Id"))
		if reqID == "" {
			reqID = "req_" + mustRandomHex(8)
		}
		w.Header().Set("X-Request-Id", reqID)

		// rate limit (per IP)
		ip := clientIP(r)
		if !rl.Allow(ip) {
			writeCORS(w, r, allowedOrigins)
			writeErr(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		start := time.Now()
		mux.ServeHTTP(w, r)
		log.Printf("%s %s ip=%s rid=%s dur=%s", r.Method, r.URL.Path, ip, reqID, time.Since(start))
	})

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("api listening :%s env=%s", port, appEnv)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

// ===== Auth middleware (stateless signed session) =====

type Claims struct {
	Sub string `json:"sub"`
	Iat int64  `json:"iat"`
	Exp int64  `json:"exp"`
}

type ctxKey int

const claimsKey ctxKey = 1

func auth(secret string, revoked *revokedStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(h, "Bearer ") {
				writeErr(w, http.StatusUnauthorized, "missing bearer token")
				return
			}
			tok := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
			c, err := verifySession(secret, tok)
			if err != nil {
				writeErr(w, http.StatusUnauthorized, "invalid token")
				return
			}
			if revoked != nil && revoked.IsRevoked(c.Sub) {
				writeErr(w, http.StatusUnauthorized, "session revoked (unlinked)")
				return
			}
			ctx := context.WithValue(r.Context(), claimsKey, c)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func mustClaims(ctx context.Context) Claims {
	v := ctx.Value(claimsKey)
	if v == nil {
		return Claims{}
	}
	return v.(Claims)
}

func signSession(secret, userID string, ttl time.Duration) string {
	now := time.Now().UTC()
	c := Claims{
		Sub: userID,
		Iat: now.Unix(),
		Exp: now.Add(ttl).Unix(),
	}
	b, _ := json.Marshal(c)
	payload := base64.RawURLEncoding.EncodeToString(b)
	sig := hmacSHA256(secret, payload)
	return "sv1." + payload + "." + sig
}

func verifySession(secret, token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != "sv1" {
		return Claims{}, errors.New("bad token format")
	}
	payload := parts[1]
	sig := parts[2]
	if !hmac.Equal([]byte(sig), []byte(hmacSHA256(secret, payload))) {
		return Claims{}, errors.New("bad signature")
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return Claims{}, err
	}
	var c Claims
	if err := json.Unmarshal(raw, &c); err != nil {
		return Claims{}, err
	}
	if c.Sub == "" || c.Exp == 0 {
		return Claims{}, errors.New("bad claims")
	}
	if time.Now().UTC().Unix() > c.Exp {
		return Claims{}, errors.New("expired")
	}
	return c, nil
}

func hmacSHA256(secret, msg string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msg))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// ===== Rate limiter (simple fixed window) =====

type rateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	buckets map[string]*bucket
}

type bucket struct {
	count int
	reset time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		limit:   limit,
		window:  window,
		buckets: map[string]*bucket{},
	}
}

func (r *rateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	b, ok := r.buckets[key]
	if !ok || now.After(b.reset) {
		r.buckets[key] = &bucket{count: 1, reset: now.Add(r.window)}
		return true
	}
	if b.count >= r.limit {
		return false
	}
	b.count++
	return true
}

// ===== Idempotency store =====


type revokedStore struct {
	mu      sync.RWMutex
	revoked map[string]time.Time
}

func newRevokedStore() *revokedStore {
	return &revokedStore{revoked: map[string]time.Time{}}
}

func (s *revokedStore) Revoke(sub string) {
	sub = strings.TrimSpace(sub)
	if sub == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revoked[sub] = time.Now().UTC()
}

func (s *revokedStore) IsRevoked(sub string) bool {
	sub = strings.TrimSpace(sub)
	if sub == "" {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.revoked[sub]
	return ok
}

type idemStore struct {
	mu    sync.Mutex
	items map[string]time.Time
}

func newIdemStore() *idemStore {
	return &idemStore{items: map[string]time.Time{}}
}

func (s *idemStore) TryUse(key string, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for k, exp := range s.items {
		if now.After(exp) {
			delete(s.items, k)
		}
	}
	if exp, ok := s.items[key]; ok && now.Before(exp) {
		return false
	}
	s.items[key] = now.Add(ttl)
	return true
}

// ===== HTTP helpers =====

// parseAllowedOrigins parses comma-separated origins or "*"
func parseAllowedOrigins(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "*" {
		return []string{"*"}
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			origins = append(origins, p)
		}
	}
	if len(origins) == 0 {
		return []string{"*"}
	}
	return origins
}

// matchOrigin checks if request origin is allowed and returns the origin to use
func matchOrigin(r *http.Request, allowedOrigins []string) string {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		if origin != "" {
			return origin
		}
		return "*"
	}
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return origin
		}
	}
	// No match - return first allowed origin (fallback)
	if len(allowedOrigins) > 0 {
		return allowedOrigins[0]
	}
	return "*"
}

func preflight(w http.ResponseWriter, r *http.Request, allowedOrigins []string) bool {
	if r.Method == http.MethodOptions {
		writeCORS(w, r, allowedOrigins)
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

func writeCORS(w http.ResponseWriter, r *http.Request, allowedOrigins []string) {
	origin := matchOrigin(r, allowedOrigins)
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-Id, Idempotency-Key")
	w.Header().Set("Access-Control-Expose-Headers", "X-Request-Id")
	// Vary header for proper caching when using dynamic origin
	if origin != "*" {
		w.Header().Set("Vary", "Origin")
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, jsonMap{"error": msg})
}

func getenv(k, def string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	return v
}

func mustRandomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}

func shortHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", sum[:6])
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
