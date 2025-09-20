CREATE TABLE moderation_actions (
   id SERIAL PRIMARY KEY,
   post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
   moderator_id VARCHAR NOT NULL,
   action VARCHAR NOT NULL,
   note TEXT,
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_moderation_actions_post_id ON moderation_actions (post_id);
CREATE INDEX idx_moderation_actions_moderator_id ON moderation_actions (moderator_id);
CREATE INDEX idx_moderation_actions_created_at ON moderation_actions (created_at);
