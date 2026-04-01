package edge

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/seikaikyo/dashai-go/internal/database"
)

var errNodeNotFound = errors.New("edge node not found")

type Store struct {
	db *database.DB
}

// scannable 統一 QueryRow 和 Rows 的 Scan 介面
type scannable interface {
	Scan(dest ...any) error
}

// Upsert 註冊或重新註冊 edge node（INSERT ON CONFLICT UPDATE）
func (s *Store) Upsert(ctx context.Context, req RegisterRequest) (*Node, error) {
	caps, err := json.Marshal(req.Capabilities)
	if err != nil {
		return nil, err
	}
	eps, err := json.Marshal(req.Endpoints)
	if err != nil {
		return nil, err
	}
	meta, err := json.Marshal(req.Metadata)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO edge_nodes (id, node_type, version, location, capabilities, endpoints, metadata, status, last_heartbeat)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'online', NOW())
		ON CONFLICT (id) DO UPDATE SET
			node_type = EXCLUDED.node_type,
			version = EXCLUDED.version,
			location = EXCLUDED.location,
			capabilities = EXCLUDED.capabilities,
			endpoints = EXCLUDED.endpoints,
			metadata = EXCLUDED.metadata,
			status = 'online',
			last_heartbeat = NOW(),
			updated_at = NOW()
		RETURNING id, node_type, version, location, capabilities, endpoints, metadata, status, last_heartbeat, registered_at, updated_at`

	row := s.db.Pool.QueryRow(ctx, query,
		req.NodeID, req.NodeType, req.Version, req.Location, caps, eps, meta,
	)
	return scanNode(row)
}

// Heartbeat 更新 node 的心跳時間、狀態、metadata
func (s *Store) Heartbeat(ctx context.Context, req HeartbeatRequest) error {
	// 將 plugins 和 system 合併到 metadata
	merged := map[string]any{}
	if req.Plugins != nil {
		merged["plugins"] = req.Plugins
	}
	if req.System != nil {
		merged["system"] = req.System
	}
	merged["uptime_seconds"] = req.UptimeSeconds

	metaJSON, err := json.Marshal(merged)
	if err != nil {
		return err
	}

	query := `
		UPDATE edge_nodes
		SET status = $2, last_heartbeat = NOW(), metadata = metadata || $3::jsonb, updated_at = NOW()
		WHERE id = $1`

	tag, err := s.db.Pool.Exec(ctx, query, req.NodeID, req.Status, metaJSON)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errNodeNotFound
	}
	return nil
}

// List 回傳所有 edge nodes，按註冊時間降序
func (s *Store) List(ctx context.Context) ([]Node, error) {
	query := `
		SELECT id, node_type, version, location, capabilities, endpoints, metadata, status, last_heartbeat, registered_at, updated_at
		FROM edge_nodes
		ORDER BY registered_at DESC`

	rows, err := s.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		n, err := scanNode(rows)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, *n)
	}
	return nodes, rows.Err()
}

// Get 取得單一 edge node
func (s *Store) Get(ctx context.Context, id string) (*Node, error) {
	query := `
		SELECT id, node_type, version, location, capabilities, endpoints, metadata, status, last_heartbeat, registered_at, updated_at
		FROM edge_nodes
		WHERE id = $1`

	row := s.db.Pool.QueryRow(ctx, query, id)
	return scanNode(row)
}

// Delete 刪除 edge node
func (s *Store) Delete(ctx context.Context, id string) error {
	tag, err := s.db.Pool.Exec(ctx, "DELETE FROM edge_nodes WHERE id = $1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errNodeNotFound
	}
	return nil
}

// MarkOffline 將超過 timeout 未心跳的 node 標記為 offline，回傳影響筆數
func (s *Store) MarkOffline(ctx context.Context, timeout time.Duration) (int, error) {
	query := `
		UPDATE edge_nodes
		SET status = 'offline', updated_at = NOW()
		WHERE last_heartbeat < NOW() - make_interval(secs => $1) AND status != 'offline'`

	tag, err := s.db.Pool.Exec(ctx, query, timeout.Seconds())
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func scanNode(row scannable) (*Node, error) {
	var n Node
	var capsJSON, epsJSON, metaJSON []byte

	err := row.Scan(
		&n.ID, &n.NodeType, &n.Version, &n.Location,
		&capsJSON, &epsJSON, &metaJSON,
		&n.Status, &n.LastHeartbeat, &n.RegisteredAt, &n.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(capsJSON, &n.Capabilities); err != nil {
		n.Capabilities = []string{}
	}
	if err := json.Unmarshal(epsJSON, &n.Endpoints); err != nil {
		n.Endpoints = map[string]string{}
	}
	if err := json.Unmarshal(metaJSON, &n.Metadata); err != nil {
		n.Metadata = map[string]any{}
	}

	return &n, nil
}
