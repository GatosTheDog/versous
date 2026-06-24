package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/GatosTheDog/versous/internal/llm"
	"github.com/GatosTheDog/versous/internal/store"
)

type Verdict struct {
	Aspect   string
	Winner   string
	Summary  string
	Evidence []string
}

func Judge(ctx context.Context, llmClient *llm.Client, aspect string, productA, productB string, commentsA, commentsB []store.Comment) (Verdict, error) {
	prompt := buildPrompt(aspect, productA, productB, commentsA, commentsB)

	result, err := llmClient.Generate(ctx, prompt)
	if err != nil {
		return Verdict{}, fmt.Errorf("judge: %w", err)
	}

	winner := productA
	lower := strings.ToLower(result)
	if strings.Contains(lower, strings.ToLower(productB)) &&
		!strings.Contains(lower, strings.ToLower(productA)) {
		winner = productB
	}

	all := make([]store.Comment, 0, len(commentsA)+len(commentsB))
	all = append(all, commentsA...)
	all = append(all, commentsB...)
	urls := make([]string, 0, len(all))
	for _, c := range all {
		urls = append(urls, c.Url)
	}

	return Verdict{
		Aspect:   aspect,
		Winner:   winner,
		Summary:  result,
		Evidence: urls,
	}, nil
}

func buildPrompt(aspect, productA, productB string, commentsA, commentsB []store.Comment) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("You are comparing %s vs %s on the aspect: %s\n\n", productA, productB, aspect))

	b.WriteString(fmt.Sprintf("User comments about %s:\n", productA))
	for _, c := range commentsA {
		b.WriteString(fmt.Sprintf("- %s\n", c.Body))
	}

	b.WriteString(fmt.Sprintf("\nUser comments about %s:\n", productB))
	for _, c := range commentsB {
		b.WriteString(fmt.Sprintf("- %s\n", c.Body))
	}

	b.WriteString(fmt.Sprintf(`
		Based on these real user comments, answer:
		1. Which product (%s or %s) is better for %s, and why?
		2. Summarise the key user sentiments in 2-3 sentences.
		Be direct. Cite specific comments where possible.`, productA, productB, aspect))

	return b.String()

}
