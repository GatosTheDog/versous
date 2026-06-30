package rag

import (
	"context"
	"fmt"

	"github.com/GatosTheDog/versous/internal/llm"
	"github.com/GatosTheDog/versous/internal/sources"
	"github.com/GatosTheDog/versous/internal/store"
)

func Ingest(ctx context.Context, src sources.CommentSource, llmClient *llm.Client, db store.Store, product string, queries []string) error {

	for _, q := range queries {
		comments, err := src.Fetch(ctx, q)
		if err != nil {
			return fmt.Errorf("fetch: %w", err)
		}
		for _, c := range comments {
			vec, err := llmClient.EmbedDocument(ctx, c.Body)
			if err != nil {
				return fmt.Errorf("embed comment %s: %w", c.ID, err)
			}
			c.Embedding = vec
			c.Product = product

			if err := db.UpsertComment(ctx, c); err != nil {
				return fmt.Errorf("upsert comment %s: %w", c.ID, err)

			}
		}
	}

	return nil
}
