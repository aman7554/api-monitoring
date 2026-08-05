package postgres

import (
	"context"
	"database/sql"
	"time"

	"pulsewatch/internal/domain"

	"github.com/google/uuid"
)

type CheckRepository struct {
	db *sql.DB
}

func NewCheckRepository(db *sql.DB) *CheckRepository {
	return &CheckRepository{db: db}
}

func (r *CheckRepository) Create(ctx context.Context, res *domain.CheckResult) error {
	query := `
		INSERT INTO check_results (id, monitor_id, status, status_code, latency_ms, dns_time_ms, ssl_days_remaining, error_message, checked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	if res.ID == uuid.Nil {
		res.ID = uuid.New()
	}
	if res.CheckedAt.IsZero() {
		res.CheckedAt = time.Now()
	}
	_, err := r.db.ExecContext(ctx, query,
		res.ID, res.MonitorID, res.Status, res.StatusCode, res.LatencyMS, res.DNSTimeMS, res.SSLDaysRemaining, res.ErrorMessage, res.CheckedAt,
	)
	return err
}

func (r *CheckRepository) ListByMonitor(ctx context.Context, monitorID uuid.UUID, limit int) ([]*domain.CheckResult, error) {
	query := `
		SELECT id, monitor_id, status, status_code, latency_ms, dns_time_ms, ssl_days_remaining, COALESCE(error_message, ''), checked_at
		FROM check_results
		WHERE monitor_id = $1
		ORDER BY checked_at DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, monitorID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.CheckResult
	for rows.Next() {
		res := &domain.CheckResult{}
		if err := rows.Scan(&res.ID, &res.MonitorID, &res.Status, &res.StatusCode, &res.LatencyMS, &res.DNSTimeMS, &res.SSLDaysRemaining, &res.ErrorMessage, &res.CheckedAt); err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	return results, nil
}

type MetricsSummary struct {
	TotalChecks   int     `json:"total_checks"`
	SuccessChecks int     `json:"success_checks"`
	UptimePercent float64 `json:"uptime_percent"`
	AvgLatencyMS  float64 `json:"avg_latency_ms"`
}

func (r *CheckRepository) GetMetricsSummary(ctx context.Context, projectID uuid.UUID, since time.Time) (*MetricsSummary, error) {
	query := `
		SELECT 
			COUNT(cr.id) as total_checks,
			COUNT(CASE WHEN cr.status = 'up' THEN 1 END) as success_checks,
			COALESCE(AVG(cr.latency_ms), 0) as avg_latency
		FROM check_results cr
		JOIN monitors m ON cr.monitor_id = m.id
		WHERE m.project_id = $1 AND cr.checked_at >= $2
	`
	summary := &MetricsSummary{}
	err := r.db.QueryRowContext(ctx, query, projectID, since).Scan(&summary.TotalChecks, &summary.SuccessChecks, &summary.AvgLatencyMS)
	if err != nil {
		return nil, err
	}
	if summary.TotalChecks > 0 {
		summary.UptimePercent = (float64(summary.SuccessChecks) / float64(summary.TotalChecks)) * 100.0
	} else {
		summary.UptimePercent = 100.0
	}
	return summary, nil
}
