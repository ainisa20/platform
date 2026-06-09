package platform

import (
	"strings"

	"github.com/gin-gonic/gin"
	"platform/internal/model/dto"
	"platform/internal/model/enum"
	"platform/internal/pkg/response"
	"platform/internal/service"
)

type AuthController struct {
	svc *service.AuthService
}

func NewAuthController(svc *service.AuthService) *AuthController {
	return &AuthController{svc: svc}
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	resp, err := ctrl.svc.PlatformLogin(req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	response.OK(c, resp)
}

func (ctrl *AuthController) Logout(c *gin.Context) {
	token := extractToken(c)
	if token == "" {
		response.Unauthorized(c, "missing token")
		return
	}

	if err := ctrl.svc.Logout(token); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OKMsg(c, "logged out")
}

func (ctrl *AuthController) UserInfo(c *gin.Context) {
	userID := c.GetUint64("user_id")
	tenantID := c.GetUint64("tenant_id")

	resp, err := ctrl.svc.GetUserInfo(userID, tenantID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, resp)
}

func (ctrl *AuthController) Permissions(c *gin.Context) {
	tree, err := ctrl.svc.GetPermissionTree(enum.SystemTypePlatform)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	if tree == nil {
		tree = []map[string]interface{}{}
	}
	response.OK(c, tree)
}

func (ctrl *AuthController) Refresh(c *gin.Context) {
	var req dto.RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	resp, err := ctrl.svc.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	response.OK(c, resp)
}

func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

func RegisterPlatformAuthRoutes(rg *gin.RouterGroup, ctrl *AuthController) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", ctrl.Login)
		auth.POST("/logout", ctrl.Logout)
		auth.GET("/userinfo", ctrl.UserInfo)
		auth.GET("/permissions", ctrl.Permissions)
		auth.POST("/refresh", ctrl.Refresh)
	}
}
