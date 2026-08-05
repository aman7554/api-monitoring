package service

import (
	"context"
	"time"

	"pulsewatch/internal/domain"
	"pulsewatch/internal/repository/postgres"

	"github.com/google/uuid"
)

type MonitorService struct {
	monRepo *postgres.MonitorRepository
}

func NewMonitorService(monRepo *postgres.MonitorRepository) *MonitorService {
	return &MonitorService{monRepo: monRepo}
}

func (s *MonitorService) CreateMonitor(ctx context.Context, m *domain.Monitor) (*domain.Monitor, error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.Type == "" {
		m.Type = domain.MonitorTypeHTTP
	}
	if m.Method == "" {
		m.Method = "GET"
	}
	if m.IntervalSeconds <= 0 {
		m.IntervalSeconds = 60
	}
	if m.TimeoutSeconds <= 0 {
		m.TimeoutSeconds = 10
	}
	if m.FailureThreshold <= 0 {
		m.FailureThreshold = 3
	}
	if m.RecoveryThreshold <= 0 {
		m.RecoveryThreshold = 2
	}
	m.Status = domain.MonitorStatusPending
	m.IsActive = true
	m.NextCheckAt = time.Now()

	if err := s.monRepo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *MonitorService) GetMonitor(ctx context.Context, id uuid.UUID) (*domain.Monitor, error) {
	return s.monRepo.GetByID(ctx, id)
}

func (s *MonitorService) ListProjectMonitors(ctx context.Context, projectID uuid.UUID) ([]*domain.Monitor, error) {
	return s.monRepo.ListByProject(ctx, projectID)
}

func (s *MonitorService) ListDueMonitors(ctx context.Context, limit int) ([]*domain.Monitor, error) {
	return s.monRepo.ListDueMonitors(ctx, limit)
}

func (s *MonitorService) DeleteMonitor(ctx context.Context, id uuid.UUID) error {
	return s.monRepo.Delete(ctx, id)
}
