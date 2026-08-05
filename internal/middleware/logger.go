package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqID := uuid.New().String()
		c.Header("X-Request-ID", reqID)

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path

		fmt.Printf(`{"time":"%s", "request_id":"%s", "ip":"%s", "method":"%s", "path":"%s", "status":%d, "latency_ms":%d}`+"\n",
			start.Format(time.RFC3339), reqID, clientIP, method, path, status, latency.Milliseconds())
	}
}
