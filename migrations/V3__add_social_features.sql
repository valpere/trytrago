-- Add social features to the dictionary
-- Comments, likes, and user interactions

-- Comments table
CREATE TABLE IF NOT EXISTS comments (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL,
  target_id UUID NOT NULL,
  target_type VARCHAR(20) NOT NULL,
  content TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for comments
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);
CREATE INDEX IF NOT EXISTS idx_comments_target_id ON comments(target_id);

-- Likes table
CREATE TABLE IF NOT EXISTS likes (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL,
  target_id UUID NOT NULL,
  target_type VARCHAR(20) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for likes
CREATE INDEX IF NOT EXISTS idx_likes_user_id ON likes(user_id);
CREATE INDEX IF NOT EXISTS idx_likes_target_id ON likes(target_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_likes_user_target ON likes(user_id, target_id, target_type);

-- Add foreign key constraints
ALTER TABLE comments
  ADD CONSTRAINT fk_comments_user
  FOREIGN KEY (user_id)
  REFERENCES users(id)
  ON DELETE CASCADE;

ALTER TABLE likes
  ADD CONSTRAINT fk_likes_user
  FOREIGN KEY (user_id)
  REFERENCES users(id)
  ON DELETE CASCADE;

-- Update existing tables to support social features
ALTER TABLE meanings ADD COLUMN likes_count INTEGER DEFAULT 0;
ALTER TABLE translations ADD COLUMN likes_count INTEGER DEFAULT 0;

-- Function to update likes count
CREATE OR REPLACE FUNCTION update_likes_count()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF NEW.target_type = 'meaning' THEN
      UPDATE meanings SET likes_count = likes_count + 1 WHERE id = NEW.target_id;
    ELSIF NEW.target_type = 'translation' THEN
      UPDATE translations SET likes_count = likes_count + 1 WHERE id = NEW.target_id;
    END IF;
  ELSIF TG_OP = 'DELETE' THEN
    IF OLD.target_type = 'meaning' THEN
      UPDATE meanings SET likes_count = likes_count - 1 WHERE id = OLD.target_id;
    ELSIF OLD.target_type = 'translation' THEN
      UPDATE translations SET likes_count = likes_count - 1 WHERE id = OLD.target_id;
    END IF;
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for likes
CREATE TRIGGER update_likes_count_trigger
AFTER INSERT OR DELETE ON likes
FOR EACH ROW EXECUTE FUNCTION update_likes_count();
