# Versous

CLI tool that compares two products using **real user sentiment** from Hacker News and YouTube, backed by a RAG pipeline with vector search. Specs are secondary — what people actually say drives the verdict.

Built to learn Go, RAG, and agentic system design from scratch. No LLM orchestration frameworks — all retrieval, embedding, and judging logic is hand-rolled.

## How it works

```
versous compare "iPhone 16 Pro" "iPhone 17 Pro"
        │
        ├── Ingest (HN + YouTube)
        │   ├── Search each source for aspect-specific queries
        │   │   e.g. "iPhone 16 Pro Battery Life", "iPhone 16 Pro Camera Quality"
        │   ├── Embed each comment → gemini-embedding-001 (3072-dim vectors)
        │   └── Upsert into Postgres + pgvector
        │
        ├── Retrieve (RAG)
        │   └── Cosine similarity search → top-k comments per product per aspect
        │
        ├── Judge (LLM)
        │   └── Gemini reads comment evidence → structured verdict per aspect
        │
        ├── Specs
        │   └── Gemini generates key specs for any product on the fly
        │
        └── Render → CLI table
```

## Stack

| Layer | Choice |
|---|---|
| Language | Go 1.22+ |
| LLM / Embeddings | Gemini (`gemini-3.1-flash-lite` + `gemini-embedding-001`) |
| Vector store | Postgres + pgvector (local or Neon free tier) |
| Sources | Hacker News Algolia API + YouTube Data API v3 |
| DB driver | `pgx/v5` |

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

Set env vars (create a `.env` file):
```
GEMINI_API_KEY=...
YOUTUBE_API_KEY=...
DATABASE_URL=postgres://localhost/versous
```

Compare two products:
```bash
set -a && source .env && set +a
go run ./cmd/versous compare "iPhone 16 Pro" "iPhone 17 Pro"
```

Custom aspects:
```bash
go run ./cmd/versous compare "AirPods Pro" "Sony WH-1000XM5" --aspects "noise cancellation,sound quality,comfort"
```

## Example output

```
=== Versous: iPhone 16 pro vs iPhone 17 pro ===

[Battery Life]  → iPhone 17 pro
Winner: iPhone 17 pro

* Strength: Significantly higher capacity — users note "better battery life than the iPhone 16."
* Weakness: Concerns about extreme thinness; "you might as well switch the phone off."

The iPhone 16 Pro has an established track record, though users report it "does not last the full day."

[Camera Quality]  → iPhone 17 pro
Winner: iPhone 17 pro

* Strength: Exceptional results after manual config — "my quality now is amazing."
* Weakness: Optical zoom marketing overstated — "it just starts to crop out the same 8x image."

The iPhone 16 Pro remains capable for specialized use despite app integration issues.

[Price]  → iPhone 16 pro
Winner: iPhone 16 pro

* Strength: High perceived value — "I'd pay full price out of pocket, no questions asked."
* Weakness: Some users rely on corporate discounts to justify the cost.

The iPhone 17 Pro discourse focuses on global pricing arbitrage rather than inherent value.

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

Integration tests require env vars and a running Postgres:

```bash
GEMINI_API_KEY=... DATABASE_URL=... YOUTUBE_API_KEY=... go test ./... -timeout 300s
```

Unit tests (no deps):
```bash
go test ./internal/sources/... -run TestHN
```

## License

MIT
