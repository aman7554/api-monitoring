package handler

import (
	"net/http"

	"pulsewatch/internal/repository/postgres"
	"pulsewatch/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ApiKeyHandler struct {
	authSvc *service.AuthService
	keyRepo *postgres.ApiKeyRepository
}

func NewApiKeyHandler(authSvc *service.AuthService, keyRepo *postgres.ApiKeyRepository) *ApiKeyHandler {
	return &ApiKeyHandler{
		authSvc: authSvc,
		keyRepo: keyRepo,
	}
}

type CreateApiKeyRequest struct {
	ProjectID uuid.UUID `json:"project_id" binding:"required"`
	Name      string    `json:"name" binding:"required"`
}

func (h *ApiKeyHandler) CreateApiKey(c *gin.Context) {
	var req CreateApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rawToken, key, err := h.authSvc.CreateApiKey(c.Request.Context(), req.ProjectID, req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"api_key": rawToken,
		"key":     key,
	})
}

func (h *ApiKeyHandler) ListProjectKeys(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id required"})
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	keys, err := h.keyRepo.ListByProject(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"keys": keys})
}
