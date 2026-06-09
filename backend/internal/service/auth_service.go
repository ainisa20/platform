package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"platform/internal/config"
	"platform/internal/model/dto"
	"platform/internal/model/entity"
	"platform/internal/model/enum"
	"platform/internal/pkg/auth"
	"platform/internal/repository"
)

type AuthService struct {
	repo *repository.AuthRepository
	rdb  *redis.Client
	cfg  *config.Config
}

func NewAuthService(repo *repository.AuthRepository, rdb *redis.Client, cfg *config.Config) *AuthService {
	return &AuthService{repo: repo, rdb: rdb, cfg: cfg}
}

func (s *AuthService) PlatformLogin(req dto.LoginReq) (*dto.LoginResp, error) {
	return s.login(req, enum.TenantPlatform, "platform")
}

func (s *AuthService) ShopLogin(req dto.LoginReq) (*dto.LoginResp, error) {
	if req.ShopCode == "" {
		return nil, errors.New("shop_code is required")
	}

	tenantID, shopStatus, err := s.repo.GetShopTenantIDByCode(req.ShopCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("shop not found")
		}
		log.Printf("[auth] query shop failed: %v", err)
		return nil, errors.New("internal error")
	}
	if shopStatus != enum.StatusEnabled {
		return nil, errors.New("shop is disabled or closed")
	}

	return s.login(req, tenantID, "shop")
}

func (s *AuthService) login(req dto.LoginReq, tenantID uint64, audience string) (*dto.LoginResp, error) {
	user, err := s.repo.GetByUsername(req.Username, tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		log.Printf("[auth] query user failed: %v", err)
		return nil, errors.New("internal error")
	}

	if user.Status != enum.StatusEnabled {
		return nil, errors.New("user is disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	dataScope, _ := s.repo.GetUserMaxDataScope(user.ID, tenantID)

	deptID := uint64(0)
	if user.DeptID != nil {
		deptID = *user.DeptID
	}

	claims := auth.JWTClaims{
		UserID:    user.ID,
		TenantID:  tenantID,
		DeptID:    deptID,
		Username:  user.Username,
		DataScope: dataScope,
	}

	jwtCfg := s.cfg.JWT
	accessToken, err := auth.GenerateAccessToken(claims, jwtCfg.Secret, jwtCfg.AccessTTL, jwtCfg.Issuer, audience)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken(claims, jwtCfg.Secret, jwtCfg.RefreshTTL, jwtCfg.Issuer, audience)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	_ = s.repo.UpdateLoginInfo(user.ID, tenantID, "")

	return &dto.LoginResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(jwtCfg.AccessTTL.Seconds()),
	}, nil
}

func (s *AuthService) Logout(tokenString string) error {
	claims, err := auth.ParseToken(tokenString, s.cfg.JWT.Secret)
	if err != nil {
		return nil
	}

	ctx := context.Background()
	blacklisted, err := auth.IsTokenBlacklisted(ctx, s.rdb, claims.ID)
	if err != nil {
		log.Printf("[auth] check blacklist failed: %v", err)
		return nil
	}
	if blacklisted {
		return nil
	}

	if claims.ExpiresAt == nil {
		return nil
	}
	if err := auth.BlacklistToken(ctx, s.rdb, claims.ID, claims.ExpiresAt.Time); err != nil {
		log.Printf("[auth] blacklist token failed: %v", err)
		return fmt.Errorf("logout failed: %w", err)
	}
	return nil
}

func (s *AuthService) RefreshToken(refreshToken string) (*dto.LoginResp, error) {
	claims, err := auth.ParseToken(refreshToken, s.cfg.JWT.Secret)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	audience := auth.ExtractAudience(claims, "platform", "shop")
	if audience == "" {
		return nil, errors.New("invalid refresh token audience")
	}

	blacklisted, err := auth.IsTokenBlacklisted(context.Background(), s.rdb, claims.ID)
	if err != nil {
		log.Printf("[auth] check blacklist failed: %v", err)
	} else if blacklisted {
		return nil, errors.New("refresh token has been revoked")
	}

	if claims.ExpiresAt != nil {
		_ = auth.BlacklistToken(context.Background(), s.rdb, claims.ID, claims.ExpiresAt.Time)
	}

	newClaims := auth.JWTClaims{
		UserID:    claims.UserID,
		TenantID:  claims.TenantID,
		DeptID:    claims.DeptID,
		Username:  claims.Username,
		DataScope: claims.DataScope,
	}

	jwtCfg := s.cfg.JWT
	accessToken, err := auth.GenerateAccessToken(newClaims, jwtCfg.Secret, jwtCfg.AccessTTL, jwtCfg.Issuer, audience)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	newRefreshToken, err := auth.GenerateRefreshToken(newClaims, jwtCfg.Secret, jwtCfg.RefreshTTL, jwtCfg.Issuer, audience)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &dto.LoginResp{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(jwtCfg.AccessTTL.Seconds()),
	}, nil
}

func (s *AuthService) GetUserInfo(userID, tenantID uint64) (*dto.UserInfoResp, error) {
	user, err := s.repo.GetUserByID(userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	roles, _ := s.repo.GetUserRoleCodes(userID, tenantID)
	perms, _ := s.repo.GetUserPermissionCodes(userID, tenantID)

	if roles == nil {
		roles = []string{}
	}
	if perms == nil {
		perms = []string{}
	}

	return &dto.UserInfoResp{
		ID:          user.ID,
		TenantID:    user.TenantID,
		Username:    user.Username,
		RealName:    user.RealName,
		Roles:       roles,
		Permissions: perms,
	}, nil
}

func (s *AuthService) GetPermissionTree(systemType string) ([]map[string]interface{}, error) {
	perms, err := s.repo.GetPermissionsBySystemType(systemType)
	if err != nil {
		return nil, err
	}
	return buildPermTree(perms, 0), nil
}

func buildPermTree(perms []entity.SysPermission, parentID uint64) []map[string]interface{} {
	var tree []map[string]interface{}
	for _, p := range perms {
		if p.ParentID != parentID {
			continue
		}
		node := map[string]interface{}{
			"id":         p.ID,
			"parent_id":  p.ParentID,
			"name":       p.Name,
			"type":       p.Type,
			"path":       p.Path,
			"component":  p.Component,
			"perms_code": p.PermsCode,
			"icon":       p.Icon,
			"sort":       p.Sort,
			"visible":    p.Visible,
			"status":     p.Status,
		}
		children := buildPermTree(perms, p.ID)
		if len(children) > 0 {
			node["children"] = children
		}
		tree = append(tree, node)
	}
	return tree
}
