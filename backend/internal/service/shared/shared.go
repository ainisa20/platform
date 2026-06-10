package shared

import (
	"errors"

	"platform/internal/model/dto"
	"platform/internal/model/entity"
)

var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUsernameExists    = errors.New("用户名已存在")
	ErrRoleNotFound      = errors.New("角色不存在")
	ErrRoleCodeExists    = errors.New("角色编码已存在")
	ErrDeptNotFound      = errors.New("部门不存在")
	ErrDeptHasUsers      = errors.New("部门下存在用户，无法删除")
	ErrInvalidPermission = errors.New("包含无效的权限ID")
	ErrSameDeptParent    = errors.New("不能将部门设置为自己的子部门")
	ErrShopNotFound      = errors.New("店铺不存在")
	ErrShopCodeExists    = errors.New("店铺编码已存在")
	ErrShopHasUsers      = errors.New("店铺下存在用户，无法删除")
	ErrCategoryNotFound  = errors.New("商品分类不存在")
	ErrCategoryNameExists = errors.New("分类名称已存在")
	ErrProductNotFound   = errors.New("商品不存在")
	ErrProductCodeExists  = errors.New("商品编号已存在")
	ErrPermissionExceeded = errors.New("不能分配超出自身权限范围的权限")
	ErrDeptOutOfScope     = errors.New("不能分配到权限范围外的部门")
	ErrRoleNotAssignable  = errors.New("不能分配超出自身权限范围的角色")
	ErrFinanceCategoryNotFound     = errors.New("收支分类不存在")
	ErrFinanceCategoryNameExists   = errors.New("收支分类名称已存在")
	ErrFinanceCategoryHasChildren  = errors.New("存在子分类，无法删除")
	ErrFinanceCategorySynced       = errors.New("已被店铺同步，无法删除")
	ErrFinanceCategoryMaxLevel     = errors.New("最多支持三级分类")
	ErrFinanceCategoryTypeMismatch = errors.New("子分类类型必须与父分类一致")
	ErrShopFinCategoryNotFound     = errors.New("店铺收支分类不存在")
	ErrShopFinCategoryAlreadySynced = errors.New("该分类已同步")
	ErrShopFinCategoryReferenced    = errors.New("该分类已被收支记录引用，无法取消同步")
	ErrShopProductNotFound          = errors.New("商品不存在")
	ErrShopProductAlreadySelected   = errors.New("该商品已选品")
	ErrShopProductNotOnShelf        = errors.New("平台商品未上架，无法选品")
	ErrShopCustomerNotFound         = errors.New("客户不存在")
	ErrShopCustomerNameExists       = errors.New("同类型下客户名称已存在")
	ErrShopCustomerHasOrders        = errors.New("客户存在订单，无法删除")
	ErrOrderNotFound                = errors.New("订单不存在")
	ErrOrderItemNotFound            = errors.New("订单明细不存在")
	ErrOrderInProgress              = errors.New("订单已进入流程，不能整单取消")
	ErrOrderCompleted               = errors.New("订单已完成，不能取消")
	ErrOrderItemCannotCancel        = errors.New("明细已完成或已取消，不能再取消")
	ErrOrderNoItems                 = errors.New("订单必须包含至少一个商品")
	ErrWorkflowEmpty                = errors.New("商品未配置流程节点，无法下单")
	ErrShopFinAccountNotFound       = errors.New("账户不存在")
	ErrShopFinAccountNameExists     = errors.New("账户名称已存在")
	ErrFinRecordNotFound             = errors.New("财务记录不存在")
	ErrFinRecordAccountInvalid       = errors.New("账户无效或不属于当前店铺")
	ErrFinRecordCategoryInvalid  = errors.New("收支分类无效或不属于当前店铺")
	ErrFinRecordCategoryNotLeaf = errors.New("只能选择最末级（第三级）收支分类")
	ErrFinRecordOrderInvalid         = errors.New("关联订单无效或不属于当前店铺")
	ErrFinRecordApproved             = errors.New("已通过的记录不可编辑或删除")
	ErrFinRecordReviewerIsCreator    = errors.New("审核人不能是创建人")
	ErrFinRecordActualAmountRequired = errors.New("审核通过时必须填写实际金额")
	ErrFinRecordInvalidAction        = errors.New("无效的审核动作")
	ErrFinRecordInvalidStatus        = errors.New("当前状态不允许此操作")
)

