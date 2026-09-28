package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/preset"
)

// "custom" isn't listed here — it has no fixed dimensions, only a request-time path.
func (h *Handler) Presets(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"presets": preset.All()})
}
