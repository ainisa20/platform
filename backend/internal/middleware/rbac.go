package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"platform/internal/pkg/response"
)

var (
	rbacDB  *gorm.DB
	rbacRDB *redis.Client
)

// InitRBAC initialises the DB and Redis clients used by PermissionMiddleware.
// Call this once during application startup.
func InitRBAC(db *gorm.DB, rdb *redis.Client) {
	rbacDB = db
	rbacRDB = rdb
}

// PermissionMiddleware checks that the authenticated user possesses the
// required permission code. Permission codes are cached in Redis with a
// 30-minute TTL; on cache miss they are loaded from the database.
func PermissionMiddleware(requiredPerm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getContextUint64(c, "user_id")
		tenantID := getContextUint64(c, "tenant_id")

		perms, err := getUserPermissionCodes(c.Request.Context(), userID, tenantID)
		if err != nil {
			response.InternalError(c, "failed to load permissions")
			c.Abort()
			return
		}

		for _, p := range perms {
			if p == requiredPerm {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "insufficient permissions")
		c.Abort()
	}
}

func getUserPermissionCodes(ctx context.Context, userID, tenantID uint64) ([]string, error) {
	cacheKey := fmt.Sprintf("perms:%d:%d", tenantID, userID)

	if rbacRDB != nil {
		val, err := rbacRDB.Get(ctx, cacheKey).Bytes()
		if err == nil {
			var cached []string
			if jsonErr := json.Unmarshal(val, &cached); jsonErr == nil {
				return cached, nil
			}
		}
	}

	var results []struct {
		PermsCode string
	}
	err := rbacDB.Table("sys_role_permission rp").
		Select("p.perms_code").
		Joins("JOIN sys_permission p ON rp.permission_id = p.id").
		Joins("JOIN sys_user_role ur ON ur.role_id = rp.role_id").
		Where("ur.user_id = ? AND rp.tenant_id = ?", userID, tenantID).
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	perms := make([]string, 0, len(results))
	for _, r := range results {
		if r.PermsCode != "" {
			perms = append(perms, r.PermsCode)
		}
	}

	if rbacRDB != nil && len(perms) > 0 {
		if data, jsonErr := json.Marshal(perms); jsonErr == nil {
			_ = rbacRDB.Set(ctx, cacheKey, data, 30*time.Minute).Err()
		}
	}

	return perms, nil
}
