package service

import (
	"context"
	"time"

	"pulsewatch/internal/domain"
	"pulsewatch/internal/repository/postgres"
)

type StatusPageService struct {
	projRepo  *postgres.ProjectRepository
	monRepo   *postgres.MonitorRepository
	incRepo   *postgres.IncidentRepository
	checkRepo *postgres.CheckRepository
}

func NewStatusPageService(
	projRepo *postgres.ProjectRepository,
	monRepo *postgres.MonitorRepository,
	incRepo *postgres.IncidentRepository,
	checkRepo *postgres.CheckRepository,
) *StatusPageService {
	return &StatusPageService{
		projRepo:  projRepo,
		monRepo:   monRepo,
		incRepo:   incRepo,
		checkRepo: checkRepo,
	}
}

type PublicStatusPageResponse struct {
	ProjectName     string             `json:"project_name"`
	Description     string             `json:"description"`
	OverallStatus   string             `json:"overall_status"` // "operational", "degraded", "outage"
	Uptime90d       float64            `json:"uptime_90d"`
	Monitors        []*domain.Monitor  `json:"monitors"`
	ActiveIncidents []*domain.Incident `json:"active_incidents"`
}

func (s *StatusPageService) GetPublicStatus(ctx context.Context, slug string) (*PublicStatusPageResponse, error) {
	proj, err := s.projRepo.GetByStatusSlug(ctx, slug)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	monitors, err := s.monRepo.ListByProject(ctx, proj.ID)
	if err != nil {
		return nil, err
	}

	incidents, err := s.incRepo.ListByProject(ctx, proj.ID, 10)
	if err != nil {
		incidents = []*domain.Incident{}
	}

	overall := "operational"
	activeIncidents := []*domain.Incident{}
	for _, inc := range incidents {
		if inc.Status != domain.IncidentStatusResolved {
			activeIncidents = append(activeIncidents, inc)
			overall = "outage"
		}
	}

	if overall == "operational" {
		for _, m := range monitors {
			if m.Status == domain.MonitorStatusDegraded {
				overall = "degraded"
				break
			}
		}
	}

	since90d := time.Now().Add(-90 * 24 * time.Hour)
	metrics, err := s.checkRepo.GetMetricsSummary(ctx, proj.ID, since90d)
	uptime90d := 100.0
	if err == nil && metrics != nil {
		uptime90d = metrics.UptimePercent
	}

	return &PublicStatusPageResponse{
		ProjectName:     proj.Name,
		Description:     proj.Description,
		OverallStatus:   overall,
		Uptime90d:       uptime90d,
		Monitors:        monitors,
		ActiveIncidents: activeIncidents,
	}, nil
}
