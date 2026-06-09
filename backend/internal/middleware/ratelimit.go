package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"platform/internal/pkg/response"
)

// RateLimitMiddleware limits the number of requests a user can make within
// a sliding time window. It uses Redis INCR + EXPIRE as the backing store.
//
//   - limit:  max requests allowed within the window (e.g. 100)
//   - window: time duration for the window (e.g. time.Minute)
//
// Key pattern: ratelimit:{user_id}:{path}
// If the user is not authenticated (user_id == 0), the client IP is used instead.
func RateLimitMiddleware(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getContextUint64(c, "user_id")
		var key string
		if userID > 0 {
			key = fmt.Sprintf("ratelimit:%d:%s", userID, c.Request.URL.Path)
		} else {
			key = fmt.Sprintf("ratelimit:%s:%s", c.ClientIP(), c.Request.URL.Path)
		}

		ctx := c.Request.Context()
		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			rdb.Expire(ctx, key, window)
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, int64(limit)-count)))

		if count > int64(limit) {
			response.Fail(c, http.StatusTooManyRequests, "rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}
