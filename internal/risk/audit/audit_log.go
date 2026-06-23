package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type AuditEntry struct {
	ID         string    `json:"id"`
	EventType  string    `json:"event_type"`
	ResourceID string    `json:"resource_id"`
	Action     string    `json:"action"`
	RiskLevel  string    `json:"risk_level"`
	Details    string    `json:"details"`
	Timestamp  time.Time `json:"timestamp"`
}

type AuditLogger struct {
	rdb *redis.Client
}

func NewAuditLogger(rdb *redis.Client) *AuditLogger {
	return &AuditLogger{rdb: rdb}
}

func (al *AuditLogger) Log(ctx context.Context, entry *AuditEntry) error {
	entry.Timestamp = time.Now()
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// Push to Redis list for async processing
	if err := al.rdb.LPush(ctx, "audit:queue", data).Err(); err != nil {
		log.Error().Err(err).Msg("failed to log audit entry")
		return err
	}

	log.Ctx(ctx).Info().
		Str("event_type", entry.EventType).
		Str("action", entry.Action).
		Str("risk_level", entry.RiskLevel).
		Msg("audit event logged")

	return nil
}

func (al *AuditLogger) GetRecent(ctx context.Context, limit int) ([]*AuditEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	items, err := al.rdb.LRange(ctx, "audit:queue", 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	var entries []*AuditEntry
	for _, item := range items {
		var entry AuditEntry
		if json.Unmarshal([]byte(item), &entry) == nil {
			entries = append(entries, &entry)
		}
	}
	return entries, nil
}
