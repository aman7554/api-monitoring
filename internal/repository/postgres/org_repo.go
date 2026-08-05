package postgres

import (
	"context"
	"database/sql"
	"errors"

	"pulsewatch/internal/domain"

	"github.com/google/uuid"
)

type OrgRepository struct {
	db *sql.DB
}

func NewOrgRepository(db *sql.DB) *OrgRepository {
	return &OrgRepository{db: db}
}

func (r *OrgRepository) CreateOrg(ctx context.Context, org *domain.Organization) error {
	query := `
		INSERT INTO organizations (id, name, slug, plan, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	if org.ID == uuid.Nil {
		org.ID = uuid.New()
	}
	return r.db.QueryRowContext(ctx, query, org.ID, org.Name, org.Slug, org.Plan).
		Scan(&org.CreatedAt, &org.UpdatedAt)
}

func (r *OrgRepository) GetOrgByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	query := `SELECT id, name, slug, plan, created_at, updated_at FROM organizations WHERE id = $1`
	org := &domain.Organization{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&org.ID, &org.Name, &org.Slug, &org.Plan, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return org, nil
}

func (r *OrgRepository) AddMember(ctx context.Context, m *domain.Member) error {
	query := `
		INSERT INTO members (id, org_id, user_id, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return r.db.QueryRowContext(ctx, query, m.ID, m.OrgID, m.UserID, m.Role).
		Scan(&m.CreatedAt, &m.UpdatedAt)
}

func (r *OrgRepository) GetMember(ctx context.Context, orgID, userID uuid.UUID) (*domain.Member, error) {
	query := `SELECT id, org_id, user_id, role, created_at, updated_at FROM members WHERE org_id = $1 AND user_id = $2`
	m := &domain.Member{}
	err := r.db.QueryRowContext(ctx, query, orgID, userID).Scan(
		&m.ID, &m.OrgID, &m.UserID, &m.Role, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return m, nil
}

func (r *OrgRepository) ListUserOrgs(ctx context.Context, userID uuid.UUID) ([]*domain.Organization, error) {
	query := `
		SELECT o.id, o.name, o.slug, o.plan, o.created_at, o.updated_at
		FROM organizations o
		JOIN members m ON o.id = m.org_id
		WHERE m.user_id = $1
		ORDER BY o.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*domain.Organization
	for rows.Next() {
		org := &domain.Organization{}
		if err := rows.Scan(&org.ID, &org.Name, &org.Slug, &org.Plan, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, nil
}
