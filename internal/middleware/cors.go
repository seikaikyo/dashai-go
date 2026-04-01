package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
	"github.com/seikaikyo/dashai-go/internal/config"
)

func CORS(cfg *config.Config) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
