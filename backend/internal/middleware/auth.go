package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"platform/internal/pkg/auth"
	"platform/internal/pkg/response"
)

// JWTClaims is an alias for the auth package's claim type so existing
// context lookups (user_id, tenant_id, etc.) keep working.
type JWTClaims = auth.JWTClaims

// JWTAuthMiddleware validates JWT tokens from the Authorization header
// and rejects tokens whose audience does not match one of allowedAudiences.
// This prevents cross-system token reuse: a shop user token cannot
// access platform endpoints and vice versa. On success it sets context
// keys: claims, claims_audience, user_id, tenant_id, dept_id,
// username, data_scope.
func JWTAuthMiddleware(jwtSecret string, allowedAudiences ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing Authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "invalid Authorization header format")
			c.Abort()
			return
		}

		claims, err := auth.ParseToken(parts[1], jwtSecret)
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		audience := auth.ExtractAudience(claims, allowedAudiences...)
		if audience == "" {
			response.Unauthorized(c, "token audience not allowed for this system")
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Set("claims_audience", audience)
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("dept_id", claims.DeptID)
		c.Set("username", claims.Username)
		c.Set("data_scope", claims.DataScope)

		c.Next()
	}
}

func getContextUint64(c *gin.Context, key string) uint64 {
	if v, ok := c.Get(key); ok {
		if val, ok := v.(uint64); ok {
			return val
		}
	}
	return 0
}

func getContextInt16(c *gin.Context, key string) int16 {
	if v, ok := c.Get(key); ok {
		if val, ok := v.(int16); ok {
			return val
		}
	}
	return 0
}

func getContextString(c *gin.Context, key string) string {
	if v, ok := c.Get(key); ok {
		if val, ok := v.(string); ok {
			return val
		}
	}
	return ""
}

// GenerateToken is kept as a thin wrapper for backward compatibility
// with code that doesn't yet go through the auth package. New code
// should call auth.GenerateAccessToken directly.
func GenerateToken(secret string, userID, tenantID, deptID uint64, username string, dataScope int16, expiresAt time.Time) (string, error) {
	claims := auth.JWTClaims{
		UserID:    userID,
		TenantID:  tenantID,
		DeptID:    deptID,
		Username:  username,
		DataScope: dataScope,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
