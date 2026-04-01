package security

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/seikaikyo/dashai-go/internal/database"
)

type Store struct {
	db *database.DB
}

// InsertReport 寫入掃描報告及附屬 alerts（transaction 確保原子性）
func (s *Store) InsertReport(ctx context.Context, req ReportRequest) (*Report, error) {
	reportID := fmt.Sprintf("rpt-%d", time.Now().UnixNano())

	summaryJSON, err := json.Marshal(req.Summary)
	if err != nil {
		return nil, fmt.Errorf("marshal summary: %w", err)
	}
	complianceJSON, err := json.Marshal(req.Compliance)
	if err != nil {
		return nil, fmt.Errorf("marshal compliance: %w", err)
	}
	devicesJSON, err := json.Marshal(req.Devices)
	if err != nil {
		return nil, fmt.Errorf("marshal devices: %w", err)
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 寫入報告
	reportQuery := `
		INSERT INTO security_reports (id, node_id, scan_id, timestamp, subnet, summary, compliance, devices)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, node_id, scan_id, timestamp, subnet, summary, compliance, devices, created_at`

	var rpt Report
	var sumBytes, compBytes, devBytes []byte

	err = tx.QueryRow(ctx, reportQuery,
		reportID, req.NodeID, req.ScanID, req.Timestamp, req.Subnet,
		summaryJSON, complianceJSON, devicesJSON,
	).Scan(
		&rpt.ID, &rpt.NodeID, &rpt.ScanID, &rpt.Timestamp, &rpt.Subnet,
		&sumBytes, &compBytes, &devBytes, &rpt.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert report: %w", err)
	}

	if err := json.Unmarshal(sumBytes, &rpt.Summary); err != nil {
		rpt.Summary = map[string]any{}
	}
	if err := json.Unmarshal(compBytes, &rpt.Compliance); err != nil {
		rpt.Compliance = map[string]any{}
	}
	if err := json.Unmarshal(devBytes, &rpt.Devices); err != nil {
		rpt.Devices = []any{}
	}

	// 寫入 alerts
	if len(req.Alerts) > 0 {
		alertQuery := `
			INSERT INTO security_alerts (id, report_id, node_id, severity, type, message, technique, device_ip)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

		for i, a := range req.Alerts {
			alertID := fmt.Sprintf("alt-%d-%d", time.Now().UnixNano(), i)
			_, err := tx.Exec(ctx, alertQuery,
				alertID, reportID, req.NodeID,
				a.Severity, a.Type, a.Message, a.Technique, a.DeviceIP,
			)
			if err != nil {
				return nil, fmt.Errorf("insert alert %d: %w", i, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &rpt, nil
}

// ListReports 列出掃描報告，可依 node_id 篩選
func (s *Store) ListReports(ctx context.Context, nodeID string, limit, offset int) ([]Report, int, error) {
	if limit <= 0 {
		limit = 20
	}

	// 計算總數
	countQuery := `SELECT COUNT(*) FROM security_reports`
	args := []any{}
	if nodeID != "" {
		countQuery += ` WHERE node_id = $1`
		args = append(args, nodeID)
	}

	var total int
	if err := s.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 查詢資料
	dataQuery := `
		SELECT id, node_id, scan_id, timestamp, subnet, summary, compliance, devices, created_at
		FROM security_reports`
	dataArgs := []any{}
	if nodeID != "" {
		dataQuery += ` WHERE node_id = $1`
		dataArgs = append(dataArgs, nodeID)
	}
	dataQuery += ` ORDER BY timestamp DESC`

	if nodeID != "" {
		dataQuery += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, 2, 3)
	} else {
		dataQuery += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, 1, 2)
	}
	dataArgs = append(dataArgs, limit, offset)

	rows, err := s.db.Pool.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reports []Report
	for rows.Next() {
		rpt, err := scanReport(rows)
		if err != nil {
			return nil, 0, err
		}
		reports = append(reports, *rpt)
	}
	if reports == nil {
		reports = []Report{}
	}
	return reports, total, rows.Err()
}

// GetReport 取得單一報告
func (s *Store) GetReport(ctx context.Context, id string) (*Report, error) {
	query := `
		SELECT id, node_id, scan_id, timestamp, subnet, summary, compliance, devices, created_at
		FROM security_reports
		WHERE id = $1`

	row := s.db.Pool.QueryRow(ctx, query, id)
	return scanReport(row)
}

// ListAlerts 列出告警，可依嚴重度和確認狀態篩選
func (s *Store) ListAlerts(ctx context.Context, severity string, ackFilter *bool, limit, offset int) ([]Alert, int, error) {
	if limit <= 0 {
		limit = 50
	}

	// 動態組合 WHERE
	where := ""
	args := []any{}
	argIdx := 1

	if severity != "" {
		where += fmt.Sprintf(" AND severity = $%d", argIdx)
		args = append(args, severity)
		argIdx++
	}
	if ackFilter != nil {
		where += fmt.Sprintf(" AND ack = $%d", argIdx)
		args = append(args, *ackFilter)
		argIdx++
	}

	// 去掉開頭的 AND
	if where != "" {
		where = " WHERE" + where[4:]
	}

	// 計算總數
	var total int
	countQuery := `SELECT COUNT(*) FROM security_alerts` + where
	if err := s.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 查詢資料
	dataQuery := fmt.Sprintf(`
		SELECT id, report_id, node_id, severity, type, message, technique, device_ip, ack, ack_at, created_at
		FROM security_alerts%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	dataArgs := append(args, limit, offset)
	rows, err := s.db.Pool.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var a Alert
		err := rows.Scan(
			&a.ID, &a.ReportID, &a.NodeID, &a.Severity, &a.Type,
			&a.Message, &a.Technique, &a.DeviceIP, &a.Ack, &a.AckAt, &a.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		alerts = append(alerts, a)
	}
	if alerts == nil {
		alerts = []Alert{}
	}
	return alerts, total, rows.Err()
}

// AckAlert 確認告警
func (s *Store) AckAlert(ctx context.Context, id string) error {
	tag, err := s.db.Pool.Exec(ctx,
		`UPDATE security_alerts SET ack = TRUE, ack_at = NOW() WHERE id = $1 AND ack = FALSE`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// GetDashboardData 彙整各節點最新掃描資訊
func (s *Store) GetDashboardData(ctx context.Context) (*Dashboard, error) {
	dash := &Dashboard{
		OverallCompliance: map[string]int{},
		NodesDetail:       []NodeSummary{},
	}

	// 取得各節點最新報告，加入 edge_nodes 取得 location 和 status
	query := `
		WITH latest AS (
			SELECT DISTINCT ON (node_id) *
			FROM security_reports
			ORDER BY node_id, timestamp DESC
		)
		SELECT
			l.node_id, l.timestamp, l.summary, l.compliance,
			COALESCE(e.location, ''), COALESCE(e.status, 'unknown')
		FROM latest l
		LEFT JOIN edge_nodes e ON l.node_id = e.id`

	rows, err := s.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type nodeData struct {
		nodeID     string
		ts         time.Time
		summary    map[string]any
		compliance map[string]any
		location   string
		status     string
	}
	var nodes []nodeData

	for rows.Next() {
		var nd nodeData
		var sumBytes, compBytes []byte
		err := rows.Scan(&nd.nodeID, &nd.ts, &sumBytes, &compBytes, &nd.location, &nd.status)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(sumBytes, &nd.summary); err != nil {
			nd.summary = map[string]any{}
		}
		if err := json.Unmarshal(compBytes, &nd.compliance); err != nil {
			nd.compliance = map[string]any{}
		}
		nodes = append(nodes, nd)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 彙整數值
	dash.Nodes = len(nodes)
	complianceSums := map[string]float64{}
	complianceCounts := map[string]int{}

	for _, nd := range nodes {
		ts := nd.ts
		if dash.LastScan == nil || ts.After(*dash.LastScan) {
			dash.LastScan = &ts
		}

		// summary 中的裝置數
		dash.TotalDevices += intFromAny(nd.summary["total_devices"])
		dash.TotalOT += intFromAny(nd.summary["ot_devices"])
		dash.TotalIT += intFromAny(nd.summary["it_devices"])

		// compliance 分數平均（值為 {"passed":N,"total":N,"score":N} 結構）
		for k, v := range nd.compliance {
			if m, ok := v.(map[string]any); ok {
				if score, ok := toFloat64(m["score"]); ok {
					complianceSums[k] += score
					complianceCounts[k]++
				}
			} else if score, ok := toFloat64(v); ok {
				complianceSums[k] += score
				complianceCounts[k]++
			}
		}

		// 每節點摘要
		critical := intFromAny(nd.summary["critical_vulns"])
		ns := NodeSummary{
			NodeID:   nd.nodeID,
			Location: nd.location,
			Devices:  intFromAny(nd.summary["total_devices"]),
			Critical: critical,
			LastScan: &ts,
			Status:   nd.status,
		}
		dash.NodesDetail = append(dash.NodesDetail, ns)
	}

	// 計算平均 compliance
	for k, sum := range complianceSums {
		if complianceCounts[k] > 0 {
			dash.OverallCompliance[k] = int(sum / float64(complianceCounts[k]))
		}
	}

	// 未確認的 critical alerts 數量
	err = s.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM security_alerts WHERE severity = 'critical' AND ack = FALSE`,
	).Scan(&dash.CriticalAlerts)
	if err != nil {
		return nil, err
	}

	return dash, nil
}

// scannable 統一 QueryRow 和 Rows 的 Scan 介面
type scannable interface {
	Scan(dest ...any) error
}

func scanReport(row scannable) (*Report, error) {
	var rpt Report
	var sumBytes, compBytes, devBytes []byte

	err := row.Scan(
		&rpt.ID, &rpt.NodeID, &rpt.ScanID, &rpt.Timestamp, &rpt.Subnet,
		&sumBytes, &compBytes, &devBytes, &rpt.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(sumBytes, &rpt.Summary); err != nil {
		rpt.Summary = map[string]any{}
	}
	if err := json.Unmarshal(compBytes, &rpt.Compliance); err != nil {
		rpt.Compliance = map[string]any{}
	}
	if err := json.Unmarshal(devBytes, &rpt.Devices); err != nil {
		rpt.Devices = []any{}
	}

	return &rpt, nil
}

// intFromAny 從 any 取出整數值（相容 float64 JSON 解碼）
func intFromAny(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	default:
		return 0
	}
}

// toFloat64 嘗試將 any 轉為 float64
func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}
