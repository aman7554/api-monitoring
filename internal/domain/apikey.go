package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ApiKey struct {
	ID         uuid.UUID       `json:"id"`
	ProjectID  uuid.UUID       `json:"project_id"`
	Name       string          `json:"name"`
	KeyPrefix  string          `json:"key_prefix"`
	KeyHash    string          `json:"-"`
	Scopes     json.RawMessage `json:"scopes"`
	LastUsedAt *time.Time      `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time      `json:"expires_at,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}
