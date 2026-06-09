# 阶段一实施差距分析（vs plan2.md）

> 基于 2026-06-08 实际部署 + 浏览器端到端测试 + API 实测
> 状态：阶段一 CRUD 部分完成，**数据权限未实现**

---

## ✅ 已完成且验证通过

| 计划章节 | 实现 | 验证 |
|---|---|---|
| 4.1 用户认证（JWT + Refresh + 黑名单） | ✅ | Playwright 登录/退出通过 |
| 4.2 多租户数据隔离（PostgreSQL RLS） | ✅ | tenant_id 写入 + 8 张表 RLS 策略已建 |
| 4.3 权限模型（RBAC + 按钮级） | ✅ | 116 条权限自动同步，`v-permission` 工作 |
| 4.4 数据范围（按部门树） | ❌ | 中间件存在但**未注册到 router** |
| 5.1 用户管理 CRUD | ✅ | 4 个用户增删查改浏览器实测通过 |
| 5.2 角色管理 CRUD | ✅ | 角色列表 + 权限树分配实测通过 |
| 5.3 部门管理 CRUD | ✅ | 树表 + 新增子部门 + 删除实测通过 |
| 6. Docker 部署 | ✅ | 5 个容器运行，nginx SPA 路由正常 |

---

## ❌ 关键差距

### 差距 1：数据权限完全未生效（**高优先级**）

**位置**：`backend/internal/middleware/datascope.go` 存在但**未被使用**

**问题**：
1. `router/platform.go` 和 `router/shop.go` 均未注册 `middleware.DataScopeMiddleware(db)`
2. 即使注册了，中间件只把 filter 存到 `c.Set("data_scope_filter", scopeFunc)`，**没有 controller 读取这个 key**
3. `userRepo.List()` 直接 `WHERE tenant_id = ?`，无视数据范围
4. 即便逻辑正确，过滤语义也错了：用 `created_by IN (user_ids)` 而非 `dept_id IN (dept_ids)`，导致管理员创建的下属用户相互不可见

**实测**：
- B (dept=技术部, data_scope=2 本部门及下级) 应该看到 B+C，实际看到全部 4 个用户
- C (dept=前端组, data_scope=2) 应该看到仅 C，实际看到全部 4 个用户

**修复方案**（约 30 行代码）：
```go
// 1. router 注册
sys.Use(middleware.JWTAuthMiddleware(jwtSecret),
       middleware.DBInjectMiddleware(db),
       middleware.DataScopeMiddleware(db))

// 2. controller List 调用
func (ctrl *SysUserCtrl) List(c *gin.Context) {
    db := c.MustGet("db").(*gorm.DB)
    if filter, ok := c.Get("data_scope_filter"); ok {
        if f, ok := filter.(func(*gorm.DB) *gorm.DB); ok {
            db = f(db)  // 关键：应用 filter
        }
    }
    // ...
}

// 3. 改写 datascope.go 的语义
// 对于 sys_user 表，按 dept_id 范围过滤
// 对于业务表（订单、商品），按 created_by 范围过滤
// 需要一个统一的 ApplyDataScope(db, scope, filterField) 辅助函数
```

---

### 差距 2：sys_dept_closure 自动维护有 bug（**高优先级**）

**位置**：`backend/internal/repository/sys_dept_repo.go:30-75` Create() 方法

**问题**：
- `Create()` 查询 `parent.Closures` 来构造新部门的祖先链
- 但 SQL 初始脚本插入的 总部 没有 `(1,1,0)` 自连接行
- 导致 总部 下的所有子部门祖先链**不包含 总部 自身**
- 实测创建 技术部 后，closure 表只有 `(2,2,0)` 缺 `(1,1,0)` 和 `(1,2,1)`

**修复方案**：
- 方案 A：在 SQL 初始脚本里补充 `INSERT INTO sys_dept_closure (tenant_id, ancestor_id, descendant_id, depth) VALUES (0, 1, 1, 0);`
- 方案 B：在 `DeptService.Create` 入口检查 `dept.Ancestors` 链是否完整，缺失时自动补齐
- 方案 C：在 `main.go` 启动时跑一次 closure 重建（已用 SQL 手写的方式）

