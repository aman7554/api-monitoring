package domain

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeEmail   NotificationType = "email"
	NotificationTypeWebhook NotificationType = "webhook"
)

type Notification struct {
	ID        uuid.UUID        `json:"id"`
	ProjectID uuid.UUID        `json:"project_id"`
	Type      NotificationType `json:"type"`
	Target    string           `json:"target"`
	IsEnabled bool             `json:"is_enabled"`
	CreatedAt time.Time        `json:"created_at"`
}
