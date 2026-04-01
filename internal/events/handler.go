package events

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"

	"github.com/seikaikyo/dashai-go/internal/response"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleIngest(s *Store, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req IngestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.Err(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.NodeID == "" || len(req.Events) == 0 {
			response.Err(w, http.StatusBadRequest, "node_id and events are required")
			return
		}

		accepted, err := s.BatchInsert(r.Context(), req.NodeID, req.Events)
		if err != nil {
			slog.Error("batch insert failed", "node_id", req.NodeID, "error", err)
			response.Err(w, http.StatusInternalServerError, "event ingestion failed")
			return
		}

		// 廣播成功寫入的事件給 WebSocket 訂閱者
		for _, e := range req.Events {
			if e.EventID == "" || e.Source == "" || e.Type == "" {
				continue
			}
			severity := e.Severity
			if severity == "" {
				severity = "info"
			}
			ts := e.Timestamp
			if ts.IsZero() {
				ts = time.Now()
			}
			hub.Broadcast(Event{
				ID:        e.EventID,
				NodeID:    req.NodeID,
				Timestamp: ts,
				Source:    e.Source,
				Type:      e.Type,
				Severity:  severity,
				Data:      e.Data,
			})
		}

		response.OK(w, map[string]int{"accepted": accepted})
	}
}

func handleList(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		nodeID := q.Get("node_id")
		eventType := q.Get("type")

		limit, _ := strconv.Atoi(q.Get("limit"))
		offset, _ := strconv.Atoi(q.Get("offset"))

		events, total, err := s.List(r.Context(), nodeID, eventType, limit, offset)
		if err != nil {
			slog.Error("event list failed", "error", err)
			response.Err(w, http.StatusInternalServerError, "failed to list events")
			return
		}

		page := 1
		if limit > 0 && offset > 0 {
			page = offset/limit + 1
		}
		response.OKPage(w, events, total, page)
	}
}

func handleStream(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Warn("websocket upgrade failed", "error", err)
			return
		}

		client := &Client{
			hub:  hub,
			conn: conn,
			send: make(chan []byte, 256),
		}
		hub.register <- client

		go client.writePump()
		go client.readPump()
	}
}

func handleStats(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := s.Stats(r.Context())
		if err != nil {
			slog.Error("event stats failed", "error", err)
			response.Err(w, http.StatusInternalServerError, "failed to get event stats")
			return
		}
		response.OK(w, stats)
	}
}
