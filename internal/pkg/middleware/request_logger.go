package middleware

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger writes structured request logs for every request.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		if rawQuery != "" {
			path = path + "?" + rawQuery
		}

		requestID := GetRequestID(c)
		status := c.Writer.Status()
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		ua := c.Request.UserAgent()
		errorMsg := strings.TrimSpace(c.Errors.ByType(gin.ErrorTypePrivate).String())
		if errorMsg == "" {
			errorMsg = strings.TrimSpace(c.Errors.String())
		}

		fields := []string{
			"event=request",
			"request_id=" + requestID,
			"method=" + method,
			"path=" + strconv.Quote(path),
			"status=" + strconv.Itoa(status),
			"latency_ms=" + strconv.FormatInt(latency.Milliseconds(), 10),
			"client_ip=" + strconv.Quote(clientIP),
			"user_agent=" + strconv.Quote(ua),
		}
		if errorMsg != "" {
			fields = append(fields, "error="+strconv.Quote(errorMsg))
		}

		log.Println(strings.Join(fields, " "))
	}
}
