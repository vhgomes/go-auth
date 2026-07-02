package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redis  *redis.Client
	limit  int
	window time.Duration
}

func NewRateLimiter(redis *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{redis: redis, limit: limit, window: window}
}

func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		key := fmt.Sprintf("ratelimit:%s:%s", c.FullPath(), c.ClientIP())

		count, err := rl.redis.Incr(ctx, key).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "service temporarily unavailable"})
			c.Abort()
			return
		}

		if count == 1 {
			rl.redis.Expire(ctx, key, rl.window)
		}

		if count > int64(rl.limit) {
			ttl, _ := rl.redis.TTL(ctx, key).Result()
			c.Header("Retry-After", fmt.Sprintf("%.0f", ttl.Seconds()))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
