package edge

import "time"

type Node struct {
	ID            string            `json:"id"`
	NodeType      string            `json:"node_type"`
	Version       string            `json:"version,omitempty"`
	Location      string            `json:"location,omitempty"`
	Capabilities  []string          `json:"capabilities"`
	Endpoints     map[string]string `json:"endpoints"`
	Metadata      map[string]any    `json:"metadata"`
	Status        string            `json:"status"`
	LastHeartbeat *time.Time        `json:"last_heartbeat,omitempty"`
	RegisteredAt  time.Time         `json:"registered_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type RegisterRequest struct {
	NodeID       string            `json:"node_id"`
	NodeType     string            `json:"node_type"`
	Version      string            `json:"version,omitempty"`
	Location     string            `json:"location,omitempty"`
	Capabilities []string          `json:"capabilities,omitempty"`
	Endpoints    map[string]string `json:"endpoints,omitempty"`
	Metadata     map[string]any    `json:"metadata,omitempty"`
}

type HeartbeatRequest struct {
	NodeID        string         `json:"node_id"`
	Status        string         `json:"status"`
	UptimeSeconds int64          `json:"uptime_seconds"`
	Plugins       map[string]any `json:"plugins,omitempty"`
	System        map[string]any `json:"system,omitempty"`
}
