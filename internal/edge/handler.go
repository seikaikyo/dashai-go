package edge

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/seikaikyo/dashai-go/internal/response"
)

type Handler struct {
	store *Store
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.NodeID == "" || req.NodeType == "" {
		response.Err(w, http.StatusBadRequest, "node_id and node_type are required")
		return
	}

	node, err := h.store.Upsert(r.Context(), req)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "registration failed")
		return
	}
	response.OK(w, node)
}

func (h *Handler) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	var req HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.NodeID == "" {
		response.Err(w, http.StatusBadRequest, "node_id is required")
		return
	}
	if req.Status == "" {
		req.Status = "online"
	}

	err := h.store.Heartbeat(r.Context(), req)
	if err != nil {
		if errors.Is(err, errNodeNotFound) {
			response.Err(w, http.StatusNotFound, "node not found, register first")
			return
		}
		response.Err(w, http.StatusInternalServerError, "heartbeat update failed")
		return
	}
	response.OK(w, map[string]string{"status": "ok"})
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.store.List(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "failed to list nodes")
		return
	}
	if nodes == nil {
		nodes = []Node{}
	}
	response.OK(w, nodes)
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	node, err := h.store.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, errNodeNotFound) {
			response.Err(w, http.StatusNotFound, "node not found")
			return
		}
		response.Err(w, http.StatusInternalServerError, "failed to get node")
		return
	}
	response.OK(w, node)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.store.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, errNodeNotFound) {
			response.Err(w, http.StatusNotFound, "node not found")
			return
		}
		response.Err(w, http.StatusInternalServerError, "failed to delete node")
		return
	}
	response.OK(w, map[string]string{"status": "deleted"})
}
