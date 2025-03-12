-- V2__add_user_tables.sql
-- User management tables for TryTraGo dictionary database

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL, -- Stored as bcrypt hash
    avatar VARCHAR(255),
    role VARCHAR(20) NOT NULL DEFAULT 'USER',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP
);

-- Add indexes for users
CREATE INDEX idx_users_username ON users (LOWER(username));
CREATE INDEX idx_users_email ON users (LOWER(email));
CREATE INDEX idx_users_role ON users (role);

-- Create auth_tokens table
CREATE TABLE IF NOT EXISTS auth_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    access_token VARCHAR(500) NOT NULL,
    refresh_token VARCHAR(500) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP,
    user_agent VARCHAR(255),
    client_ip VARCHAR(45)
);

-- Add indexes for auth_tokens
CREATE INDEX idx_auth_tokens_user_id ON auth_tokens (user_id);
CREATE INDEX idx_auth_tokens_refresh_token ON auth_tokens (refresh_token);
CREATE INDEX idx_auth_tokens_expires_at ON auth_tokens (expires_at);

-- Create user_preferences table
CREATE TABLE IF NOT EXISTS user_preferences (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    default_language VARCHAR(5) NOT NULL DEFAULT 'en' REFERENCES languages(code),
    theme_preference VARCHAR(20) NOT NULL DEFAULT 'system',
    email_notify BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create user_stats table
CREATE TABLE IF NOT EXISTS user_stats (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    entries_created INT NOT NULL DEFAULT 0,
    entries_updated INT NOT NULL DEFAULT 0,
    meanings_added INT NOT NULL DEFAULT 0,
    translations_added INT NOT NULL DEFAULT 0,
    comments_posted INT NOT NULL DEFAULT 0,
    likes_given INT NOT NULL DEFAULT 0,
    reputation_points INT NOT NULL DEFAULT 0,
    last_activity_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Update foreign keys in existing tables
-- Add foreign key constraints for created_by_id fields
ALTER TABLE entries
    ADD CONSTRAINT fk_entries_created_by FOREIGN KEY (created_by_id) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE meanings
    ADD CONSTRAINT fk_meanings_created_by FOREIGN KEY (created_by_id) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE examples
    ADD CONSTRAINT fk_examples_created_by FOREIGN KEY (created_by_id) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE translations
    ADD CONSTRAINT fk_translations_created_by FOREIGN KEY (created_by_id) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE change_history
    ADD CONSTRAINT fk_change_history_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

-- Add language foreign key constraints
ALTER TABLE entries
    ADD CONSTRAINT fk_entries_source_language FOREIGN KEY (source_language_id) REFERENCES languages(code) ON DELETE SET NULL;

-- Create admin user (password: admin123)
INSERT INTO users (
    id, 
    username, 
    email, 
    password, 
    avatar, 
    role, 
    is_active, 
    created_at, 
    updated_at
) VALUES (
    gen_random_uuid(),
    'admin',
    'admin@trytrago.com',
    '$2a$10$dBR5d8VTLjQvQOPiwbHCzuQUEVLvtvVSbG2pJUT3c4DHmfVCJNpou', -- 'admin123' hashed with bcrypt
    '',
    'ADMIN',
    TRUE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (username) DO NOTHING;

-- Create default user (password: password123)
INSERT INTO users (
    id, 
    username, 
    email, 
    password, 
    avatar, 
    role, 
    is_active, 
    created_at, 
    updated_at
) VALUES (
    gen_random_uuid(),
    'user',
    'user@trytrago.com',
    '$2a$10$dBR5d8VTLjQvQOPiwbHCzuQUEVLvtvVSbG2pJUT3c4DHmfVCJNpou', -- 'password123' hashed with bcrypt
    '',
    'USER',
    TRUE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (username) DO NOTHING;
