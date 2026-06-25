package agent

import (
	"context"
	"os"
	"testing"

	"github.com/GatosTheDog/versous/internal/llm"
	"github.com/GatosTheDog/versous/internal/sources"
	"github.com/GatosTheDog/versous/internal/store"
)

func TestAgent(t *testing.T) {
	if os.Getenv("GEMINI_API_KEY") == "" || os.Getenv("DATABASE_URL") == "" {
		t.Skip("missing env vars")
	}

	ctx := context.Background()

	llmClient, err := llm.New(ctx)
	if err != nil {
		t.Fatal(err)
	}

	db, err := store.NewPostgres(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	hn := sources.NewHN(5)
	yt := sources.NewYoutube(3, 5)

	agent := New(llmClient, db, hn, yt)

	report, err := agent.Compare(ctx, "iPhone 16", "iPhone 15")
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Aspects) != 3 {
		t.Errorf("expected 3 aspects, got %d", len(report.Aspects))
	}
	if report.Winner == "" {
		t.Errorf("expected non-empty winner")
	}
	for _, aspect := range report.Aspects {
		t.Logf("[%s] winner=%s | %s", aspect.Aspect, aspect.Winner, aspect.Summary)
	}

}
