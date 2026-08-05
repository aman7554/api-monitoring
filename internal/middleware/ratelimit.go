package middleware

import (
	"net/http"
	"time"

	"pulsewatch/internal/queue"

	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware(redisQ *queue.RedisQueue, maxReq int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		allowed, err := redisQ.AllowRateLimit(c.Request.Context(), ip, maxReq, window)
		if err != nil || !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": int(window.Seconds()),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
