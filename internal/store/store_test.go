package store

import (
	"context"
	"os"
	"testing"
)

func TestStore(t *testing.T) {
	ctx := context.Background()
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		t.Skip("no database url")
	}

	pool, err := NewPostgres(ctx, dbUrl)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	c := Comment{
		ID:        "test:post1:comment1",
		Product:   "iPhone 16",
		Source:    "reddit",
		Body:      "battery drains fast on this phone",
		Url:       "https://reddit.com/r/test/comments/abc123",
		Embedding: make([]float32, 3072), // all zeros — valid vector, just not meaningful
	}

	if err := pool.UpsertComment(ctx, c); err != nil {
		t.Fatal(err)
	}

	results, err := pool.SimilarComments(ctx, make([]float32, 3072), "iPhone 16", 5)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}

	t.Logf("got %d comments, first: %s", len(results), results[0].Body)
}
