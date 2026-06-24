package rag

import (
	"context"
	"os"
	"testing"

	"github.com/GatosTheDog/versous/internal/llm"
	"github.com/GatosTheDog/versous/internal/sources"
	"github.com/GatosTheDog/versous/internal/store"
)

func TestRAGPipeline(t *testing.T) {
	if os.Getenv("GEMINI_API_KEY") == "" || os.Getenv("DATABASE_URL") == "" {
		t.Skip("missing env vars")
	}

	ctx := context.Background()

	llmClient, err := llm.New(ctx)
	if err != nil {
		t.Fatal(err)
	}

	db, err := store.NewPostgres(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	hn := sources.NewHN(10)

	// ingest both products
	queriesA := []string{"iPhone 16", "iPhone 16 battery", "iPhone 16 camera"}
	if err := Ingest(ctx, hn, llmClient, db, "iPhone 16", queriesA); err != nil {
		t.Fatal(err)
	}

	queriesB := []string{"iPhone 15", "iPhone 15 battery", "iPhone 15 camera"}
	if err := Ingest(ctx, hn, llmClient, db, "iPhone 15", queriesB); err != nil {
		t.Fatal(err)
	}

	// retrieve for both
	commentsA, err := Retrieve(ctx, llmClient, db, "camera performance", "iPhone 16", 1)
	if err != nil {
		t.Fatal(err)
	}

	commentsB, err := Retrieve(ctx, llmClient, db, "camera performance", "iPhone 15", 1)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("commentsA (%d):", len(commentsA))
	for _, c := range commentsA {
		t.Logf("  [%s] %s", c.ID, c.Body)
	}
	t.Logf("commentsB (%d):", len(commentsB))
	for _, c := range commentsB {
		t.Logf("  [%s] %s", c.ID, c.Body)
	}

	verdict, err := Judge(ctx, llmClient, "camera", "iPhone 16", "iPhone 15", commentsA, commentsB)

	t.Logf("verdict: %s", verdict.Summary)
}
