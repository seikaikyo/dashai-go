package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOK(t *testing.T) {
	w := httptest.NewRecorder()
	OK(w, map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var body Body
	json.NewDecoder(w.Body).Decode(&body)

	if !body.Success {
		t.Error("Success should be true")
	}
}

func TestErr(t *testing.T) {
	w := httptest.NewRecorder()
	Err(w, http.StatusNotFound, "not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}

	var body Body
	json.NewDecoder(w.Body).Decode(&body)

	if body.Success {
		t.Error("Success should be false")
	}
	if body.Error != "not found" {
		t.Errorf("Error = %q, want %q", body.Error, "not found")
	}
}

func TestOKPage(t *testing.T) {
	w := httptest.NewRecorder()
	OKPage(w, []string{"a", "b"}, 10, 1)

	var body Body
	json.NewDecoder(w.Body).Decode(&body)

	if !body.Success {
		t.Error("Success should be true")
	}
	if body.Total != 10 {
		t.Errorf("Total = %d, want 10", body.Total)
	}
	if body.Page != 1 {
		t.Errorf("Page = %d, want 1", body.Page)
	}
}
