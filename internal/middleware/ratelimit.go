package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const rateWindow = time.Minute

func RateLimitMiddleware(rdb *redis.Client, rateLimit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdb == nil {
			c.Next()
			return
		}

		ip := c.ClientIP()
		key := "ratelimit:" + ip
		ctx := c.Request.Context()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			rdb.Expire(ctx, key, rateWindow)
		}

		if count > int64(rateLimit) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded, try again in a minute",
			})
			return
		}

		c.Next()
	}
}
