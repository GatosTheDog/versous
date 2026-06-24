package agent

import (
	"context"
	"fmt"

	"github.com/GatosTheDog/versous/internal/llm"
	"github.com/GatosTheDog/versous/internal/rag"
	"github.com/GatosTheDog/versous/internal/sources"
	"github.com/GatosTheDog/versous/internal/store"
)

var defaultAspects = []string{"camera", "battery", "value"}

type Agent struct {
	llm    *llm.Client
	db     store.Store
	source sources.CommentSource
}

func New(llmClient *llm.Client, db store.Store, src sources.CommentSource) *Agent {
	return &Agent{llm: llmClient, db: db, source: src}
}

func (a *Agent) Compare(ctx context.Context, productA, productB string) (Report, error) {
	queriesA := []string{productA}
	queriesB := []string{productB}

	if err := rag.Ingest(ctx, a.source, a.llm, a.db, productA, queriesA); err != nil {
		return Report{}, fmt.Errorf("ingest %s: %w", productA, err)
	}
	if err := rag.Ingest(ctx, a.source, a.llm, a.db, productB, queriesB); err != nil {
		return Report{}, fmt.Errorf("ingest %s: %w", productB, err)
	}

	var verdicts []rag.Verdict
	for _, aspect := range defaultAspects {
		commentsA, err := rag.Retrieve(ctx, a.llm, a.db, aspect, productA, 2)
		if err != nil {
			return Report{}, fmt.Errorf("retrieve %s/%s: %w", productA, aspect, err)
		}
		commentsB, err := rag.Retrieve(ctx, a.llm, a.db, aspect, productB, 2)
		if err != nil {
			return Report{}, fmt.Errorf("retrieve %s/%s: %w", productB, aspect, err)
		}

		verdict, err := rag.Judge(ctx, a.llm, aspect, productA, productB, commentsA, commentsB)
		if err != nil {
			return Report{}, fmt.Errorf("judge %s: %w", aspect, err)
		}
		verdicts = append(verdicts, verdict)
	}

	return Report{
		ProductA: productA,
		ProductB: productB,
		Aspects:  verdicts,
		Winner:   tally(verdicts, productA, productB),
	}, nil
}

func tally(verdicts []rag.Verdict, productA, productB string) string {
	votes := map[string]int{productA: 0, productB: 0}
	for _, v := range verdicts {
		votes[v.Winner]++
	}
	if votes[productA] >= votes[productB] {
		return productA
	}
	return productB
}
