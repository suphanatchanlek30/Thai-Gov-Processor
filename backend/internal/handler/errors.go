package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/apierr"
	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/service"
)

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
