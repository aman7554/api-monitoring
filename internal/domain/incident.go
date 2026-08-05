package domain

import (
	"time"

	"github.com/google/uuid"
)

type IncidentStatus string

const (
	IncidentStatusOngoing      IncidentStatus = "ongoing"
	IncidentStatusAcknowledged IncidentStatus = "acknowledged"
	IncidentStatusResolved     IncidentStatus = "resolved"
)

type Incident struct {
	ID         uuid.UUID      `json:"id"`
	MonitorID  uuid.UUID      `json:"monitor_id"`
	ProjectID  uuid.UUID      `json:"project_id"`
	Title      string         `json:"title"`
	Status     IncidentStatus `json:"status"`
	Cause      string         `json:"cause"`
	StartedAt  time.Time      `json:"started_at"`
	ResolvedAt *time.Time     `json:"resolved_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}
