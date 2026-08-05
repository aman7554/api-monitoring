package service

import (
	"context"

	"pulsewatch/internal/domain"
	"pulsewatch/internal/repository/postgres"

	"github.com/google/uuid"
)

type OrgService struct {
	orgRepo *postgres.OrgRepository
}

func NewOrgService(orgRepo *postgres.OrgRepository) *OrgService {
	return &OrgService{orgRepo: orgRepo}
}

func (s *OrgService) CreateOrg(ctx context.Context, userID uuid.UUID, name, slug, plan string) (*domain.Organization, error) {
	if plan == "" {
		plan = "free"
	}
	org := &domain.Organization{
		ID:   uuid.New(),
		Name: name,
		Slug: slug,
		Plan: plan,
	}

	if err := s.orgRepo.CreateOrg(ctx, org); err != nil {
		return nil, err
	}

	// Add creator as Org Owner
	member := &domain.Member{
		ID:     uuid.New(),
		OrgID:  org.ID,
		UserID: userID,
		Role:   domain.OrgRoleOwner,
	}

	if err := s.orgRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	return org, nil
}

func (s *OrgService) GetUserOrgs(ctx context.Context, userID uuid.UUID) ([]*domain.Organization, error) {
	return s.orgRepo.ListUserOrgs(ctx, userID)
}

func (s *OrgService) CheckOrgPermission(ctx context.Context, orgID, userID uuid.UUID, requiredRole domain.OrgRole) error {
	m, err := s.orgRepo.GetMember(ctx, orgID, userID)
	if err != nil {
		return domain.ErrForbidden
	}

	if requiredRole == domain.OrgRoleOwner && m.Role != domain.OrgRoleOwner {
		return domain.ErrForbidden
	}
	if requiredRole == domain.OrgRoleAdmin && (m.Role != domain.OrgRoleOwner && m.Role != domain.OrgRoleAdmin) {
		return domain.ErrForbidden
	}

	return nil
}
