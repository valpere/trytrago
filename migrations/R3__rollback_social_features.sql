-- R3__rollback_social_features.sql
-- Rollback script for the social features migration

-- Drop triggers first
DROP TRIGGER IF EXISTS update_user_stats_on_entry_insert ON entries;
DROP TRIGGER IF EXISTS update_user_stats_on_meaning_insert ON meanings;
DROP TRIGGER IF EXISTS update_user_stats_on_translation_insert ON translations;
DROP TRIGGER IF EXISTS update_user_stats_on_comment_insert ON comments;
DROP TRIGGER IF EXISTS update_user_stats_on_like_insert ON likes;

DROP TRIGGER IF EXISTS create_feed_item_on_entry_insert ON entries;
DROP TRIGGER IF EXISTS create_feed_item_on_meaning_insert ON meanings;
DROP TRIGGER IF EXISTS create_feed_item_on_translation_insert ON translations;
DROP TRIGGER IF EXISTS create_feed_item_on_comment_insert ON comments;
DROP TRIGGER IF EXISTS create_feed_item_on_like_insert ON likes;

-- Drop functions
DROP FUNCTION IF EXISTS update_user_stats();
DROP FUNCTION IF EXISTS create_feed_item();

-- Drop views
DROP VIEW IF EXISTS meaning_stats;
DROP VIEW IF EXISTS translation_stats;
DROP VIEW IF EXISTS user_activity;

-- Drop tables
DROP TABLE IF EXISTS feed_items CASCADE;
DROP TABLE IF EXISTS likes CASCADE;
DROP TABLE IF EXISTS comments CASCADE;

-- Drop indices
DROP INDEX IF EXISTS idx_comments_user_id CASCADE;
DROP INDEX IF EXISTS idx_comments_target_type_id CASCADE;
DROP INDEX IF EXISTS idx_comments_created_at CASCADE;
DROP INDEX IF EXISTS idx_likes_user_id CASCADE;
DROP INDEX IF EXISTS idx_likes_target_type_id CASCADE;
DROP INDEX IF EXISTS idx_likes_unique_user_target CASCADE;
DROP INDEX IF EXISTS idx_feed_items_user_id CASCADE;
DROP INDEX IF EXISTS idx_feed_items_timestamp CASCADE;
