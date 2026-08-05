package service

import (
	"context"
	"fmt"
	"time"

	"pulsewatch/internal/domain"
	"pulsewatch/internal/repository/postgres"
	"pulsewatch/internal/telemetry"

	"github.com/google/uuid"
)

type IncidentService struct {
	incRepo  *postgres.IncidentRepository
	monRepo  *postgres.MonitorRepository
	notifSvc *NotificationService
}

func NewIncidentService(incRepo *postgres.IncidentRepository, monRepo *postgres.MonitorRepository, notifSvc *NotificationService) *IncidentService {
	return &IncidentService{
		incRepo:  incRepo,
		monRepo:  monRepo,
		notifSvc: notifSvc,
	}
}

func (s *IncidentService) ProcessCheckResult(ctx context.Context, m *domain.Monitor, res *domain.CheckResult) error {
	now := time.Now()
	m.LastCheckAt = &now
	m.NextCheckAt = now.Add(time.Duration(m.IntervalSeconds) * time.Second)

	if res.Status == domain.MonitorStatusUp {
		m.ConsecutiveSuccesses++
		m.ConsecutiveFailures = 0
		m.LastSuccessAt = &now
		m.Status = domain.MonitorStatusUp

		// Check if an active incident should be resolved
		activeInc, err := s.incRepo.GetActiveByMonitor(ctx, m.ID)
		if err == nil && activeInc != nil {
			if m.ConsecutiveSuccesses >= m.RecoveryThreshold {
				if err := s.incRepo.Resolve(ctx, activeInc.ID); err == nil {
					activeInc.Status = domain.IncidentStatusResolved
					_ = s.notifSvc.NotifyIncident(ctx, activeInc, m, "incident.resolved")
					telemetry.ActiveIncidentsGauge.Dec()
				}
			}
		}
	} else {
		m.ConsecutiveFailures++
		m.ConsecutiveSuccesses = 0
		m.LastFailureAt = &now
		m.Status = res.Status

		// Check if a new incident should be created
		activeInc, err := s.incRepo.GetActiveByMonitor(ctx, m.ID)
		if err != nil || activeInc == nil {
			if m.ConsecutiveFailures >= m.FailureThreshold {
				title := fmt.Sprintf("Monitor '%s' is DOWN", m.Name)
				if res.ErrorMessage != "" {
					title = fmt.Sprintf("Monitor '%s': %s", m.Name, res.ErrorMessage)
				}
				newInc := &domain.Incident{
					ID:        uuid.New(),
					MonitorID: m.ID,
					ProjectID: m.ProjectID,
					Title:     title,
					Status:    domain.IncidentStatusOngoing,
					Cause:     res.ErrorMessage,
				}
				if err := s.incRepo.Create(ctx, newInc); err == nil {
					_ = s.notifSvc.NotifyIncident(ctx, newInc, m, "incident.opened")
					telemetry.ActiveIncidentsGauge.Inc()
				}
			}
		}
	}

	return s.monRepo.UpdateStatus(ctx, m)
}
