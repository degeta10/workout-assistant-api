package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger writes structured JSON logs for every request.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		// Process request
		c.Next()

		if rawQuery != "" {
			path = path + "?" + rawQuery
		}

		status := c.Writer.Status()
		latency := time.Since(start)

		// Create a slice of structured attributes
		attrs := []slog.Attr{
			slog.String("request_id", GetRequestID(c)),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Int64("latency_ms", latency.Milliseconds()),
			slog.String("client_ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
		}

		// Extract any errors thrown during the request
		var errorMsg string
		if len(c.Errors) > 0 {
			errorMsg = c.Errors.String()
			attrs = append(attrs, slog.String("error", errorMsg))
		}

		// Log as ERROR if status is 5xx, WARN for 4xx, INFO otherwise
		ctx := c.Request.Context()
		if status >= 500 {
			slog.LogAttrs(ctx, slog.LevelError, "Server Error", attrs...)
		} else if status >= 400 {
			slog.LogAttrs(ctx, slog.LevelWarn, "Client Error", attrs...)
		} else {
			slog.LogAttrs(ctx, slog.LevelInfo, "Request Completed", attrs...)
		}
	}
}
