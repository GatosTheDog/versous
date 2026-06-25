package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/GatosTheDog/versous/internal/agent"
	"github.com/GatosTheDog/versous/internal/llm"
	"github.com/GatosTheDog/versous/internal/render"
	"github.com/GatosTheDog/versous/internal/sources"
	"github.com/GatosTheDog/versous/internal/store"
)

func main() {

	if len(os.Args) < 4 || os.Args[1] != "compare" {
		fmt.Fprintln(os.Stderr, "usage: versous compare <productA> <productB> [--aspects camera,battery,price]")
		os.Exit(1)
	}
	productA := os.Args[2]
	productB := os.Args[3]

	fs := flag.NewFlagSet("compare", flag.ExitOnError)
	aspectsFlag := fs.String("aspects", "", "comma-separated aspects")
	fs.Parse(os.Args[4:])

	ctx := context.Background()

	var aspects []string
	if *aspectsFlag != "" {
		aspects = strings.Split(*aspectsFlag, ",")
	}

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

	report, err := a.Compare(ctx, productA, productB, aspects)
	if err != nil {
		fmt.Fprintln(os.Stderr, "compare:", err)
		os.Exit(1)
	}

	fmt.Println(render.Render(report))
}
