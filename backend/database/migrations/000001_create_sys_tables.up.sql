-- Migration: 000001_create_sys_tables
-- Description: Create shared system management tables (platform + shop)
-- Tables: sys_user, sys_role, sys_user_role, sys_dept, sys_dept_closure, sys_operation_log

BEGIN;

-- ============================================================
-- sys_user — 用户表
-- ============================================================
CREATE TABLE sys_user (
    id              BIGSERIAL       PRIMARY KEY,
    tenant_id       BIGINT          NOT NULL DEFAULT 0,
    dept_id         BIGINT,
    username        VARCHAR(64)     NOT NULL,
    password        VARCHAR(255)    NOT NULL,
    real_name       VARCHAR(64)     NOT NULL,
    phone           VARCHAR(20),
    email           VARCHAR(128),
    avatar          VARCHAR(500),
    status          SMALLINT        NOT NULL DEFAULT 1,      -- 1=启用 2=停用
    last_login_at   TIMESTAMPTZ,
    last_login_ip   VARCHAR(45),
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_by      BIGINT,
    updated_at      TIMESTAMPTZ,
    updated_by      BIGINT,
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_sys_user_tenant_id    ON sys_user (tenant_id);
CREATE INDEX idx_sys_user_dept_id      ON sys_user (dept_id);
CREATE INDEX idx_sys_user_deleted_at   ON sys_user (deleted_at);
CREATE UNIQUE INDEX uk_sys_user_tenant_username ON sys_user (tenant_id, username, deleted_at);

COMMENT ON TABLE  sys_user IS '用户表';
COMMENT ON COLUMN sys_user.tenant_id     IS '租户ID: 0=平台, N=店铺';
COMMENT ON COLUMN sys_user.status        IS '状态: 1=启用 2=停用';
COMMENT ON COLUMN sys_user.deleted_at    IS '软删除时间';

-- ============================================================
-- sys_role — 角色表
-- ============================================================
CREATE TABLE sys_role (
    id              BIGSERIAL       PRIMARY KEY,
    tenant_id       BIGINT          NOT NULL DEFAULT 0,
    role_name       VARCHAR(64)     NOT NULL,
    role_code       VARCHAR(64)     NOT NULL,
    data_scope      SMALLINT        NOT NULL DEFAULT 1,      -- 1=全部 2=本部门及以下 3=仅本部门 4=仅本人
    sort            SMALLINT        NOT NULL DEFAULT 0,
    status          SMALLINT        NOT NULL DEFAULT 1,      -- 1=启用 2=停用
    remark          VARCHAR(255),
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_by      BIGINT,
    updated_at      TIMESTAMPTZ,
    updated_by      BIGINT,
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_sys_role_tenant_id    ON sys_role (tenant_id);
CREATE INDEX idx_sys_role_deleted_at   ON sys_role (deleted_at);
CREATE UNIQUE INDEX uk_sys_role_tenant_code ON sys_role (tenant_id, role_code, deleted_at);

COMMENT ON TABLE  sys_role IS '角色表';
COMMENT ON COLUMN sys_role.data_scope  IS '数据范围: 1=全部 2=本部门及以下 3=仅本部门 4=仅本人';
COMMENT ON COLUMN sys_role.status      IS '状态: 1=启用 2=停用';

-- ============================================================
-- sys_user_role — 用户角色关联表
-- ============================================================
CREATE TABLE sys_user_role (
    id              BIGSERIAL       PRIMARY KEY,
    user_id         BIGINT          NOT NULL,
    role_id         BIGINT          NOT NULL
);

CREATE UNIQUE INDEX uk_sys_user_role ON sys_user_role (user_id, role_id);

COMMENT ON TABLE sys_user_role IS '用户角色关联表';

-- ============================================================
-- sys_dept — 部门/组织树
-- ============================================================
CREATE TABLE sys_dept (
    id              BIGSERIAL       PRIMARY KEY,
    tenant_id       BIGINT          NOT NULL DEFAULT 0,
    parent_id       BIGINT          NOT NULL DEFAULT 0,
    ancestors       VARCHAR(500),                              -- 祖先链 "0,1,5"
    dept_name       VARCHAR(64)     NOT NULL,
    sort            SMALLINT        NOT NULL DEFAULT 0,
    leader          VARCHAR(64),
    phone           VARCHAR(20),
    status          SMALLINT        NOT NULL DEFAULT 1,       -- 1=启用 2=停用
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_by      BIGINT,
    updated_at      TIMESTAMPTZ,
    updated_by      BIGINT,
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_sys_dept_tenant_id    ON sys_dept (tenant_id);
CREATE INDEX idx_sys_dept_deleted_at   ON sys_dept (deleted_at);

COMMENT ON TABLE  sys_dept IS '部门/组织树';
COMMENT ON COLUMN sys_dept.tenant_id   IS '租户ID: 0=平台, N=店铺';
COMMENT ON COLUMN sys_dept.ancestors   IS '祖先链, 逗号分隔 如 "0,1,5"';
COMMENT ON COLUMN sys_dept.status      IS '状态: 1=启用 2=停用';

-- ============================================================
-- sys_dept_closure — 部门闭包表
-- ============================================================
CREATE TABLE sys_dept_closure (
    id              BIGSERIAL       PRIMARY KEY,
    tenant_id       BIGINT          NOT NULL,
    ancestor_id     BIGINT          NOT NULL,
    descendant_id   BIGINT          NOT NULL,
    depth           SMALLINT        NOT NULL DEFAULT 0        -- 0=自身, 1=子级...
);

CREATE INDEX idx_sys_dept_closure_tenant_id    ON sys_dept_closure (tenant_id);
CREATE INDEX idx_sys_dept_closure_ancestor_id  ON sys_dept_closure (ancestor_id);
CREATE INDEX idx_sys_dept_closure_descendant_id ON sys_dept_closure (descendant_id);
CREATE UNIQUE INDEX uk_sys_dept_closure ON sys_dept_closure (ancestor_id, descendant_id);

COMMENT ON TABLE  sys_dept_closure IS '部门闭包表';
COMMENT ON COLUMN sys_dept_closure.depth IS '深度: 0=自身, 1=子级, 2=孙级...';

-- ============================================================
-- sys_operation_log — 操作日志
-- ============================================================
CREATE TABLE sys_operation_log (
    id              BIGSERIAL       PRIMARY KEY,
    tenant_id       BIGINT,
    user_id         BIGINT,
    username        VARCHAR(64),
    module          VARCHAR(64),
    action          VARCHAR(64),
    method          VARCHAR(10),                               -- HTTP 方法
    url             VARCHAR(500),
    params          TEXT,
    ip              VARCHAR(45),
    duration_ms     SMALLINT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sys_operation_log_tenant_id ON sys_operation_log (tenant_id);
CREATE INDEX idx_sys_operation_log_user_id   ON sys_operation_log (user_id);
CREATE INDEX idx_sys_operation_log_created_at ON sys_operation_log (created_at);

COMMENT ON TABLE  sys_operation_log IS '操作日志';
COMMENT ON COLUMN sys_operation_log.duration_ms IS '请求耗时(毫秒)';

COMMIT;
