package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
	"github.com/seikaikyo/dashai-go/internal/config"
)

func RateLimit(cfg *config.Config) func(http.Handler) http.Handler {
	return httprate.LimitByIP(cfg.RateLimit, time.Minute)
}
