package domain

import (
	"time"

	"github.com/google/uuid"
)

type CheckResult struct {
	ID               uuid.UUID     `json:"id"`
	MonitorID        uuid.UUID     `json:"monitor_id"`
	Status           MonitorStatus `json:"status"`
	StatusCode       int           `json:"status_code"`
	LatencyMS        int           `json:"latency_ms"`
	DNSTimeMS        int           `json:"dns_time_ms"`
	SSLDaysRemaining *int          `json:"ssl_days_remaining,omitempty"`
	ErrorMessage     string        `json:"error_message,omitempty"`
	CheckedAt        time.Time     `json:"checked_at"`
}
