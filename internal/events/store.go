package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/seikaikyo/dashai-go/internal/database"
)

type Store struct {
	db *database.DB
}

// BatchInsert 批次寫入事件，跳過缺少必要欄位的事件
func (s *Store) BatchInsert(ctx context.Context, nodeID string, events []IngestEvent) (int, error) {
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	accepted := 0
	for _, e := range events {
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

		dataJSON, err := json.Marshal(e.Data)
		if err != nil {
			dataJSON = []byte("{}")
		}

		_, err = tx.Exec(ctx,
			`INSERT INTO events (id, node_id, timestamp, source, type, severity, data)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 ON CONFLICT (id) DO NOTHING`,
			e.EventID, nodeID, ts, e.Source, e.Type, severity, dataJSON,
		)
		if err != nil {
			slog.Warn("event insert failed", "event_id", e.EventID, "error", err)
			continue
		}
		accepted++
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return accepted, nil
}

// List 查詢事件列表，支援 node_id 和 type 篩選
func (s *Store) List(ctx context.Context, nodeID, eventType string, limit, offset int) ([]Event, int, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// 動態組裝 WHERE 條件
	where := "WHERE 1=1"
	args := []any{}
	argIdx := 1

	if nodeID != "" {
		where += " AND node_id = $" + itoa(argIdx)
		args = append(args, nodeID)
		argIdx++
	}
	if eventType != "" {
		where += " AND type = $" + itoa(argIdx)
		args = append(args, eventType)
		argIdx++
	}

	// 先查總數
	var total int
	countQuery := "SELECT COUNT(*) FROM events " + where
	if err := s.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 查詢資料
	dataQuery := "SELECT id, node_id, timestamp, source, type, severity, data, created_at FROM events " +
		where + " ORDER BY timestamp DESC LIMIT $" + itoa(argIdx) + " OFFSET $" + itoa(argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.db.Pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []Event
	for rows.Next() {
		var ev Event
		var dataJSON []byte
		if err := rows.Scan(&ev.ID, &ev.NodeID, &ev.Timestamp, &ev.Source, &ev.Type, &ev.Severity, &dataJSON, &ev.CreatedAt); err != nil {
			return nil, 0, err
		}
		if err := json.Unmarshal(dataJSON, &ev.Data); err != nil {
			ev.Data = map[string]any{}
		}
		results = append(results, ev)
	}
	if results == nil {
		results = []Event{}
	}
	return results, total, rows.Err()
}

// Stats 統計事件數量（單一查詢完成）
func (s *Store) Stats(ctx context.Context) (*EventStats, error) {
	// 取得總數 + 各維度分群
	stats := &EventStats{
		ByType:     map[string]int{},
		BySeverity: map[string]int{},
		ByNode:     map[string]int{},
	}

	// 總數
	if err := s.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM events").Scan(&stats.TotalEvents); err != nil {
		return nil, err
	}

	// by_type
	rows, err := s.db.Pool.Query(ctx, "SELECT type, COUNT(*) FROM events GROUP BY type")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var k string
		var v int
		if err := rows.Scan(&k, &v); err != nil {
			rows.Close()
			return nil, err
		}
		stats.ByType[k] = v
	}
	rows.Close()

	// by_severity
	rows, err = s.db.Pool.Query(ctx, "SELECT severity, COUNT(*) FROM events GROUP BY severity")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var k string
		var v int
		if err := rows.Scan(&k, &v); err != nil {
			rows.Close()
			return nil, err
		}
		stats.BySeverity[k] = v
	}
	rows.Close()

	// by_node
	rows, err = s.db.Pool.Query(ctx, "SELECT node_id, COUNT(*) FROM events GROUP BY node_id")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var k string
		var v int
		if err := rows.Scan(&k, &v); err != nil {
			rows.Close()
			return nil, err
		}
		stats.ByNode[k] = v
	}
	rows.Close()

	return stats, nil
}

// itoa 簡易整數轉字串（避免引入 strconv 只為了這個）
func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}
