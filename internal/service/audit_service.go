package service

import (
	"context"

	"pulsewatch/internal/domain"
	"pulsewatch/internal/repository/postgres"

	"github.com/google/uuid"
)

type AuditService struct {
	repo *postgres.AuditRepository
}

func NewAuditService(repo *postgres.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Log(ctx context.Context, orgID uuid.UUID, actorID *uuid.UUID, action, entityType string, entityID *uuid.UUID, ipAddress string) {
	log := &domain.AuditLog{
		OrgID:      orgID,
		ActorID:    actorID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		IPAddress:  ipAddress,
	}
	_ = s.repo.Create(ctx, log)
}

func (s *AuditService) ListOrgLogs(ctx context.Context, orgID uuid.UUID, limit int) ([]*domain.AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListByOrg(ctx, orgID, limit)
}
