package postgres

import (
	"context"
	"database/sql"
	"errors"

	"pulsewatch/internal/domain"

	"github.com/google/uuid"
)

type ApiKeyRepository struct {
	db *sql.DB
}

func NewApiKeyRepository(db *sql.DB) *ApiKeyRepository {
	return &ApiKeyRepository{db: db}
}

func (r *ApiKeyRepository) Create(ctx context.Context, k *domain.ApiKey) error {
	query := `
		INSERT INTO api_keys (id, project_id, name, key_prefix, key_hash, scopes, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING created_at
	`
	if k.ID == uuid.Nil {
		k.ID = uuid.New()
	}
	if k.Scopes == nil {
		k.Scopes = []byte(`["read", "write"]`)
	}
	return r.db.QueryRowContext(ctx, query, k.ID, k.ProjectID, k.Name, k.KeyPrefix, k.KeyHash, k.Scopes, k.ExpiresAt).Scan(&k.CreatedAt)
}

func (r *ApiKeyRepository) GetByKeyHash(ctx context.Context, hash string) (*domain.ApiKey, error) {
	query := `
		SELECT id, project_id, name, key_prefix, key_hash, scopes, last_used_at, expires_at, created_at
		FROM api_keys WHERE key_hash = $1
	`
	k := &domain.ApiKey{}
	err := r.db.QueryRowContext(ctx, query, hash).Scan(
		&k.ID, &k.ProjectID, &k.Name, &k.KeyPrefix, &k.KeyHash, &k.Scopes, &k.LastUsedAt, &k.ExpiresAt, &k.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return k, nil
}

func (r *ApiKeyRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *ApiKeyRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]*domain.ApiKey, error) {
	query := `
		SELECT id, project_id, name, key_prefix, scopes, last_used_at, expires_at, created_at
		FROM api_keys WHERE project_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*domain.ApiKey
	for rows.Next() {
		k := &domain.ApiKey{}
		if err := rows.Scan(&k.ID, &k.ProjectID, &k.Name, &k.KeyPrefix, &k.Scopes, &k.LastUsedAt, &k.ExpiresAt, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}
