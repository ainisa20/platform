-- Migration: 000003_enable_rls_policies
-- Description: Enable Row Level Security on all tenant-scoped system tables
-- and create tenant_isolation policies for mandatory multi-tenant data isolation.
--
-- RLS ensures that even if application code omits WHERE tenant_id = ?,
-- the database will automatically filter rows to the current tenant context.
-- Tenant context is set via: SET app.tenant_id = <value> per request.

BEGIN;

-- ============================================================
-- Enable RLS on system tables
-- ============================================================

ALTER TABLE sys_user ENABLE ROW LEVEL SECURITY;
ALTER TABLE sys_role ENABLE ROW LEVEL SECURITY;
ALTER TABLE sys_user_role ENABLE ROW LEVEL SECURITY;
ALTER TABLE sys_role_permission ENABLE ROW LEVEL SECURITY;
ALTER TABLE sys_dept ENABLE ROW LEVEL SECURITY;
ALTER TABLE sys_dept_closure ENABLE ROW LEVEL SECURITY;

-- ============================================================
-- Create tenant_isolation policies
-- Policy: each tenant can only see rows where tenant_id matches
-- the session variable app.tenant_id set by the application middleware.
-- ============================================================

CREATE POLICY tenant_isolation ON sys_user
    FOR ALL
    USING (tenant_id = current_setting('app.tenant_id', true)::bigint)
    WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::bigint);

CREATE POLICY tenant_isolation ON sys_role
    FOR ALL
    USING (tenant_id = current_setting('app.tenant_id', true)::bigint)
    WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::bigint);

CREATE POLICY tenant_isolation ON sys_user_role
    FOR ALL
    USING (user_id IN (
        SELECT u.id FROM sys_user u WHERE u.tenant_id = current_setting('app.tenant_id', true)::bigint
    ))
    WITH CHECK (user_id IN (
        SELECT u.id FROM sys_user u WHERE u.tenant_id = current_setting('app.tenant_id', true)::bigint
    ));

CREATE POLICY tenant_isolation ON sys_role_permission
    FOR ALL
    USING (tenant_id = current_setting('app.tenant_id', true)::bigint)
    WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::bigint);

CREATE POLICY tenant_isolation ON sys_dept
    FOR ALL
    USING (tenant_id = current_setting('app.tenant_id', true)::bigint)
    WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::bigint);

CREATE POLICY tenant_isolation ON sys_dept_closure
    FOR ALL
    USING (tenant_id = current_setting('app.tenant_id', true)::bigint)
    WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::bigint);

-- ============================================================
-- Grant BYPASSRLS to postgres superuser
-- Platform users (tenant_id=0) connect as a BYPASSRLS role
-- to manage cross-tenant platform data without RLS filtering.
-- ============================================================

ALTER ROLE postgres BYPASSRLS;

COMMENT ON POLICY tenant_isolation ON sys_user IS
    'RLS tenant isolation: restricts rows to current tenant via app.tenant_id session variable';

COMMIT;
