package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/google/uuid"

	"habitcashback/internal/auth"
	"habitcashback/internal/httpx"
	"habitcashback/internal/store"
	"habitcashback/internal/toss"
)

type Server struct {
	Env       string
	JWTSecret string
	Store     store.Store
	Toss      *toss.Client
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // tighten in production (allow your toss app origin only)
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Idempotency-Key"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_ = s.Store.Ping(r.Context())
		httpx.WriteJSON(w, 200, map[string]any{"ok": true})
	})

	r.Route("/v1", func(r chi.Router) {
		r.Post("/auth/exchange", s.handleAuthExchange)

		r.Group(func(r chi.Router) {
			r.Use(s.authMiddleware)
			r.Post("/payments/create", s.handlePaymentCreate)
			r.Post("/payments/execute", s.handlePaymentExecute)

			r.Post("/payouts/issue", s.handlePayoutIssue)
			r.Post("/payouts/result", s.handlePayoutResult)

			r.Post("/messages/send", s.handleMessageSend)
		})
	})

	return r
}

type AuthExchangeReq struct {
	AuthorizationCode string `json:"authorizationCode"`
	Referrer          string `json:"referrer"`
}

func (s *Server) handleAuthExchange(w http.ResponseWriter, r *http.Request) {
	var req AuthExchangeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "invalid_json", err.Error())
		return
	}
	if req.AuthorizationCode == "" || req.Referrer == "" {
		httpx.BadRequest(w, "missing_fields", map[string]string{"authorizationCode": "required", "referrer": "required"})
		return
	}

	ctx := r.Context()

	tok, rawTok, err := s.Toss.GenerateToken(ctx, toss.GenerateTokenReq{
		AuthorizationCode: req.AuthorizationCode,
		Referrer:          req.Referrer,
	})
	if err != nil {
		httpx.ServerError(w, "toss_generate_token_failed")
		return
	}

	me, rawMe, err := s.Toss.LoginMe(ctx, tok.AccessToken)
	if err != nil {
		httpx.ServerError(w, "toss_login_me_failed")
		return
	}

	tossUserKey := int64ToString(me.UserKey)
	userID, err := s.Store.UpsertUser(ctx, tossUserKey)
	if err != nil {
		httpx.ServerError(w, "db_upsert_user_failed")
		return
	}

	appToken, err := auth.MintToken(s.JWTSecret, userID, tossUserKey, 24*time.Hour)
	if err != nil {
		httpx.ServerError(w, "mint_token_failed")
		return
	}

	resp := map[string]any{
		"token":   appToken,
		"userKey": tossUserKey,
	}
	if s.Env == "dev" {
		resp["rawToken"] = json.RawMessage(rawTok)
		resp["rawLoginMe"] = json.RawMessage(rawMe)
	}
	httpx.WriteJSON(w, 200, resp)
}

type ctxKey string

