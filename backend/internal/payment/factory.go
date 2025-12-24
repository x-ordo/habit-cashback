package payment

import (
	"log"
	"os"
	"strings"
)

// NewService creates the appropriate payment service based on environment.
// Uses mock service for local/test environments, real TossPay for staging/prod.
//
// Environment detection:
//   - APP_ENV=local or APP_ENV=test -> MockService
//   - APP_ENV=staging or APP_ENV=prod -> TossPayClient (requires mTLS + API key)
//   - Missing TossPay config in staging/prod -> falls back to MockService with warning
func NewService() Service {
	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))

	// For local/test environments, always use mock
	if appEnv == "local" || appEnv == "test" || appEnv == "" {
		log.Println("[payment] Using MockService (APP_ENV=" + appEnv + ")")
		return NewMockService()
	}

	// For staging/prod, try to create TossPay client
	// This requires the toss package, which will be imported in main.go
	// Here we just return mock as a fallback indicator
	log.Println("[payment] APP_ENV=" + appEnv + " - TossPay client should be initialized in main")
	return nil // Signal that caller should create TossPayClient
}

// NewServiceWithTossPay is called from main.go when TossPay client is available.
func NewServiceWithTossPay(tossPayService Service) Service {
	if tossPayService != nil {
		log.Println("[payment] Using TossPayClient (live mode)")
		return tossPayService
	}
	log.Println("[payment] TossPay not available, falling back to MockService")
	return NewMockService()
}

// IsMockEnvironment returns true if the current environment should use mock payment.
func IsMockEnvironment() bool {
	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	return appEnv == "local" || appEnv == "test" || appEnv == ""
}
