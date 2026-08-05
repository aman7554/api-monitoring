package postgres

import (
	"context"
	"database/sql"

	"pulsewatch/internal/domain"

	"github.com/google/uuid"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, org_id, actor_id, action, entity_type, entity_id, metadata, ip_address, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING created_at
	`
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.Metadata == nil {
		log.Metadata = []byte("{}")
	}
	return r.db.QueryRowContext(ctx, query, log.ID, log.OrgID, log.ActorID, log.Action, log.EntityType, log.EntityID, log.Metadata, log.IPAddress).
		Scan(&log.CreatedAt)
}

func (r *AuditRepository) ListByOrg(ctx context.Context, orgID uuid.UUID, limit int) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, org_id, actor_id, action, entity_type, entity_id, metadata, COALESCE(ip_address, ''), created_at
		FROM audit_logs WHERE org_id = $1 ORDER BY created_at DESC LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, orgID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		log := &domain.AuditLog{}
		if err := rows.Scan(&log.ID, &log.OrgID, &log.ActorID, &log.Action, &log.EntityType, &log.EntityID, &log.Metadata, &log.IPAddress, &log.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}
