package security

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/seikaikyo/dashai-go/internal/response"
)

func handleReport(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.Err(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.NodeID == "" || req.ScanID == "" {
			response.Err(w, http.StatusBadRequest, "node_id and scan_id are required")
			return
		}
		if req.Summary == nil {
			response.Err(w, http.StatusBadRequest, "summary is required")
			return
		}
		if req.Compliance == nil {
			response.Err(w, http.StatusBadRequest, "compliance is required")
			return
		}

		rpt, err := s.InsertReport(r.Context(), req)
		if err != nil {
			response.Err(w, http.StatusInternalServerError, "failed to insert report")
			return
		}
		response.OK(w, rpt)
	}
}

func handleListReports(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodeID := r.URL.Query().Get("node_id")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		reports, total, err := s.ListReports(r.Context(), nodeID, limit, offset)
		if err != nil {
			response.Err(w, http.StatusInternalServerError, "failed to list reports")
			return
		}

		page := 1
		if limit > 0 && offset >= 0 {
			page = (offset / limit) + 1
		}
		response.OKPage(w, reports, total, page)
	}
}

func handleGetReport(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		rpt, err := s.GetReport(r.Context(), id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				response.Err(w, http.StatusNotFound, "report not found")
				return
			}
			response.Err(w, http.StatusInternalServerError, "failed to get report")
			return
		}
		response.OK(w, rpt)
	}
}

func handleDashboard(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dash, err := s.GetDashboardData(r.Context())
		if err != nil {
			response.Err(w, http.StatusInternalServerError, "failed to get dashboard data")
			return
		}
		response.OK(w, dash)
	}
}

func handleListAlerts(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		severity := r.URL.Query().Get("severity")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		var ackFilter *bool
		if ackStr := r.URL.Query().Get("ack"); ackStr != "" {
			v := ackStr == "true"
			ackFilter = &v
		}

		alerts, total, err := s.ListAlerts(r.Context(), severity, ackFilter, limit, offset)
		if err != nil {
			response.Err(w, http.StatusInternalServerError, "failed to list alerts")
			return
		}

		page := 1
		if limit > 0 && offset >= 0 {
			page = (offset / limit) + 1
		}
		response.OKPage(w, alerts, total, page)
	}
}

func handleAckAlert(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		err := s.AckAlert(r.Context(), id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				response.Err(w, http.StatusNotFound, "alert not found or already acknowledged")
				return
			}
			response.Err(w, http.StatusInternalServerError, "failed to acknowledge alert")
			return
		}
		response.OK(w, map[string]string{"status": "acknowledged"})
	}
}
