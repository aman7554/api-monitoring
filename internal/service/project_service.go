package service

import (
	"context"

	"pulsewatch/internal/domain"
	"pulsewatch/internal/repository/postgres"

	"github.com/google/uuid"
)

type ProjectService struct {
	projectRepo *postgres.ProjectRepository
}

func NewProjectService(projectRepo *postgres.ProjectRepository) *ProjectService {
	return &ProjectService{projectRepo: projectRepo}
}

func (s *ProjectService) CreateProject(ctx context.Context, orgID uuid.UUID, name, slug, description string, isPublic bool, statusSlug string) (*domain.Project, error) {
	p := &domain.Project{
		ID:                 uuid.New(),
		OrgID:              orgID,
		Name:               name,
		Slug:               slug,
		Description:        description,
		IsPublicStatusPage: isPublic,
		StatusPageSlug:     statusSlug,
	}

	if err := s.projectRepo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProjectService) GetProjectByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	return s.projectRepo.GetByID(ctx, id)
}

func (s *ProjectService) ListOrgProjects(ctx context.Context, orgID uuid.UUID) ([]*domain.Project, error) {
	return s.projectRepo.ListByOrg(ctx, orgID)
}
