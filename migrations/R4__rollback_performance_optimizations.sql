-- R4__rollback_performance_optimizations.sql
-- Rollback script for performance optimizations

-- Drop materialized view and its indices
DROP MATERIALIZED VIEW IF EXISTS common_translations CASCADE;

-- Drop custom functions
DROP FUNCTION IF EXISTS refresh_common_translations();
DROP FUNCTION IF EXISTS analyze_dictionary_tables();
DROP FUNCTION IF EXISTS first_letter(TEXT);

-- Drop all the letter-based partial indices
DROP INDEX IF EXISTS idx_entries_word_a;
DROP INDEX IF EXISTS idx_entries_word_b;
DROP INDEX IF EXISTS idx_entries_word_c;
DROP INDEX IF EXISTS idx_entries_word_d;
DROP INDEX IF EXISTS idx_entries_word_e;
DROP INDEX IF EXISTS idx_entries_word_f;
DROP INDEX IF EXISTS idx_entries_word_g;
DROP INDEX IF EXISTS idx_entries_word_h;
DROP INDEX IF EXISTS idx_entries_word_i;
DROP INDEX IF EXISTS idx_entries_word_j;
DROP INDEX IF EXISTS idx_entries_word_k;
DROP INDEX IF EXISTS idx_entries_word_l;
DROP INDEX IF EXISTS idx_entries_word_m;
DROP INDEX IF EXISTS idx_entries_word_n;
DROP INDEX IF EXISTS idx_entries_word_o;
DROP INDEX IF EXISTS idx_entries_word_p;
DROP INDEX IF EXISTS idx_entries_word_q;
DROP INDEX IF EXISTS idx_entries_word_r;
DROP INDEX IF EXISTS idx_entries_word_s;
DROP INDEX IF EXISTS idx_entries_word_t;
DROP INDEX IF EXISTS idx_entries_word_u;
DROP INDEX IF EXISTS idx_entries_word_v;
DROP INDEX IF EXISTS idx_entries_word_w;
DROP INDEX IF EXISTS idx_entries_word_x;
DROP INDEX IF EXISTS idx_entries_word_y;
DROP INDEX IF EXISTS idx_entries_word_z;
DROP INDEX IF EXISTS idx_entries_word_other;

-- Drop other optimization indices
DROP INDEX IF EXISTS idx_entries_word_type;
DROP INDEX IF EXISTS idx_translations_language_meaning;
DROP INDEX IF EXISTS idx_meanings_active;
DROP INDEX IF EXISTS idx_entries_updated_id;
DROP INDEX IF EXISTS idx_meanings_entry_created;
DROP INDEX IF EXISTS idx_translations_meaning_created;
DROP INDEX IF EXISTS idx_entries_word_tsvector;
DROP INDEX IF EXISTS idx_common_translations_source;
DROP INDEX IF EXISTS idx_common_translations_language;
DROP INDEX IF EXISTS idx_common_translations_translation;

-- Drop PostgreSQL extension (if necessary)
-- Note: This might require admin privileges
DROP EXTENSION IF EXISTS pg_stat_statements;
