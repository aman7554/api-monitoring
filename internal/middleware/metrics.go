package middleware

import (
	"strconv"
	"time"

	"pulsewatch/internal/telemetry"

	"github.com/gin-gonic/gin"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		telemetry.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		telemetry.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}
