package middleware

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TenantScope returns a GORM scope function that appends
// WHERE tenant_id = ? for shop queries, or WHERE tenant_id = 0 for platform.
// Use this in the repository layer to guarantee tenant filtering at the ORM level.
func TenantScope(c *gin.Context) func(*gorm.DB) *gorm.DB {
	tenantID := getContextUint64(c, "tenant_id")
	return func(db *gorm.DB) *gorm.DB {
		if tenantID > 0 {
			return db.Where("tenant_id = ?", tenantID)
		}
		return db.Where("tenant_id = 0")
	}
}
