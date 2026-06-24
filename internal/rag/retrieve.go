package rag

import (
	"context"
	"fmt"

	"github.com/GatosTheDog/versous/internal/llm"
	"github.com/GatosTheDog/versous/internal/store"
)

func Retrieve(ctx context.Context, llmClient *llm.Client, db store.Store, query string, product string, limit int) ([]store.Comment, error) {
	vec, err := llmClient.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	comments, err := db.SimilarComments(ctx, vec, product, limit)
	if err != nil {
		return nil, fmt.Errorf("retrieve: %w", err)
	}

	return comments, nil
}
