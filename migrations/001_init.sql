CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS comments (
    id          TEXT PRIMARY KEY,        -- source:post_id:comment_id
    product     TEXT NOT NULL,           -- e.g. "iPhone 16"
    source      TEXT NOT NULL,           -- "reddit", "youtube", "hn"
    body        TEXT NOT NULL,           -- raw comment text
    url         TEXT NOT NULL,           -- link to original comment
    fetched_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    embedding   vector(3072)             -- gemini-embedding-001 output
);
