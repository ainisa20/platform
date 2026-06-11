package permission

import "platform/internal/model/enum"

var ShopManifest = Manifest{
	SystemType: enum.SystemTypeShop,
	Nodes: []Node{
		{Name: "系统管理", Type: enum.PermTypeDirectory, Icon: "setting", Sort: 1, Children: []Node{
			{Name: "用户管理", Type: enum.PermTypeMenu, Path: "/system/user",
				Component: "shop/system/user/index", Sort: 1,
				Buttons: CRUDWith("shop:user", BAssign, BReset)},
			{Name: "角色管理", Type: enum.PermTypeMenu, Path: "/system/role",
				Component: "shop/system/role/index", Sort: 2,
				Buttons: CRUDWith("shop:role", BAssign)},
			{Name: "部门管理", Type: enum.PermTypeMenu, Path: "/system/dept",
				Component: "shop/system/dept/index", Sort: 3,
				Buttons: CRUD("shop:dept")},
		}},
		{Name: "商品管理", Type: enum.PermTypeDirectory, Icon: "goods", Sort: 2, Children: []Node{
			{Name: "选品管理", Type: enum.PermTypeMenu, Path: "/product",
				Component: "shop/product/index", Sort: 1,
				Buttons: []Button{
					{Name: "查询", Code: "shop:product:list"},
					{Name: "选品", Code: "shop:product:select"},
					{Name: "改价", Code: "shop:product:price"},
					{Name: "上架/下架", Code: "shop:product:status"},
					{Name: "取消选品", Code: "shop:product:delete"},
				}},
		}},
		{Name: "客户管理", Type: enum.PermTypeDirectory, Icon: "user", Sort: 3, Children: []Node{
			{Name: "客户列表", Type: enum.PermTypeMenu, Path: "/customer",
				Component: "shop/customer/index", Sort: 1,
				Buttons: CRUDWith("shop:customer", BExport)},
		}},
		{Name: "订单管理", Type: enum.PermTypeDirectory, Icon: "order", Sort: 4, Children: []Node{
			{Name: "订单列表", Type: enum.PermTypeMenu, Path: "/order",
				Component: "shop/order/index", Sort: 1,
				Buttons: []Button{
					{Name: "查询", Code: "shop:order:list"},
					{Name: "创建", Code: "shop:order:create"},
					{Name: "取消", Code: "shop:order:cancel"},
					{Name: "流程推进", Code: "shop:order:advance"},
					{Name: "上传附件", Code: "shop:order:upload"},
					{Name: "导出", Code: "shop:order:export"},
				}},
		}},
		{Name: "财务管理", Type: enum.PermTypeDirectory, Icon: "finance", Sort: 5, Children: []Node{
			{Name: "收支账户", Type: enum.PermTypeMenu, Path: "/finance/account",
				Component: "shop/finance/account/index", Sort: 1,
				Buttons: CRUD("shop:finance:account")},
			{Name: "收支分类", Type: enum.PermTypeMenu, Path: "/finance/category",
				Component: "shop/finance/category/index", Sort: 2,
				Buttons: []Button{
					{Name: "查询", Code: "shop:finance:category:list"},
					{Name: "同步", Code: "shop:finance:category:sync"},
					{Name: "取消同步", Code: "shop:finance:category:delete"},
				}},
			{Name: "收支记录", Type: enum.PermTypeMenu, Path: "/finance/record",
				Component: "shop/finance/record/index", Sort: 3,
				Buttons: []Button{
					{Name: "查询", Code: "shop:finance:record:list"},
					{Name: "新增", Code: "shop:finance:record:create"},
					{Name: "编辑", Code: "shop:finance:record:update"},
					{Name: "删除", Code: "shop:finance:record:delete"},
					{Name: "审核", Code: "shop:finance:record:audit"},
					{Name: "上传附件", Code: "shop:finance:record:upload"},
					{Name: "导出", Code: "shop:finance:record:export"},
					{Name: "导出附件", Code: "shop:finance:record:export-zip"},
				}},
			{Name: "财务报表", Type: enum.PermTypeMenu, Path: "/finance/report",
				Component: "shop/finance/report/index", Sort: 4,
				Buttons: []Button{
					{Name: "查询", Code: "shop:finance:report:list"},
					{Name: "导出", Code: "shop:finance:report:export"},
				}},
		}},
	},
}
