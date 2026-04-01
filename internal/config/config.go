package config

import (
	"log/slog"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port  int  `envconfig:"PORT" default:"8101"`
	Debug bool `envconfig:"DEBUG" default:"false"`

	// Database (Neon PostgreSQL, optional)
	DatabaseURL string `envconfig:"DATABASE_URL"`

	// Logto JWT (shared, per-module opt-in)
	LogtoEndpoint    string `envconfig:"LOGTO_ENDPOINT"`
	LogtoAPIResource string `envconfig:"LOGTO_API_RESOURCE"`

	// CORS
	CORSOrigins []string `envconfig:"CORS_ORIGINS"`

	// Rate Limiting (requests per minute)
	RateLimit int `envconfig:"RATE_LIMIT" default:"30"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	if cfg.Debug {
		slog.Info("debug mode enabled")
		cfg.CORSOrigins = append(cfg.CORSOrigins,
			"http://localhost:5173",
			"http://localhost:5174",
			"http://localhost:5175",
			"http://localhost:5176",
			"http://localhost:8101",
		)
	}

	return &cfg, nil
}
