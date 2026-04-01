package demo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/seikaikyo/dashai-go/internal/config"
	"github.com/seikaikyo/dashai-go/internal/response"
)

func TestPing(t *testing.T) {
	cfg := &config.Config{}
	r := Router(cfg, nil)

	req := httptest.NewRequest("GET", "/api/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var body response.Body
	json.NewDecoder(w.Body).Decode(&body)

	if !body.Success {
		t.Error("Success should be true")
	}
}

func TestStatusNoDB(t *testing.T) {
	cfg := &config.Config{}
	r := Router(cfg, nil)

	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Database string `json:"database"`
		} `json:"data"`
	}
	json.NewDecoder(w.Body).Decode(&body)

	if body.Data.Database != "not configured" {
		t.Errorf("database = %q, want %q", body.Data.Database, "not configured")
	}
}

func TestProtectedNoToken(t *testing.T) {
	cfg := &config.Config{
		LogtoEndpoint: "https://example.logto.app",
	}
	r := Router(cfg, nil)

	req := httptest.NewRequest("GET", "/api/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}
