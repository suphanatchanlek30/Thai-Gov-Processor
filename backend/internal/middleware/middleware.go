// Package middleware provides minimal request logging and panic recovery
// for gin.New() (as opposed to gin.Default(), which pulls in gin's own
// logger that writes to os.Stdout with ANSI color codes and assumptions
// that don't fit a read-only, structured-logging production container).
package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger writes one line per request to the standard library logger:
// method, path, status, and latency. Kept intentionally minimal — no file
// writes, no external deps.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		log.Printf("%s %s %d %s", c.Request.Method, path, c.Writer.Status(), time.Since(start))
	}
}

// Recovery recovers from any panic in a handler, logs it, and responds
// with a generic 500 JSON body instead of letting the connection die.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic recovered: %v", r)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
		}()
		c.Next()
	}
}
