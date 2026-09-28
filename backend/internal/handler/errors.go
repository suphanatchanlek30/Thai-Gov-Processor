package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/apierr"
	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/service"
)

// writeServiceError maps an error returned by a PhotoService/DocumentService
// call to the right HTTP status, per docs/api-reference.md:
//   - service.ErrNotImplemented -> 501 (stub, temporary)
//   - service.ErrCannotMeetTarget -> 422 (can't compress under target)
//   - anything else -> 500
//
// This is shared by /photos/preset, /documents/merge-pdf and /selftest so
// all three behave identically once the real service replaces the stub.
func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotImplemented):
		apierr.NotImplemented(c, err.Error())
	case errors.Is(err, service.ErrCannotMeetTarget):
		apierr.Unprocessable(c, err.Error())
	default:
		apierr.Internal(c, err.Error())
	}
}
