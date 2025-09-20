CREATE TABLE votes
(
    id       SERIAL PRIMARY KEY,
    post_id  INTEGER NOT NULL REFERENCES posts (id) ON DELETE CASCADE,
    voter_id VARCHAR NOT NULL,
    vote     INTEGER NOT NULL CHECK (vote IN (1, -1)),
    CONSTRAINT votes_post_voter_unique UNIQUE (post_id, voter_id)
);

CREATE INDEX idx_votes_voter_id ON votes (voter_id);
CREATE INDEX idx_votes_post_id ON votes (post_id);
