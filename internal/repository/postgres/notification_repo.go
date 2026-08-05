package postgres

import (
	"context"
	"database/sql"

	"pulsewatch/internal/domain"

	"github.com/google/uuid"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	query := `
		INSERT INTO notifications (id, project_id, type, target, is_enabled, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING created_at
	`
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return r.db.QueryRowContext(ctx, query, n.ID, n.ProjectID, n.Type, n.Target, n.IsEnabled).Scan(&n.CreatedAt)
}

func (r *NotificationRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]*domain.Notification, error) {
	query := `
		SELECT id, project_id, type, target, is_enabled, created_at
		FROM notifications WHERE project_id = $1 AND is_enabled = TRUE
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Notification
	for rows.Next() {
		n := &domain.Notification{}
		if err := rows.Scan(&n.ID, &n.ProjectID, &n.Type, &n.Target, &n.IsEnabled, &n.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, n)
	}
	return list, nil
}
