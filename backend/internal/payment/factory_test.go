package payment

import (
	"os"
	"testing"
)

func TestIsMockEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		appEnv   string
		expected bool
	}{
		{"Empty APP_ENV", "", true},
		{"local", "local", true},
		{"LOCAL (uppercase)", "LOCAL", true},
		{"Local (mixed)", "Local", true},
		{"test", "test", true},
		{"TEST (uppercase)", "TEST", true},
		{"staging", "staging", false},
		{"STAGING (uppercase)", "STAGING", false},
		{"prod", "prod", false},
		{"PROD (uppercase)", "PROD", false},
		{"production", "production", false},
		{"development", "development", false},
		{"with spaces", "  local  ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			original := os.Getenv("APP_ENV")
			defer os.Setenv("APP_ENV", original)

			os.Setenv("APP_ENV", tt.appEnv)

			result := IsMockEnvironment()
			if result != tt.expected {
				t.Errorf("IsMockEnvironment() with APP_ENV=%q: expected %v, got %v",
					tt.appEnv, tt.expected, result)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	t.Run("Local Environment", func(t *testing.T) {
		original := os.Getenv("APP_ENV")
		defer os.Setenv("APP_ENV", original)

		os.Setenv("APP_ENV", "local")
		svc := NewService()

		if svc == nil {
			t.Fatal("expected non-nil service for local environment")
		}
		if svc.Mode() != "mock" {
			t.Errorf("expected mode 'mock', got '%s'", svc.Mode())
		}
	})

	t.Run("Test Environment", func(t *testing.T) {
		original := os.Getenv("APP_ENV")
		defer os.Setenv("APP_ENV", original)

		os.Setenv("APP_ENV", "test")
		svc := NewService()

		if svc == nil {
			t.Fatal("expected non-nil service for test environment")
		}
		if svc.Mode() != "mock" {
			t.Errorf("expected mode 'mock', got '%s'", svc.Mode())
		}
	})

	t.Run("Empty Environment", func(t *testing.T) {
		original := os.Getenv("APP_ENV")
		defer os.Setenv("APP_ENV", original)

		os.Setenv("APP_ENV", "")
		svc := NewService()

		if svc == nil {
			t.Fatal("expected non-nil service for empty environment")
		}
		if svc.Mode() != "mock" {
			t.Errorf("expected mode 'mock', got '%s'", svc.Mode())
		}
	})

	t.Run("Staging Environment", func(t *testing.T) {
		original := os.Getenv("APP_ENV")
		defer os.Setenv("APP_ENV", original)

		os.Setenv("APP_ENV", "staging")
		svc := NewService()

		// Should return nil to signal that TossPay client should be created
		if svc != nil {
			t.Error("expected nil for staging environment (caller should create TossPay)")
		}
	})

	t.Run("Prod Environment", func(t *testing.T) {
		original := os.Getenv("APP_ENV")
		defer os.Setenv("APP_ENV", original)

		os.Setenv("APP_ENV", "prod")
		svc := NewService()

		// Should return nil to signal that TossPay client should be created
		if svc != nil {
			t.Error("expected nil for prod environment (caller should create TossPay)")
		}
	})
}

func TestNewServiceWithTossPay(t *testing.T) {
	t.Run("With TossPay Service", func(t *testing.T) {
		mockTossPay := NewMockService() // Using mock as stand-in for TossPay

		svc := NewServiceWithTossPay(mockTossPay)
		if svc == nil {
			t.Fatal("expected non-nil service")
		}
		if svc != mockTossPay {
			t.Error("expected to return the provided TossPay service")
		}
	})

	t.Run("Without TossPay Service", func(t *testing.T) {
		svc := NewServiceWithTossPay(nil)

		if svc == nil {
			t.Fatal("expected non-nil service (fallback to mock)")
		}
		if svc.Mode() != "mock" {
			t.Errorf("expected fallback to 'mock' mode, got '%s'", svc.Mode())
		}
	})
}
