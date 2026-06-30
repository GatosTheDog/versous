package agent

import (
	"context"
	"fmt"

	"github.com/GatosTheDog/versous/internal/llm"
	"github.com/GatosTheDog/versous/internal/rag"
	"github.com/GatosTheDog/versous/internal/sources"
	"github.com/GatosTheDog/versous/internal/specs"
	"github.com/GatosTheDog/versous/internal/store"
	"golang.org/x/sync/errgroup"
)

var defaultAspects = []string{"performance", "value", "user experience"}

type Agent struct {
	llm     *llm.Client
	db      store.Store
	sources []sources.CommentSource
}

func New(llmClient *llm.Client, db store.Store, srcs ...sources.CommentSource) *Agent {
	return &Agent{llm: llmClient, db: db, sources: srcs}
}

func (a *Agent) Compare(ctx context.Context, productA, productB string, aspects []string) (Report, error) {
	if len(aspects) == 0 {
		aspects = defaultAspects
	}
	g, gctx := errgroup.WithContext(ctx)

	for _, source := range a.sources {
		src := source
		g.Go(func() error {
			return rag.Ingest(gctx, src, a.llm, a.db, productA, buildQueries(productA, aspects))
		})
		g.Go(func() error {
			return rag.Ingest(gctx, src, a.llm, a.db, productB, buildQueries(productB, aspects))
		})
	}

	if err := g.Wait(); err != nil {
		return Report{}, err
	}

	specA, _ := specs.Fetch(ctx, a.llm, productA)
	specB, _ := specs.Fetch(ctx, a.llm, productB)

	var verdicts []rag.Verdict
	for _, aspect := range aspects {
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
