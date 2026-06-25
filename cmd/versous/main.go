package main

import (
	"context"
	"fmt"
	"os"

	"github.com/GatosTheDog/versous/internal/agent"
	"github.com/GatosTheDog/versous/internal/llm"
	"github.com/GatosTheDog/versous/internal/render"
	"github.com/GatosTheDog/versous/internal/sources"
	"github.com/GatosTheDog/versous/internal/store"
)

func main() {
	if len(os.Args) != 4 || os.Args[1] != "compare" {
		fmt.Fprintln(os.Stderr, "usage: versous compare <productA> <productB>")
		os.Exit(1)
	}

	productA := os.Args[2]
	productB := os.Args[3]

	ctx := context.Background()

	llmClient, err := llm.New(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "llm:", err)
		os.Exit(1)
	}

	db, err := store.NewPostgres(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "db:", err)
		os.Exit(1)
	}
	defer db.Close()

	a := agent.New(llmClient, db, sources.NewHN(5), sources.NewYoutube(3, 5))

	report, err := a.Compare(ctx, productA, productB)
	if err != nil {
		fmt.Fprintln(os.Stderr, "compare:", err)
		os.Exit(1)
	}

	fmt.Println(render.Render(report))
}
