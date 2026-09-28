package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/preset"
)

// Presets lists every built-in preset (ocsc, passport, teacher). "custom"
// is not listed here since it has no fixed dimensions/size — it's a
// request-time path, not a lookup entry.
func (h *Handler) Presets(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"presets": preset.All()})
}
