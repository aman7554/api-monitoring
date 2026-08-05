package domain

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID                  uuid.UUID `json:"id"`
	OrgID               uuid.UUID `json:"org_id"`
	Name                string    `json:"name"`
	Slug                string    `json:"slug"`
	Description         string    `json:"description"`
	IsPublicStatusPage  bool      `json:"is_public_status_page"`
	StatusPageSlug      string    `json:"status_page_slug,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