func UserToResp(u *entity.SysUser, roles []dto.RoleResp) *dto.UserResp {
	if roles == nil {
		roles = []dto.RoleResp{}
	}
	return &dto.UserResp{
		ID:          u.ID,
		TenantID:    u.TenantID,
		DeptID:      u.DeptID,
		Username:    u.Username,
		RealName:    u.RealName,
		Phone:       u.Phone,
		Email:       u.Email,
		Avatar:      u.Avatar,
		Status:      u.Status,
		LastLoginAt: u.LastLoginAt,
		LastLoginIP: u.LastLoginIP,
		CreatedAt:   u.CreatedAt,
		CreatedBy:   u.CreatedBy,
		UpdatedAt:   u.UpdatedAt,
		UpdatedBy:   u.UpdatedBy,
		Roles:       roles,
	}
}

func RoleToResp(r *entity.SysRole, perms []dto.PermissionResp) dto.RoleResp {
	if perms == nil {
		perms = []dto.PermissionResp{}
	}
	return dto.RoleResp{
		ID:          r.ID,
		TenantID:    r.TenantID,
		RoleName:    r.RoleName,
		RoleCode:    r.RoleCode,
		DataScope:   r.DataScope,
		Sort:        r.Sort,
		Status:      r.Status,
		Remark:      r.Remark,
		CreatedAt:   r.CreatedAt,
		CreatedBy:   r.CreatedBy,
		UpdatedAt:   r.UpdatedAt,
		UpdatedBy:   r.UpdatedBy,
		Permissions: perms,
	}
}

func PermissionToResp(p *entity.SysPermission) dto.PermissionResp {
	return dto.PermissionResp{
		ID:         p.ID,
		ParentID:   p.ParentID,
		SystemType: p.SystemType,
		Name:       p.Name,
		Type:       p.Type,
		Path:       p.Path,
		Component:  p.Component,
		PermsCode:  p.PermsCode,
		Icon:       p.Icon,
		Sort:       p.Sort,
		Visible:    p.Visible,
		Status:     p.Status,
	}
}

func DeptToResp(d *entity.SysDept) dto.DeptResp {
	return dto.DeptResp{
		ID:        d.ID,
		TenantID:  d.TenantID,
		ParentID:  d.ParentID,
		Ancestors: d.Ancestors,
		DeptName:  d.DeptName,
		Sort:      d.Sort,
		Leader:    d.Leader,
		Phone:     d.Phone,
		Status:    d.Status,
		CreatedAt: d.CreatedAt,
		CreatedBy: d.CreatedBy,
		UpdatedAt: d.UpdatedAt,
		UpdatedBy: d.UpdatedBy,
	}
}

func BuildDeptTree(flat []dto.DeptResp, parentID uint64) []dto.DeptResp {
	visibleIDs := make(map[uint64]struct{}, len(flat))
	for _, d := range flat {
		visibleIDs[d.ID] = struct{}{}
	}
	var tree []dto.DeptResp
	for _, d := range flat {
		if d.ParentID == parentID {
			d.Children = BuildDeptTree(flat, d.ID)
			if d.Children == nil {
				d.Children = []dto.DeptResp{}
			}
			tree = append(tree, d)
			continue
		}
		if parentID == 0 {
			if _, parentVisible := visibleIDs[d.ParentID]; !parentVisible && d.ParentID != 0 {
				d.Children = BuildDeptTree(flat, d.ID)
				if d.Children == nil {
					d.Children = []dto.DeptResp{}
				}
				tree = append(tree, d)
			}
		}
	}
	return tree
}

func BuildPermissionTree(flat []dto.PermissionResp, parentID uint64) []dto.PermissionResp {
	var tree []dto.PermissionResp
	for _, p := range flat {
		if p.ParentID == parentID {
			p.Children = BuildPermissionTree(flat, p.ID)
			if p.Children == nil {
				p.Children = []dto.PermissionResp{}
			}
			tree = append(tree, p)
		}
	}
	return tree
}
