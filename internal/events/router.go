package events

import (
	"github.com/go-chi/chi/v5"

	"github.com/seikaikyo/dashai-go/internal/database"
)

func Router(db *database.DB) chi.Router {
	s := &Store{db: db}
	hub := NewHub()
	go hub.Run()

	r := chi.NewRouter()
	r.Post("/ingest", handleIngest(s, hub))
	r.Get("/", handleList(s))
	r.Get("/stream", handleStream(hub))
	r.Get("/stats", handleStats(s))
	return r
}
