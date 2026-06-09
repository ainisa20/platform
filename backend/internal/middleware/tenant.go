package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"platform/internal/pkg/response"
)

// TenantRLSMiddleware sets PostgreSQL RLS context for tenant isolation.
// For tenant_id > 0 (shop requests): begins a transaction and executes
// SELECT set_config('app.tenant_id', $1, true) to activate RLS policies.
// For tenant_id == 0 (platform requests): passes the raw DB through.
func TenantRLSMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsVal, exists := c.Get("claims")
		if !exists {
			response.Unauthorized(c, "missing auth claims")
			c.Abort()
			return
		}
		claims, ok := claimsVal.(*JWTClaims)
		if !ok {
			response.Unauthorized(c, "invalid auth claims")
			c.Abort()
			return
		}

		tenantID := claims.TenantID
		c.Set("tenant_id", tenantID)
		c.Set("user_id", claims.UserID)
		c.Set("dept_id", claims.DeptID)
		c.Set("data_scope", claims.DataScope)

		if tenantID > 0 {
			tx := db.Begin()
			if tx.Error != nil {
				response.InternalError(c, "failed to begin transaction")
				c.Abort()
				return
			}

			if err := tx.Exec("SELECT set_config('app.tenant_id', $1, true)",
				strconv.FormatUint(tenantID, 10)).Error; err != nil {
				tx.Rollback()
				response.InternalError(c, "failed to set tenant context")
				c.Abort()
				return
			}

			c.Set("db", tx)

			defer func() {
				if c.IsAborted() {
					tx.Rollback()
				} else {
					tx.Commit()
				}
			}()
		} else {
			c.Set("db", db)
		}

		c.Next()
	}
}
