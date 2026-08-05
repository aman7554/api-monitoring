package postgres

import (
	"context"
	"database/sql"
	"errors"

	"pulsewatch/internal/domain"

	"github.com/google/uuid"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, p *domain.Project) error {
	query := `
		INSERT INTO projects (id, org_id, name, slug, description, is_public_status_page, status_page_slug, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	var statusSlug *string
	if p.StatusPageSlug != "" {
		statusSlug = &p.StatusPageSlug
	}

	return r.db.QueryRowContext(ctx, query, p.ID, p.OrgID, p.Name, p.Slug, p.Description, p.IsPublicStatusPage, statusSlug).
		Scan(&p.CreatedAt, &p.UpdatedAt)
}

func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	query := `
		SELECT id, org_id, name, slug, description, is_public_status_page, COALESCE(status_page_slug, ''), created_at, updated_at
		FROM projects WHERE id = $1
	`
	p := &domain.Project{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.OrgID, &p.Name, &p.Slug, &p.Description, &p.IsPublicStatusPage, &p.StatusPageSlug, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *ProjectRepository) GetByStatusSlug(ctx context.Context, slug string) (*domain.Project, error) {
	query := `
		SELECT id, org_id, name, slug, description, is_public_status_page, COALESCE(status_page_slug, ''), created_at, updated_at
		FROM projects WHERE status_page_slug = $1 AND is_public_status_page = TRUE
	`
	p := &domain.Project{}
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&p.ID, &p.OrgID, &p.Name, &p.Slug, &p.Description, &p.IsPublicStatusPage, &p.StatusPageSlug, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *ProjectRepository) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*domain.Project, error) {
	query := `
		SELECT id, org_id, name, slug, description, is_public_status_page, COALESCE(status_page_slug, ''), created_at, updated_at
		FROM projects WHERE org_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		p := &domain.Project{}
		if err := rows.Scan(&p.ID, &p.OrgID, &p.Name, &p.Slug, &p.Description, &p.IsPublicStatusPage, &p.StatusPageSlug, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}
