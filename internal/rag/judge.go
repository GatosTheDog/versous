package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/GatosTheDog/versous/internal/llm"
	"github.com/GatosTheDog/versous/internal/store"
	"google.golang.org/genai"
)

type Verdict struct {
	Aspect      string
	Winner      string
	Summary     string
	Weakness    string
	LoserUpside string
	Evidence    []store.Comment
}

type verdictJSON struct {
	Winner      string `json:"winner"`
	Strength    string `json:"strength"`
	Weakness    string `json:"weakness"`
	LoserUpside string `json:"loser_upside"`
}

func Judge(ctx context.Context, llmClient *llm.Client, aspect string, productA, productB string, commentsA, commentsB []store.Comment) (Verdict, error) {
	prompt := buildPrompt(aspect, productA, productB, commentsA, commentsB)

	result, err := llmClient.GenerateStructured(ctx, prompt, buildSchema())
	if err != nil {
		return Verdict{}, fmt.Errorf("judge: %w", err)
	}

	var v verdictJSON
	if err := json.Unmarshal([]byte(result), &v); err != nil {
		return Verdict{}, fmt.Errorf("judge: parse response: %w", err)
	}

	var winner string
	switch v.Winner {
	case "B":
		winner = productB
	case "Tie":
		winner = "Tie"
	default:
		winner = productA
	}

	all := make([]store.Comment, 0, len(commentsA)+len(commentsB))
	all = append(all, commentsA...)
	all = append(all, commentsB...)

	return Verdict{
		Aspect:      aspect,
		Winner:      winner,
		Summary:     v.Strength,
		Weakness:    v.Weakness,
		LoserUpside: v.LoserUpside,
		Evidence:    all,
	}, nil
}

func buildPrompt(aspect, productA, productB string, commentsA, commentsB []store.Comment) string {
	var b strings.Builder

	fmt.Fprintf(&b, "You are comparing %s vs %s on the aspect: %s\n\n", productA, productB, aspect)

	fmt.Fprintf(&b, "User comments about %s:\n", productA)
	for _, c := range commentsA {
		fmt.Fprintf(&b, "- %s\n", c.Body)
	}

	fmt.Fprintf(&b, "\nUser comments about %s:\n", productB)
	for _, c := range commentsB {
		fmt.Fprintf(&b, "- %s\n", c.Body)
	}

	fmt.Fprintf(&b, `
		Based on these real user comments, return a JSON object with:
		- "winner": "A" if %s wins, "B" if %s wins, or "Tie"
		- "strength": one strength of the winner, quoting a real comment
		- "weakness": one weakness of the winner, quoting a real comment
		- "loser_upside": one sentence on what the losing product does better
		Be brutally concise. No intros, no disclaimers.`, productA, productB)

	return b.String()
}

func buildSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"winner":       {Type: genai.TypeString, Enum: []string{"A", "B", "Tie"}},
			"strength":     {Type: genai.TypeString},
			"weakness":     {Type: genai.TypeString},
			"loser_upside": {Type: genai.TypeString},
		},
		Required: []string{"winner", "strength", "weakness", "loser_upside"},
	}
}
