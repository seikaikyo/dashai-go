package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Clear env to test defaults
	os.Unsetenv("PORT")
	os.Unsetenv("DEBUG")
	os.Unsetenv("DATABASE_URL")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Port != 8101 {
		t.Errorf("Port = %d, want 8101", cfg.Port)
	}
	if cfg.Debug {
		t.Error("Debug should be false by default")
	}
	if cfg.DatabaseURL != "" {
		t.Errorf("DatabaseURL = %q, want empty", cfg.DatabaseURL)
	}
	if cfg.RateLimit != 30 {
		t.Errorf("RateLimit = %d, want 30", cfg.RateLimit)
	}
}

func TestLoadDebugCORS(t *testing.T) {
	os.Setenv("DEBUG", "true")
	defer os.Unsetenv("DEBUG")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if !cfg.Debug {
		t.Error("Debug should be true")
	}

	found := false
	for _, o := range cfg.CORSOrigins {
		if o == "http://localhost:5173" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Debug mode should add localhost CORS origins")
	}
}
