-- V1__initial_schema.sql
-- Initial schema for TryTraGo dictionary database

-- Create parts_of_speech table
CREATE TABLE IF NOT EXISTS parts_of_speech (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create entries table
CREATE TABLE IF NOT EXISTS entries (
    id UUID PRIMARY KEY,
    word VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('WORD', 'COMPOUND_WORD', 'PHRASE')),
    pronunciation VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by_id UUID,
    source_language_id VARCHAR(5)
);

-- Add index for word lookups (case-insensitive)
CREATE INDEX idx_entries_word ON entries (LOWER(word));
CREATE INDEX idx_entries_type ON entries (type);
CREATE INDEX idx_entries_created_at ON entries (created_at DESC);

-- Create meanings table
CREATE TABLE IF NOT EXISTS meanings (
    id UUID PRIMARY KEY,
    entry_id UUID NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    part_of_speech_id UUID NOT NULL REFERENCES parts_of_speech(id),
    description TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by_id UUID
);

-- Add indexes for meanings
CREATE INDEX idx_meanings_entry_id ON meanings (entry_id);
CREATE INDEX idx_meanings_part_of_speech_id ON meanings (part_of_speech_id);

-- Create examples table
CREATE TABLE IF NOT EXISTS examples (
    id UUID PRIMARY KEY,
    meaning_id UUID NOT NULL REFERENCES meanings(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    context TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by_id UUID
);

-- Add index for examples
CREATE INDEX idx_examples_meaning_id ON examples (meaning_id);

-- Create translations table
CREATE TABLE IF NOT EXISTS translations (
    id UUID PRIMARY KEY,
    meaning_id UUID NOT NULL REFERENCES meanings(id) ON DELETE CASCADE,
    language_id VARCHAR(5) NOT NULL,  -- ISO 639-1 code
    text TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by_id UUID
);

-- Add indexes for translations
CREATE INDEX idx_translations_meaning_id ON translations (meaning_id);
CREATE INDEX idx_translations_language_id ON translations (language_id);
CREATE INDEX idx_translations_text ON translations USING gin(to_tsvector('english', text));

-- Create languages table for reference
CREATE TABLE IF NOT EXISTS languages (
    code VARCHAR(5) PRIMARY KEY,  -- ISO 639-1 code
    name VARCHAR(100) NOT NULL,
    native_name VARCHAR(100) NOT NULL,
    rtl BOOLEAN NOT NULL DEFAULT FALSE,
    active BOOLEAN NOT NULL DEFAULT TRUE
);

-- Create change_history table
CREATE TABLE IF NOT EXISTS change_history (
    id UUID PRIMARY KEY,
    entry_id UUID NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    action VARCHAR(20) NOT NULL,
    data JSONB NOT NULL,
    user_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add index for change history
CREATE INDEX idx_change_history_entry_id ON change_history (entry_id);
CREATE INDEX idx_change_history_created_at ON change_history (created_at DESC);

-- Insert default parts of speech
INSERT INTO parts_of_speech (id, name, created_at, updated_at) VALUES 
    (gen_random_uuid(), 'noun', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'verb', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'adjective', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'adverb', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'pronoun', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'preposition', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'conjunction', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'interjection', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'article', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'numeral', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'determiner', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (gen_random_uuid(), 'particle', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (name) DO NOTHING;

-- Insert default languages
INSERT INTO languages (code, name, native_name, rtl, active) VALUES
    ('en', 'English', 'English', FALSE, TRUE),
    ('es', 'Spanish', 'Español', FALSE, TRUE),
    ('fr', 'French', 'Français', FALSE, TRUE),
    ('de', 'German', 'Deutsch', FALSE, TRUE),
    ('it', 'Italian', 'Italiano', FALSE, TRUE),
    ('pt', 'Portuguese', 'Português', FALSE, TRUE),
    ('ru', 'Russian', 'Русский', FALSE, TRUE),
    ('zh', 'Chinese', '中文', FALSE, TRUE),
    ('ja', 'Japanese', '日本語', FALSE, TRUE),
    ('ko', 'Korean', '한국어', FALSE, TRUE),
    ('ar', 'Arabic', 'العربية', TRUE, TRUE),
    ('hi', 'Hindi', 'हिन्दी', FALSE, TRUE),
    ('tr', 'Turkish', 'Türkçe', FALSE, TRUE),
    ('nl', 'Dutch', 'Nederlands', FALSE, TRUE),
    ('sv', 'Swedish', 'Svenska', FALSE, TRUE),
    ('pl', 'Polish', 'Polski', FALSE, TRUE),
    ('uk', 'Ukrainian', 'Українська', FALSE, TRUE)
ON CONFLICT (code) DO NOTHING;
