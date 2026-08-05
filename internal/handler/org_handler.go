package handler

import (
	"net/http"

	"pulsewatch/internal/middleware"
	"pulsewatch/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrgHandler struct {
	orgSvc   *service.OrgService
	auditSvc *service.AuditService
}

func NewOrgHandler(orgSvc *service.OrgService, auditSvc *service.AuditService) *OrgHandler {
	return &OrgHandler{
		orgSvc:   orgSvc,
		auditSvc: auditSvc,
	}
}

type CreateOrgRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
	Plan string `json:"plan"`
}

func (h *OrgHandler) CreateOrg(c *gin.Context) {
	userIDVal, _ := c.Get(middleware.CtxUserIDKey)
	userID := userIDVal.(uuid.UUID)

	var req CreateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	org, err := h.orgSvc.CreateOrg(c.Request.Context(), userID, req.Name, req.Slug, req.Plan)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.auditSvc.Log(c.Request.Context(), org.ID, &userID, "org.create", "organization", &org.ID, c.ClientIP())

	c.JSON(http.StatusCreated, org)
}

func (h *OrgHandler) ListUserOrgs(c *gin.Context) {
	userIDVal, _ := c.Get(middleware.CtxUserIDKey)
	userID := userIDVal.(uuid.UUID)

	orgs, err := h.orgSvc.GetUserOrgs(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"organizations": orgs})
}
