package permission

import "platform/internal/model/enum"

var PlatformManifest = Manifest{
	SystemType: enum.SystemTypePlatform,
	Nodes: []Node{
		{Name: "系统管理", Type: enum.PermTypeDirectory, Icon: "setting", Sort: 1, Children: []Node{
			{Name: "用户管理", Type: enum.PermTypeMenu, Path: "/system/user",
				Component: "platform/system/user/index", Sort: 1,
				Buttons: CRUDWith("platform:user", BAssign, BReset)},
			{Name: "角色管理", Type: enum.PermTypeMenu, Path: "/system/role",
				Component: "platform/system/role/index", Sort: 2,
				Buttons: CRUDWith("platform:role", BAssign)},
			{Name: "部门管理", Type: enum.PermTypeMenu, Path: "/system/dept",
				Component: "platform/system/dept/index", Sort: 3,
				Buttons: CRUD("platform:dept")},
		}},
		{Name: "店铺管理", Type: enum.PermTypeDirectory, Icon: "shop", Sort: 2, Children: []Node{
			{Name: "店铺列表", Type: enum.PermTypeMenu, Path: "/shop/list",
				Component: "platform/shop/list/index", Sort: 1,
				Buttons: []Button{
					{Name: "查询", Code: "platform:shop:list"},
					{Name: "新增", Code: "platform:shop:create"},
					{Name: "编辑", Code: "platform:shop:update"},
					{Name: "删除", Code: "platform:shop:delete"},
					{Name: "重置管理员密码", Code: "platform:shop:reset"},
					{Name: "启用/停用", Code: "platform:shop:status"},
				}},
		}},
		{Name: "商品管理", Type: enum.PermTypeDirectory, Icon: "goods", Sort: 3, Children: []Node{
			{Name: "商品分类", Type: enum.PermTypeMenu, Path: "/product/category",
				Component: "platform/product/category/index", Sort: 1,
				Buttons: CRUD("platform:product:category")},
			{Name: "商品列表", Type: enum.PermTypeMenu, Path: "/product/list",
				Component: "platform/product/list/index", Sort: 2,
				Buttons: []Button{
					{Name: "查询", Code: "platform:product:list"},
					{Name: "新增", Code: "platform:product:create"},
					{Name: "编辑", Code: "platform:product:update"},
					{Name: "删除", Code: "platform:product:delete"},
					{Name: "上架/下架", Code: "platform:product:status"},
				}},
		}},
		{Name: "财务管理", Type: enum.PermTypeDirectory, Icon: "finance", Sort: 4, Children: []Node{
			{Name: "收支分类", Type: enum.PermTypeMenu, Path: "/finance/category",
				Component: "platform/finance/category/index", Sort: 1,
				Buttons: CRUD("platform:finance:category")},
			{Name: "财务报表", Type: enum.PermTypeMenu, Path: "/finance/report",
				Component: "platform/finance/report/index", Sort: 2,
				Buttons: []Button{
					{Name: "查询", Code: "platform:finance:report:list"},
					{Name: "导出", Code: "platform:finance:report:export"},
				}},
		}},
	},
}
