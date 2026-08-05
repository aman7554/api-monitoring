package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type MonitorType string

const (
	MonitorTypeHTTP MonitorType = "http"
	MonitorTypeDNS  MonitorType = "dns"
	MonitorTypeSSL  MonitorType = "ssl"
)

type MonitorStatus string

const (
	MonitorStatusUp       MonitorStatus = "up"
	MonitorStatusDown     MonitorStatus = "down"
	MonitorStatusDegraded MonitorStatus = "degraded"
	MonitorStatusPending  MonitorStatus = "pending"
)

type Monitor struct {
	ID                   uuid.UUID       `json:"id"`
	ProjectID            uuid.UUID       `json:"project_id"`
	Name                 string          `json:"name"`
	Type                 MonitorType     `json:"type"`
	URL                  string          `json:"url"`
	Method               string          `json:"method"`
	Headers              json.RawMessage `json:"headers"`
	Body                 string          `json:"body,omitempty"`
	AuthConfig           json.RawMessage `json:"auth_config,omitempty"`
	IntervalSeconds      int             `json:"interval_seconds"`
	TimeoutSeconds       int             `json:"timeout_seconds"`
	ExpectedStatusCode   int             `json:"expected_status_code"`
	ResponseKeyword      string          `json:"response_keyword,omitempty"`
	Status               MonitorStatus   `json:"status"`
	ConsecutiveFailures  int             `json:"consecutive_failures"`
	ConsecutiveSuccesses int             `json:"consecutive_successes"`
	FailureThreshold     int             `json:"failure_threshold"`
	RecoveryThreshold    int             `json:"recovery_threshold"`
	LastCheckAt          *time.Time      `json:"last_check_at,omitempty"`
	NextCheckAt          time.Time       `json:"next_check_at"`
	LastSuccessAt        *time.Time      `json:"last_success_at,omitempty"`
	LastFailureAt        *time.Time      `json:"last_failure_at,omitempty"`
	IsActive             bool            `json:"is_active"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}
