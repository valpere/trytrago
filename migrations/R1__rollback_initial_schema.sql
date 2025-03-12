-- R1__rollback_initial_schema.sql
-- Rollback script for the initial schema migration

-- Drop all tables in reverse order of creation to respect foreign key constraints
DROP TABLE IF EXISTS change_history CASCADE;
DROP TABLE IF EXISTS languages CASCADE;
DROP TABLE IF EXISTS translations CASCADE;
DROP TABLE IF EXISTS examples CASCADE;
DROP TABLE IF EXISTS meanings CASCADE;
DROP TABLE IF EXISTS entries CASCADE;
DROP TABLE IF EXISTS parts_of_speech CASCADE;

-- Drop any indices that might not be automatically dropped with the tables
DROP INDEX IF EXISTS idx_entries_word CASCADE;
DROP INDEX IF EXISTS idx_entries_type CASCADE;
DROP INDEX IF EXISTS idx_entries_created_at CASCADE;
DROP INDEX IF EXISTS idx_meanings_entry_id CASCADE;
DROP INDEX IF EXISTS idx_meanings_part_of_speech_id CASCADE;
DROP INDEX IF EXISTS idx_examples_meaning_id CASCADE;
DROP INDEX IF EXISTS idx_translations_meaning_id CASCADE;
DROP INDEX IF EXISTS idx_translations_language_id CASCADE;
DROP INDEX IF EXISTS idx_translations_text CASCADE;
DROP INDEX IF EXISTS idx_change_history_entry_id CASCADE;
DROP INDEX IF EXISTS idx_change_history_created_at CASCADE;
