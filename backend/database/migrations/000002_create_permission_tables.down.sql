-- Migration: 000002_create_permission_tables (rollback)

BEGIN;

DROP TABLE IF EXISTS sys_role_permission;
DROP TABLE IF EXISTS sys_permission;

COMMIT;
