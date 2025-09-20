CREATE TABLE post_contents (
   id SERIAL PRIMARY KEY,
   post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
   content_type VARCHAR NOT NULL,
   content_data TEXT NOT NULL,
   is_link BOOLEAN NOT NULL DEFAULT TRUE,
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   updated_at TIMESTAMPTZ
);

CREATE INDEX idx_post_contents_post_id ON post_contents (post_id);
