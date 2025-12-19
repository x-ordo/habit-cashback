package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"habitcashback/internal/config"
	"habitcashback/internal/server"
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

	s := &server.Server{
		Env:       cfg.Env,
		JWTSecret: cfg.JWTSecret,
		Store:     st,
		Toss:      tc,
	}

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           s.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("api listening on %s env=%s", cfg.HTTPAddr, cfg.Env)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Println(err)
		os.Exit(1)
	}
}
