package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/GatosTheDog/versous/internal/llm"
	"github.com/GatosTheDog/versous/internal/rag"
	"github.com/GatosTheDog/versous/internal/sources"
	"github.com/GatosTheDog/versous/internal/specs"
	"github.com/GatosTheDog/versous/internal/store"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
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

	hn := sources.NewHN(5)
	yt := sources.NewYoutube(3, 2)

	s := server.NewMCPServer("versous", "1.0.0")

	s.AddTool(
		mcp.NewTool("search_comments",
			mcp.WithDescription("Search for user comments about a product from HN and YouTube"),
			mcp.WithString("product", mcp.Required(), mcp.Description("product name, e.g. iPhone 16 Pro")),
			mcp.WithString("query", mcp.Required(), mcp.Description("search query, e.g. battery life")),
			mcp.WithNumber("limit", mcp.Description("max comments to return, default 5")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			product, err := req.RequireString("product")
			if err != nil {
				return nil, err
			}
			query, err := req.RequireString("query")
			if err != nil {
				return nil, err
			}
			limit := req.GetInt("limit", 5)

			comments, err := rag.Retrieve(ctx, llmClient, db, query, product, limit)
			if err != nil {
				return nil, err
			}

			var b strings.Builder
			for _, c := range comments {
				fmt.Fprintf(&b, "- %s\n  source: %s\n\n", c.Body, c.Url)
			}
			return mcp.NewToolResultText(b.String()), nil
		},
	)

	s.AddTool(
		mcp.NewTool("ingest",
			mcp.WithDescription("Fetch and embed comments about a product into the vector store"),
			mcp.WithString("product", mcp.Required(), mcp.Description("canonical product name, e.g. iPhone 16 Pro")),
			mcp.WithString("queries", mcp.Required(), mcp.Description("comma-separated search queries, e.g. iPhone 16 Pro camera,iPhone 16 Pro battery")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			product, err := req.RequireString("product")
			if err != nil {
				return nil, err
			}
			queriesRaw, err := req.RequireString("queries")
			if err != nil {
				return nil, err
			}
			queries := strings.Split(queriesRaw, ",")

			for _, src := range []sources.CommentSource{hn, yt} {
				if err := rag.Ingest(ctx, src, llmClient, db, product, queries); err != nil {
					return nil, err
				}
			}
			return mcp.NewToolResultText("ingested " + product), nil
		},
	)

	s.AddTool(
		mcp.NewTool("get_specs",
			mcp.WithDescription("Get key technical specs for a product"),
			mcp.WithString("product", mcp.Required(), mcp.Description("product name, e.g. iPhone 16 Pro")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			product, err := req.RequireString("product")
			if err != nil {
				return nil, err
			}
			spec, err := specs.Fetch(ctx, llmClient, product)
			if err != nil {
				return nil, err
			}
			out := fmt.Sprintf("Display: %s\nProcessor: %s\nRAM: %s\nBattery: %s\nCamera: %s\nPrice: %s",
				spec.Display, spec.Processor, spec.RAM, spec.Battery, spec.Camera, spec.Price)
			return mcp.NewToolResultText(out), nil
		},
	)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
