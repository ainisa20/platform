# 多租户 B2B 服务管理 SaaS 平台 — 实施计划

> **版本:** 2.0  
> **日期:** 2026-06-08  
> **状态:** 待评审  
> **变更说明:** v2.0 — 数据库迁移至 PostgreSQL（RLS 行级安全）；权限树改为代码声明+自动同步；按钮权限全可配

---

## 目录

- [1. 项目概述](#1-项目概述)
- [2. 技术架构](#2-技术架构)
- [3. 数据库设计](#3-数据库设计)
- [4. 数据权限设计](#4-数据权限设计)
- [5. 权限体系设计（自动加载）](#5-权限体系设计自动加载)
- [6. 功能模块详细设计](#6-功能模块详细设计)
- [7. 核心技术方案](#7-核心技术方案)
- [8. 安全设计](#8-安全设计)
- [9. 开发阶段与里程碑](#9-开发阶段与里程碑)
- [10. 部署与运维](#10-部署与运维)
- [11. 风险与应对](#11-风险与应对)
- [12. 待确认事项](#12-待确认事项)

---

## 1. 项目概述

### 1.1 业务模型

本项目是一个 **B2B 服务管理 SaaS 平台**，包含**两套独立系统**，部署在同一个项目目录（monorepo）中：

- **平台系统（Platform）**：平台运营方使用。管理整个平台 — 创建店铺/租户、定义标准商品目录与服务流程模板、管理收支分类、查看平台级报表。
- **店铺系统（Shop）**：每个店铺/租户独立使用。管理自己的客户、从平台商品库选品并定价、创建订单（按客户下单、按商品拆分为多条订单项）、跟进服务流程、管理自己的财务账户与收支记录。

两套系统**共享基础设施**（同一数据库、同一后端服务进程、同一前端项目），但**逻辑完全独立**：独立的登录入口与认证、独立的用户/角色/部门体系、独立的数据作用域、独立的前端路由与界面。

### 1.2 核心业务流程

```
平台系统                              店铺系统
────────                              ────────
创建店铺 ────────────────────► 店铺管理员账户初始化
                                     │
定义商品分类                          │
定义商品 + 服务流程模板                │
定义收支分类（三级） ───────────► 选品（从平台商品库选择，可自定义价格）
                                     │
                                  创建客户
                                     │
                                  创建订单（选客户 + 选商品 → 自动拆分）
                                     │
                                  服务流程跟进（联系→下单→创建账号→部署→实施→完成）
                                  每步可填备注、传附件、查看历史
                                     │
                                  财务管理（收支账户、收支记录、审核、报表）
```

### 1.3 设计原则

| 原则 | 说明 |
|---|---|
| **两套系统，一个代码库** | Platform 和 Shop 逻辑独立，共享 monorepo、基础设施、通用组件 |
| **服务型商品，非电商** | 无购物车、无库存、无物流。核心是服务交付流程跟踪 |
| **多租户四层隔离** | PostgreSQL RLS（数据库层）+ GORM Scopes（ORM 层）+ 中间件注入 + 请求校验 |
| **权限代码驱动** | 权限树在 Go 代码中声明，启动时自动同步到数据库，开发者无需手动维护菜单 |
| **按钮权限全可配** | 每个操作（增删改查审导出等）都是独立权限点，角色管理界面可勾选 |
| **简单流程引擎** | 固定顺序节点状态机，不做 BPMN、不做条件分支、不做并行 |
| **平台控制标准** | 商品库、服务流程、收支分类由平台统一定义，店铺只能选用不能自建 |
| **财务不可篡改** | 审核通过后记录不可编辑/删除，所有变更有审计日志 |

---

## 2. 技术架构

### 2.1 系统架构图

```
                         ┌───────────────────────────────────────┐
                         │           Nginx（反向代理）              │
                         │   :80/:443 → 静态资源 + API 路由        │
                         └──────────┬──────────────┬──────────────┘
                                    │              │
                  ┌─────────────────┘              └──────────────────┐
                  ▼                                                   ▼
        ┌─────────────────────┐                           ┌─────────────────────┐
        │   前端（Vue3）        │                           │    后端（Go）         │
        │   vue-pure-admin    │                           │    Gin + GORM        │
        │   Element Plus      │        HTTP/JSON           │    PostgreSQL        │
        │   TypeScript        │◄──────────────────────────►│                      │
        │                     │                           │  ┌────────────────┐ │
        │  ┌───────────────┐  │                           │  │   中间件链      │ │
        │  │ 平台系统       │  │                           │  │ • JWT 认证      │ │
        │  │ /platform/*   │  │                           │  │ • 租户上下文    │ │
        │  ├───────────────┤  │                           │  │ • RLS 事务      │ │
        │  │ 店铺系统       │  │                           │  │ • 数据权限      │ │
        │  │ /shop/*       │  │                           │  │ • RBAC 鉴权     │ │
        │  └───────────────┘  │                           │  │ • 操作审计      │ │
        │                     │                           │  │ • 限流          │ │
        │  构建: Vite         │                           │  └────────────────┘ │
        └─────────────────────┘                           │                      │
                                                          │  ┌────────────────┐ │
                                                          │  │   业务服务      │ │
                                                          │  │ • 权限自动同步  │ │
                                                          │  │ • 工作流引擎    │ │
                                                          │  │ • 订单拆分      │ │
                                                          │  │ • 财务审核      │ │
                                                          │  │ • 文件管理      │ │
                                                          │  └───────┬────────┘ │
                                                          └──────────┼──────────┘
                                                                     │
                                    ┌────────────────────────────────┼──────────────┐
                                    ▼                                ▼              ▼
                           ┌─────────────────┐          ┌────────────────┐ ┌──────────────┐
                           │  PostgreSQL 16   │          │  Redis 7.x     │ │  MinIO (S3)  │
                           │  RLS 行级安全    │          │  缓存/会话     │ │  附件存储     │
                           │  JSONB 灵活字段  │          │  Token 黑名单  │ │              │
                           └─────────────────┘          └────────────────┘ └──────────────┘
```

### 2.2 技术选型

| 层级 | 技术 | 选型理由 |
|---|---|---|
| **后端语言** | Go 1.22+ | 高性能、高并发、强类型 |
| **Web 框架** | Gin v1.10+ | 最流行 Go Web 框架（88K+ Stars），中间件生态丰富 |
| **ORM** | GORM v2 | Scopes 机制适合租户过滤；pgx v5 驱动成熟 |
| **状态机** | looplab/fsm | 轻量级有限状态机，支持回调，完美匹配顺序流程 |
| **数据库** | **PostgreSQL 16** | 原生 RLS 行级安全（数据库层强制租户隔离）、JSONB（支付配置灵活字段）、复杂查询性能更优 |
| **缓存** | Redis 7.x | 会话管理、接口限流、Token 黑名单、权限缓存 |
| **文件存储** | MinIO（S3 兼容） | 自部署对象存储，管理工作流附件与财务凭证 |
| **前端框架** | Vue 3.5+（Composition API） | 响应式、组件化 |
| **前端模板** | vue-pure-admin（Element Plus） | 成熟 RBAC、中文文档优秀、代码简洁 |
| **前端语言** | TypeScript 5.6+ | 类型安全 |
| **状态管理** | Pinia | Vue 官方推荐 |
| **构建工具** | Vite 7+ | 快速 HMR |
| **反向代理** | Nginx | 静态资源 + API 代理 + TLS |
| **容器化** | Docker + Docker Compose（MVP）→ K8s（扩展） | 环境一致性 |

> **为何选 PostgreSQL 而非 MySQL：**
> 1. **RLS 行级安全**：数据库层强制租户隔离，即使应用代码遗漏 WHERE 也不会泄露数据。MySQL 无此能力。
> 2. **JSONB**：收支账户配置（对公/微信/支付宝各有不同字段）用一列 JSONB 搞定，无需大量可空字段。
> 3. **报表性能**：复杂聚合查询快 2-13 倍（财务报表场景）。
> 4. **递归 CTE**：MySQL 8.0 也支持，此项不再是差异点。

### 2.3 多租户隔离架构

**策略：共享 PostgreSQL 数据库 + RLS 行级安全 + `tenant_id` 字段**

```
┌──────────────────────────────────────────────────────────────┐
│                   单一 PostgreSQL 数据库                       │
│                                                              │
│  四层隔离保障（由内到外）：                                     │
│                                                              │
│  ❶ 数据库层（RLS）★ 本项目独有保障                             │
│     每个店铺级表启用 RLS 策略：                                 │
│     USING (tenant_id = current_setting('app.tenant_id')::bigint)
│     → 即使应用代码忘了 WHERE，数据库也会自动过滤                │
│                                                              │
│  ❷ ORM 层（GORM Scopes）                                      │
│     每个查询自动追加 WHERE tenant_id = ?                        │
│     → 正常路径的双重保障                                       │
│                                                              │
│  ❸ 中间件层                                                    │
│     JWT → 提取 tenant_id → SET set_config('app.tenant_id')   │
│     → GORM Scope + RLS 上下文同时注入                          │
│                                                              │
│  ❹ 校验层                                                      │
│     请求体中的 tenant_id 必须与 JWT 中一致                      │
│     → 防止越权注入                                             │
│                                                              │
│  平台级表：不启用 RLS，tenant_id = 0 或无 tenant_id            │
│  店铺级表：启用 RLS，tenant_id > 0                             │
└──────────────────────────────────────────────────────────────┘
```

### 2.4 项目目录结构

```
platform/
├── backend/
│   ├── cmd/server/main.go              # 入口
│   ├── internal/
│   │   ├── config/                     # 配置（Viper）
│   │   ├── controller/
│   │   │   ├── platform/               # ★ 平台系统接口
│   │   │   └── shop/                   # ★ 店铺系统接口
│   │   ├── service/
│   │   │   ├── platform/
│   │   │   └── shop/
│   │   ├── repository/
│   │   │   ├── platform/
│   │   │   └── shop/
│   │   ├── model/
│   │   │   ├── entity/                 # 数据库模型
│   │   │   ├── dto/                    # 请求/响应结构体
│   │   │   └── enum/                   # 枚举常量
│   │   ├── middleware/
│   │   │   ├── auth.go                 # JWT 认证
│   │   │   ├── tenant.go               # 租户上下文 + RLS 事务
│   │   │   ├── datascope.go            # 数据权限（部门树）
│   │   │   ├── rbac.go                 # 权限校验
│   │   │   ├── audit.go                # 操作日志
│   │   │   ├── cors.go
│   │   │   └── ratelimit.go
│   │   ├── router/
│   │   │   ├── platform.go             # ★ 平台路由组
│   │   │   └── shop.go                 # ★ 店铺路由组
│   │   ├── permission/                 # ★ 权限自动加载
│   │   │   ├── manifest_platform.go    #   平台权限清单
│   │   │   ├── manifest_shop.go        #   店铺权限清单
│   │   │   ├── preset.go               #   按钮权限预设
│   │   │   ├── types.go                #   类型定义
│   │   │   └── sync.go                 #   自动同步逻辑
│   │   └── pkg/
│   │       ├── auth/                   # JWT
│   │       ├── tenant/                 # 租户 Scope + RLS
│   │       ├── datascope/              # 数据权限
│   │       ├── workflow/               # 状态机
│   │       ├── storage/                # MinIO
│   │       ├── snowflake/              # ID 生成
│   │       ├── logger/                 # zap
│   │       └── response/               # 统一响应
│   ├── database/
│   │   ├── migrations/                 # SQL 迁移文件
│   │   └── seeds/                      # 种子数据
│   ├── go.mod
│   └── Makefile
├── frontend/
│   ├── src/
│   │   ├── api/
│   │   │   ├── platform/
│   │   │   └── shop/
│   │   ├── views/
│   │   │   ├── platform/
│   │   │   └── shop/
│   │   ├── router/modules/
│   │   │   ├── platform/
│   │   │   └── shop/
│   │   ├── store/
│   │   ├── components/
│   │   ├── utils/
│   │   └── types/
│   └── package.json
├── docker-compose.yml
├── nginx/nginx.conf
├── docs/
│   └── plan2.md
└── README.md
```

---

## 3. 数据库设计

### 3.1 设计规范

| 规范 | 说明 |
|---|---|
| 引擎 | PostgreSQL 16 |
| 主键 | `id BIGSERIAL`（自增 bigint） |
| 外键 | `BIGINT`（PG 无 UNSIGNED） |
| 状态/类型 | `SMALLINT`（替代 TINYINT） |
| 布尔 | `BOOLEAN` |
| 金额 | `NUMERIC(12,2)` |
| 时间 | `TIMESTAMPTZ`（带时区） |
| 长文本 | `TEXT`（PG 高效处理，无需 VARCHAR 大小限制） |
| 短文本 | `VARCHAR(n)` |
| 灵活配置 | `JSONB`（GIN 索引支持高效查询） |
| 软删除 | `deleted_at TIMESTAMPTZ NULL`（索引） |
| 审计字段 | `created_at`、`created_by`、`updated_at`、`updated_by` |
| 租户字段 | 店铺级表含 `tenant_id BIGINT NOT NULL`（索引），平台级 `tenant_id = 0` |
| 字符集 | UTF-8（PG 默认） |

### 3.2 平台系统表

#### `shop` — 店铺/租户

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | 店铺 ID（即 tenant_id） |
| `shop_code` | VARCHAR(64) | UNIQUE NOT NULL | 店铺编码 |
| `shop_name` | VARCHAR(128) | NOT NULL | 店铺名称 |
| `contact_person` | VARCHAR(64) | | 联系人 |
| `contact_phone` | VARCHAR(20) | | 联系电话 |
| `address` | VARCHAR(255) | | 地址 |
| `status` | SMALLINT | DEFAULT 1 | 1=启用 2=停用 3=关闭 |
| `expire_at` | TIMESTAMPTZ | | 到期时间 |
| `remark` | VARCHAR(500) | | |
| `created_at` | TIMESTAMPTZ | NOT NULL | |
| `created_by` | BIGINT | | |
| `updated_at` | TIMESTAMPTZ | | |
| `updated_by` | BIGINT | | |
| `deleted_at` | TIMESTAMPTZ | INDEX | 软删除 |

#### `product_category` — 商品分类

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `category_code` | VARCHAR(64) | | 分类编码 |
| `category_name` | VARCHAR(128) | NOT NULL | 分类名称 |
| `price` | NUMERIC(12,2) | DEFAULT 0 | 参考价格 |
| `sort` | SMALLINT | DEFAULT 0 | |
| `status` | SMALLINT | DEFAULT 1 | 1=启用 2=停用 |
| + 审计字段 | | | |

#### `product` — 商品

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `product_code` | VARCHAR(64) | UNIQUE NOT NULL | 商品编号 |
| `product_name` | VARCHAR(128) | NOT NULL | 商品名称 |
| `category_id` | BIGINT | INDEX | FK → product_category |
| `price` | NUMERIC(12,2) | NOT NULL | 标准价格 |
| `sort` | SMALLINT | DEFAULT 0 | |
| `status` | SMALLINT | DEFAULT 1 | 1=上架 2=下架 |
| `mall_product_code` | VARCHAR(64) | | 保留字段，前端默认不展示 |
| `description` | TEXT | | |
| + 审计字段 | | | |

#### `product_workflow_node` — 商品流程模板

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `product_id` | BIGINT | NOT NULL INDEX | FK → product |
| `node_index` | SMALLINT | NOT NULL | 节点顺序（1,2,3...） |
| `node_code` | VARCHAR(32) | NOT NULL | contact/order/account/deploy/implement/complete |
| `node_name` | VARCHAR(64) | NOT NULL | 联系客户/下单/创建账号/部署/实施/完成 |
| + 审计字段 | | | |

**索引：** `UNIQUE(product_id, node_index)`

#### `finance_category` — 收支分类（三级树）

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `parent_id` | BIGINT | DEFAULT 0 | 父级 ID |
| `level` | SMALLINT | NOT NULL | 1/2/3 |
| `category_type` | SMALLINT | NOT NULL | 1=收入 2=支出 |
| `category_code` | VARCHAR(64) | | |
| `category_name` | VARCHAR(128) | NOT NULL | |
| `finance_code` | VARCHAR(64) | | 财务编号（默认空） |
| `sort` | SMALLINT | DEFAULT 0 | |
| + 审计字段 | | | |

### 3.3 共用系统管理表（平台 + 店铺）

以下表通过 `tenant_id` 区分平台（`tenant_id=0`）和各店铺（`tenant_id=N`）。

#### `sys_user`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `tenant_id` | BIGINT | NOT NULL DEFAULT 0 INDEX | 0=平台, N=店铺 |
| `dept_id` | BIGINT | INDEX | FK → sys_dept |
| `username` | VARCHAR(64) | NOT NULL | 登录账号 |
| `password` | VARCHAR(255) | NOT NULL | bcrypt 哈希 |
| `real_name` | VARCHAR(64) | NOT NULL | |
| `phone` | VARCHAR(20) | | |
| `email` | VARCHAR(128) | | |
| `avatar` | VARCHAR(500) | | MinIO 路径 |
| `status` | SMALLINT | DEFAULT 1 | 1=启用 2=停用 |
| `last_login_at` | TIMESTAMPTZ | | |
| `last_login_ip` | VARCHAR(45) | | |
| + 审计字段 | | | |

**索引：** `UNIQUE(tenant_id, username, deleted_at)`

#### `sys_role`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `tenant_id` | BIGINT | NOT NULL DEFAULT 0 INDEX | |
| `role_name` | VARCHAR(64) | NOT NULL | |
| `role_code` | VARCHAR(64) | NOT NULL | |
| `data_scope` | SMALLINT | DEFAULT 1 | **数据范围：** 1=全部 2=本部门及以下 3=仅本部门 4=仅本人 |
| `sort` | SMALLINT | DEFAULT 0 | |
| `status` | SMALLINT | DEFAULT 1 | |
| `remark` | VARCHAR(255) | | |
| + 审计字段 | | | |

**索引：** `UNIQUE(tenant_id, role_code, deleted_at)`

#### `sys_permission` — 全局权限定义表（★ 代码自动同步，无 tenant_id）

> **核心变更：** 权限树不再按租户复制，而是全局唯一定义，从代码自动同步。两套系统通过 `system_type` 区分。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `parent_id` | BIGINT | DEFAULT 0 | 父级 ID |
| `system_type` | VARCHAR(16) | NOT NULL | `platform` 或 `shop` |
| `name` | VARCHAR(64) | NOT NULL | 显示名称 |
| `type` | SMALLINT | NOT NULL | 1=目录 2=菜单 3=按钮 |
| `path` | VARCHAR(255) | | 路由路径（目录/菜单） |
| `component` | VARCHAR(255) | | Vue 组件路径（菜单） |
| `perms_code` | VARCHAR(100) | | 权限标识（按钮），如 `shop:order:create` |
| `icon` | VARCHAR(64) | | 图标 |
| `sort` | SMALLINT | DEFAULT 0 | |
| `visible` | BOOLEAN | DEFAULT TRUE | |
| `status` | SMALLINT | DEFAULT 1 | |
| `auto_synced` | BOOLEAN | DEFAULT TRUE | 从代码同步的条目不可在 UI 编辑/删除 |
| `updated_at` | TIMESTAMPTZ | | 同步时间 |

**索引：** `UNIQUE(system_type, perms_code)` WHERE perms_code != ''，`INDEX(system_type, parent_id)`

#### `sys_role_permission` — 角色权限关联（租户隔离）

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `tenant_id` | BIGINT | NOT NULL INDEX | 0=平台 |
| `role_id` | BIGINT | NOT NULL | FK → sys_role |
| `permission_id` | BIGINT | NOT NULL | FK → sys_permission |

**索引：** `UNIQUE(tenant_id, role_id, permission_id)`

#### `sys_user_role`

| 字段 | 类型 | 约束 |
|---|---|---|
| `id` | BIGSERIAL | PK |
| `user_id` | BIGINT | NOT NULL |
| `role_id` | BIGINT | NOT NULL |

**索引：** `UNIQUE(user_id, role_id)`

#### `sys_dept` — 部门/组织树

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `tenant_id` | BIGINT | NOT NULL DEFAULT 0 INDEX | |
| `parent_id` | BIGINT | DEFAULT 0 | 父部门 |
| `ancestors` | VARCHAR(500) | | 祖先链 "0,1,5" |
| `dept_name` | VARCHAR(64) | NOT NULL | |
| `sort` | SMALLINT | DEFAULT 0 | |
| `leader` | VARCHAR(64) | | 负责人 |
| `phone` | VARCHAR(20) | | |
| `status` | SMALLINT | DEFAULT 1 | |
| + 审计字段 | | | |

#### `sys_dept_closure` — 部门闭包表

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `tenant_id` | BIGINT | NOT NULL INDEX | |
| `ancestor_id` | BIGINT | NOT NULL INDEX | 祖先部门 |
| `descendant_id` | BIGINT | NOT NULL INDEX | 后代部门 |
| `depth` | SMALLINT | NOT NULL DEFAULT 0 | 0=自身, 1=子级... |

**索引：** `UNIQUE(ancestor_id, descendant_id)`

#### `sys_operation_log` — 操作日志

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | BIGSERIAL PK | |
| `tenant_id` | BIGINT | |
| `user_id` | BIGINT | 操作人 |
| `username` | VARCHAR(64) | |
| `module` | VARCHAR(64) | 模块 |
| `action` | VARCHAR(64) | 操作 |
| `method` | VARCHAR(10) | HTTP 方法 |
| `url` | VARCHAR(500) | |
| `params` | TEXT | 请求参数 |
| `ip` | VARCHAR(45) | |
| `duration_ms` | SMALLINT | 耗时 |
| `created_at` | TIMESTAMPTZ | |

### 3.4 店铺系统表

#### `shop_product` — 店铺商品（从平台选品）

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `tenant_id` | BIGINT | NOT NULL INDEX | |
| `platform_product_id` | BIGINT | NOT NULL | FK → product |
| `product_code` | VARCHAR(64) | | 快照 |
| `product_name` | VARCHAR(128) | | 快照 |
| `platform_price` | NUMERIC(12,2) | | 平台原价快照 |
| `shop_price` | NUMERIC(12,2) | NOT NULL | 店铺售价 |
| `status` | SMALLINT | DEFAULT 1 | |
| + 审计字段 | | | |

**索引：** `UNIQUE(tenant_id, platform_product_id, deleted_at)`

#### `shop_customer` — 客户

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `tenant_id` | BIGINT | NOT NULL INDEX | |
| `customer_name` | VARCHAR(128) | NOT NULL | |
| `customer_type` | SMALLINT | NOT NULL | 1=个人 2=企业 |
| `contact_person` | VARCHAR(64) | | |
| `contact_phone` | VARCHAR(20) | | |
| `contact_email` | VARCHAR(128) | | |
| `address` | VARCHAR(255) | | |
| `remark` | VARCHAR(500) | | |
| `status` | SMALLINT | DEFAULT 1 | |
| + 审计字段 | | | |

#### `order_group` — 订单主表

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `tenant_id` | BIGINT | NOT NULL INDEX | |
| `order_no` | VARCHAR(64) | UNIQUE NOT NULL | 订单号 `ORD20260608143025001` |
| `customer_id` | BIGINT | NOT NULL | |
| `customer_name` | VARCHAR(128) | | 快照 |
| `total_amount` | NUMERIC(12,2) | | 总金额 |
| `order_status` | SMALLINT | DEFAULT 1 | 1=待处理 2=进行中 3=已完成 4=已取消 |
| `remark` | VARCHAR(500) | | |
| + 审计字段 | | | |

#### `order_item` — 订单明细

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `tenant_id` | BIGINT | NOT NULL INDEX | |
| `order_group_id` | BIGINT | NOT NULL INDEX | FK → order_group |
| `shop_product_id` | BIGINT | NOT NULL | |
| `product_name` | VARCHAR(128) | | 快照 |
| `quantity` | SMALLINT | DEFAULT 1 | |
| `unit_price` | NUMERIC(12,2) | NOT NULL | 下单时快照 |
| `total_price` | NUMERIC(12,2) | | |
| `current_node_index` | SMALLINT | DEFAULT 0 | 当前流程节点 |
| `item_status` | SMALLINT | DEFAULT 1 | 1=待处理 2=进行中 3=已完成 4=已取消 |
| + 审计字段 | | | |

#### `order_workflow_log` — 流程跟进记录

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `tenant_id` | BIGINT | NOT NULL INDEX | |
| `order_item_id` | BIGINT | NOT NULL INDEX | |
| `node_index` | SMALLINT | NOT NULL | |
| `node_code` | VARCHAR(32) | | |
| `node_name` | VARCHAR(64) | | |
| `notes` | TEXT | | 服务备注 |
| `operator_id` | BIGINT | | 操作人 ID |
| `operator_name` | VARCHAR(64) | | 操作人姓名 |
| `operated_at` | TIMESTAMPTZ | | 操作时间 |
| `created_at` | TIMESTAMPTZ | | |

#### `order_attachment` — 订单附件

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `tenant_id` | BIGINT | NOT NULL INDEX | |
| `order_item_id` | BIGINT | NOT NULL INDEX | |
| `workflow_log_id` | BIGINT | INDEX | FK → order_workflow_log |
| `file_name` | VARCHAR(255) | | |
| `file_path` | VARCHAR(500) | | MinIO 路径 |
| `file_size` | BIGINT | | 字节 |
| `file_type` | VARCHAR(64) | | MIME |
| + 审计字段 | | | |

#### `shop_finance_account` — 收支账户（★ JSONB 灵活配置）

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `tenant_id` | BIGINT | NOT NULL INDEX | |
| `account_name` | VARCHAR(128) | NOT NULL | 账户名称 |
| `account_type` | SMALLINT | NOT NULL | 1=对公账户 2=微信 3=支付宝 |
| `account_no` | VARCHAR(128) | | 账号 |
| `initial_balance` | NUMERIC(12,2) | DEFAULT 0 | 初始余额 |
| `config` | **JSONB** | | 按类型不同存不同结构（见下方说明） |
| `status` | SMALLINT | DEFAULT 1 | |
| + 审计字段 | | | |

**JSONB config 结构示例：**
```json
// 对公账户 (account_type=1)
{"bank_name": "招商银行", "branch": "深圳南山支行"}

// 微信 (account_type=2)
{"mch_id": "1234567890", "appid": "wx1234", "api_key": "xxx"}

// 支付宝 (account_type=3)
{"app_id": "2021xxx", "merchant_id": "2088xxx", "private_key_path": "/keys/alipay.pem"}
```

**索引：** `INDEX(tenant_id, account_type)`

#### `shop_finance_category` — 店铺收支分类（从平台同步）

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `tenant_id` | BIGINT | NOT NULL INDEX | |
| `platform_category_id` | BIGINT | NOT NULL | FK → finance_category |
| `category_code` | VARCHAR(64) | | |
| `category_name` | VARCHAR(128) | | |
| `category_type` | SMALLINT | | 1=收入 2=支出 |
| `level` | SMALLINT | | |
| `parent_id` | BIGINT | | |
| + 审计字段 | | | |

#### `finance_record` — 收支记录

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `tenant_id` | BIGINT | NOT NULL INDEX | |
| `record_no` | VARCHAR(64) | UNIQUE NOT NULL | |
| `account_id` | BIGINT | NOT NULL INDEX | FK → shop_finance_account |
| `category_id` | BIGINT | NOT NULL | FK → shop_finance_category |
| `record_type` | SMALLINT | NOT NULL | 1=收入 2=支出 |
| `amount` | NUMERIC(12,2) | NOT NULL | 申请金额 |
| `actual_amount` | NUMERIC(12,2) | | 审核时填写实际金额 |
| `order_group_id` | BIGINT | INDEX | 可选关联订单 |
| `review_status` | SMALLINT | DEFAULT 1 | 1=待审核 2=已通过 3=已驳回 |
| `review_by` | BIGINT | | 审核人 |
| `review_at` | TIMESTAMPTZ | | 审核时间 |
| `review_notes` | VARCHAR(500) | | |
| `record_date` | DATE | NOT NULL | 记账日期 |
| `remark` | VARCHAR(500) | | |
| + 审计字段 | | | |

> **不可变规则：** `review_status = 2`（已通过）后禁止编辑/删除。

#### `finance_attachment` — 财务附件

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | |
| `tenant_id` | BIGINT | NOT NULL INDEX | |
| `finance_record_id` | BIGINT | NOT NULL INDEX | |
| `file_name` | VARCHAR(255) | | |
| `file_path` | VARCHAR(500) | | |
| `file_size` | BIGINT | | |
| `file_type` | VARCHAR(64) | | |
| + 审计字段 | | | |

### 3.5 表关系总览

```
平台系统:
  shop ─────────────────────────────────────────────────────────────┐
  product_category ←── product ←── product_workflow_node           │
  finance_category (三级树)                                         │
                                                                   │
全局权限（无 tenant_id）:                                            │
  sys_permission (platform/shop 两棵树, 代码自动同步)                │
                                                                   │
共用系统表 (tenant_id 区分):                                        │
  sys_user ←── sys_user_role ──→ sys_role ←── sys_role_permission ──→ sys_permission
  sys_user ──→ sys_dept ←──→ sys_dept_closure                      │
                                                                   │
店铺系统 (tenant_id = shop.id, 启用 RLS):                           │
  shop_product (→product)     ←───────┐                            │
  shop_customer               ←──┐    │                            │
  order_group (→customer)        │    │                            │
    └── order_item (→shop_product)    │                            │
          ├── order_workflow_log      │                            │
          └── order_attachment        │                            │
  shop_finance_account (JSONB)  ←────────┐                        │
  shop_finance_category (→finance_category)                       │
  finance_record (→account, →category, →order_group)              │
    └── finance_attachment                                         │
                                                                   │
  sys_operation_log (审计) ◄───────────────────────────────────────┘
```

---

## 4. 数据权限设计

### 4.1 两层数据可见性

| 维度 | 规则 | 适用数据 |
|---|---|---|
| **部门树层级** | 上级可见下级（通过角色 `data_scope` 配置） | 客户、订单、收支记录等业务数据 |
| **店铺级共享** | 同店铺全员可见 | 选品、收支账户、收支分类 |

### 4.2 data_scope 配置

| 值 | 含义 | SQL 过滤 |
|---|---|---|
| 1 | 全部数据 | 无过滤 |
| 2 | 本部门及以下 | `created_by IN (SELECT id FROM sys_user WHERE dept_id IN (闭包表查子部门))` |
| 3 | 仅本部门 | `created_by IN (SELECT id FROM sys_user WHERE dept_id = 当前部门)` |
| 4 | 仅本人 | `created_by = 当前用户` |

### 4.3 共享数据例外（不受部门限制）

| 数据 | 表 | 理由 |
|---|---|---|
| 店铺选品 | `shop_product` | 所有销售需看到可售商品 |
| 收支账户 | `shop_finance_account` | 所有财务人员需看到 |
| 收支分类 | `shop_finance_category` | 记账时需选分类 |
| 商品分类 | `product_category` | 平台级，全局共享 |
| 系统配置 | `sys_permission`、`sys_role`、`sys_dept` | 系统管理（需管理员角色） |

### 4.4 PostgreSQL RLS 策略

所有店铺级表启用 RLS，数据库层强制隔离：

```sql
-- 对每个店铺级表执行：
ALTER TABLE shop_customer ENABLE ROW LEVEL SECURITY;
ALTER TABLE shop_product ENABLE ROW LEVEL SECURITY;
ALTER TABLE order_group ENABLE ROW LEVEL SECURITY;
ALTER TABLE order_item ENABLE ROW LEVEL SECURITY;
ALTER TABLE order_workflow_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE order_attachment ENABLE ROW LEVEL SECURITY;
ALTER TABLE shop_finance_account ENABLE ROW LEVEL SECURITY;
ALTER TABLE shop_finance_category ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_record ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_attachment ENABLE ROW LEVEL SECURITY;
ALTER TABLE sys_user ENABLE ROW LEVEL SECURITY;
ALTER TABLE sys_role ENABLE ROW LEVEL SECURITY;
ALTER TABLE sys_user_role ENABLE ROW LEVEL SECURITY;
ALTER TABLE sys_role_permission ENABLE ROW LEVEL SECURITY;
ALTER TABLE sys_dept ENABLE ROW LEVEL SECURITY;
ALTER TABLE sys_dept_closure ENABLE ROW LEVEL SECURITY;

-- 策略模板（每表一条，替换表名即可）：
CREATE POLICY tenant_isolation ON shop_customer
    FOR ALL
    USING (tenant_id = current_setting('app.tenant_id', true)::bigint)
    WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::bigint);

-- 平台用户（tenant_id=0）需绕过 RLS，设为 BYPASSRLS 角色
ALTER ROLE platform_admin BYPASSRLS;
```

**平台级表（shop、product、product_category 等）不启用 RLS。**

---

## 5. 权限体系设计（自动加载）

### 5.1 核心思想

> **权限定义 = 代码产物，不是数据库产物。**

开发者新增功能时，只需在 Go 代码中添加权限声明。应用启动时自动同步到 `sys_permission` 表。管理后台只能**查看权限树 + 分配给角色**，不能新增/编辑/删除权限定义。

### 5.2 权限清单声明

```go
// internal/permission/types.go
type (
    Manifest struct {
        SystemType string
        Nodes      []Node
    }

    Node struct {
        Name      string
        Type      smallint // 1=目录 2=菜单 3=按钮
        Path      string   // 路由路径
        Component string   // Vue 组件路径
        Icon      string
        Sort      int
        Children  []Node   // 子目录/子菜单
        Buttons   []Button // 菜单下的按钮权限
    }

    Button struct {
        Name string // 按钮名称
        Code string // 权限标识
    }
)
```

### 5.3 按钮权限预设

```go
// internal/permission/preset.go

// 标准五件套：查询、详情、新增、编辑、删除
func CRUD(prefix string) []Button {
    return []Button{
        {Name: "查询", Code: prefix + ":list"},
        {Name: "详情", Code: prefix + ":detail"},
        {Name: "新增", Code: prefix + ":create"},
        {Name: "编辑", Code: prefix + ":update"},
        {Name: "删除", Code: prefix + ":delete"},
    }
}

// 可组合的额外按钮（Code 前缀自动拼接）
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

// 组合使用示例
func CRUDWith(prefix string, extras ...func(string) Button) []Button {
    btns := CRUD(prefix)
    for _, e := range extras {
        btns = append(btns, e(prefix))
    }
    return btns
}
```

### 5.4 平台系统权限清单

```go
// internal/permission/manifest_platform.go
var PlatformManifest = Manifest{
    SystemType: "platform",
    Nodes: []Node{
        {Name: "系统管理", Type: 1, Icon: "setting", Sort: 1, Children: []Node{
            {Name: "用户管理", Type: 2, Path: "/system/user",
                Component: "platform/system/user/index", Sort: 1,
                Buttons: CRUDWith("platform:user", BAssign, BReset)},
            {Name: "角色管理", Type: 2, Path: "/system/role",
                Component: "platform/system/role/index", Sort: 2,
                Buttons: CRUDWith("platform:role", BAssign)},
            {Name: "部门管理", Type: 2, Path: "/system/dept",
                Component: "platform/system/dept/index", Sort: 3,
                Buttons: CRUD("platform:dept")},
        }},
        {Name: "店铺管理", Type: 1, Icon: "shop", Sort: 2, Children: []Node{
            {Name: "店铺列表", Type: 2, Path: "/shop/list",
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
        {Name: "商品管理", Type: 1, Icon: "goods", Sort: 3, Children: []Node{
            {Name: "商品分类", Type: 2, Path: "/product/category",
                Component: "platform/product/category/index", Sort: 1,
                Buttons: CRUD("platform:product:category")},
            {Name: "商品列表", Type: 2, Path: "/product/list",
                Component: "platform/product/list/index", Sort: 2,
                Buttons: []Button{
                    {Name: "查询", Code: "platform:product:list"},
                    {Name: "详情", Code: "platform:product:detail"},
                    {Name: "新增", Code: "platform:product:create"},
                    {Name: "编辑", Code: "platform:product:update"},
                    {Name: "删除", Code: "platform:product:delete"},
                    {Name: "上架/下架", Code: "platform:product:status"},
                }},
        }},
        {Name: "财务管理", Type: 1, Icon: "finance", Sort: 4, Children: []Node{
            {Name: "收支分类", Type: 2, Path: "/finance/category",
                Component: "platform/finance/category/index", Sort: 1,
                Buttons: CRUD("platform:finance:category")},
            {Name: "财务报表", Type: 2, Path: "/finance/report",
                Component: "platform/finance/report/index", Sort: 2,
                Buttons: []Button{
                    {Name: "查询", Code: "platform:finance:report:list"},
                    {Name: "导出", Code: "platform:finance:report:export"},
                }},
        }},
    },
}
```

### 5.5 店铺系统权限清单

```go
// internal/permission/manifest_shop.go
var ShopManifest = Manifest{
    SystemType: "shop",
    Nodes: []Node{
        {Name: "系统管理", Type: 1, Icon: "setting", Sort: 1, Children: []Node{
            {Name: "用户管理", Type: 2, Path: "/system/user",
                Component: "shop/system/user/index", Sort: 1,
                Buttons: CRUDWith("shop:user", BAssign, BReset)},
            {Name: "角色管理", Type: 2, Path: "/system/role",
                Component: "shop/system/role/index", Sort: 2,
                Buttons: CRUDWith("shop:role", BAssign)},
            {Name: "部门管理", Type: 2, Path: "/system/dept",
                Component: "shop/system/dept/index", Sort: 3,
                Buttons: CRUD("shop:dept")},
        }},
        {Name: "商品管理", Type: 1, Icon: "goods", Sort: 2, Children: []Node{
            {Name: "选品管理", Type: 2, Path: "/product",
                Component: "shop/product/index", Sort: 1,
                Buttons: []Button{
                    {Name: "查询", Code: "shop:product:list"},
                    {Name: "选品", Code: "shop:product:select"},
                    {Name: "改价", Code: "shop:product:price"},
                    {Name: "上架/下架", Code: "shop:product:status"},
                    {Name: "取消选品", Code: "shop:product:delete"},
                }},
        }},
        {Name: "客户管理", Type: 1, Icon: "user", Sort: 3, Children: []Node{
            {Name: "客户列表", Type: 2, Path: "/customer",
                Component: "shop/customer/index", Sort: 1,
                Buttons: CRUDWith("shop:customer", BExport)},
        }},
        {Name: "订单管理", Type: 1, Icon: "order", Sort: 4, Children: []Node{
            {Name: "订单列表", Type: 2, Path: "/order",
                Component: "shop/order/index", Sort: 1,
                Buttons: []Button{
                    {Name: "查询", Code: "shop:order:list"},
                    {Name: "详情", Code: "shop:order:detail"},
                    {Name: "创建", Code: "shop:order:create"},
                    {Name: "取消", Code: "shop:order:cancel"},
                    {Name: "流程推进", Code: "shop:order:advance"},
                    {Name: "上传附件", Code: "shop:order:upload"},
                    {Name: "导出", Code: "shop:order:export"},
                }},
        }},
        {Name: "财务管理", Type: 1, Icon: "finance", Sort: 5, Children: []Node{
            {Name: "收支账户", Type: 2, Path: "/finance/account",
                Component: "shop/finance/account/index", Sort: 1,
                Buttons: CRUD("shop:finance:account")},
            {Name: "收支分类", Type: 2, Path: "/finance/category",
                Component: "shop/finance/category/index", Sort: 2,
                Buttons: []Button{
                    {Name: "查询", Code: "shop:finance:category:list"},
                    {Name: "同步", Code: "shop:finance:category:sync"},
                    {Name: "取消同步", Code: "shop:finance:category:delete"},
                }},
            {Name: "收支记录", Type: 2, Path: "/finance/record",
                Component: "shop/finance/record/index", Sort: 3,
                Buttons: []Button{
                    {Name: "查询", Code: "shop:finance:record:list"},
                    {Name: "详情", Code: "shop:finance:record:detail"},
                    {Name: "新增", Code: "shop:finance:record:create"},
                    {Name: "编辑", Code: "shop:finance:record:update"},
                    {Name: "删除", Code: "shop:finance:record:delete"},
                    {Name: "审核", Code: "shop:finance:record:audit"},
                    {Name: "上传附件", Code: "shop:finance:record:upload"},
                    {Name: "导出", Code: "shop:finance:record:export"},
                }},
            {Name: "财务报表", Type: 2, Path: "/finance/report",
                Component: "shop/finance/report/index", Sort: 4,
                Buttons: []Button{
                    {Name: "查询", Code: "shop:finance:report:list"},
                    {Name: "导出", Code: "shop:finance:report:export"},
                }},
        }},
    },
}
```

### 5.6 自动同步机制

```go
// internal/permission/sync.go

// SyncPermissions 启动时将代码中的权限清单同步到 sys_permission 表
// 逻辑：按 perms_code 做 upsert，不删除数据库中多余的条目（避免误删角色已分配的权限）
func SyncPermissions(db *gorm.DB, manifests ...Manifest) error {
    for _, m := range manifests {
        flatNodes := flatten(m.Nodes, 0, m.SystemType) // 展平树为列表
        for _, node := range flatNodes {
            var existing SysPermission
            result := db.Where("system_type = ? AND perms_code = ?",
                node.SystemType, node.PermsCode).First(&existing)

            if result.Error == gorm.ErrRecordNotFound {
                // 新增
                db.Create(&node)
            } else {
                // 更新（名称、路径、排序等可能变化）
                db.Model(&existing).Updates(map[string]interface{}{
                    "name":       node.Name,
                    "parent_id":  node.ParentID,
                    "path":       node.Path,
                    "component":  node.Component,
                    "icon":       node.Icon,
                    "sort":       node.Sort,
                    "updated_at": time.Now(),
                })
            }
        }
    }
    return nil
}

// main.go
func main() {
    db := initDB()
    permission.SyncPermissions(db,
        permission.PlatformManifest,
        permission.ShopManifest,
    )
    // ... start server
}
```

**同步策略：**
- **新增**：代码有、数据库无 → INSERT
- **更新**：代码有、数据库有 → UPDATE 名称/路径/排序等
- **不删除**：数据库有、代码无 → 不动（可能是正在使用的权限）
- **手动标记**：管理员可在后台将废弃权限的 `status` 设为 2（停用）

### 5.7 前端加载流程

```
用户登录
  ↓
GET /auth/permissions → 返回当前系统（platform/shop）的完整权限树
  ↓                         ↓
GET /auth/userinfo    → 返回用户已分配的权限码列表 ["shop:order:list", "shop:order:create", ...]
  ↓
前端：
  - 权限树 → 构建动态路由（只展示有权限的菜单）
  - 权限码 → 控制按钮显示（v-permission="'shop:order:create'"）
```

### 5.8 角色管理 UI

角色管理页面中，权限分配区域展示**从代码同步的权限树**（只读树形结构），管理员勾选/取消勾选哪些权限分配给该角色。保存后写入 `sys_role_permission`。

**新增功能流程（开发者视角）：**
1. 在 `manifest_shop.go` 中添加新的菜单/按钮声明
2. 重启应用（或自动热加载）
3. 新权限自动出现在所有店铺的角色管理 UI 中
4. 管理员勾选后，对应角色的用户即可看到新菜单/按钮

---

## 6. 功能模块详细设计

### 6.1 认证与授权

| 接口 | 方法 | 路径 | 说明 |
|---|---|---|---|
| 平台登录 | POST | `/api/v1/platform/auth/login` | |
| 店铺登录 | POST | `/api/v1/shop/auth/login` | 需指定店铺编码 |
| 登出 | POST | `/api/v1/*/auth/logout` | Token 加入 Redis 黑名单 |
| 用户信息+权限码 | GET | `/api/v1/*/auth/userinfo` | 返回用户、角色、权限码列表 |
| 权限树 | GET | `/api/v1/*/auth/permissions` | 返回当前系统的完整权限树（已分配的部分标记） |

**JWT Payload：**
```json
{
  "user_id": 123,
  "tenant_id": 5,
  "dept_id": 3,
  "username": "admin",
  "data_scope": 2,
  "exp": 1717987200
}
```

**业务规则：** 密码 bcrypt 加密；Access Token 2h，Refresh Token 7d；登录失败 5 次锁定 30 分钟；店铺登录校验店铺状态和到期时间。

### 6.2 平台系统 — 系统管理

#### 用户管理

| 接口 | 方法 | 路径 | 权限码 |
|---|---|---|---|
| 用户列表 | GET | `/api/v1/platform/sys/users` | `platform:user:list` |
| 用户详情 | GET | `/api/v1/platform/sys/users/:id` | `platform:user:detail` |
| 创建用户 | POST | `/api/v1/platform/sys/users` | `platform:user:create` |
| 编辑用户 | PUT | `/api/v1/platform/sys/users/:id` | `platform:user:update` |
| 删除用户 | DELETE | `/api/v1/platform/sys/users/:id` | `platform:user:delete` |
| 分配角色 | PUT | `/api/v1/platform/sys/users/:id/roles` | `platform:user:assign` |
| 重置密码 | PUT | `/api/v1/platform/sys/users/:id/password` | `platform:user:reset` |

> **后续所有接口表格统一增加「权限码」列，省略不写。每个接口都经过 `PermissionMiddleware("对应权限码")` 校验。**

#### 角色管理

| 接口 | 方法 | 路径 |
|---|---|---|
| 角色列表 | GET | `/api/v1/platform/sys/roles` |
| 创建角色 | POST | `/api/v1/platform/sys/roles` |
| 编辑角色 | PUT | `/api/v1/platform/sys/roles/:id` |
| 删除角色 | DELETE | `/api/v1/platform/sys/roles/:id` |
| 分配权限 | PUT | `/api/v1/platform/sys/roles/:id/permissions` |

**分配权限接口：**
```json
// PUT /api/v1/platform/sys/roles/:id/permissions
{ "permission_ids": [1, 2, 3, 10, 11, 12, ...] }
```

#### 部门管理

| 接口 | 方法 | 路径 |
|---|---|---|
| 部门树 | GET | `/api/v1/platform/sys/depts` |
| 创建部门 | POST | `/api/v1/platform/sys/depts` |
| 编辑部门 | PUT | `/api/v1/platform/sys/depts/:id` |
| 删除部门 | DELETE | `/api/v1/platform/sys/depts/:id` |

**业务规则：** 部门增删改时自动维护闭包表（`sys_dept_closure`）和祖先链（`ancestors`）。

#### 权限树查看（只读）

| 接口 | 方法 | 路径 | 说明 |
|---|---|---|---|
| 查看权限树 | GET | `/api/v1/platform/sys/permissions` | 返回 system_type=platform 的权限树（只读） |

> **注意：** 权限树由代码自动同步，这里只有查看接口，没有 CRUD。新增功能在代码中声明即可。

### 6.3 平台系统 — 店铺管理

| 接口 | 方法 | 路径 |
|---|---|---|
| 店铺列表 | GET | `/api/v1/platform/shops` |
| 店铺详情 | GET | `/api/v1/platform/shops/:id` |
| 创建店铺 | POST | `/api/v1/platform/shops` |
| 编辑店铺 | PUT | `/api/v1/platform/shops/:id` |
| 删除店铺 | DELETE | `/api/v1/platform/shops/:id` |
| 停用/启用 | PUT | `/api/v1/platform/shops/:id/status` |
| 重置管理员密码 | PUT | `/api/v1/platform/shops/:id/admin-password` |

**创建店铺时自动初始化：**
1. 创建店铺管理员账户（`sys_user` with `tenant_id = shop.id`）
2. 创建店铺管理员角色（`sys_role` with `tenant_id = shop.id`，`data_scope = 1`）
3. 将 `sys_permission` 中 `system_type = 'shop'` 的**全部权限**分配给管理员角色
4. 创建默认部门（根部门）

### 6.4 平台系统 — 商品管理

#### 商品分类

| 接口 | 方法 | 路径 |
|---|---|---|
| 分类列表 | GET | `/api/v1/platform/product-categories` |
| 创建分类 | POST | `/api/v1/platform/product-categories` |
| 编辑分类 | PUT | `/api/v1/platform/product-categories/:id` |
| 删除分类 | DELETE | `/api/v1/platform/product-categories/:id` |

#### 商品管理

| 接口 | 方法 | 路径 |
|---|---|---|
| 商品列表 | GET | `/api/v1/platform/products` |
| 商品详情 | GET | `/api/v1/platform/products/:id` |
| 创建商品 | POST | `/api/v1/platform/products` |
| 编辑商品 | PUT | `/api/v1/platform/products/:id` |
| 删除商品 | DELETE | `/api/v1/platform/products/:id` |
| 上架/下架 | PUT | `/api/v1/platform/products/:id/status` |
| 流程模板查看 | GET | `/api/v1/platform/products/:id/workflow` |
| 保存流程模板 | PUT | `/api/v1/platform/products/:id/workflow` |

**流程模板配置规则：** 每个商品可配置一组顺序节点（默认：联系客户→下单→创建账号→部署→实施→完成）。可增减节点、自定义名称、拖拽排序。`mall_product_code` 保留字段前端不展示。

### 6.5 平台系统 — 财务管理

#### 收支分类

| 接口 | 方法 | 路径 |
|---|---|---|
| 分类树 | GET | `/api/v1/platform/finance-categories` |
| 创建分类 | POST | `/api/v1/platform/finance-categories` |
| 编辑分类 | PUT | `/api/v1/platform/finance-categories/:id` |
| 删除分类 | DELETE | `/api/v1/platform/finance-categories/:id` |

**规则：** 最多三级，支持收入/支出。删除前校验：有关联子分类或已被店铺同步的不可删除。

#### 财务报表

| 接口 | 方法 | 路径 |
|---|---|---|
| 收支汇总 | GET | `/api/v1/platform/finance/reports/summary` |
| 利润报表 | GET | `/api/v1/platform/finance/reports/profit` |
| 导出 | GET | `/api/v1/platform/finance/reports/export` |

### 6.6 店铺系统 — 系统管理

与平台系统管理功能完全一致，区别：路由前缀 `/api/v1/shop/sys/*`，数据隔离在当前 `tenant_id`。权限树查看接口返回 `system_type = 'shop'` 的权限。

### 6.7 店铺系统 — 商品管理（选品）

| 接口 | 方法 | 路径 |
|---|---|---|
| 平台商品库（可选） | GET | `/api/v1/shop/products/platform` |
| 已选商品列表 | GET | `/api/v1/shop/products` |
| 选品（加入） | POST | `/api/v1/shop/products` |
| 修改售价 | PUT | `/api/v1/shop/products/:id/price` |
| 上架/下架 | PUT | `/api/v1/shop/products/:id/status` |
| 取消选品 | DELETE | `/api/v1/shop/products/:id` |

**规则：** 只能从平台上架商品选；每商品每店铺只能选一次；售价默认=平台价；流程复用平台配置不可改；**数据权限：共享，全员可见。**

### 6.8 店铺系统 — 客户管理

| 接口 | 方法 | 路径 |
|---|---|---|
| 客户列表 | GET | `/api/v1/shop/customers` |
| 客户详情 | GET | `/api/v1/shop/customers/:id` |
| 创建客户 | POST | `/api/v1/shop/customers` |
| 编辑客户 | PUT | `/api/v1/shop/customers/:id` |
| 删除客户 | DELETE | `/api/v1/shop/customers/:id` |
| 导出 | GET | `/api/v1/shop/customers/export` |
| 客户订单记录 | GET | `/api/v1/shop/customers/:id/orders` |

**数据权限：受部门树限制。**

### 6.9 店铺系统 — 订单管理

#### 订单接口

| 接口 | 方法 | 路径 |
|---|---|---|
| 订单列表 | GET | `/api/v1/shop/orders` |
| 订单详情 | GET | `/api/v1/shop/orders/:id` |
| 创建订单 | POST | `/api/v1/shop/orders` |
| 取消订单 | PUT | `/api/v1/shop/orders/:id/cancel` |
| 导出 | GET | `/api/v1/shop/orders/export` |

#### 流程跟进

| 接口 | 方法 | 路径 |
|---|---|---|
| 流程节点列表 | GET | `/api/v1/shop/orders/:id/items/:itemId/workflow` |
| 跟进记录列表 | GET | `/api/v1/shop/orders/:id/items/:itemId/workflow/logs` |
| 推进流程 | POST | `/api/v1/shop/orders/:id/items/:itemId/workflow/advance` |
| 上传附件 | POST | `/api/v1/shop/orders/:id/items/:itemId/attachments` |
| 下载附件 | GET | `/api/v1/shop/orders/:id/items/:itemId/attachments/:attId` |

#### 创建订单逻辑

```
POST /api/v1/shop/orders
{
  "customer_id": 123,
  "remark": "大客户订单",
  "items": [
    {"shop_product_id": 1, "quantity": 1},
    {"shop_product_id": 3, "quantity": 2}
  ]
}

处理（单事务）：
1. 生成订单号 ORD + yyyyMMddHHmmss + 3位序号
2. 创建 order_group
3. 遍历 items：
   a. 从 shop_product 快照售价 → unit_price
   b. 从 product_workflow_node 获取流程模板
   c. 创建 order_item (current_node_index=0)
4. 更新 order_group.total_amount
5. 全部成功 → commit；任一失败 → rollback
```

#### 流程推进逻辑

```
POST /api/v1/shop/orders/100/items/201/workflow/advance
{ "notes": "已联系客户，确认需求" }

处理：
1. 获取 order_item 当前节点
2. 从 product_workflow_node 获取下一节点
3. 创建 order_workflow_log（节点信息 + 备注 + 操作人 + 时间）
4. 更新 order_item.current_node_index++
5. 若到达最后节点 → item_status = 3（已完成）
6. 检查同 order_group 下所有 item → 全部完成则 order_status = 3
```

**数据权限：受部门树限制。**

### 6.10 店铺系统 — 财务管理

#### 收支账户

| 接口 | 方法 | 路径 |
|---|---|---|
| 账户列表 | GET | `/api/v1/shop/finance/accounts` |
| 创建账户 | POST | `/api/v1/shop/finance/accounts` |
| 编辑账户 | PUT | `/api/v1/shop/finance/accounts/:id` |
| 删除账户 | DELETE | `/api/v1/shop/finance/accounts/:id` |

**创建请求示例（JSONB 灵活配置）：**
```json
// 对公账户
{ "account_name": "招行基本户", "account_type": 1, "account_no": "6225xxxx",
  "config": {"bank_name": "招商银行", "branch": "深圳南山支行"} }

// 微信
{ "account_name": "微信收款", "account_type": 2, "account_no": "商户号xxx",
  "config": {"mch_id": "1234567890", "appid": "wx1234", "api_key": "xxx"} }

// 支付宝
{ "account_name": "支付宝收款", "account_type": 3, "account_no": "2088xxx",
  "config": {"app_id": "2021xxx", "merchant_id": "2088xxx"} }
```

**数据权限：共享，全员可见。**

#### 收支分类（从平台同步）

| 接口 | 方法 | 路径 |
|---|---|---|
| 已同步分类 | GET | `/api/v1/shop/finance/categories` |
| 可选分类（平台） | GET | `/api/v1/shop/finance/categories/available` |
| 同步分类 | POST | `/api/v1/shop/finance/categories/sync` |
| 取消同步 | DELETE | `/api/v1/shop/finance/categories/:id` |

**规则：** 只能从平台选择，不能新增编辑。已被收支记录引用的不可取消。**数据权限：共享。**

#### 收支记录

| 接口 | 方法 | 路径 |
|---|---|---|
| 记录列表 | GET | `/api/v1/shop/finance/records` |
| 记录详情 | GET | `/api/v1/shop/finance/records/:id` |
| 创建记录 | POST | `/api/v1/shop/finance/records` |
| 编辑记录 | PUT | `/api/v1/shop/finance/records/:id` |
| 删除记录 | DELETE | `/api/v1/shop/finance/records/:id` |
| 审核 | POST | `/api/v1/shop/finance/records/:id/review` |
| 上传附件 | POST | `/api/v1/shop/finance/records/:id/attachments` |
| 导出 | GET | `/api/v1/shop/finance/records/export` |

**审核请求：**
```json
{ "action": "approve",        // approve / reject
  "actual_amount": 4980.00,   // 审核通过时填真实金额
  "notes": "扣除手续费",
  "attachment_ids": [1, 2] }
```

**规则：** 待审核可编辑/删除；**已通过不可编辑/删除**；已驳回可编辑后重提；审核人和创建人不能同一人；**数据权限：受部门树限制。**

#### 财务报表

| 接口 | 方法 | 路径 |
|---|---|---|
| 收支汇总 | GET | `/api/v1/shop/finance/reports/summary` |
| 利润报表 | GET | `/api/v1/shop/finance/reports/profit` |
| 账户余额 | GET | `/api/v1/shop/finance/reports/balance` |
| 导出 | GET | `/api/v1/shop/finance/reports/export` |

**报表维度：** 时间范围（日/周/月/季/年/自定义）、按分类汇总、按账户汇总、利润 = 收入 - 支出。支持导出 Excel。

---

## 7. 核心技术方案

### 7.1 多租户 RLS + GORM

```go
// middleware/tenant.go

// TenantRLSMiddleware 为每个店铺请求设置 RLS 上下文
func TenantRLSMiddleware(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        claims := c.MustGet("claims").(*auth.JWTClaims)
        tenantID := claims.TenantID

        c.Set("tenant_id", tenantID)
        c.Set("user_id", claims.UserID)
        c.Set("dept_id", claims.DeptID)
        c.Set("data_scope", claims.DataScope)

        if tenantID > 0 {
            // 店铺请求：在事务中设置 RLS 上下文
            tx := db.Begin()
            tx.Exec("SELECT set_config('app.tenant_id', $1, true)",
                strconv.FormatUint(tenantID, 10))
            c.Set("db", tx) // 后续 handler 使用这个 tx

            defer func() {
                // 中间件结束时 commit（或 Abort 时 rollback）
                if c.IsAborted() {
                    tx.Rollback()
                } else {
                    tx.Commit()
                }
            }()
        } else {
            c.Set("db", db) // 平台请求：不设 RLS
        }
        c.Next()
    }
}

// GORM 租户 Scope（应用层双重保障）
func TenantScope(c *gin.Context) func(*gorm.DB) *gorm.DB {
    tenantID := c.GetUint64("tenant_id")
    return func(db *gorm.DB) *gorm.DB {
        if tenantID > 0 {
            return db.Where("tenant_id = ?", tenantID)
        }
        return db.Where("tenant_id = 0")
    }
}
```

### 7.2 RBAC 权限校验

```go
// middleware/rbac.go

func PermissionMiddleware(requiredPerm string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetUint64("user_id")
        tenantID := c.GetUint64("tenant_id")

        perms := getUserPermissionCodes(userID, tenantID) // Redis 缓存
        if !contains(perms, requiredPerm) {
            response.Forbidden(c, "无权限")
            c.Abort()
            return
        }
        c.Next()
    }
}

// 路由注册示例
shopRoute.GET("/customers", PermissionMiddleware("shop:customer:list"), ctrl.List)
shopRoute.POST("/customers", PermissionMiddleware("shop:customer:create"), ctrl.Create)
shopRoute.PUT("/customers/:id", PermissionMiddleware("shop:customer:update"), ctrl.Update)
shopRoute.DELETE("/customers/:id", PermissionMiddleware("shop:customer:delete"), ctrl.Delete)
shopRoute.GET("/customers/export", PermissionMiddleware("shop:customer:export"), ctrl.Export)
```

**前端按钮级权限：**
```vue
<el-button v-permission="'shop:order:create'">创建订单</el-button>
<el-button v-permission="'shop:finance:record:audit'">审核</el-button>
<el-button v-permission="'shop:customer:export'">导出</el-button>
```

### 7.3 工作流状态机（looplab/fsm）

```go
import "github.com/looplab/fsm"

func NewOrderWorkflow(nodes []ProductWorkflowNode) *fsm.FSM {
    events := fsm.Events{}
    for i := 0; i < len(nodes)-1; i++ {
        events = append(events, fsm.EventConfig{
            Name: "advance",
            Src:  []string{nodes[i].NodeCode},
            Dst:  nodes[i+1].NodeCode,
        })
    }
    return fsm.NewFSM(nodes[0].NodeCode, events, fsm.Callbacks{})
}

// 推进流程
func (s *OrderService) AdvanceWorkflow(itemID uint64, notes string) error {
    item := s.repo.GetOrderItem(itemID)
    nodes := s.repo.GetWorkflowNodes(item.ProductID)
    fsm := NewOrderWorkflow(nodes)
    fsm.SetState(item.CurrentNodeCode())

    if err := fsm.Event(context.Background(), "advance"); err != nil {
        return errors.New("无法推进到下一节点")
    }

    // 记录日志
    s.repo.CreateWorkflowLog(&OrderWorkflowLog{
        OrderItemID: itemID, NodeCode: fsm.Current(), Notes: notes,
        OperatorID: item.UpdatedBy, OperatedAt: time.Now(),
    })

    // 更新状态
    item.CurrentNodeIndex++
    if item.CurrentNodeIndex >= len(nodes) {
        item.ItemStatus = enum.ItemCompleted
    }
    return s.repo.UpdateOrderItem(item)
}
```

### 7.4 订单拆分逻辑

```go
func (s *OrderService) CreateOrder(req *CreateOrderReq) (*OrderGroup, error) {
    orderNo := generateOrderNo()

    return s.repo.Transaction(func(tx *gorm.DB) error {
        group := &OrderGroup{OrderNo: orderNo, CustomerID: req.CustomerID, ...}
        if err := tx.Create(group).Error; err != nil { return err }

        var totalAmount decimal.Decimal
        for _, item := range req.Items {
            product := s.repo.GetShopProduct(item.ShopProductID)
            totalPrice := product.ShopPrice.Mul(decimal.New(int64(item.Quantity), 0))
            totalAmount = totalAmount.Add(totalPrice)

            orderItem := &OrderItem{
                OrderGroupID: group.ID, ShopProductID: item.ShopProductID,
                ProductName: product.ProductName, Quantity: item.Quantity,
                UnitPrice: product.ShopPrice, TotalPrice: totalPrice,
                CurrentNodeIndex: 0, ItemStatus: 1,
                TenantID: req.TenantID, CreatedBy: req.UserID,
            }
            if err := tx.Create(orderItem).Error; err != nil { return err }
        }
        return tx.Model(group).Update("total_amount", totalAmount).Error
    })
}
```

### 7.5 文件上传服务

```go
func (s *StorageService) Upload(file *multipart.FileHeader, tenantID uint64) (*FileInfo, error) {
    bucket := fmt.Sprintf("tenant-%d", tenantID)
    s.ensureBucket(bucket)

    ext := filepath.Ext(file.Filename)
    objectName := fmt.Sprintf("attachments/%s%s", snowflake.Generate(), ext)

    src, _ := file.Open()
    defer src.Close()
    _, err := s.minio.PutObject(ctx, bucket, objectName, src, file.Size, minio.PutObjectOptions{})
    return &FileInfo{FileName: file.Filename, FilePath: bucket + "/" + objectName, ...}, err
}

func (s *StorageService) GetDownloadURL(filePath string) (string, error) {
    parts := strings.SplitN(filePath, "/", 2)
    url, _ := s.minio.PresignedGetObject(ctx, parts[0], parts[1], time.Hour, nil)
    return url.String(), nil
}
```

### 7.6 操作审计中间件

```go
func AuditMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        log := &OperationLog{
            TenantID: c.GetUint64("tenant_id"), UserID: c.GetUint64("user_id"),
            Method: c.Request.Method, URL: c.Request.URL.Path,
            IP: c.ClientIP(), DurationMs: int(time.Since(start).Milliseconds()),
        }
        go auditRepo.Create(log) // 异步写入
    }
}
```

---

## 8. 安全设计

### 8.1 四层数据隔离保障

| 保障层 | 措施 | 说明 |
|---|---|---|
| ❶ **数据库层** | **PostgreSQL RLS** | 每个店铺级表启用 RLS 策略，数据库层强制 `tenant_id` 过滤。**即使应用代码遗漏 WHERE，也不会泄露数据** |
| ❷ **ORM 层** | GORM Scopes | 每个查询自动追加 `WHERE tenant_id = ?`，正常路径的双重保障 |
| ❸ **中间件层** | JWT → RLS 事务 | 每个请求在事务中 `set_config('app.tenant_id')`，RLS 上下文自动生效 |
| ❹ **校验层** | 请求体验证 | 请求体中的 tenant_id 必须与 JWT 一致，防止越权 |

### 8.2 接口安全

| 措施 | 说明 |
|---|---|
| JWT 认证 | 所有接口必须携带有效 Token |
| 权限码校验 | 每个接口绑定权限码（如 `shop:order:create`），中间件自动校验 |
| 限流 | Redis 令牌桶，100 次/分钟/用户 |
| CORS | 白名单域名 |
| 参数校验 | Gin Binding + validator |
| SQL 注入 | GORM 参数化查询 |
| XSS | 前端输出转义 |

### 8.3 数据安全

| 措施 | 说明 |
|---|---|
| 密码 | bcrypt（cost=10） |
| 敏感字段 | 账户号等脱敏展示 |
| 软删除 | 所有业务数据软删除 |
| 操作审计 | 写操作记录到 `sys_operation_log` |
| 财务不可变 | 审核通过后禁止编辑/删除 |
| Token | Access 2h，Refresh 7d，登出入黑名单 |

### 8.4 文件上传安全

| 措施 | 说明 |
|---|---|
| 类型限制 | 白名单：jpg/png/pdf/doc/xlsx/zip |
| 大小限制 | 单文件 10MB |
| 文件名 | Snowflake ID 重命名 |
| 存储隔离 | 每租户独立 Bucket |
| 访问控制 | 预签名 URL，有效期 1 小时 |

---

## 9. 开发阶段与里程碑

### 阶段一：基础设施 + 认证 + RBAC + 多租户底座（2-3 周）

**目标：** 搭建项目骨架，完成两套系统的登录认证和权限管理。

**后端：**
- [ ] Go 项目初始化（Gin + GORM + PG 驱动 pgx）
- [ ] 配置管理（Viper：PG/Redis/MinIO/JWT）
- [ ] 数据库迁移（golang-migrate）：系统管理表 + RLS 策略
- [ ] **权限自动加载**：manifest 定义 + 启动同步到 `sys_permission`
- [ ] JWT 认证（生成/解析/黑名单）
- [ ] 平台登录 + 店铺登录接口
- [ ] 多租户中间件 + **RLS 事务管理**
- [ ] GORM 租户 Scope（应用层双重保障）
- [ ] 数据权限中间件（部门树闭包表）
- [ ] RBAC 权限校验中间件（基于 `sys_role_permission`）
- [ ] 用户/角色/部门 CRUD（平台+店铺各一套）
- [ ] **角色权限分配 UI**（勾选权限树 → 写入 `sys_role_permission`）
- [ ] 操作审计中间件
- [ ] 统一响应格式 + 错误处理
- [ ] 种子数据（平台超管、默认部门）

**前端：**
- [ ] Vue3 + TypeScript + vue-pure-admin 初始化
- [ ] 两套登录页面（平台/店铺）
- [ ] 路由守卫（认证 + 动态路由 + 权限码控制）
- [ ] Axios 封装（Token 注入、错误拦截）
- [ ] 布局组件
- [ ] 用户/角色/部门管理页面（平台+店铺）
- [ ] **角色权限分配页面**（权限树展示 + 勾选）
- [ ] 按钮级权限指令（`v-permission`）

**基础设施：**
- [ ] Docker Compose（PostgreSQL + Redis + MinIO + Nginx）
- [ ] Makefile（dev/test/migrate/seed）

**交付物：** 两套系统可独立登录，完整的用户/角色/权限/部门管理。

---

### 阶段二：平台业务模块（3-4 周）

**目标：** 完成平台系统的全部业务功能。

**后端：**
- [ ] 店铺管理 CRUD + 初始化（管理员/角色/全部权限分配）
- [ ] 商品分类 CRUD
- [ ] 商品 CRUD + 上下架
- [ ] 商品流程模板配置
- [ ] 收支分类管理（三级树）
- [ ] 平台财务报表

**前端：**
- [ ] 店铺管理页面（含创建时自动初始化流程）
- [ ] 商品分类 + 商品管理 + 流程节点配置页面
- [ ] 收支分类管理页面（三级树）
- [ ] 平台报表页面

**交付物：** 平台系统功能完整，可创建店铺、管理商品和流程模板。

---

### 阶段三：店铺业务模块 — 选品 + 客户 + 订单（3-4 周）

**目标：** 完成店铺核心业务流程。

**后端：**
- [ ] 选品管理（从平台选品、定价、上下架）
- [ ] 客户管理 CRUD
- [ ] 订单创建（多商品拆分，单事务）
- [ ] 订单列表/详情/取消
- [ ] 订单流程跟进（looplab/fsm）
- [ ] 跟进记录 + 附件上传
- [ ] MinIO 文件上传/下载

**前端：**
- [ ] 选品管理页面（平台商品库 + 选品 + 定价）
- [ ] 客户管理页面
- [ ] 订单创建页面（选客户 + 多选商品）
- [ ] 订单列表 + 详情 + 流程跟进时间线
- [ ] 附件上传/预览/下载组件

**交付物：** 店铺核心流程可用：选品 → 客户 → 下单 → 流程跟进。

---

### 阶段四：店铺财务模块（2-3 周）

**目标：** 完成店铺财务管理全功能。

**后端：**
- [ ] 收支账户管理（JSONB 灵活配置）
- [ ] 收支分类同步（从平台选择，只读）
- [ ] 收支记录 CRUD
- [ ] 审核流程（通过/驳回 + 附件 + 实际金额）
- [ ] 财务报表（汇总/利润/余额）
- [ ] 报表导出（Excel）

**前端：**
- [ ] 收支账户页面（按类型动态展示不同配置字段）
- [ ] 收支分类同步页面
- [ ] 收支记录页面 + 审核操作
- [ ] 财务报表页面（图表 + 表格 + 导出）

**交付物：** 店铺财务全流程：账户 → 分类 → 记录 → 审核 → 报表。

---

### 阶段五：测试 + 优化 + 部署（2-3 周）

- [ ] 集成测试（核心业务端到端）
- [ ] 多租户隔离测试（跨租户数据安全，含 RLS 验证）
- [ ] 数据权限测试（部门树过滤）
- [ ] 性能优化（索引、N+1 排查、Redis 缓存、PG 慢查询分析）
- [ ] 前端优化（路由懒加载、按需加载）
- [ ] Nginx 配置（HTTPS、缓存、API 代理）
- [ ] Docker 生产镜像
- [ ] PG 备份脚本（pg_dump + WAL archiving）
- [ ] 部署文档

**总工期：** 约 12-17 周（3-4 个月），可根据团队规模并行压缩。

---

## 10. 部署与运维

### 10.1 Docker Compose（MVP）

```yaml
version: '3.8'
services:
  backend:
    build: ./backend
    ports: ["8080:8080"]
    depends_on: [postgres, redis, minio]
    environment:
      - DB_DSN=host=postgres user=${DB_USER} password=${DB_PASSWORD} dbname=platform sslmode=disable TimeZone=Asia/Shanghai
      - REDIS_ADDR=redis:6379
      - MINIO_ENDPOINT=minio:9000
      - JWT_SECRET=${JWT_SECRET}

  frontend:
    build: ./frontend

  nginx:
    image: nginx:alpine
    ports: ["80:80", "443:443"]
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./frontend/dist:/usr/share/nginx/html
    depends_on: [backend]

  postgres:
    image: postgres:16-alpine
    volumes: ["pg_data:/var/lib/postgresql/data"]
    environment:
      POSTGRES_DB: platform
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    ports: ["5432:5432"]

  redis:
    image: redis:7-alpine
    volumes: ["redis_data:/data"]

  minio:
    image: minio/minio
    command: server /data --console-address ":9001"
    ports: ["9000:9000", "9001:9001"]
    volumes: ["minio_data:/data"]
    environment:
      MINIO_ROOT_USER: ${MINIO_ACCESS_KEY}
      MINIO_ROOT_PASSWORD: ${MINIO_SECRET_KEY}

volumes:
  pg_data:
  redis_data:
  minio_data:
```

### 10.2 Nginx 配置

```nginx
server {
    listen 80;
    server_name _;

    location / {
        root /usr/share/nginx/html;
        try_files $uri $uri/ /index.html;
    }

    location /api/v1/platform/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /api/v1/shop/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    client_max_body_size 50m;
}
```

### 10.3 数据库迁移

```bash
make migrate-up      # 执行迁移
make migrate-down    # 回滚
make seed            # 种子数据 + 权限同步
```

迁移文件结构：
```
database/migrations/
  000001_create_sys_tables.up.sql         # 系统管理表
  000001_create_sys_tables.down.sql
  000002_create_permission_tables.up.sql  # sys_permission + sys_role_permission
  000002_create_permission_tables.down.sql
  000003_create_platform_tables.up.sql    # shop, product, finance_category
  000003_create_platform_tables.down.sql
  000004_create_shop_tables.up.sql        # shop_product, customer, order*
  000004_create_shop_tables.down.sql
  000005_create_finance_tables.up.sql     # finance_account, record, attachment
  000005_create_finance_tables.down.sql
  000006_enable_rls_policies.up.sql       # ★ 所有店铺级表启用 RLS
  000006_enable_rls_policies.down.sql
```

### 10.4 备份策略

| 策略 | 频率 | 保留 | 工具 |
|---|---|---|---|
| PG 全量备份 | 每日 02:00 | 30 天 | `pg_dump` + cron |
| WAL 归档 | 实时 | 7 天 | `archive_mode=on` |
| MinIO 数据 | 每日 03:00 | 30 天 | `mc mirror` |
| Redis RDB | 每小时 | 24 份 | 自动 |

### 10.5 生产环境扩展（K8s）

- 后端：Deployment + HPA
- PostgreSQL：迁移到云托管 RDS（如 RDS for PostgreSQL / Supabase / Neon）
- Redis：云托管 ElastiCache
- MinIO：迁移到 S3 或 MinIO 集群

---

## 11. 风险与应对

| # | 风险 | 影响 | 概率 | 应对 |
|---|---|---|---|---|
| 1 | **RLS 配置错误** | 严重 — 租户数据泄露 | 低 | RLS 策略在迁移脚本中统一定义，集成测试覆盖跨租户场景。RLS + GORM Scopes 双保险 |
| 2 | **RLS 遗忘启用** | 严重 — 新表没有 RLS | 中 | 迁移脚本模板中包含 `ENABLE ROW LEVEL SECURITY`，Code Review 检查新表是否启用 RLS |
| 3 | **权限同步遗漏** | 中 — 新功能权限未声明 | 中 | CI 检查：路由注册的权限码必须在 manifest 中存在 |
| 4 | **工作流过度设计** | 高 — 浪费时间 | 高 | 严格限制为顺序状态机，使用 looplab/fsm，禁止 BPMN |
| 5 | **财务数据不一致** | 严重 | 低 | 审核通过后禁止编辑/删除，PG 约束 + 应用层校验 |
| 6 | **订单拆分事务失败** | 高 | 中 | 单事务包裹，失败全部回滚 |
| 7 | **部门树查询性能** | 中 | 中 | 闭包表预计算 + Redis 缓存 |
| 8 | **JSONB 查询性能** | 低 | 低 | 按类型+租户索引即可，JSONB 内容不参与 WHERE |
| 9 | **PG 运维经验不足** | 中 | 中 | 使用 Docker 部署，pg_dump 备份足够简单。复杂场景迁移云托管 |

---

## 12. 待确认事项

| # | 事项 | 当前假设 | 需确认 |
|---|---|---|---|
| 1 | **团队规模** | 按全栈 2-3 人估算 | 实际配置？ |
| 2 | **部署环境** | Docker Compose 自部署 | 云服务商偏好？ |
| 3 | **域名规划** | 单域名，路径区分 `/platform` `/shop` | 是否需要子域名？ |
| 4 | **店铺创建** | 平台管理员手动创建 | 是否需要自助注册？ |
| 5 | **消息通知** | MVP 不做 | 订单状态变更是否需要通知？ |
| 6 | **数据导出** | 仅财务报表 | 其他模块是否需要导出？ |
| 7 | **平台-店铺结算** | MVP 不做 | 平台是否抽成？ |
| 8 | **客户跨店** | 不共享 | 同客户能否多店铺有记录？ |
| 9 | **线上商城对接** | `mall_product_code` 保留字段 | 商城系统时间计划？ |

---

> **文档结束。** 本计划为实施指导文档，开发过程中如遇需求变更或技术决策调整，请同步更新本文档。
