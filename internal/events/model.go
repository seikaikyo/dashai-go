package events

import "time"

type Event struct {
	ID        string         `json:"id"`
	NodeID    string         `json:"node_id"`
	Timestamp time.Time      `json:"timestamp"`
	Source    string         `json:"source"`
	Type      string         `json:"type"`
	Severity  string         `json:"severity"`
	Data      map[string]any `json:"data"`
	CreatedAt time.Time      `json:"created_at"`
}

type IngestRequest struct {
	NodeID string        `json:"node_id"`
	Events []IngestEvent `json:"events"`
}

type IngestEvent struct {
	EventID   string         `json:"event_id"`
	Timestamp time.Time      `json:"timestamp"`
	Source    string         `json:"source"`
	Type      string         `json:"type"`
	Severity  string         `json:"severity,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
}

type EventStats struct {
	TotalEvents int            `json:"total_events"`
	ByType      map[string]int `json:"by_type"`
	BySeverity  map[string]int `json:"by_severity"`
	ByNode      map[string]int `json:"by_node"`
}
