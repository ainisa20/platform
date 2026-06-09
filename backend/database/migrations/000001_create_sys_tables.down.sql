-- Migration: 000001_create_sys_tables (rollback)
-- Drop all system management tables in reverse dependency order

BEGIN;

DROP TABLE IF EXISTS sys_operation_log;
DROP TABLE IF EXISTS sys_dept_closure;
DROP TABLE IF EXISTS sys_dept;
DROP TABLE IF EXISTS sys_user_role;
DROP TABLE IF EXISTS sys_role;
DROP TABLE IF EXISTS sys_user;

COMMIT;
