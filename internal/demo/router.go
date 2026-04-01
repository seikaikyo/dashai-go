package demo

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/seikaikyo/dashai-go/internal/config"
	"github.com/seikaikyo/dashai-go/internal/database"
	"github.com/seikaikyo/dashai-go/internal/middleware/auth"
	"github.com/seikaikyo/dashai-go/internal/response"
)

func Router(cfg *config.Config, db *database.DB) chi.Router {
	r := chi.NewRouter()

	r.Get("/api/ping", handlePing)
	r.Get("/api/status", handleStatus(db))

	// Protected (requires Logto JWT)
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireJWT(cfg))
		r.Get("/api/protected", handleProtected)
	})

	return r
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	response.OK(w, map[string]any{
		"pong":      true,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func handleProtected(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	response.OK(w, map[string]any{
		"message": "authenticated",
		"user_id": userID,
	})
}

func handleStatus(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "not configured"
		if db != nil {
			if err := db.Pool.Ping(r.Context()); err != nil {
				dbStatus = "error: " + err.Error()
			} else {
				dbStatus = "connected"
			}
		}

		response.OK(w, map[string]any{
			"module":   "demo",
			"database": dbStatus,
		})
	}
}
