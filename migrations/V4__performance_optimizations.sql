-- Performance optimization migration

-- Add active column to entries
ALTER TABLE entries ADD COLUMN IF NOT EXISTS active BOOLEAN DEFAULT TRUE;

-- Add indexes for faster lookup
CREATE INDEX IF NOT EXISTS idx_entries_word ON entries(word);
CREATE INDEX IF NOT EXISTS idx_entries_word_type ON entries(word, type);
CREATE INDEX IF NOT EXISTS idx_meanings_entry_id ON meanings(entry_id);
CREATE INDEX IF NOT EXISTS idx_translations_meaning_id ON translations(meaning_id);
CREATE INDEX IF NOT EXISTS idx_translations_language_id ON translations(language_id);
CREATE INDEX IF NOT EXISTS idx_comments_target_id ON comments(target_id);
CREATE INDEX IF NOT EXISTS idx_likes_target_id ON likes(target_id);

-- Add partial indexes for common queries
CREATE INDEX IF NOT EXISTS idx_entries_active ON entries(created_at) WHERE active = true;

-- Add column for caching counts
ALTER TABLE entries ADD COLUMN IF NOT EXISTS meaning_count INTEGER DEFAULT 0;
ALTER TABLE meanings ADD COLUMN IF NOT EXISTS example_count INTEGER DEFAULT 0;
ALTER TABLE meanings ADD COLUMN IF NOT EXISTS translation_count INTEGER DEFAULT 0;
ALTER TABLE meanings ADD COLUMN IF NOT EXISTS comments_count INTEGER DEFAULT 0;
ALTER TABLE translations ADD COLUMN IF NOT EXISTS comments_count INTEGER DEFAULT 0;

-- Function to update meaning counts
CREATE OR REPLACE FUNCTION update_entry_meaning_count()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    UPDATE entries SET meaning_count = meaning_count + 1 WHERE id = NEW.entry_id;
  ELSIF TG_OP = 'DELETE' THEN
    UPDATE entries SET meaning_count = meaning_count - 1 WHERE id = OLD.entry_id;
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Function to update example counts
CREATE OR REPLACE FUNCTION update_meaning_example_count()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    UPDATE meanings SET example_count = example_count + 1 WHERE id = NEW.meaning_id;
  ELSIF TG_OP = 'DELETE' THEN
    UPDATE meanings SET example_count = example_count - 1 WHERE id = OLD.meaning_id;
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Function to update translation counts
CREATE OR REPLACE FUNCTION update_meaning_translation_count()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    UPDATE meanings SET translation_count = translation_count + 1 WHERE id = NEW.meaning_id;
  ELSIF TG_OP = 'DELETE' THEN
    UPDATE meanings SET translation_count = translation_count - 1 WHERE id = OLD.meaning_id;
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create triggers
CREATE TRIGGER update_entry_meaning_count_trigger
AFTER INSERT OR DELETE ON meanings
FOR EACH ROW EXECUTE FUNCTION update_entry_meaning_count();

CREATE TRIGGER update_meaning_example_count_trigger
AFTER INSERT OR DELETE ON examples
FOR EACH ROW EXECUTE FUNCTION update_meaning_example_count();

CREATE TRIGGER update_meaning_translation_count_trigger
AFTER INSERT OR DELETE ON translations
FOR EACH ROW EXECUTE FUNCTION update_meaning_translation_count();

-- Add updated_at column triggers
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_entries_modified
BEFORE UPDATE ON entries
FOR EACH ROW EXECUTE FUNCTION update_modified_column();

CREATE TRIGGER update_meanings_modified
BEFORE UPDATE ON meanings
FOR EACH ROW EXECUTE FUNCTION update_modified_column();

CREATE TRIGGER update_translations_modified
BEFORE UPDATE ON translations
FOR EACH ROW EXECUTE FUNCTION update_modified_column();
