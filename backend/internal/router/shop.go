package router

import (
	"time"

	"platform/internal/controller/shop"
	"platform/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterShopRoutes(
	r *gin.Engine,
	db *gorm.DB,
	rdb *redis.Client,
	jwtSecret string,
	authLogin, authLogout, authUserInfo, authPermissions, authRefresh gin.HandlerFunc,
	userCtrl *shop.SysUserCtrl,
	roleCtrl *shop.SysRoleCtrl,
	deptCtrl *shop.SysDeptCtrl,
	permCtrl *shop.SysPermissionCtrl,
	finCatCtrl *shop.ShopFinCategoryCtrl,
	shopProductCtrl *shop.ShopProductCtrl,
	customerCtrl *shop.ShopCustomerCtrl,
	orderCtrl *shop.OrderCtrl,
	finAccountCtrl *shop.ShopFinAccountCtrl,
	recordCtrl *shop.RecordCtrl,
	reportCtrl *shop.FinanceReportCtrl,
) {
	v1 := r.Group("/api/v1/shop")
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
	authed.Use(middleware.JWTAuthMiddleware(jwtSecret, "shop"))
	{
		authed.GET("/auth/userinfo", authUserInfo)
		authed.GET("/auth/permissions", authPermissions)
		authed.POST("/auth/refresh", authRefresh)
	}

	sys := v1.Group("/sys")
	sys.Use(
		middleware.JWTAuthMiddleware(jwtSecret, "shop"),
		middleware.DBInjectMiddleware(db),
		middleware.TenantRLSMiddleware(db),
		middleware.DataScopeMiddleware(db),
	)
	{
		users := sys.Group("/users")
		{
			users.GET("", middleware.PermissionMiddleware("shop:user:list"), userCtrl.List)
			users.POST("", middleware.PermissionMiddleware("shop:user:create"), userCtrl.Create)
			users.PUT("/:id", middleware.PermissionMiddleware("shop:user:update"), userCtrl.Update)
			users.DELETE("/:id", middleware.PermissionMiddleware("shop:user:delete"), userCtrl.Delete)
			users.PUT("/:id/roles", middleware.PermissionMiddleware("shop:user:assign"), userCtrl.AssignRoles)
			users.PUT("/:id/password", middleware.PermissionMiddleware("shop:user:reset"), userCtrl.ResetPassword)
		}

		roles := sys.Group("/roles")
		{
			roles.GET("", middleware.PermissionMiddleware("shop:role:list"), roleCtrl.List)
			roles.GET("/assignable", middleware.PermissionMiddleware("shop:role:list"), roleCtrl.AssignableRoles)
			roles.GET("/:id", middleware.PermissionMiddleware("shop:role:list"), roleCtrl.GetByID)
			roles.POST("", middleware.PermissionMiddleware("shop:role:create"), roleCtrl.Create)
			roles.PUT("/:id", middleware.PermissionMiddleware("shop:role:update"), roleCtrl.Update)
			roles.DELETE("/:id", middleware.PermissionMiddleware("shop:role:delete"), roleCtrl.Delete)
			roles.PUT("/:id/permissions", middleware.PermissionMiddleware("shop:role:assign"), roleCtrl.AssignPermissions)
		}

		depts := sys.Group("/depts")
		{
			depts.GET("", middleware.PermissionMiddleware("shop:dept:list"), deptCtrl.List)
			depts.POST("", middleware.PermissionMiddleware("shop:dept:create"), deptCtrl.Create)
			depts.PUT("/:id", middleware.PermissionMiddleware("shop:dept:update"), deptCtrl.Update)
			depts.DELETE("/:id", middleware.PermissionMiddleware("shop:dept:delete"), deptCtrl.Delete)
		}

		sys.GET("/permissions", middleware.PermissionMiddleware("shop:role:assign"), permCtrl.GetTree)
	}

	finance := v1.Group("/finance")
	finance.Use(
		middleware.JWTAuthMiddleware(jwtSecret, "shop"),
		middleware.DBInjectMiddleware(db),
		middleware.TenantRLSMiddleware(db),
	)
	{
		finance.GET("/categories/available", middleware.PermissionMiddleware("shop:finance:category:list"), finCatCtrl.Available)
		finance.GET("/categories", middleware.PermissionMiddleware("shop:finance:category:list"), finCatCtrl.List)
		finance.POST("/categories/sync", middleware.PermissionMiddleware("shop:finance:category:sync"), finCatCtrl.Sync)
		finance.DELETE("/categories/:id", middleware.PermissionMiddleware("shop:finance:category:delete"), finCatCtrl.CancelSync)

		accounts := finance.Group("/accounts")
		{
			accounts.GET("", middleware.PermissionMiddleware("shop:finance:account:list"), finAccountCtrl.List)
			accounts.POST("", middleware.PermissionMiddleware("shop:finance:account:create"), finAccountCtrl.Create)
			accounts.PUT("/:id", middleware.PermissionMiddleware("shop:finance:account:update"), finAccountCtrl.Update)
			accounts.DELETE("/:id", middleware.PermissionMiddleware("shop:finance:account:delete"), finAccountCtrl.Delete)
		}

		reports := finance.Group("/reports")
		{
			reports.GET("/summary", middleware.PermissionMiddleware("shop:finance:report:list"), reportCtrl.Summary)
			reports.GET("/trend", middleware.PermissionMiddleware("shop:finance:report:list"), reportCtrl.Trend)
			reports.GET("/profit-loss", middleware.PermissionMiddleware("shop:finance:report:list"), reportCtrl.ProfitLoss)
		}
	}

	records := v1.Group("/finance/records")
	records.Use(
		middleware.JWTAuthMiddleware(jwtSecret, "shop"),
		middleware.DBInjectMiddleware(db),
		middleware.TenantRLSMiddleware(db),
		middleware.DataScopeMiddleware(db),
	)
	{
	records.GET("/export", middleware.PermissionMiddleware("shop:finance:record:export"), recordCtrl.Export)
	records.GET("/export-zip", middleware.PermissionMiddleware("shop:finance:record:export"), recordCtrl.ExportZip)
	records.GET("/export-zip/:task_id", middleware.PermissionMiddleware("shop:finance:record:export"), recordCtrl.GetExportTask)
	records.GET("/export-zip/:task_id/download", middleware.PermissionMiddleware("shop:finance:record:export"), recordCtrl.DownloadExport)
	records.GET("", middleware.PermissionMiddleware("shop:finance:record:list"), recordCtrl.List)
		records.GET("/:id", middleware.PermissionMiddleware("shop:finance:record:list"), recordCtrl.Get)
		records.GET("/:id/attachments", middleware.PermissionMiddleware("shop:finance:record:list"), recordCtrl.ListAttachments)
		records.POST("", middleware.PermissionMiddleware("shop:finance:record:create"), recordCtrl.Create)
		records.POST("/:id/review", middleware.PermissionMiddleware("shop:finance:record:audit"), recordCtrl.Review)
		records.POST("/:id/attachments", middleware.PermissionMiddleware("shop:finance:record:upload"), recordCtrl.CreateAttachment)
		records.GET("/:id/attachments/:attId/download", middleware.PermissionMiddleware("shop:finance:record:list"), recordCtrl.DownloadAttachment)
		records.PUT("/:id", middleware.PermissionMiddleware("shop:finance:record:update"), recordCtrl.Update)
		records.DELETE("/:id", middleware.PermissionMiddleware("shop:finance:record:delete"), recordCtrl.Delete)
	}

	products := v1.Group("/products")
	products.Use(
		middleware.JWTAuthMiddleware(jwtSecret, "shop"),
		middleware.DBInjectMiddleware(db),
		middleware.TenantRLSMiddleware(db),
	)
	{
		products.GET("/platform", middleware.PermissionMiddleware("shop:product:list"), shopProductCtrl.ListPlatform)
		products.GET("", middleware.PermissionMiddleware("shop:product:list"), shopProductCtrl.List)
		products.POST("", middleware.PermissionMiddleware("shop:product:select"), shopProductCtrl.Select)
		products.PUT("/:id/price", middleware.PermissionMiddleware("shop:product:price"), shopProductCtrl.UpdatePrice)
		products.PUT("/:id/status", middleware.PermissionMiddleware("shop:product:status"), shopProductCtrl.UpdateStatus)
		products.DELETE("/:id", middleware.PermissionMiddleware("shop:product:delete"), shopProductCtrl.DeleteSelection)
	}

	customers := v1.Group("/customers")
	customers.Use(
		middleware.JWTAuthMiddleware(jwtSecret, "shop"),
		middleware.DBInjectMiddleware(db),
		middleware.TenantRLSMiddleware(db),
		middleware.DataScopeMiddleware(db),
	)
	{
		customers.GET("", middleware.PermissionMiddleware("shop:customer:list"), customerCtrl.List)
		customers.GET("/export", middleware.PermissionMiddleware("shop:customer:export"), customerCtrl.Export)
		customers.GET("/:id", middleware.PermissionMiddleware("shop:customer:list"), customerCtrl.Get)
		customers.GET("/:id/orders", middleware.PermissionMiddleware("shop:customer:list"), customerCtrl.ListOrders)
		customers.POST("", middleware.PermissionMiddleware("shop:customer:create"), customerCtrl.Create)
		customers.PUT("/:id", middleware.PermissionMiddleware("shop:customer:update"), customerCtrl.Update)
		customers.DELETE("/:id", middleware.PermissionMiddleware("shop:customer:delete"), customerCtrl.Delete)
	}

	orders := v1.Group("/orders")
	orders.Use(
		middleware.JWTAuthMiddleware(jwtSecret, "shop"),
		middleware.DBInjectMiddleware(db),
		middleware.TenantRLSMiddleware(db),
		middleware.DataScopeMiddleware(db),
	)
	{
		orders.GET("", middleware.PermissionMiddleware("shop:order:list"), orderCtrl.List)
		orders.GET("/export", middleware.PermissionMiddleware("shop:order:export"), orderCtrl.Export)
		orders.GET("/:id", middleware.PermissionMiddleware("shop:order:list"), orderCtrl.Get)
		orders.POST("", middleware.PermissionMiddleware("shop:order:create"), orderCtrl.Create)
		orders.PUT("/:id/cancel", middleware.PermissionMiddleware("shop:order:cancel"), orderCtrl.CancelGroup)
		orders.PUT("/:id/items/:itemId/cancel", middleware.PermissionMiddleware("shop:order:cancel"), orderCtrl.CancelItem)
		orders.GET("/:id/items/:itemId/workflow", middleware.PermissionMiddleware("shop:order:list"), orderCtrl.GetItemWorkflow)
		orders.GET("/:id/items/:itemId/workflow/logs", middleware.PermissionMiddleware("shop:order:list"), orderCtrl.GetItemWorkflowLogs)
		orders.POST("/:id/items/:itemId/workflow/advance", middleware.PermissionMiddleware("shop:order:advance"), orderCtrl.AdvanceItemWorkflow)
		orders.GET("/:id/items/:itemId/attachments", middleware.PermissionMiddleware("shop:order:list"), orderCtrl.ListItemAttachments)
		orders.POST("/:id/items/:itemId/attachments", middleware.PermissionMiddleware("shop:order:upload"), orderCtrl.CreateItemAttachment)
		orders.GET("/:id/items/:itemId/attachments/:attId", middleware.PermissionMiddleware("shop:order:list"), orderCtrl.GetItemAttachment)
	}
}
