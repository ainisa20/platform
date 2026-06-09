package middleware

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"platform/internal/model/enum"
)

// DataScopeMiddleware computes the set of users and depts visible to
// the current user based on their role's data_scope, and stores both
// slices in the gin context under "data_scope_user_ids" and
// "data_scope_dept_ids" respectively.
//
//   - DataScopeAll (1):        user_ids = nil (no filter); dept_ids = nil
//   - DataScopeDeptAndSub (2): user_ids = users in dept + descendants;
//                              dept_ids = dept + descendants
//   - DataScopeDeptOnly (3):   user_ids = users in dept only;
//                              dept_ids = dept only
//   - DataScopeSelfOnly (4):   user_ids = [self]; dept_ids = [self dept]
//
// Downstream controllers should call ApplyUserScope(db), ApplyDeptScope(db),
// or ApplyUserIDScope(db) from this package to filter their queries.
func DataScopeMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		dataScope := getContextInt16(c, "data_scope")
		userID := getContextUint64(c, "user_id")
		deptID := getContextUint64(c, "dept_id")
		tenantID := getContextUint64(c, "tenant_id")

		if dataScope == enum.DataScopeAll {
			c.Next()
			return
		}

		var deptIDs []uint64
		switch dataScope {
		case enum.DataScopeDeptAndSub:
			err := db.Table("sys_dept_closure").
				Where("ancestor_id = ? AND tenant_id = ?", deptID, tenantID).
				Pluck("descendant_id", &deptIDs).Error
			if err != nil || len(deptIDs) == 0 {
				deptIDs = []uint64{deptID}
			}
		case enum.DataScopeDeptOnly:
			deptIDs = []uint64{deptID}
		case enum.DataScopeSelfOnly:
			deptIDs = []uint64{deptID}
		default:
			deptIDs = []uint64{deptID}
		}

		var userIDs []uint64
		if len(deptIDs) > 0 {
			_ = db.Table("sys_user").
				Where("dept_id IN ? AND tenant_id = ?", deptIDs, tenantID).
				Pluck("id", &userIDs).Error
		}
		if len(userIDs) == 0 {
			userIDs = []uint64{userID}
		}

		c.Set("data_scope_user_ids", userIDs)
		c.Set("data_scope_dept_ids", deptIDs)
		c.Next()
	}
}

func ApplyUserScope(c *gin.Context, db *gorm.DB) *gorm.DB {
	v, ok := c.Get("data_scope_user_ids")
	if !ok || v == nil {
		return db
	}
	ids, ok := v.([]uint64)
	if !ok || len(ids) == 0 {
		return db
	}
	return db.Where("created_by IN ?", ids)
}

func ApplyDeptScope(c *gin.Context, db *gorm.DB) *gorm.DB {
	v, ok := c.Get("data_scope_dept_ids")
	if !ok || v == nil {
		return db
	}
	ids, ok := v.([]uint64)
	if !ok || len(ids) == 0 {
		return db
	}
	return db.Where("id IN ?", ids)
}

func ApplyUserIDScope(c *gin.Context, db *gorm.DB) *gorm.DB {
	v, ok := c.Get("data_scope_user_ids")
	if !ok || v == nil {
		return db
	}
	ids, ok := v.([]uint64)
	if !ok || len(ids) == 0 {
		return db
	}
	return db.Where("id IN ?", ids)
}
