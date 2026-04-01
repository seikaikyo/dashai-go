package edge

import (
	"github.com/go-chi/chi/v5"

	"github.com/seikaikyo/dashai-go/internal/database"
)

func Router(db *database.DB) chi.Router {
	s := &Store{db: db}
	h := &Handler{store: s}

	r := chi.NewRouter()
	r.Post("/register", h.handleRegister)
	r.Post("/heartbeat", h.handleHeartbeat)
	r.Get("/nodes", h.handleList)
	r.Get("/nodes/{id}", h.handleGet)
	r.Delete("/nodes/{id}", h.handleDelete)
	return r
}
