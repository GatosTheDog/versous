package agent

import (
	"context"
	"fmt"

	"github.com/GatosTheDog/versous/internal/llm"
	"github.com/GatosTheDog/versous/internal/rag"
	"github.com/GatosTheDog/versous/internal/sources"
	"github.com/GatosTheDog/versous/internal/specs"
	"github.com/GatosTheDog/versous/internal/store"
)

var defaultAspects = []string{"Battery Life", "Camera Quality", "Performance", "Price"}

type Agent struct {
	llm     *llm.Client
	db      store.Store
	sources []sources.CommentSource
}

func New(llmClient *llm.Client, db store.Store, srcs ...sources.CommentSource) *Agent {
	return &Agent{llm: llmClient, db: db, sources: srcs}
}

func (a *Agent) Compare(ctx context.Context, productA, productB string) (Report, error) {

	for _, source := range a.sources {
		if err := rag.Ingest(ctx, source, a.llm, a.db, productA, buildQueries(productA, defaultAspects)); err != nil {
			return Report{}, fmt.Errorf("ingest %s: %w", productA, err)
		}
		if err := rag.Ingest(ctx, source, a.llm, a.db, productB, buildQueries(productB, defaultAspects)); err != nil {
			return Report{}, fmt.Errorf("ingest %s: %w", productB, err)
		}
	}

	specA, _ := specs.Lookup(productA)
	specB, _ := specs.Lookup(productB)

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
		SpecA:    specA,
		SpecB:    specB,

		Winner: tally(verdicts, productA, productB),
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

func buildQueries(product string, aspects []string) []string {
	query := make([]string, 0, len(aspects))
	for _, aspect := range aspects {
		query = append(query, fmt.Sprintf("%s %s", product, aspect))
	}
	return query
}
