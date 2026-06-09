package dto

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	ShopCode string `json:"shop_code"`
}

type LoginResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UserInfoResp struct {
	ID          uint64   `json:"id"`
	TenantID    uint64   `json:"tenant_id"`
	Username    string   `json:"username"`
	RealName    string   `json:"real_name"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}
