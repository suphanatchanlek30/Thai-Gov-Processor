// Package apierr provides consistent JSON error responses for the API.
// Every error body has the shape {"error": "..."}.
package apierr

import "github.com/gin-gonic/gin"

// Body is the JSON shape returned on any API error.
type Body struct {
	Error string `json:"error"`
}

// Send writes a JSON error body with the given HTTP status and aborts the
// gin context so later handlers/middleware don't also write a response.
func Send(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, Body{Error: message})
}

// BadRequest sends a 400 with the given message (bad file type, bad preset,
// missing/invalid field).
func BadRequest(c *gin.Context, message string) {
	Send(c, 400, message)
}

// PayloadTooLarge sends a 413 (upload exceeds the configured max size).
func PayloadTooLarge(c *gin.Context, message string) {
	Send(c, 413, message)
}

// Unprocessable sends a 422 (the processor could not hit the target size
// without an unacceptable quality loss).
func Unprocessable(c *gin.Context, message string) {
	Send(c, 422, message)
}

// NotImplemented sends a 501. Used while PhotoProcessor/PDFMerger are still
// stub implementations.
func NotImplemented(c *gin.Context, message string) {
	Send(c, 501, message)
}

// Internal sends a 500 for unexpected server-side failures.
func Internal(c *gin.Context, message string) {
	Send(c, 500, message)
}
