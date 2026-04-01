package edge

import (
	"context"
	"log/slog"
	"time"

	"github.com/seikaikyo/dashai-go/internal/database"
)

// StartMonitor 定期檢查並標記超時未心跳的 edge node 為 offline
func StartMonitor(ctx context.Context, db *database.DB, logger *slog.Logger) {
	s := &Store{db: db}
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			count, err := s.MarkOffline(ctx, 90*time.Second)
			if err != nil {
				logger.Error("edge monitor error", "error", err)
			}
			if count > 0 {
				logger.Info("edge nodes marked offline", "count", count)
			}
		case <-ctx.Done():
			return
		}
	}
}
