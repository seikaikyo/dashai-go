package security

import "time"

type Report struct {
	ID         string         `json:"id"`
	NodeID     string         `json:"node_id"`
	ScanID     string         `json:"scan_id"`
	Timestamp  time.Time      `json:"timestamp"`
	Subnet     string         `json:"subnet,omitempty"`
	Summary    map[string]any `json:"summary"`
	Compliance map[string]any `json:"compliance"`
	Devices    []any          `json:"devices"`
	CreatedAt  time.Time      `json:"created_at"`
}

type ReportRequest struct {
	NodeID     string         `json:"node_id"`
	ScanID     string         `json:"scan_id"`
	Timestamp  time.Time      `json:"timestamp"`
	Subnet     string         `json:"subnet,omitempty"`
	Summary    map[string]any `json:"summary"`
	Compliance map[string]any `json:"compliance"`
	Alerts     []AlertInput   `json:"alerts,omitempty"`
	Devices    []any          `json:"devices,omitempty"`
}

type AlertInput struct {
	Severity  string `json:"severity"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	Technique string `json:"technique,omitempty"`
	DeviceIP  string `json:"device_ip,omitempty"`
}

type Alert struct {
	ID        string     `json:"id"`
	ReportID  string     `json:"report_id"`
	NodeID    string     `json:"node_id"`
	Severity  string     `json:"severity"`
	Type      string     `json:"type"`
	Message   string     `json:"message"`
	Technique string     `json:"technique,omitempty"`
	DeviceIP  string     `json:"device_ip,omitempty"`
	Ack       bool       `json:"ack"`
	AckAt     *time.Time `json:"ack_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type Dashboard struct {
	Nodes             int            `json:"nodes"`
	LastScan          *time.Time     `json:"last_scan,omitempty"`
	TotalDevices      int            `json:"total_devices"`
	TotalOT           int            `json:"total_ot"`
	TotalIT           int            `json:"total_it"`
	CriticalAlerts    int            `json:"critical_alerts"`
	OverallCompliance map[string]int `json:"overall_compliance"`
	NodesDetail       []NodeSummary  `json:"nodes_detail"`
}

type NodeSummary struct {
	NodeID   string     `json:"node_id"`
	Location string     `json:"location,omitempty"`
	Devices  int        `json:"devices"`
	Critical int        `json:"critical"`
	LastScan *time.Time `json:"last_scan,omitempty"`
	Status   string     `json:"status"`
}
