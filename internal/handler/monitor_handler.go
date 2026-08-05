package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"pulsewatch/internal/domain"
	"pulsewatch/internal/repository/postgres"
	"pulsewatch/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MonitorHandler struct {
	monSvc   *service.MonitorService
	checkRepo *postgres.CheckRepository
	auditSvc *service.AuditService
}

func NewMonitorHandler(monSvc *service.MonitorService, checkRepo *postgres.CheckRepository, auditSvc *service.AuditService) *MonitorHandler {
	return &MonitorHandler{
		monSvc:    monSvc,
		checkRepo: checkRepo,
		auditSvc:  auditSvc,
	}
}

type CreateMonitorRequest struct {
	ProjectID          uuid.UUID         `json:"project_id" binding:"required"`
	Name               string            `json:"name" binding:"required"`
	Type               domain.MonitorType `json:"type"`
	URL                string            `json:"url" binding:"required"`
	Method             string            `json:"method"`
	Headers            map[string]string `json:"headers"`
	Body               string            `json:"body"`
	AuthConfig         map[string]any    `json:"auth_config"`
	IntervalSeconds    int               `json:"interval_seconds"`
	TimeoutSeconds     int               `json:"timeout_seconds"`
	ExpectedStatusCode int               `json:"expected_status_code"`
	ResponseKeyword    string            `json:"response_keyword"`
}

func (h *MonitorHandler) CreateMonitor(c *gin.Context) {
	var req CreateMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	headersBytes, _ := json.Marshal(req.Headers)
	authBytes, _ := json.Marshal(req.AuthConfig)

	m := &domain.Monitor{
		ProjectID:          req.ProjectID,
		Name:               req.Name,
		Type:               req.Type,
		URL:                req.URL,
		Method:             req.Method,
		Headers:            headersBytes,
		Body:               req.Body,
		AuthConfig:         authBytes,
		IntervalSeconds:    req.IntervalSeconds,
		TimeoutSeconds:     req.TimeoutSeconds,
		ExpectedStatusCode: req.ExpectedStatusCode,
		ResponseKeyword:    req.ResponseKeyword,
	}

	created, err := h.monSvc.CreateMonitor(c.Request.Context(), m)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *MonitorHandler) ListProjectMonitors(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id parameter required"})
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	monitors, err := h.monSvc.ListProjectMonitors(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"monitors": monitors})
}

func (h *MonitorHandler) GetMonitor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid monitor ID"})
		return
	}

	m, err := h.monSvc.GetMonitor(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, m)
}

func (h *MonitorHandler) DeleteMonitor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid monitor ID"})
		return
	}

	if err := h.monSvc.DeleteMonitor(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "monitor deleted successfully"})
}

func (h *MonitorHandler) ListMonitorChecks(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid monitor ID"})
		return
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	checks, err := h.checkRepo.ListByMonitor(c.Request.Context(), id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"checks": checks})
}