const ctxClaimsKey ctxKey = "claims"

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h == "" || !strings.HasPrefix(strings.ToLower(h), "bearer ") {
			httpx.Unauthorized(w, "missing_bearer_token")
			return
		}
		tokenStr := strings.TrimSpace(h[len("Bearer "):])
		claims, err := auth.ParseToken(s.JWTSecret, tokenStr)
		if err != nil {
			httpx.Unauthorized(w, "invalid_token")
			return
		}
		ctx := context.WithValue(r.Context(), ctxClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func claimsFrom(r *http.Request) *auth.Claims {
	if v := r.Context().Value(ctxClaimsKey); v != nil {
		if c, ok := v.(*auth.Claims); ok {
			return c
		}
	}
	return nil
}

type PaymentCreateReq struct {
	OrderNo       string `json:"orderNo,omitempty"`
	ProductDesc   string `json:"productDesc"`
	Amount        int64  `json:"amount"`
	IsTestPayment bool   `json:"isTestPayment"`
}

func (s *Server) handlePaymentCreate(w http.ResponseWriter, r *http.Request) {
	claims := claimsFrom(r)
	if claims == nil {
		httpx.Unauthorized(w, "missing_claims")
		return
	}

	var req PaymentCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "invalid_json", err.Error())
		return
	}
	if req.ProductDesc == "" || req.Amount <= 0 {
		httpx.BadRequest(w, "invalid_fields", nil)
		return
	}

	if req.OrderNo == "" {
		req.OrderNo = "order-" + uuid.NewString()
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		idemKey = "mp-" + uuid.NewString()
	}

	// Global idempotency: if key repeats, return cached response
	if found, cached, err := s.Store.GetIdempotency(r.Context(), "make-payment", idemKey); err == nil && found {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200)
		_, _ = w.Write(cached)
		return
	}

	success, raw, err := s.Toss.MakePayment(r.Context(), claims.TossUserKey, idemKey, toss.MakePaymentReq{
		OrderNo:       req.OrderNo,
		ProductDesc:   req.ProductDesc,
		Amount:        req.Amount,
		AmountTaxFree: 0,
		IsTestPayment: req.IsTestPayment,
	})
	if err != nil {
		httpx.ServerError(w, "toss_make_payment_failed")
		return
	}

	_ = s.Store.InsertPayment(r.Context(), store.Payment{
		UserID:   claims.UserID,
		OrderNo:  req.OrderNo,
		PayToken: success.PayToken,
		Amount:   req.Amount,
		Status:   "CREATED",
		RawJSON:  raw,
	})

	resp := map[string]any{
		"orderNo":  req.OrderNo,
		"payToken": success.PayToken,
	}

	respBytes, _ := json.Marshal(resp)
	_ = s.Store.PutIdempotency(r.Context(), "make-payment", idemKey, respBytes)

	httpx.WriteJSON(w, 200, resp)
}

type PaymentExecuteReq struct {
	OrderNo       string `json:"orderNo"`
	PayToken      string `json:"payToken"`
	IsTestPayment bool   `json:"isTestPayment"`
}

func (s *Server) handlePaymentExecute(w http.ResponseWriter, r *http.Request) {
	claims := claimsFrom(r)
	if claims == nil {
		httpx.Unauthorized(w, "missing_claims")
		return
	}
	var req PaymentExecuteReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "invalid_json", err.Error())
		return
	}
	if req.OrderNo == "" || req.PayToken == "" {
		httpx.BadRequest(w, "missing_fields", nil)
		return
	}

	success, raw, err := s.Toss.ExecutePayment(r.Context(), claims.TossUserKey, toss.ExecutePaymentReq{
		PayToken:      req.PayToken,
		OrderNo:       req.OrderNo,
		IsTestPayment: req.IsTestPayment,
	})
	if err != nil {
		httpx.ServerError(w, "toss_execute_payment_failed")
		return
	}

	// Record status update as raw log (simple MVP)
	_ = s.Store.InsertPayment(r.Context(), store.Payment{
		UserID:   claims.UserID,
		OrderNo:  req.OrderNo,
		PayToken: req.PayToken,
		Amount:   0,
		Status:   "EXECUTED",
		RawJSON:  raw,
	})

	httpx.WriteJSON(w, 200, map[string]any{
		"result": success,
		"raw":    json.RawMessage(raw),
	})
}

type PayoutIssueReq struct {
	PromotionCode string `json:"promotionCode"`
	AmountPoints  int64  `json:"amountPoints"`
}

