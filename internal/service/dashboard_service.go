package service

import (
	"context"
	"time"

	"pulsewatch/internal/domain"
	"pulsewatch/internal/repository/postgres"

	"github.com/google/uuid"
)

type DashboardService struct {
	monRepo  *postgres.MonitorRepository
	checkRepo *postgres.CheckRepository
	incRepo  *postgres.IncidentRepository
}

func NewDashboardService(
	monRepo *postgres.MonitorRepository,
	checkRepo *postgres.CheckRepository,
	incRepo *postgres.IncidentRepository,
) *DashboardService {
	return &DashboardService{
		monRepo:   monRepo,
		checkRepo: checkRepo,
		incRepo:   incRepo,
	}
}

type ProjectDashboardData struct {
	TotalMonitors   int                      `json:"total_monitors"`
	UpMonitors      int                      `json:"up_monitors"`
	DownMonitors    int                      `json:"down_monitors"`
	DegradedMonitors int                     `json:"degraded_monitors"`
	Uptime24h       float64                  `json:"uptime_24h_percent"`
	AvgLatency24hMS float64                  `json:"avg_latency_24h_ms"`
	RecentIncidents []*domain.Incident       `json:"recent_incidents"`
	Monitors        []*domain.Monitor        `json:"monitors"`
}

func (s *DashboardService) GetProjectDashboard(ctx context.Context, projectID uuid.UUID) (*ProjectDashboardData, error) {
	monitors, err := s.monRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	upCount, downCount, degradedCount := 0, 0, 0
	for _, m := range monitors {
		switch m.Status {
		case domain.MonitorStatusUp:
			upCount++
		case domain.MonitorStatusDown:
			downCount++
		case domain.MonitorStatusDegraded:
			degradedCount++
		}
	}

	since := time.Now().Add(-24 * time.Hour)
	summary, err := s.checkRepo.GetMetricsSummary(ctx, projectID, since)
	if err != nil {
		summary = &postgres.MetricsSummary{UptimePercent: 100.0, AvgLatencyMS: 0}
	}

	incidents, err := s.incRepo.ListByProject(ctx, projectID, 5)
	if err != nil {
		incidents = []*domain.Incident{}
	}

	return &ProjectDashboardData{
		TotalMonitors:    len(monitors),
		UpMonitors:       upCount,
		DownMonitors:     downCount,
		DegradedMonitors: degradedCount,
		Uptime24h:        summary.UptimePercent,
		AvgLatency24hMS:  summary.AvgLatencyMS,
		RecentIncidents:  incidents,
		Monitors:         monitors,
	}, nil
}
