package health

import (
	"github.com/degeta10/workout-assistant-api/internal/pkg/responses"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/health", h.Check)
}

// Check godoc
// @Summary Health check status
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *Handler) Check(c *gin.Context) {
	health := h.svc.Check(c.Request.Context())

	responses.OK(c, "Health check successful", gin.H{
		"status":      health.Status,
		"version":     health.Version,
		"release_id":  health.ReleaseID,
		"description": health.Description,
		"checks": gin.H{
			"database": gin.H{
				"status":         health.DBStatus,
				"component_type": "datastore",
				"time":           health.CheckedAt,
			},
		},
	})
}