func (s *Server) handlePayoutIssue(w http.ResponseWriter, r *http.Request) {
	claims := claimsFrom(r)
	if claims == nil {
		httpx.Unauthorized(w, "missing_claims")
		return
	}

	var req PayoutIssueReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "invalid_json", err.Error())
		return
	}
	if req.PromotionCode == "" || req.AmountPoints <= 0 {
		httpx.BadRequest(w, "invalid_fields", nil)
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		idemKey = "po-" + uuid.NewString()
	}

	if found, cached, err := s.Store.GetIdempotency(r.Context(), "payout-issue", idemKey); err == nil && found {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200)
		_, _ = w.Write(cached)
		return
	}

	keyRes, rawKey, err := s.Toss.PromotionGetKey(r.Context(), claims.TossUserKey, toss.GetKeyReq{PromotionCode: req.PromotionCode})
	if err != nil {
		httpx.ServerError(w, "toss_get_key_failed")
		return
	}

	execRes, rawExec, err := s.Toss.PromotionExecute(r.Context(), claims.TossUserKey, idemKey, toss.ExecutePromotionReq{
		Key:   keyRes.Key,
		Value: req.AmountPoints,
	})
	if err != nil {
		httpx.ServerError(w, "toss_execute_promotion_failed")
		return
	}

	// Insert payout row as REQUESTED (worker will finalize via execution-result)
	_ = s.Store.InsertPayout(r.Context(), store.Payout{
		UserID:        claims.UserID,
		PromotionCode: req.PromotionCode,
		PromotionKey:  keyRes.Key,
		AmountPoints:  req.AmountPoints,
		Status:        "REQUESTED",
		RawJSON:       mergeRaw(rawKey, rawExec),
	})

	resp := map[string]any{
		"promotionKey": keyRes.Key,
		"execute":      execRes,
		"rawKey":       json.RawMessage(rawKey),
		"rawExecute":   json.RawMessage(rawExec),
	}
	respBytes, _ := json.Marshal(resp)
	_ = s.Store.PutIdempotency(r.Context(), "payout-issue", idemKey, respBytes)

	httpx.WriteJSON(w, 200, resp)
}

type PayoutResultReq struct {
	PromotionKey string `json:"promotionKey"`
}

func (s *Server) handlePayoutResult(w http.ResponseWriter, r *http.Request) {
	claims := claimsFrom(r)
	if claims == nil {
		httpx.Unauthorized(w, "missing_claims")
		return
	}
	var req PayoutResultReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "invalid_json", err.Error())
		return
	}
	if req.PromotionKey == "" {
		httpx.BadRequest(w, "missing_fields", nil)
		return
	}

	res, raw, err := s.Toss.PromotionExecutionResult(r.Context(), claims.TossUserKey, toss.ExecutionResultReq{Key: req.PromotionKey})
	if err != nil {
		httpx.ServerError(w, "toss_execution_result_failed")
		return
	}

	// Update status best-effort
	status := normalizeResultType(res.ResultType)
	_ = s.Store.UpdatePayoutStatus(r.Context(), req.PromotionKey, status, raw)

	httpx.WriteJSON(w, 200, map[string]any{
		"status": status,
		"raw":    json.RawMessage(raw),
	})
}

type MessageSendReq struct {
	TemplateSetCode string         `json:"templateSetCode"`
	Context         map[string]any `json:"context"`
}

func (s *Server) handleMessageSend(w http.ResponseWriter, r *http.Request) {
	claims := claimsFrom(r)
	if claims == nil {
		httpx.Unauthorized(w, "missing_claims")
		return
	}
	var req MessageSendReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "invalid_json", err.Error())
		return
	}
	if req.TemplateSetCode == "" || req.Context == nil {
		httpx.BadRequest(w, "missing_fields", nil)
		return
	}

	success, raw, err := s.Toss.SendMessage(r.Context(), claims.TossUserKey, toss.SendMessageReq{
		TemplateSetCode: req.TemplateSetCode,
		Context:         req.Context,
	})
	if err != nil {
		httpx.ServerError(w, "toss_send_message_failed")
		return
	}

	httpx.WriteJSON(w, 200, map[string]any{
		"result": success,
		"raw":    json.RawMessage(raw),
	})
}

func int64ToString(n int64) string {
	return strconv.FormatInt(n, 10)
}

func normalizeResultType(rt string) string {
	switch strings.ToUpper(rt) {
	case "SUCCESS":
		return "SUCCESS"
	case "FAIL", "EXECUTION_FAIL", "INTERNAL_ERROR":
		return "FAIL"
	case "HTTP_TIMEOUT", "NETWORK_ERROR", "INTERRUPTED":
		return "PENDING"
	default:
		return "PENDING"
	}
}

func mergeRaw(a, b []byte) []byte {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	out, _ := json.Marshal(map[string]json.RawMessage{
		"key":     json.RawMessage(a),
		"execute": json.RawMessage(b),
	})
	return out
}
