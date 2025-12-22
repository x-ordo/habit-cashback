package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"habitcashback/internal/store"
)

func main() {
	// Parse command line flags
	runOnce := flag.Bool("once", false, "Run all jobs once and exit")
	jobName := flag.String("job", "", "Run specific job: close-participations, update-settlements, cleanup-idempotency, cleanup-sessions, stats")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("[worker] starting habitcashback batch worker")

	// Check DATABASE_URL
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		log.Fatal("[worker] DATABASE_URL is required")
	}

	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	db, err := store.New(ctx)
	cancel()
	if err != nil {
		log.Fatalf("[worker] database connection failed: %v", err)
	}
	defer db.Close()
	log.Println("[worker] database connected")

	// Run specific job if requested
	if *jobName != "" {
		runJob(db, *jobName)
		return
	}

	// Run all jobs once if requested
	if *runOnce {
		runAllJobs(db)
		return
	}

	// Start scheduled jobs
	log.Println("[worker] starting scheduled jobs")
	log.Println("[worker] - close-participations: every day at 00:05")
	log.Println("[worker] - update-settlements: every day at 00:10")
	log.Println("[worker] - cleanup-idempotency: every hour")
	log.Println("[worker] - cleanup-sessions: every day at 03:00")

	// Create stop channel
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Start job runners
	go runDailyJob(db, "close-participations", 0, 5, closeParticipations)
	go runDailyJob(db, "update-settlements", 0, 10, updateSettlements)
	go runHourlyJob(db, "cleanup-idempotency", cleanupIdempotency)
	go runDailyJob(db, "cleanup-sessions", 3, 0, cleanupSessions)

	// Wait for shutdown signal
	<-stop
	log.Println("[worker] shutting down...")
}

func runJob(db *store.Store, jobName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	switch jobName {
	case "close-participations":
		closeParticipations(ctx, db)
	case "update-settlements":
		updateSettlements(ctx, db)
	case "cleanup-idempotency":
		cleanupIdempotency(ctx, db)
	case "cleanup-sessions":
		cleanupSessions(ctx, db)
	case "stats":
		showStats(ctx, db)
	default:
		log.Fatalf("[worker] unknown job: %s", jobName)
	}
}

func runAllJobs(db *store.Store) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	log.Println("[worker] running all jobs once")
	closeParticipations(ctx, db)
	updateSettlements(ctx, db)
	cleanupIdempotency(ctx, db)
	cleanupSessions(ctx, db)
	showStats(ctx, db)
	log.Println("[worker] all jobs completed")
}

func runDailyJob(db *store.Store, name string, hour, minute int, fn func(context.Context, *store.Store)) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}
		wait := next.Sub(now)
		log.Printf("[worker] %s: next run at %s (in %s)", name, next.Format("2006-01-02 15:04:05"), wait.Round(time.Second))

		time.Sleep(wait)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		fn(ctx, db)
		cancel()
	}
}

func runHourlyJob(db *store.Store, name string, fn func(context.Context, *store.Store)) {
	// Run immediately on startup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	fn(ctx, db)
	cancel()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		fn(ctx, db)
		cancel()
	}
}

func closeParticipations(ctx context.Context, db *store.Store) {
	log.Println("[job:close-participations] starting")
	result, err := db.CloseExpiredParticipations(ctx)
	if err != nil {
		log.Printf("[job:close-participations] error: %v", err)
		return
	}
	log.Printf("[job:close-participations] completed: processed=%d, failed=%d", result.Processed, result.Failed)
	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			log.Printf("[job:close-participations] error detail: %s", e)
		}
	}
}

func updateSettlements(ctx context.Context, db *store.Store) {
	log.Println("[job:update-settlements] starting")
	result, err := db.UpdateSettlementStatuses(ctx)
	if err != nil {
		log.Printf("[job:update-settlements] error: %v", err)
		return
	}
	log.Printf("[job:update-settlements] completed: processed=%d", result.Processed)
}

func cleanupIdempotency(ctx context.Context, db *store.Store) {
	log.Println("[job:cleanup-idempotency] starting")
	result, err := db.CleanupExpiredIdempotencyKeys(ctx)
	if err != nil {
		log.Printf("[job:cleanup-idempotency] error: %v", err)
		return
	}
	log.Printf("[job:cleanup-idempotency] completed: deleted=%d", result.Processed)
}

func cleanupSessions(ctx context.Context, db *store.Store) {
	log.Println("[job:cleanup-sessions] starting")
	result, err := db.CleanupOldRevokedSessions(ctx)
	if err != nil {
		log.Printf("[job:cleanup-sessions] error: %v", err)
		return
	}
	log.Printf("[job:cleanup-sessions] completed: deleted=%d", result.Processed)
}

func showStats(ctx context.Context, db *store.Store) {
	log.Println("[job:stats] fetching batch statistics")
	stats, err := db.GetBatchStats(ctx)
	if err != nil {
		log.Printf("[job:stats] error: %v", err)
		return
	}
	log.Printf("[job:stats] active_participations=%d, running_settlements=%d, idempotency_keys=%d, revoked_sessions=%d",
		stats.ActiveParticipations, stats.RunningSettlements, stats.PendingIdempotencyKeys, stats.RevokedSessions)
}
