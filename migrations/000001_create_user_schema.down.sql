-- Drop tables in reverse dependency order, then enums, triggers, and schema.

DROP TABLE IF EXISTS "user".artist_profile;
DROP TABLE IF EXISTS "user".user_role;
DROP TABLE IF EXISTS "user".user_management;
DROP TABLE IF EXISTS "user".permission;
DROP TABLE IF EXISTS "user".role;
DROP TABLE IF EXISTS "user".submodule;
DROP TABLE IF EXISTS "user".module;
DROP TABLE IF EXISTS "user".application;

DROP FUNCTION IF EXISTS "user".set_updated_at();

DROP TYPE IF EXISTS "user".primary_service;
DROP TYPE IF EXISTS "user".user_role_status;

DROP SCHEMA IF EXISTS "user";
