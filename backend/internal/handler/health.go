package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Healthz is a liveness/readiness probe. Always 200 while the process is up.
func (h *Handler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
