package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/apierr"
	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/processor"
)

// writeProcessorError maps an error returned by a PhotoProcessor/PDFMerger
// call to the right HTTP status, per docs/api-reference.md:
//   - processor.ErrNotImplemented -> 501 (stub, temporary)
//   - processor.ErrCannotMeetTarget -> 422 (can't compress under target)
//   - anything else -> 500
//
// This is shared by /photos/preset, /documents/merge-pdf and /selftest so
// all three behave identically once the real processor replaces the stub.
func writeProcessorError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, processor.ErrNotImplemented):
		apierr.NotImplemented(c, err.Error())
	case errors.Is(err, processor.ErrCannotMeetTarget):
		apierr.Unprocessable(c, err.Error())
	default:
		apierr.Internal(c, err.Error())
	}
}
