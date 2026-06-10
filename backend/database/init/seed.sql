-- Seed: Platform initial data
-- Superadmin user, admin role, root department, and role assignment

BEGIN;

-- Platform root department
INSERT INTO sys_dept (tenant_id, parent_id, ancestors, dept_name, sort, leader, status, created_at)
VALUES (0, 0, '0', '总部', 0, NULL, 1, NOW());

-- Closure self-row for the root department (created via SQL, bypasses repo Create())
INSERT INTO sys_dept_closure (tenant_id, ancestor_id, descendant_id, depth)
SELECT 0, id, id, 0 FROM sys_dept WHERE parent_id = 0 AND NOT EXISTS (
    SELECT 1 FROM sys_dept_closure WHERE ancestor_id = sys_dept.id AND descendant_id = sys_dept.id
);

-- Platform admin role
INSERT INTO sys_role (tenant_id, role_name, role_code, data_scope, sort, status, created_at)
VALUES (0, '平台管理员', 'platform_admin', 1, 0, 1, NOW());

-- Platform superadmin user
-- password = bcrypt('admin123') = $2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE8ByOhJIrdAu2
INSERT INTO sys_user (tenant_id, dept_id, username, password, real_name, status, created_at)
VALUES (0, (SELECT id FROM sys_dept WHERE tenant_id = 0 AND dept_name = '总部' LIMIT 1),
        'admin', '$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE8ByOhJIrdAu2',
        '平台管理员', 1, NOW());

-- Assign admin user to platform_admin role
INSERT INTO sys_user_role (user_id, role_id)
VALUES (
    (SELECT id FROM sys_user WHERE tenant_id = 0 AND username = 'admin' LIMIT 1),
    (SELECT id FROM sys_role WHERE tenant_id = 0 AND role_code = 'platform_admin' LIMIT 1)
);

-- Assign ALL platform permissions to platform_admin role
-- This must run AFTER permission sync has populated sys_permission.
-- Use ON CONFLICT DO NOTHING so it is idempotent.
INSERT INTO sys_role_permission (tenant_id, role_id, permission_id)
SELECT
    0,
    (SELECT id FROM sys_role WHERE tenant_id = 0 AND role_code = 'platform_admin' LIMIT 1),
    p.id
FROM sys_permission p
WHERE p.system_type = 'platform'
  AND p.perms_code != ''
ON CONFLICT DO NOTHING;

COMMIT;
