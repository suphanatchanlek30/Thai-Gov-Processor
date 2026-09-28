package handler

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/preset"
)

// Synthetic fixture, not a real photo — used to smoke-test the pipeline
// (see docs/architecture.md) without a real user upload.
//
//go:embed testdata/sample.jpg
var sampleJPEG []byte

func (h *Handler) Selftest(c *gin.Context) {
	ocsc, ok := preset.Lookup(preset.OCSC)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ocsc preset not found"})
		return
	}

	result, err := h.PhotoService.ProcessPreset(sampleJPEG, preset.OCSC, nil)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	if result.Width != ocsc.Width || result.Height != ocsc.Height {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":    "selftest failed: unexpected output dimensions",
			"expected": gin.H{"width": ocsc.Width, "height": ocsc.Height},
			"actual":   gin.H{"width": result.Width, "height": result.Height},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"width":   result.Width,
		"height":  result.Height,
		"size_kb": result.SizeKB,
		"quality": result.Quality,
	})
}
