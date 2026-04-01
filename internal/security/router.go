package security

import (
	"github.com/go-chi/chi/v5"

	"github.com/seikaikyo/dashai-go/internal/database"
)

func Router(db *database.DB) chi.Router {
	s := &Store{db: db}
	r := chi.NewRouter()

	r.Post("/report", handleReport(s))
	r.Get("/reports", handleListReports(s))
	r.Get("/reports/{id}", handleGetReport(s))
	r.Get("/dashboard", handleDashboard(s))
	r.Get("/alerts", handleListAlerts(s))
	r.Post("/alerts/{id}/ack", handleAckAlert(s))

	return r
}
