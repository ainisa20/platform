package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AllowedOrigins holds the set of origins permitted by CORS.
// Modify this slice before calling CORSMiddleware to customise allowed origins.
// A value of ["*"] permits every origin.
var AllowedOrigins = []string{"*"}

// CORSMiddleware adds standard CORS headers to every response and handles
// OPTIONS preflight requests.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		allowOrigin := ""
		for _, o := range AllowedOrigins {
			if o == "*" || strings.EqualFold(o, origin) {
				allowOrigin = origin
				if o == "*" {
					allowOrigin = "*"
				}
				break
			}
		}

		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
