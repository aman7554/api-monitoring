package handler

import (
	"net/http"

	"pulsewatch/internal/middleware"
	"pulsewatch/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	projSvc  *service.ProjectService
	auditSvc *service.AuditService
}

func NewProjectHandler(projSvc *service.ProjectService, auditSvc *service.AuditService) *ProjectHandler {
	return &ProjectHandler{
		projSvc:  projSvc,
		auditSvc: auditSvc,
	}
}

type CreateProjectRequest struct {
	OrgID              uuid.UUID `json:"org_id" binding:"required"`
	Name               string    `json:"name" binding:"required"`
	Slug               string    `json:"slug" binding:"required"`
	Description        string    `json:"description"`
	IsPublicStatusPage bool      `json:"is_public_status_page"`
	StatusPageSlug     string    `json:"status_page_slug"`
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, err := h.projSvc.CreateProject(c.Request.Context(), req.OrgID, req.Name, req.Slug, req.Description, req.IsPublicStatusPage, req.StatusPageSlug)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if userIDVal, exists := c.Get(middleware.CtxUserIDKey); exists {
		userID := userIDVal.(uuid.UUID)
		h.auditSvc.Log(c.Request.Context(), req.OrgID, &userID, "project.create", "project", &p.ID, c.ClientIP())
	}

	c.JSON(http.StatusCreated, p)
}

func (h *ProjectHandler) ListOrgProjects(c *gin.Context) {
	orgIDStr := c.Query("org_id")
	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "org_id query parameter required"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org_id"})
		return
	}

	projects, err := h.projSvc.ListOrgProjects(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"projects": projects})
}

func (h *ProjectHandler) GetProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	p, err := h.projSvc.GetProjectByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, p)
}
