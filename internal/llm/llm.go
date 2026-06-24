package llm

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/genai"
)

type Client struct {
	inner *genai.Client
}

func New(ctx context.Context) (*Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not set")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("new client: %w", err)
	}

	return &Client{inner: client}, nil
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	result, err := c.inner.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(prompt), nil)

	if err != nil {
		return "", fmt.Errorf("generate: %w", err)
	}

	return result.Text(), nil
}

func (c *Client) Embed(ctx context.Context, text string) ([]float32, error) {
	contents := []*genai.Content{
		genai.NewContentFromText(text, "user"),
	}

	resp, err := c.inner.Models.EmbedContent(ctx, "gemini-embedding-001", contents, nil)
	if err != nil {
		return nil, fmt.Errorf("embed: %w", err)
	}

	if len(resp.Embeddings) == 0 {
		return nil, fmt.Errorf("embed: empty response")
	}

	return resp.Embeddings[0].Values, nil
}
