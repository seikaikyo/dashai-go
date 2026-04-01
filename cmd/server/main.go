package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/seikaikyo/dashai-go/internal/config"
	"github.com/seikaikyo/dashai-go/internal/database"
	"github.com/seikaikyo/dashai-go/internal/demo"
	"github.com/seikaikyo/dashai-go/internal/middleware"
	"github.com/seikaikyo/dashai-go/internal/response"
	"github.com/seikaikyo/dashai-go/internal/scanner"
)

var version = "0.1.0"

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	// Logger
	level := slog.LevelInfo
	if cfg.Debug {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))

	// Database (optional)
	var db *database.DB
	if cfg.DatabaseURL != "" {
		db, err = database.Connect(context.Background(), cfg.DatabaseURL)
		if err != nil {
			slog.Error("database connect failed", "error", err)
			os.Exit(1)
		}
		defer db.Close()
		slog.Info("database connected")
	} else {
		slog.Info("no DATABASE_URL, running without database")
	}

	// Router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.CORS(cfg))
	r.Use(middleware.RateLimit(cfg))
	r.Use(chimw.Recoverer)

	// Health (UptimeRobot ping target)
	r.MethodFunc("GET", "/health", func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, map[string]string{"status": "ok", "app": "DashAI Go Gateway"})
	})
	r.MethodFunc("HEAD", "/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Root
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, map[string]any{
			"app":      "DashAI Go Gateway",
			"version":  version,
			"services": []string{"/demo", "/scanner"},
		})
	})

	// Mount sub-modules
	r.Mount("/demo", demo.Router(cfg, db))
	r.Mount("/scanner", scanner.Router())

	// Server
	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "addr", addr, "version", version)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	slog.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("server stopped")
}
