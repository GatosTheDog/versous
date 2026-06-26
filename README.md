# Versous

Compares two products using **real user sentiment** from Hacker News and YouTube, backed by a hand-rolled RAG pipeline with pgvector. Two modes: CLI (Gemini judges) and MCP server (your AI judges).

Built to learn Go, RAG, and agentic system design from scratch. No LLM orchestration frameworks.

## How it works

### CLI mode
```
versous compare "iPhone 16 Pro" "iPhone 17 Pro"
        │
        ├── Ingest (HN + YouTube)
        │   ├── Aspect-specific queries per product
        │   ├── Embed each comment → gemini-embedding-001 (3072-dim vectors)
        │   └── Upsert into Postgres + pgvector
        │
        ├── Retrieve (RAG)
        │   └── Cosine similarity search → top-k comments per product per aspect
        │
        ├── Judge
        │   └── Gemini reads evidence → structured verdict per aspect
        │
        ├── Specs
        │   └── Gemini generates key specs for any product on the fly
        │
        └── Render → CLI table
```

### MCP mode
Exposes three tools to any MCP-compatible AI (Claude, Cursor, Zed):

| Tool | What it does |
|---|---|
| `ingest` | Fetch + embed comments for a product into pgvector |
| `search_comments` | Cosine similarity retrieval — returns raw comment text + URLs |
| `get_specs` | LLM-generated specs for any product |

The AI calls tools in whatever order it decides, reads the raw evidence, and makes its own judgment. No hardcoded aspect list — the AI picks what matters.

## Stack

| Layer | Choice |
|---|---|
| Language | Go 1.22+ |
| LLM / Embeddings | Gemini (`gemini-3.1-flash-lite` + `gemini-embedding-001`) |
| Vector store | Postgres + pgvector (local or Neon free tier) |
| Sources | Hacker News Algolia API + YouTube Data API v3 |
| DB driver | `pgx/v5` |
| MCP | `mark3labs/mcp-go` |

All free tier. Zero infra cost.

## Quick start

**Prerequisites:** Go 1.22+, Postgres with pgvector extension, API keys.

```bash
git clone https://github.com/GatosTheDog/versous
cd versous
```

Run the migration:
```bash
psql $DATABASE_URL -f migrations/001_init.sql
```

Create a `.env` file:
```
GEMINI_API_KEY=...
YOUTUBE_API_KEY=...
DATABASE_URL=postgres://localhost/versous
```

### CLI

```bash
set -a && source .env && set +a
go run ./cmd/versous compare "iPhone 16 Pro" "Pixel 9 Pro"
```

Custom aspects:
```bash
go run ./cmd/versous compare "AirPods Pro 2" "Sony WH-1000XM5" --aspects "noise cancellation,sound quality,comfort"
```

### MCP server

Build and register with Claude Code:
```bash
go build -o versous-mcp ./cmd/versous-mcp
claude mcp add versous ./versous-mcp \
  -e GEMINI_API_KEY=... \
  -e YOUTUBE_API_KEY=... \
  -e DATABASE_URL=...
```

Then prompt Claude:
```
Use the versous tools to compare "AirPods Pro 2" vs "Sony WH-1000XM5".
Ingest both products, search comments on noise cancellation and sound quality, then give me your verdict.
```

## Example output (CLI)

```
=== Versous: iPhone 16 pro vs iPhone 17 pro ===

[Battery Life]  → iPhone 17 pro
Winner: iPhone 17 pro

* Strength: Significantly higher capacity — "better battery life than the iPhone 16."
* Weakness: Extreme thinness concerns; "you might as well switch the phone off."

The iPhone 16 Pro has established track record, though users report it "does not last the full day."

[Camera Quality]  → iPhone 17 pro
Winner: iPhone 17 pro

* Strength: Exceptional results after manual config — "my quality now is amazing."
* Weakness: Optical zoom overstated — "it just starts to crop out the same 8x image."

The iPhone 16 Pro remains capable for specialized use.

[Price]  → iPhone 16 pro
Winner: iPhone 16 pro

* Strength: High perceived value — "I'd pay full price out of pocket, no questions asked."
* Weakness: Some users rely on corporate discounts to justify the cost.

OVERALL WINNER: iPhone 17 pro

── Specs ──────────────────────────────
               iPhone 16 pro          iPhone 17 pro
Processor      A18 Pro chip           Apple A19 Pro
RAM            8GB                    12GB
Battery        3,582 mAh              4,252 mAh
Camera         48MP Fusion, 48MP      48MP Main, 48MP Ultra
               Ultra Wide, 12MP 5x    Wide, 48MP 5x
               Telephoto              Telephoto
Price          Starting at $999       Starting at $1,099
```

## Project layout

```
cmd/versous/        CLI entrypoint
cmd/versous-mcp/    MCP server entrypoint
internal/
  agent/            orchestrator — ingest → retrieve → judge → report
  llm/              Gemini wrapper (generate + embed, retry + backoff)
  rag/              ingest, retrieve, judge
  render/           CLI output formatter
  sources/          CommentSource interface + HN + YouTube implementations
  specs/            LLM-powered spec fetcher
  store/            pgvector repo (upsert + cosine similarity search)
migrations/         SQL schema
```

## Running tests

```bash
GEMINI_API_KEY=... DATABASE_URL=... YOUTUBE_API_KEY=... go test ./... -timeout 300s
```

## License

MIT
