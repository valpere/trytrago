-- V4__performance_optimizations.sql
-- Performance optimizations for scaling to 60 million entries

-- Add table partitioning for the entries table
-- This assumes PostgreSQL 10+ which has declarative partitioning
-- For large dictionaries, we'll partition by the first letter of the word

-- First, create a function to extract the first letter
CREATE OR REPLACE FUNCTION first_letter(word TEXT) 
RETURNS TEXT AS $$
BEGIN
    RETURN LOWER(LEFT(word, 1));
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Create partial indices for each common first letter to optimize queries
-- This approach is more flexible than partitioning and works with all database types
CREATE INDEX idx_entries_word_a ON entries (word) WHERE first_letter(word) = 'a';
CREATE INDEX idx_entries_word_b ON entries (word) WHERE first_letter(word) = 'b';
CREATE INDEX idx_entries_word_c ON entries (word) WHERE first_letter(word) = 'c';
CREATE INDEX idx_entries_word_d ON entries (word) WHERE first_letter(word) = 'd';
CREATE INDEX idx_entries_word_e ON entries (word) WHERE first_letter(word) = 'e';
CREATE INDEX idx_entries_word_f ON entries (word) WHERE first_letter(word) = 'f';
CREATE INDEX idx_entries_word_g ON entries (word) WHERE first_letter(word) = 'g';
CREATE INDEX idx_entries_word_h ON entries (word) WHERE first_letter(word) = 'h';
CREATE INDEX idx_entries_word_i ON entries (word) WHERE first_letter(word) = 'i';
CREATE INDEX idx_entries_word_j ON entries (word) WHERE first_letter(word) = 'j';
CREATE INDEX idx_entries_word_k ON entries (word) WHERE first_letter(word) = 'k';
CREATE INDEX idx_entries_word_l ON entries (word) WHERE first_letter(word) = 'l';
CREATE INDEX idx_entries_word_m ON entries (word) WHERE first_letter(word) = 'm';
CREATE INDEX idx_entries_word_n ON entries (word) WHERE first_letter(word) = 'n';
CREATE INDEX idx_entries_word_o ON entries (word) WHERE first_letter(word) = 'o';
CREATE INDEX idx_entries_word_p ON entries (word) WHERE first_letter(word) = 'p';
CREATE INDEX idx_entries_word_q ON entries (word) WHERE first_letter(word) = 'q';
CREATE INDEX idx_entries_word_r ON entries (word) WHERE first_letter(word) = 'r';
CREATE INDEX idx_entries_word_s ON entries (word) WHERE first_letter(word) = 's';
CREATE INDEX idx_entries_word_t ON entries (word) WHERE first_letter(word) = 't';
CREATE INDEX idx_entries_word_u ON entries (word) WHERE first_letter(word) = 'u';
CREATE INDEX idx_entries_word_v ON entries (word) WHERE first_letter(word) = 'v';
CREATE INDEX idx_entries_word_w ON entries (word) WHERE first_letter(word) = 'w';
CREATE INDEX idx_entries_word_x ON entries (word) WHERE first_letter(word) = 'x';
CREATE INDEX idx_entries_word_y ON entries (word) WHERE first_letter(word) = 'y';
CREATE INDEX idx_entries_word_z ON entries (word) WHERE first_letter(word) = 'z';
CREATE INDEX idx_entries_word_other ON entries (word) 
    WHERE first_letter(word) NOT BETWEEN 'a' AND 'z';

-- Create a composite index for entry lookups by word and type
CREATE INDEX idx_entries_word_type ON entries (word, type);

-- Create a composite index for translations by language and meaning
CREATE INDEX idx_translations_language_meaning ON translations (language_id, meaning_id);

-- Optimize for common queries using partial indices
CREATE INDEX idx_meanings_active ON meanings (entry_id, id) 
    WHERE id IN (SELECT meaning_id FROM translations GROUP BY meaning_id HAVING COUNT(*) > 0);

-- Add index for pagination queries
CREATE INDEX idx_entries_updated_id ON entries (updated_at DESC, id);

-- Improve foreign key lookup performance
CREATE INDEX idx_meanings_entry_created ON meanings (entry_id, created_at DESC);
CREATE INDEX idx_translations_meaning_created ON translations (meaning_id, created_at DESC);

-- Optimize text search for entries
CREATE INDEX idx_entries_word_tsvector ON entries USING gin(to_tsvector('english', word));

-- Create materialized view for common language pairs
CREATE MATERIALIZED VIEW common_translations AS
SELECT 
    e.word AS source_word,
    e.id AS entry_id,
    m.id AS meaning_id,
    m.description,
    t.language_id,
    t.text AS translation
FROM 
    entries e
JOIN 
    meanings m ON e.id = m.entry_id
JOIN 
    translations t ON m.id = t.meaning_id
WHERE 
    t.language_id IN ('en', 'es', 'fr', 'de', 'ru', 'zh')
WITH DATA;

-- Add indices to the materialized view
CREATE INDEX idx_common_translations_source ON common_translations (source_word);
CREATE INDEX idx_common_translations_language ON common_translations (language_id);
CREATE INDEX idx_common_translations_translation ON common_translations USING gin(to_tsvector('english', translation));

-- Add database statistics collection for query optimization
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Create a function to refresh the materialized view periodically
CREATE OR REPLACE FUNCTION refresh_common_translations()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY common_translations;
END;
$$ LANGUAGE plpgsql;

-- Create a function to analyze tables periodically
CREATE OR REPLACE FUNCTION analyze_dictionary_tables()
RETURNS void AS $$
BEGIN
    ANALYZE entries;
    ANALYZE meanings;
    ANALYZE translations;
    ANALYZE common_translations;
END;
$$ LANGUAGE plpgsql;

-- Add database configuration for large dataset
-- These would typically be set in postgresql.conf, but we include them here for documentation
-- ALTER SYSTEM SET shared_buffers = '1GB';
-- ALTER SYSTEM SET effective_cache_size = '3GB';
-- ALTER SYSTEM SET maintenance_work_mem = '256MB';
-- ALTER SYSTEM SET work_mem = '20MB';
-- ALTER SYSTEM SET max_connections = '200';
-- ALTER SYSTEM SET random_page_cost = '1.1';
-- ALTER SYSTEM SET effective_io_concurrency = '200';

-- Add comment to document optimization
COMMENT ON DATABASE :dbname IS 'TryTraGo dictionary database optimized for 60M entries';
