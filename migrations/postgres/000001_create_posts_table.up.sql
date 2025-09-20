CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    author_id VARCHAR NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    preview_url TEXT,
    post_type VARCHAR NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    last_moderation_id INTEGER,
    rating INTEGER NOT NULL DEFAULT 0,
    comments_count INTEGER NOT NULL DEFAULT 0,
    favorites_count INTEGER NOT NULL DEFAULT 0,
    search_vector tsvector GENERATED ALWAYS AS (
        to_tsvector('simple', coalesce(title, '') || ' ' || coalesce(description, ''))
    ) STORED
);

CREATE INDEX idx_posts_author_id ON posts (author_id);
CREATE INDEX idx_posts_post_type ON posts (post_type);
CREATE INDEX idx_posts_created_at ON posts (created_at);
CREATE INDEX idx_posts_updated_at ON posts (updated_at);
CREATE INDEX idx_posts_deleted_at ON posts (deleted_at);
CREATE INDEX idx_posts_search_vector ON posts USING GIN (search_vector);