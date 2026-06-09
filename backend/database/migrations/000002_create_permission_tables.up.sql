-- Migration: 000002_create_permission_tables
-- Description: Create permission definition and role-permission mapping tables
-- Tables: sys_permission (global), sys_role_permission (tenant-scoped)

BEGIN;

-- ============================================================
-- sys_permission — 全局权限定义表
-- 代码自动同步, 无 tenant_id
-- ============================================================
CREATE TABLE sys_permission (
    id              BIGSERIAL       PRIMARY KEY,
    parent_id       BIGINT          NOT NULL DEFAULT 0,
    system_type     VARCHAR(16)     NOT NULL,                 -- platform / shop
    name            VARCHAR(64)     NOT NULL,
    type            SMALLINT        NOT NULL,                 -- 1=目录 2=菜单 3=按钮
    path            VARCHAR(255),
    component       VARCHAR(255),
    perms_code      VARCHAR(100)    NOT NULL DEFAULT '',
    icon            VARCHAR(64),
    sort            SMALLINT        NOT NULL DEFAULT 0,
    visible         BOOLEAN         NOT NULL DEFAULT TRUE,
    status          SMALLINT        NOT NULL DEFAULT 1,       -- 1=启用 2=停用
    auto_synced     BOOLEAN         NOT NULL DEFAULT TRUE,    -- 代码同步的条目不可在 UI 编辑/删除
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uk_sys_permission_sys_type_code ON sys_permission (system_type, perms_code)
    WHERE perms_code != '';
CREATE INDEX idx_sys_permission_sys_type_parent ON sys_permission (system_type, parent_id);

COMMENT ON TABLE  sys_permission IS '全局权限定义表 (代码自动同步)';
COMMENT ON COLUMN sys_permission.system_type  IS '系统类型: platform / shop';
COMMENT ON COLUMN sys_permission.type         IS '类型: 1=目录 2=菜单 3=按钮';
COMMENT ON COLUMN sys_permission.perms_code   IS '权限标识, 如 platform:user:create';
COMMENT ON COLUMN sys_permission.auto_synced  IS '是否从代码自动同步 (true=不可UI编辑)';

-- ============================================================
-- sys_role_permission — 角色权限关联表 (租户隔离)
-- ============================================================
CREATE TABLE sys_role_permission (
    id              BIGSERIAL       PRIMARY KEY,
    tenant_id       BIGINT          NOT NULL,                 -- 0=平台
    role_id         BIGINT          NOT NULL,
    permission_id   BIGINT          NOT NULL
);

CREATE INDEX idx_sys_role_permission_tenant_id ON sys_role_permission (tenant_id);
CREATE UNIQUE INDEX uk_sys_role_permission ON sys_role_permission (tenant_id, role_id, permission_id);

COMMENT ON TABLE sys_role_permission IS '角色权限关联表 (租户隔离)';

COMMIT;
