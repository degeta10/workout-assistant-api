package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	HeaderRequestID  = "X-Request-ID"
	ContextRequestID = "request_id"
)

// RequestID ensures every request has a request ID and exposes it via
// response header + gin context for logging and debugging.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader(HeaderRequestID))
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(ContextRequestID, requestID)
		c.Writer.Header().Set(HeaderRequestID, requestID)
		c.Next()
	}
}

func GetRequestID(c *gin.Context) string {
	if value, ok := c.Get(ContextRequestID); ok {
		if requestID, ok := value.(string); ok {
			return requestID
		}
	}
	return ""
}
