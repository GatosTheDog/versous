package llm

import (
	"context"
	"fmt"
	"os"
	"time"

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
	var lastErr error
	for attempt := range 3 {
		callCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		result, err := c.inner.Models.GenerateContent(callCtx, "gemini-3.1-flash-lite", genai.Text(prompt), nil)
		cancel()

		if err == nil {
			if result == nil {
				lastErr = fmt.Errorf("nil response from model")
				continue
			}
			text := result.Text()
			if text == "" {
				lastErr = fmt.Errorf("empty response from model")
				continue
			}
			return text, nil
		}

		lastErr = err

		backoff := time.Duration(1<<(attempt+1)) * time.Second
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	return "", fmt.Errorf("generate (3 attempts): %w", lastErr)
}

func (c *Client) EmbedDocument(ctx context.Context, text string) ([]float32, error) {
	return c.embed(ctx, text, "RETRIEVAL_DOCUMENT")
}

func (c *Client) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return c.embed(ctx, text, "RETRIEVAL_QUERY")
}

func (c *Client) embed(ctx context.Context, text, taskType string) ([]float32, error) {
	contents := []*genai.Content{
		genai.NewContentFromText(text, "user"),
	}

	var lastErr error
	for attempt := range 3 {
		resp, err := c.inner.Models.EmbedContent(ctx, "gemini-embedding-001", contents, &genai.EmbedContentConfig{TaskType: taskType})
		if err == nil {
			if len(resp.Embeddings) == 0 {
				lastErr = fmt.Errorf("embed: empty response")
				continue
			}
			return resp.Embeddings[0].Values, nil
		}

		lastErr = err

		backoff := time.Duration(1<<(attempt+1)) * time.Second
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("embed (3 attempts): %w", lastErr)
}

func (c *Client) GenerateStructured(ctx context.Context, prompt string, schema *genai.Schema) (string, error) {
	var lastErr error
	for attempt := range 3 {
		callCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		result, err := c.inner.Models.GenerateContent(callCtx, "gemini-3.1-flash-lite", genai.Text(prompt), &genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema:   schema,
		})
		cancel()

		if err == nil {
			if result == nil {
				lastErr = fmt.Errorf("nil response from model")
				continue
			}
			text := result.Text()
			if text == "" {
				lastErr = fmt.Errorf("empty response from model")
				continue
			}
			return text, nil
		}

		lastErr = err

		backoff := time.Duration(1<<(attempt+1)) * time.Second
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	return "", fmt.Errorf("generate (3 attempts): %w", lastErr)
}
