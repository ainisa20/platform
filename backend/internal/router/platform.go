package router

import (
	"time"

	"platform/internal/controller/platform"
	"platform/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterPlatformRoutes(
	r *gin.Engine,
	db *gorm.DB,
	rdb *redis.Client,
	jwtSecret string,
	authLogin, authLogout, authUserInfo, authPermissions, authRefresh gin.HandlerFunc,
	userCtrl *platform.SysUserCtrl,
	roleCtrl *platform.SysRoleCtrl,
	deptCtrl *platform.SysDeptCtrl,
	permCtrl *platform.SysPermissionCtrl,
	shopCtrl *platform.SysShopCtrl,
	categoryCtrl *platform.ProductCategoryCtrl,
	productCtrl *platform.ProductCtrl,
	financeCategoryCtrl *platform.FinanceCategoryCtrl,
	reportCtrl *platform.FinanceReportCtrl,
) {
	v1 := r.Group("/api/v1/platform")
	v1.Use(
		middleware.CORSMiddleware(),
		middleware.RateLimitMiddleware(rdb, 100, time.Minute),
		middleware.AuditMiddleware(),
	)

	auth := v1.Group("/auth")
	{
		auth.POST("/login", authLogin)
		auth.POST("/logout", authLogout)
	}

	authed := v1.Group("")
	authed.Use(middleware.JWTAuthMiddleware(jwtSecret, "platform"), middleware.DBInjectMiddleware(db))
	{
		authed.GET("/auth/userinfo", authUserInfo)
		authed.GET("/auth/permissions", authPermissions)
		authed.POST("/auth/refresh", authRefresh)
	}

	sys := v1.Group("/sys")
	sys.Use(
		middleware.JWTAuthMiddleware(jwtSecret, "platform"),
		middleware.DBInjectMiddleware(db),
		middleware.DataScopeMiddleware(db),
	)
	{
		users := sys.Group("/users")
		{
			users.GET("", middleware.PermissionMiddleware("platform:user:list"), userCtrl.List)
			users.POST("", middleware.PermissionMiddleware("platform:user:create"), userCtrl.Create)
			users.PUT("/:id", middleware.PermissionMiddleware("platform:user:update"), userCtrl.Update)
			users.DELETE("/:id", middleware.PermissionMiddleware("platform:user:delete"), userCtrl.Delete)
			users.PUT("/:id/roles", middleware.PermissionMiddleware("platform:user:assign"), userCtrl.AssignRoles)
			users.PUT("/:id/password", middleware.PermissionMiddleware("platform:user:reset"), userCtrl.ResetPassword)
		}

		roles := sys.Group("/roles")
		{
			roles.GET("", middleware.PermissionMiddleware("platform:role:list"), roleCtrl.List)
			roles.GET("/assignable", middleware.PermissionMiddleware("platform:role:list"), roleCtrl.AssignableRoles)
			roles.GET("/:id", middleware.PermissionMiddleware("platform:role:list"), roleCtrl.GetByID)
			roles.POST("", middleware.PermissionMiddleware("platform:role:create"), roleCtrl.Create)
			roles.PUT("/:id", middleware.PermissionMiddleware("platform:role:update"), roleCtrl.Update)
			roles.DELETE("/:id", middleware.PermissionMiddleware("platform:role:delete"), roleCtrl.Delete)
			roles.PUT("/:id/permissions", middleware.PermissionMiddleware("platform:role:assign"), roleCtrl.AssignPermissions)
		}

		depts := sys.Group("/depts")
		{
			depts.GET("", middleware.PermissionMiddleware("platform:dept:list"), deptCtrl.List)
			depts.POST("", middleware.PermissionMiddleware("platform:dept:create"), deptCtrl.Create)
			depts.PUT("/:id", middleware.PermissionMiddleware("platform:dept:update"), deptCtrl.Update)
			depts.DELETE("/:id", middleware.PermissionMiddleware("platform:dept:delete"), deptCtrl.Delete)
		}

		sys.GET("/permissions", middleware.PermissionMiddleware("platform:role:assign"), permCtrl.GetTree)

		shops := v1.Group("/shops")
		shops.Use(
			middleware.JWTAuthMiddleware(jwtSecret, "platform"),
			middleware.DBInjectMiddleware(db),
		)
		{
			shops.GET("", middleware.PermissionMiddleware("platform:shop:list"), shopCtrl.List)
			shops.GET("/:id", middleware.PermissionMiddleware("platform:shop:list"), shopCtrl.Get)
			shops.POST("", middleware.PermissionMiddleware("platform:shop:create"), shopCtrl.Create)
			shops.PUT("/:id", middleware.PermissionMiddleware("platform:shop:update"), shopCtrl.Update)
			shops.DELETE("/:id", middleware.PermissionMiddleware("platform:shop:delete"), shopCtrl.Delete)
			shops.PUT("/:id/status", middleware.PermissionMiddleware("platform:shop:status"), shopCtrl.UpdateStatus)
			shops.PUT("/:id/admin-password", middleware.PermissionMiddleware("platform:shop:reset"), shopCtrl.ResetAdminPassword)
		}
	}

	productCategories := v1.Group("/product-categories")
	productCategories.Use(
		middleware.JWTAuthMiddleware(jwtSecret, "platform"),
		middleware.DBInjectMiddleware(db),
	)
	{
		productCategories.GET("", middleware.PermissionMiddleware("platform:product:category:list"), categoryCtrl.List)
		productCategories.POST("", middleware.PermissionMiddleware("platform:product:category:create"), categoryCtrl.Create)
		productCategories.PUT("/:id", middleware.PermissionMiddleware("platform:product:category:update"), categoryCtrl.Update)
		productCategories.DELETE("/:id", middleware.PermissionMiddleware("platform:product:category:delete"), categoryCtrl.Delete)
	}

	products := v1.Group("/products")
	products.Use(
		middleware.JWTAuthMiddleware(jwtSecret, "platform"),
		middleware.DBInjectMiddleware(db),
	)
	{
		products.GET("", middleware.PermissionMiddleware("platform:product:list"), productCtrl.List)
		products.GET("/:id", middleware.PermissionMiddleware("platform:product:list"), productCtrl.Get)
		products.POST("", middleware.PermissionMiddleware("platform:product:create"), productCtrl.Create)
		products.PUT("/:id", middleware.PermissionMiddleware("platform:product:update"), productCtrl.Update)
		products.DELETE("/:id", middleware.PermissionMiddleware("platform:product:delete"), productCtrl.Delete)
		products.PUT("/:id/status", middleware.PermissionMiddleware("platform:product:status"), productCtrl.UpdateStatus)
		products.GET("/:id/workflow", middleware.PermissionMiddleware("platform:product:list"), productCtrl.GetWorkflow)
		products.PUT("/:id/workflow", middleware.PermissionMiddleware("platform:product:update"), productCtrl.SaveWorkflow)
	}

	financeCategories := v1.Group("/finance-categories")
	financeCategories.Use(
		middleware.JWTAuthMiddleware(jwtSecret, "platform"),
		middleware.DBInjectMiddleware(db),
	)
	{
		financeCategories.GET("", middleware.PermissionMiddleware("platform:finance:category:list"), financeCategoryCtrl.List)
		financeCategories.POST("", middleware.PermissionMiddleware("platform:finance:category:create"), financeCategoryCtrl.Create)
		financeCategories.PUT("/:id", middleware.PermissionMiddleware("platform:finance:category:update"), financeCategoryCtrl.Update)
		financeCategories.DELETE("/:id", middleware.PermissionMiddleware("platform:finance:category:delete"), financeCategoryCtrl.Delete)
	}

	financeReports := v1.Group("/finance/reports")
	financeReports.Use(
		middleware.JWTAuthMiddleware(jwtSecret, "platform"),
		middleware.DBInjectMiddleware(db),
	)
	{
		financeReports.GET("/summary", middleware.PermissionMiddleware("platform:finance:report:list"), reportCtrl.Summary)
		financeReports.GET("/trend", middleware.PermissionMiddleware("platform:finance:report:list"), reportCtrl.Trend)
		financeReports.GET("/profit-loss", middleware.PermissionMiddleware("platform:finance:report:list"), reportCtrl.ProfitLoss)
		financeReports.GET("/shops", middleware.PermissionMiddleware("platform:finance:report:list"), reportCtrl.PerShop)
	}
}
