package permission

func CRUD(prefix string) []Button {
	return []Button{
		{Name: "查询", Code: prefix + ":list"},
		{Name: "新增", Code: prefix + ":create"},
		{Name: "编辑", Code: prefix + ":update"},
		{Name: "删除", Code: prefix + ":delete"},
	}
}

var (
	BAudit   = func(prefix string) Button { return Button{Name: "审核", Code: prefix + ":audit"} }
	BExport  = func(prefix string) Button { return Button{Name: "导出", Code: prefix + ":export"} }
	BImport  = func(prefix string) Button { return Button{Name: "导入", Code: prefix + ":import"} }
	BReset   = func(prefix string) Button { return Button{Name: "重置密码", Code: prefix + ":reset"} }
	BAssign  = func(prefix string) Button { return Button{Name: "分配权限", Code: prefix + ":assign"} }
	BStatus  = func(prefix string) Button { return Button{Name: "启用/停用", Code: prefix + ":status"} }
	BPrice   = func(prefix string) Button { return Button{Name: "改价", Code: prefix + ":price"} }
	BAdvance = func(prefix string) Button { return Button{Name: "流程推进", Code: prefix + ":advance"} }
	BUpload  = func(prefix string) Button { return Button{Name: "上传附件", Code: prefix + ":upload"} }
	BSync    = func(prefix string) Button { return Button{Name: "同步", Code: prefix + ":sync"} }
	BCancel  = func(prefix string) Button { return Button{Name: "取消", Code: prefix + ":cancel"} }
)

func CRUDWith(prefix string, extras ...func(string) Button) []Button {
	btns := CRUD(prefix)
	for _, e := range extras {
		btns = append(btns, e(prefix))
	}
	return btns
}
