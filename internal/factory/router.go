package factory

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dashfactory/go-factory-io/pkg/studio"
)

// Router returns a chi.Router that serves the SECSGEM Studio UI at /factory/*.
// It starts an embedded equipment simulator and host connection.
func Router(ctx context.Context) chi.Router {
	r := chi.NewRouter()

	srv := studio.NewServer(studio.Config{SessionID: 1}, slog.Default())

	// Start embedded equipment simulator
	addr, err := srv.StartEquipment(ctx)
	if err != nil {
		slog.Error("factory: start equipment failed", "error", err)
		// Serve without simulator — UI still loads
		r.Handle("/*", srv.Handler())
		return r
	}

	// Connect host to simulator
	if err := srv.ConnectHost(ctx, addr); err != nil {
		slog.Error("factory: connect host failed", "error", err)
	}

	// Mount studio handler under /factory/*
	r.Handle("/*", http.StripPrefix("/factory", srv.Handler()))

	return r
}
