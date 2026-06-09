package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"platform/internal/model/entity"
)

var auditDB *gorm.DB

// InitAudit sets the *gorm.DB used by AuditMiddleware for writing
// operation logs. Call this once during application startup.
func InitAudit(db *gorm.DB) {
	auditDB = db
}

// AuditMiddleware records an entry in sys_operation_log for every request.
// The log is written asynchronously so it never blocks the response.
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		if auditDB == nil {
			return
		}

		params := c.Request.URL.RawQuery
		if len(params) > 2000 {
			params = params[:2000]
		}

		log := &entity.SysOperationLog{
			TenantID:   getContextUint64(c, "tenant_id"),
			UserID:     getContextUint64(c, "user_id"),
			Username:   getContextString(c, "username"),
			Method:     c.Request.Method,
			URL:        c.Request.URL.Path,
			Params:     params,
			IP:         c.ClientIP(),
			DurationMs: int(time.Since(start).Milliseconds()),
		}

		go func(entry *entity.SysOperationLog) {
			_ = auditDB.Create(entry).Error
		}(log)
	}
}

// extractModule derives a short module name from the URL path for logging.
func extractModule(urlPath string) string {
	parts := strings.Split(strings.Trim(urlPath, "/"), "/")
	if len(parts) >= 4 {
		return parts[3]
	}
	return ""
}
