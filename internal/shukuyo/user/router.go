package user

import (
	"github.com/go-chi/chi/v5"
	"github.com/seikaikyo/dashai-go/internal/config"
	"github.com/seikaikyo/dashai-go/internal/database"
	"github.com/seikaikyo/dashai-go/internal/middleware/auth"
)

// Router returns a chi.Router for the shukuyo user endpoints.
func Router(cfg *config.Config, db *database.DB) chi.Router {
	s := &Store{db: db}
	h := &Handler{store: s}

	r := chi.NewRouter()

	// Protected routes (require Logto JWT)
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireJWT(cfg))
		r.Get("/profile", h.handleGetProfile)
		r.Post("/profile/sync", h.handleSyncProfile)
		r.Get("/profile/full", h.handleGetFullProfile)
		r.Post("/profile/sync-full", h.handleSyncFull)
	})

	// Public routes (company cache)
	r.Get("/company-cache/{country}/{name}", h.handleGetCompanyCache)
	r.Post("/company-cache/batch", h.handleBatchCompanyCache)
	r.Post("/company-cache", h.handleSaveCompanyCache)

	return r
}
