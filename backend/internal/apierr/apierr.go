// Package apierr provides consistent JSON error responses for the API.
// Every error body has the shape {"error": "..."}.
package apierr

import "github.com/gin-gonic/gin"

type Body struct {
	Error string `json:"error"`
}

func Send(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, Body{Error: message})
}

func BadRequest(c *gin.Context, message string) {
	Send(c, 400, message)
}

func PayloadTooLarge(c *gin.Context, message string) {
	Send(c, 413, message)
}

func Unprocessable(c *gin.Context, message string) {
	Send(c, 422, message)
}

func NotImplemented(c *gin.Context, message string) {
	Send(c, 501, message)
}

func Internal(c *gin.Context, message string) {
	Send(c, 500, message)
}
