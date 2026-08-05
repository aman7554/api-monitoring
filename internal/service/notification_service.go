package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"pulsewatch/internal/domain"
	"pulsewatch/internal/repository/postgres"
)

type NotificationService struct {
	repo       *postgres.NotificationRepository
	httpClient *http.Client
}

func NewNotificationService(repo *postgres.NotificationRepository) *NotificationService {
	return &NotificationService{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type IncidentAlertPayload struct {
	EventType string          `json:"event_type"` // "incident.opened" or "incident.resolved"
	Incident  *domain.Incident `json:"incident"`
	Monitor   *domain.Monitor  `json:"monitor"`
	Timestamp time.Time        `json:"timestamp"`
}

func (s *NotificationService) NotifyIncident(ctx context.Context, inc *domain.Incident, m *domain.Monitor, eventType string) error {
	targets, err := s.repo.ListByProject(ctx, inc.ProjectID)
	if err != nil {
		return err
	}

	payload := IncidentAlertPayload{
		EventType: eventType,
		Incident:  inc,
		Monitor:   m,
		Timestamp: time.Now(),
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	for _, target := range targets {
		if !target.IsEnabled {
			continue
		}

		switch target.Type {
		case domain.NotificationTypeWebhook:
			go s.sendWebhook(target.Target, bodyBytes)
		case domain.NotificationTypeEmail:
			go s.sendEmailAlert(target.Target, payload)
		}
	}
	return nil
}

func (s *NotificationService) sendWebhook(url string, payload []byte) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("[NotificationService] Webhook req error: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PulseWatch-AlertWebhook/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		fmt.Printf("[NotificationService] Webhook send error to %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("[NotificationService] Webhook dispatched to %s with status %d\n", url, resp.StatusCode)
}

func (s *NotificationService) sendEmailAlert(email string, payload IncidentAlertPayload) {
	// Simulated email alert dispatch
	fmt.Printf("[NotificationService] [EMAIL ALERT] To: %s | Event: %s | Monitor: %s | Cause: %s\n",
		email, payload.EventType, payload.Monitor.Name, payload.Incident.Cause)
}
