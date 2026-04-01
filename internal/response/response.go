package response

import (
	"encoding/json"
	"net/http"
)

type Body struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Total   int    `json:"total,omitempty"`
	Page    int    `json:"page,omitempty"`
}

func OK(w http.ResponseWriter, data any) {
	write(w, http.StatusOK, Body{Success: true, Data: data})
}

func OKPage(w http.ResponseWriter, data any, total, page int) {
	write(w, http.StatusOK, Body{Success: true, Data: data, Total: total, Page: page})
}

func Err(w http.ResponseWriter, status int, msg string) {
	write(w, status, Body{Success: false, Error: msg})
}

func write(w http.ResponseWriter, status int, body Body) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
