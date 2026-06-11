package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"platform/internal/config"
	"platform/internal/controller/platform"
	"platform/internal/controller/shop"
	"platform/internal/middleware"
	"platform/internal/model/entity"
	"platform/internal/permission"
	platformrepo "platform/internal/repository/platform"
	shoprepo "platform/internal/repository/shop"
	"platform/internal/repository"
	"platform/internal/router"
	"platform/internal/service"
	platformsvc "platform/internal/service/platform"
	shopsvc "platform/internal/service/shop"
	"platform/internal/pkg/storage"
)

func main() {
	cfg := config.Load("config.yaml")

	db := initDB(cfg)
	rdb := initRedis(cfg)

	middleware.InitRBAC(db, rdb)

	if err := autoMigrate(db); err != nil {
		log.Fatalf("Auto-migrate failed: %v", err)
	}

	if err := platformrepo.ValidateDeptClosure(db); err != nil {
		log.Printf("[WARN] closure validation failed: %v", err)
	} else {
		log.Println("Dept closure validated")
	}

	if err := permission.SyncPermissions(db, permission.PlatformManifest, permission.ShopManifest); err != nil {
		log.Fatalf("Permission sync failed: %v", err)
	}

	authRepo := repository.NewAuthRepository(db)

	minioStorage, err := storage.NewMinIOStorage(cfg.MinIO)
	if err != nil {
		log.Fatalf("Failed to connect to MinIO: %v", err)
	}
	log.Println("MinIO connected")

	// Platform-side repositories
	platUserRepo := platformrepo.NewUserRepository()
	platRoleRepo := platformrepo.NewRoleRepository()
	platDeptRepo := platformrepo.NewDeptRepository()
	platShopRepo := platformrepo.NewShopRepository()
	platCategoryRepo := platformrepo.NewCategoryRepository()
	platProductRepo := platformrepo.NewProductRepository()
	platWorkflowRepo := platformrepo.NewWorkflowRepository()
	platFinCatRepo := platformrepo.NewFinanceCategoryRepository()

	// Shop-side repositories
	shopUserRepo := shoprepo.NewUserRepository()
	shopRoleRepo := shoprepo.NewRoleRepository()
	shopDeptRepo := shoprepo.NewDeptRepository()
	shopFinCatRepo := shoprepo.NewShopFinCategoryRepository()
	shopFinAccountRepo := shoprepo.NewShopFinAccountRepository()
	shopProductRepo := shoprepo.NewShopProductRepository()
	shopCustomerRepo := shoprepo.NewShopCustomerRepository()
	shopOrderRepo := shoprepo.NewOrderRepository()
	shopRecordRepo := shoprepo.NewRecordRepository()

	// Platform-side services
	platUserService := platformsvc.NewUserService(platUserRepo, platRoleRepo)
	platRoleService := platformsvc.NewRoleService(platRoleRepo)
	platDeptService := platformsvc.NewDeptService(platDeptRepo, platUserRepo)
	platShopService := platformsvc.NewShopService(platShopRepo, platUserRepo, platRoleRepo, platDeptRepo)
	platCategoryService := platformsvc.NewCategoryService(platCategoryRepo)
	platProductService := platformsvc.NewProductService(platProductRepo, platWorkflowRepo)
	platFinCatService := platformsvc.NewFinanceCategoryService(platFinCatRepo)

	// Shop-side services
	shopUserService := shopsvc.NewUserService(shopUserRepo, shopRoleRepo)
	shopRoleService := shopsvc.NewRoleService(shopRoleRepo)
	shopDeptService := shopsvc.NewDeptService(shopDeptRepo, shopUserRepo)
	shopFinCatService := shopsvc.NewShopFinCategoryService(shopFinCatRepo, platFinCatRepo)
	shopFinAccountService := shopsvc.NewShopFinAccountService(shopFinAccountRepo, shopUserRepo)
	shopProductService := shopsvc.NewShopProductService(shopProductRepo, platProductRepo, platCategoryRepo)
	shopCustomerService := shopsvc.NewShopCustomerService(shopCustomerRepo)
	shopOrderService := shopsvc.NewOrderService(shopOrderRepo, shopCustomerRepo, shopProductRepo, platProductRepo, platWorkflowRepo, shopUserRepo, minioStorage)
	shopRecordService := shopsvc.NewRecordService(shopRecordRepo, shopFinAccountRepo, shopFinCatRepo, shopOrderRepo, shopUserRepo, minioStorage, cfg.Database.DSN())

	// Auth is shared (both endpoints use the same login service)
	authService := service.NewAuthService(authRepo, rdb, cfg)

	platformUserCtrl := platform.NewSysUserCtrl(platUserService)
	platformRoleCtrl := platform.NewSysRoleCtrl(platRoleService)
	platformDeptCtrl := platform.NewSysDeptCtrl(platDeptService)
	platformPermCtrl := platform.NewSysPermissionCtrl(platRoleService)
	platformShopCtrl := platform.NewSysShopCtrl(platShopService)
	platformCategoryCtrl := platform.NewProductCategoryCtrl(platCategoryService)
	platformProductCtrl := platform.NewProductCtrl(platProductService)
	platformFinCatCtrl := platform.NewFinanceCategoryCtrl(platFinCatService)
	platformFinReportCtrl := platform.NewFinanceReportCtrl(platformsvc.NewFinanceReportService())
	platformAuthCtrl := platform.NewAuthController(authService)

	shopUserCtrl := shop.NewSysUserCtrl(shopUserService)
	shopRoleCtrl := shop.NewSysRoleCtrl(shopRoleService)
	shopDeptCtrl := shop.NewSysDeptCtrl(shopDeptService)
	shopPermCtrl := shop.NewSysPermissionCtrl(shopRoleService)
	shopAuthCtrl := shop.NewAuthController(authService)
	shopFinCatCtrl := shop.NewShopFinCategoryCtrl(shopFinCatService)
	shopFinAccountCtrl := shop.NewShopFinAccountCtrl(shopFinAccountService)
	shopProductCtrl := shop.NewShopProductCtrl(shopProductService)
	shopCustomerCtrl := shop.NewShopCustomerCtrl(shopCustomerService)
	shopOrderCtrl := shop.NewOrderCtrl(shopOrderService)
	shopRecordCtrl := shop.NewRecordCtrl(shopRecordService)
	shopReportCtrl := shop.NewFinanceReportCtrl()

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.RegisterPlatformRoutes(
		r, db, rdb, cfg.JWT.Secret,
		platformAuthCtrl.Login, platformAuthCtrl.Logout, platformAuthCtrl.UserInfo, platformAuthCtrl.Permissions, platformAuthCtrl.Refresh,
		platformUserCtrl, platformRoleCtrl, platformDeptCtrl, platformPermCtrl, platformShopCtrl,
		platformCategoryCtrl, platformProductCtrl,
		platformFinCatCtrl,
		platformFinReportCtrl,
	)

	router.RegisterShopRoutes(
		r, db, rdb, cfg.JWT.Secret,
		shopAuthCtrl.Login, shopAuthCtrl.Logout, shopAuthCtrl.UserInfo, shopAuthCtrl.Permissions, shopAuthCtrl.Refresh,
		shopUserCtrl, shopRoleCtrl, shopDeptCtrl, shopPermCtrl,
		shopFinCatCtrl,
		shopProductCtrl,
		shopCustomerCtrl,
		shopOrderCtrl,
		shopFinAccountCtrl,
		shopRecordCtrl,
		shopReportCtrl,
	)

	startServer(r, cfg)
}

func initDB(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}

	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connected")
	return db
}

func initRedis(cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis connected")
	return rdb
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.SysUser{},
		&entity.SysRole{},
		&entity.SysPermission{},
		&entity.SysRolePermission{},
		&entity.SysUserRole{},
		&entity.SysDept{},
		&entity.SysDeptClosure{},
		&entity.SysOperationLog{},
		&entity.SysShop{},
		&entity.ProductCategory{},
		&entity.Product{},
		&entity.ProductWorkflowNode{},
		&entity.FinanceCategory{},
		&entity.ShopFinanceCategory{},
		&entity.ShopFinanceAccount{},
		&entity.ShopProduct{},
		&entity.ShopCustomer{},
		&entity.OrderGroup{},
		&entity.OrderItem{},
		&entity.OrderItemNode{},
		&entity.OrderWorkflowLog{},
		&entity.OrderAttachment{},
		&entity.FinanceRecord{},
		&entity.FinanceAttachment{},
		&entity.ExportTask{},
	)
}

func startServer(r *gin.Engine, cfg *config.Config) {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("Server starting on :%d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
