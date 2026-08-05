package handler

import (
	"net/http"

	"pulsewatch/internal/service"

	"github.com/gin-gonic/gin"
)

type StatusHandler struct {
	statusSvc *service.StatusPageService
}

func NewStatusHandler(statusSvc *service.StatusPageService) *StatusHandler {
	return &StatusHandler{statusSvc: statusSvc}
}

func (h *StatusHandler) GetPublicStatus(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status page slug required"})
		return
	}

	res, err := h.statusSvc.GetPublicStatus(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
