ALTER TABLE comments ALTER COLUMN embedding TYPE halfvec(3072);

CREATE INDEX ON comments USING hnsw (embedding halfvec_cosine_ops);
