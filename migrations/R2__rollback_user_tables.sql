-- R2__rollback_user_tables.sql
-- Rollback script for the user tables migration

-- First remove foreign key constraints
ALTER TABLE IF EXISTS entries DROP CONSTRAINT IF EXISTS fk_entries_created_by;
ALTER TABLE IF EXISTS meanings DROP CONSTRAINT IF EXISTS fk_meanings_created_by;
ALTER TABLE IF EXISTS examples DROP CONSTRAINT IF EXISTS fk_examples_created_by;
ALTER TABLE IF EXISTS translations DROP CONSTRAINT IF EXISTS fk_translations_created_by;
ALTER TABLE IF EXISTS change_history DROP CONSTRAINT IF EXISTS fk_change_history_user;
ALTER TABLE IF EXISTS entries DROP CONSTRAINT IF EXISTS fk_entries_source_language;

-- Drop user-related tables
DROP TABLE IF EXISTS user_stats CASCADE;
DROP TABLE IF EXISTS user_preferences CASCADE;
DROP TABLE IF EXISTS auth_tokens CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Drop any indices that might not be automatically dropped with the tables
DROP INDEX IF EXISTS idx_users_username CASCADE;
DROP INDEX IF EXISTS idx_users_email CASCADE;
DROP INDEX IF EXISTS idx_users_role CASCADE;
DROP INDEX IF EXISTS idx_auth_tokens_user_id CASCADE;
DROP INDEX IF EXISTS idx_auth_tokens_refresh_token CASCADE;
DROP INDEX IF EXISTS idx_auth_tokens_expires_at CASCADE;
