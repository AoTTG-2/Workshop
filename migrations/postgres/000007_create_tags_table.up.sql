CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL UNIQUE
);

CREATE UNIQUE INDEX idx_tags_name ON tags (name);
