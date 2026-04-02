package user

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/seikaikyo/dashai-go/internal/middleware/auth"
	"github.com/seikaikyo/dashai-go/internal/response"
)

// Handler holds HTTP handler methods for user endpoints.
type Handler struct {
	store *Store
}

func (h *Handler) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	authID := auth.GetUserID(r.Context())
	u, err := h.store.GetOrCreateUser(r.Context(), authID)
	if err != nil {
		slog.Error("get profile", "error", err)
		response.Err(w, http.StatusInternalServerError, "failed to get profile")
		return
	}
	response.OK(w, u)
}

func (h *Handler) handleSyncProfile(w http.ResponseWriter, r *http.Request) {
	authID := auth.GetUserID(r.Context())

	var req ProfileSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "invalid request body")
		return
	}

	u, err := h.store.UpdateUserSync(r.Context(), authID, req)
	if err != nil {
		slog.Error("sync profile", "error", err)
		response.Err(w, http.StatusInternalServerError, "failed to sync profile")
		return
	}
	response.OK(w, u)
}

func (h *Handler) handleGetFullProfile(w http.ResponseWriter, r *http.Request) {
	authID := auth.GetUserID(r.Context())

	u, partners, companies, jobSeekers, hrCandidates, err := h.store.GetFullProfile(r.Context(), authID)
	if err != nil {
		slog.Error("get full profile", "error", err)
		response.Err(w, http.StatusInternalServerError, "failed to get full profile")
		return
	}

	response.OK(w, map[string]any{
		"id":                u.ID,
		"auth_id":           u.AuthID,
		"email":             u.Email,
		"display_name":      u.DisplayName,
		"birth_date":        u.BirthDate,
		"plan":              u.Plan,
		"credits_remaining": u.CreditsRemaining,
		"preferences":       u.Preferences,
		"hr_company":        u.HrCompany,
		"partners":          partners,
		"companies":         companies,
		"job_seekers":       jobSeekers,
		"hr_candidates":     hrCandidates,
	})
}

func (h *Handler) handleSyncFull(w http.ResponseWriter, r *http.Request) {
	authID := auth.GetUserID(r.Context())

	var req FullSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.store.SyncFull(r.Context(), authID, req); err != nil {
		slog.Error("sync full", "error", err)
		response.Err(w, http.StatusInternalServerError, "failed to sync full profile")
		return
	}
	response.OK(w, map[string]bool{"synced": true})
}

func (h *Handler) handleGetCompanyCache(w http.ResponseWriter, r *http.Request) {
	country := chi.URLParam(r, "country")
	name := chi.URLParam(r, "name")

	if country == "" || name == "" {
		response.Err(w, http.StatusBadRequest, "country and name are required")
		return
	}

	entry, err := h.store.GetCompanyCache(r.Context(), country, name)
	if err != nil {
		slog.Error("get company cache", "error", err)
		response.Err(w, http.StatusInternalServerError, "failed to query company cache")
		return
	}
	response.OK(w, entry) // nil = not found, returns null data
}

func (h *Handler) handleBatchCompanyCache(w http.ResponseWriter, r *http.Request) {
	var req CompanyCacheBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.Names) == 0 {
		response.Err(w, http.StatusBadRequest, "names is required")
		return
	}
	if len(req.Names) > 50 {
		response.Err(w, http.StatusBadRequest, "max 50 names per batch")
		return
	}
	if req.Country == "" {
		response.Err(w, http.StatusBadRequest, "country is required")
		return
	}

	result, err := h.store.BatchGetCompanyCache(r.Context(), req.Country, req.Names)
	if err != nil {
		slog.Error("batch company cache", "error", err)
		response.Err(w, http.StatusInternalServerError, "failed to batch query")
		return
	}
	response.OK(w, result)
}

func (h *Handler) handleSaveCompanyCache(w http.ResponseWriter, r *http.Request) {
	var req CompanyCacheSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Country == "" {
		response.Err(w, http.StatusBadRequest, "name and country are required")
		return
	}

	if err := h.store.UpsertCompanyCache(r.Context(), req); err != nil {
		slog.Error("save company cache", "error", err)
		response.Err(w, http.StatusInternalServerError, "failed to save company cache")
		return
	}
	response.OK(w, map[string]bool{"saved": true})
}
