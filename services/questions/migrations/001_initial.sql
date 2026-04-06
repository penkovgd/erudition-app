CREATE TABLE IF NOT EXISTS topics (
    slug VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT
);
CREATE TABLE IF NOT EXISTS questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic_slug VARCHAR(255) REFERENCES topics(slug) ON DELETE CASCADE,
    bloom_level INT NOT NULL,
    anchor VARCHAR(255) NOT NULL,
    context_path TEXT NOT NULL,
    context_hash VARCHAR(64) UNIQUE NOT NULL,
    stem TEXT NOT NULL,
    options JSONB NOT NULL,
    correct_key CHAR(1) NOT NULL,
    justification TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_questions_topic ON questions(topic_slug);
CREATE INDEX IF NOT EXISTS idx_questions_bloom ON questions(bloom_level);