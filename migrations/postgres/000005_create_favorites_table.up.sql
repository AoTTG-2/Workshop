CREATE TABLE favorites
(
    id         SERIAL PRIMARY KEY,
    post_id    INTEGER     NOT NULL REFERENCES posts (id) ON DELETE CASCADE,
    user_id    VARCHAR     NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT favorites_user_post_unique UNIQUE (user_id, post_id)
);

CREATE INDEX idx_favorites_post_id ON favorites (post_id);
