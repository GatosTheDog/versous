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

	// ingest iPhone 16 comments
	if err := Ingest(ctx, hn, llmClient, db, "iPhone 16"); err != nil {
		t.Fatal(err)
	}

	// retrieve comments about battery
	comments, err := Retrieve(ctx, llmClient, db, "iPhone 16 battery life", "iPhone 16", 10)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("retrieved %d comments for battery query", len(comments))

	// judge battery aspect (single product for now — full compare in M6)
	verdict, err := Judge(ctx, llmClient, "iPhone 16 battery life", "iPhone 16", "Pixel 9", comments, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("verdict: %s", verdict.Summary)
}
