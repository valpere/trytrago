-- V3__add_social_features.sql
-- Social features tables for TryTraGo dictionary database

-- Create comments table
CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type VARCHAR(20) NOT NULL CHECK (target_type IN ('meaning', 'translation')),
    target_id UUID NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Add indexes for comments
CREATE INDEX idx_comments_user_id ON comments (user_id);
CREATE INDEX idx_comments_target_type_id ON comments (target_type, target_id);
CREATE INDEX idx_comments_created_at ON comments (created_at DESC);

-- Create likes table
CREATE TABLE IF NOT EXISTS likes (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type VARCHAR(20) NOT NULL CHECK (target_type IN ('meaning', 'translation')),
    target_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Add indexes for likes
CREATE INDEX idx_likes_user_id ON likes (user_id);
CREATE INDEX idx_likes_target_type_id ON likes (target_type, target_id);
-- Add a unique constraint to prevent duplicate likes
CREATE UNIQUE INDEX idx_likes_unique_user_target ON likes (user_id, target_type, target_id) WHERE deleted_at IS NULL;

-- Create feed_items table for user activity feed
CREATE TABLE IF NOT EXISTS feed_items (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('entry', 'meaning', 'translation', 'comment', 'like')),
    reference_id UUID NOT NULL,
    content JSONB NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for feed items
CREATE INDEX idx_feed_items_user_id ON feed_items (user_id);
CREATE INDEX idx_feed_items_timestamp ON feed_items (timestamp DESC);

-- Create views for common queries

-- View for meaning statistics (comments and likes count)
CREATE OR REPLACE VIEW meaning_stats AS
SELECT 
    m.id AS meaning_id,
    COUNT(DISTINCT c.id) AS comments_count,
    COUNT(DISTINCT l.id) AS likes_count
FROM 
    meanings m
LEFT JOIN 
    comments c ON c.target_id = m.id AND c.target_type = 'meaning' AND c.deleted_at IS NULL
LEFT JOIN 
    likes l ON l.target_id = m.id AND l.target_type = 'meaning' AND l.deleted_at IS NULL
GROUP BY 
    m.id;

-- View for translation statistics (comments and likes count)
CREATE OR REPLACE VIEW translation_stats AS
SELECT 
    t.id AS translation_id,
    COUNT(DISTINCT c.id) AS comments_count,
    COUNT(DISTINCT l.id) AS likes_count
FROM 
    translations t
LEFT JOIN 
    comments c ON c.target_id = t.id AND c.target_type = 'translation' AND c.deleted_at IS NULL
LEFT JOIN 
    likes l ON l.target_id = t.id AND l.target_type = 'translation' AND l.deleted_at IS NULL
GROUP BY 
    t.id;

-- View for user activity
CREATE OR REPLACE VIEW user_activity AS
SELECT
    u.id AS user_id,
    u.username,
    COUNT(DISTINCT e.id) AS entries_count,
    COUNT(DISTINCT m.id) AS meanings_count,
    COUNT(DISTINCT t.id) AS translations_count,
    COUNT(DISTINCT c.id) AS comments_count,
    COUNT(DISTINCT l.id) AS likes_count
FROM
    users u
LEFT JOIN 
    entries e ON e.created_by_id = u.id
LEFT JOIN 
    meanings m ON m.created_by_id = u.id
LEFT JOIN 
    translations t ON t.created_by_id = u.id
LEFT JOIN 
    comments c ON c.user_id = u.id AND c.deleted_at IS NULL
LEFT JOIN 
    likes l ON l.user_id = u.id AND l.deleted_at IS NULL
GROUP BY
    u.id, u.username;

-- Create triggers to update user_stats whenever a user creates content

-- Function to update user stats
CREATE OR REPLACE FUNCTION update_user_stats()
RETURNS TRIGGER AS $$
BEGIN
    -- Create stats record if it doesn't exist
    INSERT INTO user_stats (id, user_id, last_activity_at, created_at, updated_at)
    VALUES (gen_random_uuid(), NEW.created_by_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    ON CONFLICT (user_id) DO NOTHING;

    -- Update appropriate counter based on the table being modified
    CASE TG_TABLE_NAME
        WHEN 'entries' THEN
            UPDATE user_stats SET 
                entries_created = entries_created + 1,
                last_activity_at = CURRENT_TIMESTAMP,
                updated_at = CURRENT_TIMESTAMP
            WHERE user_id = NEW.created_by_id;
        WHEN 'meanings' THEN
            UPDATE user_stats SET 
                meanings_added = meanings_added + 1,
                last_activity_at = CURRENT_TIMESTAMP,
                updated_at = CURRENT_TIMESTAMP
            WHERE user_id = NEW.created_by_id;
        WHEN 'translations' THEN
            UPDATE user_stats SET 
                translations_added = translations_added + 1,
                last_activity_at = CURRENT_TIMESTAMP,
                updated_at = CURRENT_TIMESTAMP
            WHERE user_id = NEW.created_by_id;
        WHEN 'comments' THEN
            UPDATE user_stats SET 
                comments_posted = comments_posted + 1,
                last_activity_at = CURRENT_TIMESTAMP,
                updated_at = CURRENT_TIMESTAMP
            WHERE user_id = NEW.user_id;
        WHEN 'likes' THEN
            UPDATE user_stats SET 
                likes_given = likes_given + 1,
                last_activity_at = CURRENT_TIMESTAMP,
                updated_at = CURRENT_TIMESTAMP
            WHERE user_id = NEW.user_id;
    END CASE;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for each table
CREATE TRIGGER update_user_stats_on_entry_insert
AFTER INSERT ON entries
FOR EACH ROW
WHEN (NEW.created_by_id IS NOT NULL)
EXECUTE FUNCTION update_user_stats();

CREATE TRIGGER update_user_stats_on_meaning_insert
AFTER INSERT ON meanings
FOR EACH ROW
WHEN (NEW.created_by_id IS NOT NULL)
EXECUTE FUNCTION update_user_stats();

CREATE TRIGGER update_user_stats_on_translation_insert
AFTER INSERT ON translations
FOR EACH ROW
WHEN (NEW.created_by_id IS NOT NULL)
EXECUTE FUNCTION update_user_stats();

CREATE TRIGGER update_user_stats_on_comment_insert
AFTER INSERT ON comments
FOR EACH ROW
EXECUTE FUNCTION update_user_stats();

CREATE TRIGGER update_user_stats_on_like_insert
AFTER INSERT ON likes
FOR EACH ROW
EXECUTE FUNCTION update_user_stats();

-- Create function for feed item creation
CREATE OR REPLACE FUNCTION create_feed_item()
RETURNS TRIGGER AS $$
BEGIN
    -- Insert a feed item based on the content type
    CASE TG_TABLE_NAME
        WHEN 'entries' THEN
            INSERT INTO feed_items (id, user_id, type, reference_id, content, timestamp)
            VALUES (
                gen_random_uuid(), 
                NEW.created_by_id, 
                'entry', 
                NEW.id, 
                jsonb_build_object('word', NEW.word, 'type', NEW.type),
                CURRENT_TIMESTAMP
            );
        WHEN 'meanings' THEN
            INSERT INTO feed_items (id, user_id, type, reference_id, content, timestamp)
            VALUES (
                gen_random_uuid(), 
                NEW.created_by_id, 
                'meaning', 
                NEW.id, 
                jsonb_build_object('description', NEW.description, 'entry_id', NEW.entry_id),
                CURRENT_TIMESTAMP
            );
        WHEN 'translations' THEN
            INSERT INTO feed_items (id, user_id, type, reference_id, content, timestamp)
            VALUES (
                gen_random_uuid(), 
                NEW.created_by_id, 
                'translation', 
                NEW.id, 
                jsonb_build_object('text', NEW.text, 'language_id', NEW.language_id, 'meaning_id', NEW.meaning_id),
                CURRENT_TIMESTAMP
            );
        WHEN 'comments' THEN
            INSERT INTO feed_items (id, user_id, type, reference_id, content, timestamp)
            VALUES (
                gen_random_uuid(), 
                NEW.user_id, 
                'comment', 
                NEW.id, 
                jsonb_build_object('content', NEW.content, 'target_type', NEW.target_type, 'target_id', NEW.target_id),
                CURRENT_TIMESTAMP
            );
        WHEN 'likes' THEN
            INSERT INTO feed_items (id, user_id, type, reference_id, content, timestamp)
            VALUES (
                gen_random_uuid(), 
                NEW.user_id, 
                'like', 
                NEW.id, 
                jsonb_build_object('target_type', NEW.target_type, 'target_id', NEW.target_id),
                CURRENT_TIMESTAMP
            );
    END CASE;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for feed item creation
CREATE TRIGGER create_feed_item_on_entry_insert
AFTER INSERT ON entries
FOR EACH ROW
WHEN (NEW.created_by_id IS NOT NULL)
EXECUTE FUNCTION create_feed_item();

CREATE TRIGGER create_feed_item_on_meaning_insert
AFTER INSERT ON meanings
FOR EACH ROW
WHEN (NEW.created_by_id IS NOT NULL)
EXECUTE FUNCTION create_feed_item();

CREATE TRIGGER create_feed_item_on_translation_insert
AFTER INSERT ON translations
FOR EACH ROW
WHEN (NEW.created_by_id IS NOT NULL)
EXECUTE FUNCTION create_feed_item();

CREATE TRIGGER create_feed_item_on_comment_insert
AFTER INSERT ON comments
FOR EACH ROW
EXECUTE FUNCTION create_feed_item();

CREATE TRIGGER create_feed_item_on_like_insert
AFTER INSERT ON likes
FOR EACH ROW
EXECUTE FUNCTION create_feed_item();
