package handler

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/preset"
)

// sampleJPEG is a solid-color 800x600 JPEG fixture embedded at build time,
// used purely as smoke-test input for the ocsc preset below. It carries no
// real user data.
//
//go:embed testdata/sample.jpg
var sampleJPEG []byte

// Selftest runs the embedded sample image through the ocsc preset and
// checks the output comes back at the preset's exact target dimensions.
// It exists so a Kubernetes PostSync smoke test (see docs/architecture.md)
// can verify the whole photo pipeline — resize + compress + storage — end
// to end without a real user upload.
//
// While PhotoService is still the stub implementation this always returns
// 501; once the real service lands it will resize the fixture to 200x230
// and return 200.
func (h *Handler) Selftest(c *gin.Context) {
	ocsc, ok := preset.Lookup(preset.OCSC)
	if !ok {
		// Should be unreachable: ocsc is a compiled-in built-in preset.
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
