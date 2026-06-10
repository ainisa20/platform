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

func InitRBAC(db *gorm.DB, rdb *redis.Client) {
	rbacDB = db
	rbacRDB = rdb
}

// permImplies maps a permission to the permissions it automatically implies.
// When a user holds the key permission, they can also pass checks for the
// implied permissions without explicit assignment.
var permImplies = map[string][]string{
	"platform:user:list":         {"platform:role:list", "platform:dept:list"},
	"platform:role:list":         {"platform:dept:list"},
	"platform:product:list":      {"platform:product:category:list"},
	"platform:product:create":    {"platform:product:category:list"},
	"platform:product:update":    {"platform:product:category:list"},
	"shop:user:list":             {"shop:role:list", "shop:dept:list"},
	"shop:role:list":             {"shop:dept:list"},
	"shop:product:list":          {"shop:finance:category:list"},
	"shop:finance:record:list":   {"shop:finance:category:list", "shop:finance:account:list"},
	"shop:finance:record:create": {"shop:finance:category:list", "shop:finance:account:list"},
	"shop:order:create":          {"shop:customer:list", "shop:product:list"},
}

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

		if hasPermission(perms, requiredPerm) {
			c.Next()
			return
		}

		response.Forbidden(c, "insufficient permissions")
		c.Abort()
	}
}

func hasPermission(perms []string, required string) bool {
	permSet := make(map[string]struct{}, len(perms))
	for _, p := range perms {
		permSet[p] = struct{}{}
	}

	if _, ok := permSet[required]; ok {
		return true
	}

	for p := range permSet {
		if implied, exists := permImplies[p]; exists {
			for _, imp := range implied {
				if imp == required {
					return true
				}
			}
		}
	}

	return false
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