**影响**：
- 直接影响差距 1（数据权限）的 dept 范围查询
- 影响所有依赖 closure 的部门树查询

---

### 差距 3：死代码：GET /:id 端点已删除但 controller 方法未清理

**位置**：`backend/internal/controller/platform/sys_user.go:41` `Get()` 方法

**问题**：
- 路由已删除 `users.GET("/:id", ...)` 但 `func (ctrl *SysUserCtrl) Get(c *gin.Context)` 仍在
- `internal/service/sys_user_service.go:129` `GetByID` 也仍在
- 不会被编译报错（Go 不警告未使用方法），但属于死代码

**修复**：删除 `Get` controller 方法和 `GetByID` service 方法

---

### 差距 4：店铺系统未实测（**中优先级**）

**位置**：`router/shop.go` 整条链路

**问题**：
- 平台已实测完整，店铺系统**完全没有跑过任何测试**
- `TenantRLSMiddleware` 依赖 `claims.TenantID > 0` 启动事务，理论上能让店铺用户只能看到自己店铺的数据
- 但目前数据库里 0 个店铺（`sys_dept_closure` 没有 tenant_id > 0 的行），无法验证
- 阶段二/三需要先创建店铺，然后验证：
  - 店铺 A 管理员能否看到店铺 B 的数据（应该不能）
  - 同一店铺内，部门树数据权限是否生效

**影响**：店铺 RLS 逻辑可能与平台不一致或有未发现的 bug

---

### 差距 5：审计日志未生效（**低优先级**）

**位置**：`backend/internal/middleware/audit.go`，`backend/internal/model/entity/entity.go:135` `SysOperationLog`

**问题**：
- 路由里注册了 `middleware.AuditMiddleware()`
- 但该中间件是空实现（`c.Next()` 之后什么都不做）
- `sys_operation_log` 表存在但永远为空

**修复**：在 `AuditMiddleware` 里把请求 method/path/user_id/timestamp 写入 `sys_operation_log`

---

### 差距 6：文件上传未实现（**低优先级，阶段二需要**）

**位置**：plan2.md 第 5.x 章节多处提到「上传附件」

**现状**：
- 容器里有 minio 但没有接入
- 没有任何 upload 端点
- 前端没有 el-upload 配置

**修复**：阶段二做订单/商品时一起做

---

## 📊 完成度统计

| 模块 | 完成度 | 备注 |
|---|---|---|
| 基础设施（DB/Redis/容器） | 100% | 5 个容器运行健康 |
| 认证（JWT/Refresh/黑名单） | 100% | 含 1 个 panic bug 已修复 |
| RBAC（角色/权限/按钮控制） | 100% | `v-permission` 工作正常 |
| 多租户（RLS） | 100% | SQL 策略 + tenant 中间件 |
| 用户管理 CRUD | 100% | 4 个用户实测通过 |
| 角色管理 CRUD | 100% | 权限树分配实测通过 |
| 部门管理 CRUD | 100% | 树表 + 展开折叠实测通过 |
| **数据范围权限** | **0%** | 关键差距，影响所有未来业务模块 |
| 店铺 RLS 验证 | 0% | 0 个店铺，无法实测 |
| 审计日志 | 0% | 中间件空实现 |
| 文件上传 | 0% | 阶段二需要 |

**阶段一总体完成度：约 75%**（CRUD + 认证 + RBAC 完成，数据权限 + 审计缺失）

---

## 🎯 建议优先级（推进到阶段二之前）

1. **必修**：实现数据权限（差距 1）— 否则阶段二做订单/商品时，所有人都能看所有店铺的数据
2. **必修**：修复 closure 维护 bug（差距 2）— 与差距 1 强耦合
3. **建议**：建 1-2 个测试店铺 + 几个店铺用户，端到端验证店铺 RLS（差距 4）
4. **建议**：清理死代码（差距 3）
5. **可选**：实现审计日志（差距 5）— 阶段二再做也行
6. **延后**：文件上传（差距 6）— 阶段二业务模块自然引入
