# Versous

Compares two products using **real user sentiment** from Hacker News and YouTube, backed by a hand-rolled RAG pipeline with pgvector. Two modes: CLI (Gemini judges) and MCP server (your AI judges).

Built to learn Go, RAG, and agentic system design from scratch. No LLM orchestration frameworks.

## Demo

[▶️ Watch the demo (MP4)](https://raw.githubusercontent.com/GatosTheDog/versous/main/demo.mp4)

## Try it live (no setup required)

Add to Claude Desktop (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "versous": {
      "command": "npx",
      "args": ["-y", "mcp-remote", "https://versous.fly.dev/mcp"]
    }
  }
}
```

Restart Claude Desktop, then ask:
> "Use versous to compare iPhone 16 Pro vs Pixel 9 Pro — ingest both, then judge on battery, camera, and value."

No API keys or local setup needed. The server runs on fly.io with a shared Neon vector store.

For MCP clients with native HTTP support (Cursor, Zed): connect directly to `https://versous.fly.dev/mcp`.

## How it works

### CLI mode
```
versous compare "AirPods Pro 2" vs "Sony WH-1000XM5"
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
go run ./cmd/versous compare "AirPods Pro 2" vs "Sony WH-1000XM5"
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
=== Versous: AirPods Pro 2 vs Sony WH-1000XM5 ===

[noise cancellation]  → AirPods Pro 2
+ Airpods Pro 2 seems great at canceling noise
- during the call i heard some hech hech noise that is very annoying
  Loser upside: Sony WH-1000XM5 benefits from physical over-ear isolation combined with traditional earplugs for a more robust multi-layered noise mitigation strategy.

  Evidence:
  · "I m not getting the off option in mu noise cantrol please help and also during the call i heard some hech hech noise tha…"
    https://youtube.com/watch?v=3ozTYEcg-O4 (youtube)
  · "I completely understand the sentiments of the author. It's easy to fall into the trap of \"it's too easy to be true; let …"
    https://news.ycombinator.com/item?id=43347805 (hn)
  · "Any recommendation for best noise cancelling headphones? I'm looking at bose quiet comfort 45 or Sony WH-1000XM5, but I …"
    https://news.ycombinator.com/item?id=37048192 (hn)
  · "> noise-cancelling headphones + 3m earplugs might work togetherSure, any closed back headphones work and cancel out nois…"
    https://news.ycombinator.com/item?id=43349417 (hn)

[sound quality]  → AirPods Pro 2
+ AirPods Pro 2 are amazing, and have solid, respectable sound
- they have a V shaped tuning, with various levels of bad
  Loser upside: The Sony WH-1000XM5 benefits from a larger over-ear form factor that allows for a different soundstage and bass delivery compared to TWS earbuds.

  Evidence:
  · "What AirPods are you talking about? The wired AirPods that sound pretty bad have been overtaken by wireless Bluetooth Ai…"
    https://news.ycombinator.com/item?id=48494917 (hn)
  · "True, but I'd argue that the effect only starts to kick in at some point around $400 or so for a pair of headphones. Abo…"
    https://news.ycombinator.com/item?id=36378534 (hn)
  · "Что ж у них с басами,никогда не понимал сони,они не упроги а размыл…"
    https://youtube.com/watch?v=rNq-UIpm9Hk (youtube)
  · "I have question; does anybody know if there is much difference in sound quality between the Sony CH720 and the Sony XM5?"
    https://youtube.com/watch?v=6WTHBCZBt_E (youtube)

OVERALL WINNER: AirPods Pro 2

── Specs ──────────────────────────────
               AirPods Pro 2          Sony WH-1000XM5
Processor      Apple H2 chip          Integrated Processor V1 and HD Noise Cancelling Processor QN1
RAM            N/A                    N/A
Battery        Up to 6 hours (up to 30 hours with case) Up to 30 hours (ANC on) / 40 hours (ANC off)
Camera         N/A                    N/A
Price          $249                   $399.99
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
