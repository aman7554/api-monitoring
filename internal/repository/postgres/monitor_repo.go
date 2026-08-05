package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"pulsewatch/internal/domain"

	"github.com/google/uuid"
)

type MonitorRepository struct {
	db *sql.DB
}

func NewMonitorRepository(db *sql.DB) *MonitorRepository {
	return &MonitorRepository{db: db}
}

func (r *MonitorRepository) Create(ctx context.Context, m *domain.Monitor) error {
	query := `
		INSERT INTO monitors (
			id, project_id, name, type, url, method, headers, body, auth_config,
			interval_seconds, timeout_seconds, expected_status_code, response_keyword,
			status, consecutive_failures, consecutive_successes, failure_threshold, recovery_threshold,
			next_check_at, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9,
			$10, $11, $12, $13,
			$14, $15, $16, $17, $18,
			$19, $20, NOW(), NOW()
		)
		RETURNING created_at, updated_at
	`
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.NextCheckAt.IsZero() {
		m.NextCheckAt = time.Now()
	}
	if m.Headers == nil {
		m.Headers = []byte("{}")
	}
	if m.AuthConfig == nil {
		m.AuthConfig = []byte("{}")
	}

	return r.db.QueryRowContext(ctx, query,
		m.ID, m.ProjectID, m.Name, m.Type, m.URL, m.Method, m.Headers, m.Body, m.AuthConfig,
		m.IntervalSeconds, m.TimeoutSeconds, m.ExpectedStatusCode, m.ResponseKeyword,
		m.Status, m.ConsecutiveFailures, m.ConsecutiveSuccesses, m.FailureThreshold, m.RecoveryThreshold,
		m.NextCheckAt, m.IsActive,
	).Scan(&m.CreatedAt, &m.UpdatedAt)
}

func (r *MonitorRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Monitor, error) {
	query := `
		SELECT id, project_id, name, type, url, method, headers, COALESCE(body, ''), auth_config,
		       interval_seconds, timeout_seconds, expected_status_code, COALESCE(response_keyword, ''),
		       status, consecutive_failures, consecutive_successes, failure_threshold, recovery_threshold,
		       last_check_at, next_check_at, last_success_at, last_failure_at, is_active, created_at, updated_at
		FROM monitors WHERE id = $1
	`
	m := &domain.Monitor{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.ProjectID, &m.Name, &m.Type, &m.URL, &m.Method, &m.Headers, &m.Body, &m.AuthConfig,
		&m.IntervalSeconds, &m.TimeoutSeconds, &m.ExpectedStatusCode, &m.ResponseKeyword,
		&m.Status, &m.ConsecutiveFailures, &m.ConsecutiveSuccesses, &m.FailureThreshold, &m.RecoveryThreshold,
		&m.LastCheckAt, &m.NextCheckAt, &m.LastSuccessAt, &m.LastFailureAt, &m.IsActive, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return m, nil
}

func (r *MonitorRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]*domain.Monitor, error) {
	query := `
		SELECT id, project_id, name, type, url, method, headers, COALESCE(body, ''), auth_config,
		       interval_seconds, timeout_seconds, expected_status_code, COALESCE(response_keyword, ''),
		       status, consecutive_failures, consecutive_successes, failure_threshold, recovery_threshold,
		       last_check_at, next_check_at, last_success_at, last_failure_at, is_active, created_at, updated_at
		FROM monitors WHERE project_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monitors []*domain.Monitor
	for rows.Next() {
		m := &domain.Monitor{}
		if err := rows.Scan(
			&m.ID, &m.ProjectID, &m.Name, &m.Type, &m.URL, &m.Method, &m.Headers, &m.Body, &m.AuthConfig,
			&m.IntervalSeconds, &m.TimeoutSeconds, &m.ExpectedStatusCode, &m.ResponseKeyword,
			&m.Status, &m.ConsecutiveFailures, &m.ConsecutiveSuccesses, &m.FailureThreshold, &m.RecoveryThreshold,
			&m.LastCheckAt, &m.NextCheckAt, &m.LastSuccessAt, &m.LastFailureAt, &m.IsActive, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		monitors = append(monitors, m)
	}
	return monitors, nil
}

func (r *MonitorRepository) ListDueMonitors(ctx context.Context, limit int) ([]*domain.Monitor, error) {
	query := `
		SELECT id, project_id, name, type, url, method, headers, COALESCE(body, ''), auth_config,
		       interval_seconds, timeout_seconds, expected_status_code, COALESCE(response_keyword, ''),
		       status, consecutive_failures, consecutive_successes, failure_threshold, recovery_threshold,
		       last_check_at, next_check_at, last_success_at, last_failure_at, is_active, created_at, updated_at
		FROM monitors
		WHERE is_active = TRUE AND next_check_at <= NOW()
		ORDER BY next_check_at ASC
		LIMIT $1
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monitors []*domain.Monitor
	for rows.Next() {
		m := &domain.Monitor{}
		if err := rows.Scan(
			&m.ID, &m.ProjectID, &m.Name, &m.Type, &m.URL, &m.Method, &m.Headers, &m.Body, &m.AuthConfig,
			&m.IntervalSeconds, &m.TimeoutSeconds, &m.ExpectedStatusCode, &m.ResponseKeyword,
			&m.Status, &m.ConsecutiveFailures, &m.ConsecutiveSuccesses, &m.FailureThreshold, &m.RecoveryThreshold,
			&m.LastCheckAt, &m.NextCheckAt, &m.LastSuccessAt, &m.LastFailureAt, &m.IsActive, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		monitors = append(monitors, m)
	}
	return monitors, nil
}

func (r *MonitorRepository) UpdateStatus(ctx context.Context, m *domain.Monitor) error {
	query := `
		UPDATE monitors SET
			status = $1,
			consecutive_failures = $2,
			consecutive_successes = $3,
			last_check_at = $4,
			next_check_at = $5,
			last_success_at = COALESCE($6, last_success_at),
			last_failure_at = COALESCE($7, last_failure_at),
			updated_at = NOW()
		WHERE id = $8
	`
	_, err := r.db.ExecContext(ctx, query,
		m.Status, m.ConsecutiveFailures, m.ConsecutiveSuccesses,
		m.LastCheckAt, m.NextCheckAt, m.LastSuccessAt, m.LastFailureAt, m.ID,
	)
	return err
}

func (r *MonitorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM monitors WHERE id = $1`, id)
	return err
}
