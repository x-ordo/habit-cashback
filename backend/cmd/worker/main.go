package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"habitcashback/internal/config"
	"habitcashback/internal/store/postgres"
	"habitcashback/internal/toss"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required (Postgres).")
	}

	st, err := postgres.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	tc, err := toss.NewClient(cfg.TossAPIBaseURL, cfg.TossPayBaseURL, cfg.HTTPTimeout)
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Printf("worker started (poll payouts) interval=%s", 30*time.Second)

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_ = reconcilePayouts(ctx, st, tc)
		cancel()
	}
}

func reconcilePayouts(ctx context.Context, st *postgres.Postgres, tc *toss.Client) error {
	payouts, err := st.ListPendingPayouts(ctx, 50)
	if err != nil {
		return err
	}
	for _, p := range payouts {
		res, raw, err := tc.PromotionExecutionResult(ctx, p.TossUserKey, toss.ExecutionResultReq{Key: p.PromotionKey})
		if err != nil {
			continue
		}
		status := strings.ToUpper(res.ResultType)
		if status == "" {
			status = "PENDING"
		}
		_ = st.UpdatePayoutStatus(ctx, p.PromotionKey, status, raw)
	}
	return nil
}
