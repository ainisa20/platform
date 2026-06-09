-- Migration: 000003_enable_rls_policies (rollback)
-- Drop all tenant_isolation policies and disable RLS

BEGIN;

DROP POLICY IF EXISTS tenant_isolation ON sys_user;
DROP POLICY IF EXISTS tenant_isolation ON sys_role;
DROP POLICY IF EXISTS tenant_isolation ON sys_user_role;
DROP POLICY IF EXISTS tenant_isolation ON sys_role_permission;
DROP POLICY IF EXISTS tenant_isolation ON sys_dept;
DROP POLICY IF EXISTS tenant_isolation ON sys_dept_closure;

ALTER TABLE sys_user DISABLE ROW LEVEL SECURITY;
ALTER TABLE sys_role DISABLE ROW LEVEL SECURITY;
ALTER TABLE sys_user_role DISABLE ROW LEVEL SECURITY;
ALTER TABLE sys_role_permission DISABLE ROW LEVEL SECURITY;
ALTER TABLE sys_dept DISABLE ROW LEVEL SECURITY;
ALTER TABLE sys_dept_closure DISABLE ROW LEVEL SECURITY;

COMMIT;
