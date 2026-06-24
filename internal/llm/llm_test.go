package llm

import (
	"context"
	"os"
	"testing"
)

func TestGenerate(t *testing.T) {
	ctx := context.Background()
	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("no api key")
	}

	client, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.Generate(ctx, "What is 2+2")
	if err != nil {
		t.Fatal(err)
	}

	if result == "" {
		t.Fatal("expected non-empty response")
	}

	t.Logf("response: %s", result)
}

func TestEmbed(t *testing.T) {
	ctx := context.Background()
	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("no API key")
	}

	client, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.Embed(ctx, "battery drains fast")
	if err != nil {
		t.Fatal(err)
	}

	if len(result) == 0 {
		t.Fatal("expected non-empty response")
	}

	t.Logf("vector length: %d", len(result))
}
