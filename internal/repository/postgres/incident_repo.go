package postgres

import (
	"context"
	"database/sql"
	"errors"

	"pulsewatch/internal/domain"

	"github.com/google/uuid"
)

type IncidentRepository struct {
	db *sql.DB
}

func NewIncidentRepository(db *sql.DB) *IncidentRepository {
	return &IncidentRepository{db: db}
}

func (r *IncidentRepository) Create(ctx context.Context, inc *domain.Incident) error {
	query := `
		INSERT INTO incidents (id, monitor_id, project_id, title, status, cause, started_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW(), NOW())
		RETURNING started_at, created_at, updated_at
	`
	if inc.ID == uuid.Nil {
		inc.ID = uuid.New()
	}
	return r.db.QueryRowContext(ctx, query, inc.ID, inc.MonitorID, inc.ProjectID, inc.Title, inc.Status, inc.Cause).
		Scan(&inc.StartedAt, &inc.CreatedAt, &inc.UpdatedAt)
}

func (r *IncidentRepository) GetActiveByMonitor(ctx context.Context, monitorID uuid.UUID) (*domain.Incident, error) {
	query := `
		SELECT id, monitor_id, project_id, title, status, cause, started_at, resolved_at, created_at, updated_at
		FROM incidents
		WHERE monitor_id = $1 AND status != 'resolved'
		ORDER BY started_at DESC LIMIT 1
	`
	inc := &domain.Incident{}
	err := r.db.QueryRowContext(ctx, query, monitorID).Scan(
		&inc.ID, &inc.MonitorID, &inc.ProjectID, &inc.Title, &inc.Status, &inc.Cause, &inc.StartedAt, &inc.ResolvedAt, &inc.CreatedAt, &inc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return inc, nil
}

func (r *IncidentRepository) Resolve(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE incidents SET status = 'resolved', resolved_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *IncidentRepository) ListByProject(ctx context.Context, projectID uuid.UUID, limit int) ([]*domain.Incident, error) {
	query := `
		SELECT id, monitor_id, project_id, title, status, cause, started_at, resolved_at, created_at, updated_at
		FROM incidents
		WHERE project_id = $1
		ORDER BY started_at DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []*domain.Incident
	for rows.Next() {
		inc := &domain.Incident{}
		if err := rows.Scan(&inc.ID, &inc.MonitorID, &inc.ProjectID, &inc.Title, &inc.Status, &inc.Cause, &inc.StartedAt, &inc.ResolvedAt, &inc.CreatedAt, &inc.UpdatedAt); err != nil {
			return nil, err
		}
		incidents = append(incidents, inc)
	}
	return incidents, nil
}
