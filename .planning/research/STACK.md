# Stack Research

**Domain:** Locally-run personal AI assistant with agent capabilities
**Researched:** 2026-03-11
**Confidence:** MEDIUM-HIGH (core stack verified, some library versions estimated)

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.26.1 | Backend language | Current stable (Feb 2026). Upgrade from 1.22 in go.mod. Single-binary deploys, low memory, goroutines for concurrent agent tasks. |
| go-zero | v1.9.0 | Backend framework | Already decided. Built-in SSE support for LLM streaming, goctl code generation from .api files, middleware chain, JWT auth, rate limiting. Active maintenance (Feb 2026 release). |
| Next.js | 16.1 | Frontend framework | Already decided. App Router stable, React Server Components, Turbopack stable for dev. Canary 16.2 available but use stable. |
| TypeScript | 5.7+ | Frontend language | Strict typing for complex chat UI state, AI SDK integration. Use strict mode. |
| SQLite (via ncruces/go-sqlite3) | v0.30.1 (SQLite 3.51.0) | Primary database | Single-user local app -- no need for Postgres. Zero-config, file-based, embeds in Go binary. ncruces wraps SQLite via WASM/wazero -- no CGO required, cross-compiles cleanly. |

### LLM Integration

| Library | Version | Purpose | Why Recommended |
|---------|---------|---------|-----------------|
| openai/openai-go | latest (beta) | OpenAI API client | Official SDK. Direct access to OpenAI Responses API. Use for OpenAI and any OpenAI-compatible endpoint (Ollama exposes OpenAI-compatible API). |
| Vercel AI SDK | v6 | Frontend streaming UI | useChat hook, streaming text rendering, Server Actions integration. Handles SSE consumption, token-by-token rendering, message state. Works with any OpenAI-compatible backend via custom provider. |

**LLM provider strategy:** Use `openai/openai-go` with configurable base URL. Ollama, LM Studio, and many providers expose OpenAI-compatible endpoints. For Anthropic, use their Messages API format with a thin adapter in Go. Do NOT use `any-llm-go` (Mozilla) -- it adds abstraction over what is fundamentally just switching base URLs for OpenAI-compatible providers. Keep it simple.

### Memory & Search

| Library | Version | Purpose | Why Recommended |
|---------|---------|---------|-----------------|
| ncruces/go-sqlite3 | v0.30.1 | SQLite driver (no CGO) | Pure Go via WASM. Supports FTS5 out of the box. Single dependency for both relational and full-text search. |
| sqlite-vec (via sqlite-vec-go-bindings/ncruces) | latest | Vector similarity search | WASM bindings for ncruces driver -- no CGO. Hybrid search: FTS5 for keyword recall + sqlite-vec for semantic similarity. Same approach ZeroClaw uses successfully. |
| Embedding model (OpenAI / Ollama) | - | Text-to-vector | Use same LLM provider config. OpenAI text-embedding-3-small or Ollama nomic-embed-text for local. |

### Agent Skills (Browser, Shell, Files)

| Library | Version | Purpose | Why Recommended |
|---------|---------|---------|-----------------|
| go-rod/rod | latest | Web browsing skill | Better API than chromedp: decode-on-demand (faster), simpler iframe handling, built-in Chrome version management. No external dependencies. |
| os/exec (stdlib) | - | Shell command execution | Standard library. Wrap with timeout, working directory, output capture. No library needed. |
| os + io/fs (stdlib) | - | File read/write skill | Standard library. Add sandboxing via allowed-paths config. |

### Frontend UI

| Library | Version | Purpose | Why Recommended |
|---------|---------|---------|-----------------|
| shadcn/ui | CLI v4 | Component library | Not a dependency -- copies components into your codebase. Radix UI primitives + Tailwind. CLI v4 (March 2026) scaffolds full Next.js projects. |
| Tailwind CSS | v4 | Styling | CSS-first config (no tailwind.config.js). 5x faster builds via Rust engine. Next.js 16 has first-class support. |
| React | 19 (via Next.js 16) | UI framework | Bundled with Next.js 16. Server Components, Server Actions for AI inference. |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| goctl | Go code generation | Part of go-zero. Generate handlers, models, types from .api files. Install: `go install github.com/zeromicro/go-zero/tools/goctl@latest` |
| pnpm | Node package manager | Faster, disk-efficient vs npm. Strict dependency resolution avoids phantom deps. |
| Docker | Local dev + deployment | Single Dockerfile: multi-stage build (Go binary + Next.js standalone). Include chromium for go-rod. |
| Air | Go hot reload | `github.com/air-verse/air` -- watches .go files, rebuilds on change. |

## Installation

```bash
# Go backend
go mod init open-jarvis
go get github.com/zeromicro/go-zero@v1.9.0
go get github.com/ncruces/go-sqlite3@latest
go get github.com/asg017/sqlite-vec-go-bindings/ncruces@latest
go get github.com/openai/openai-go@latest
go get github.com/go-rod/rod@latest

# Go dev tools
go install github.com/zeromicro/go-zero/tools/goctl@latest
go install github.com/air-verse/air@latest

# Frontend
pnpm create next-app@latest frontend --typescript --tailwind --app --src-dir
cd frontend
pnpm dlx shadcn@latest init
pnpm add ai @ai-sdk/openai  # Vercel AI SDK v6

# Dev dependencies
pnpm add -D @types/node
```

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| openai/openai-go | any-llm-go (Mozilla) | If you need first-class Anthropic/Gemini native APIs beyond OpenAI-compatible mode. But adds complexity -- prefer direct openai-go with base URL switching. |
| openai/openai-go | langchaingo | If building complex multi-step chains with built-in vector store abstractions. Overkill for this project -- we control the agent loop directly. |
| ncruces/go-sqlite3 (no CGO) | mattn/go-sqlite3 (CGO) | If you need maximum SQLite performance. CGO version is ~2x faster but requires C toolchain and complicates cross-compilation. Not worth it for single-user app. |
| go-rod | chromedp | If you want minimal dependencies and only need basic page loading. chromedp is lower-level but more verbose. Rod has better DX for agent-style browsing. |
| SQLite + sqlite-vec | PostgreSQL + pgvector | If you plan multi-user or need concurrent writes. Single-user local app does not need this. |
| Vercel AI SDK | Custom SSE client | If AI SDK abstractions get in the way. But useChat saves significant boilerplate for streaming chat UI. |
| pnpm | npm | If team is more familiar with npm. pnpm is strictly better for monorepo and disk usage, but npm works fine too. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| langchaingo | Over-abstraction for a project that owns its agent loop. Hides control flow, makes debugging harder. Opinionated chain patterns conflict with custom skill system. | Direct openai/openai-go + custom agent loop |
| any-llm-go | Unnecessary abstraction layer. Most providers (Ollama, Groq, DeepSeek) already expose OpenAI-compatible APIs. Adds 8 provider SDKs as transitive deps. | openai/openai-go with configurable baseURL |
| mattn/go-sqlite3 | Requires CGO. Breaks `go build` on clean machines, complicates Docker multi-stage builds, prevents easy cross-compilation. | ncruces/go-sqlite3 (WASM, no CGO) |
| WebSocket for LLM streaming | SSE is simpler, unidirectional (which is what LLM streaming needs), and go-zero has built-in SSE support. WebSocket adds reconnection complexity for no benefit here. | SSE via go-zero built-in support |
| Tailwind CSS v3 | v4 is stable, significantly faster, CSS-first config is cleaner. v3 is legacy. Next.js 16 + shadcn CLI v4 default to Tailwind v4. | Tailwind CSS v4 |
| chromedp | More verbose API, slower (decodes every browser message), harder iframe handling. go-rod wraps the same DevTools Protocol with better DX. | go-rod |
| Electron / Tauri | Scope creep. Web UI first, validate the agent loop. Desktop wrapper adds packaging complexity for zero user value at this stage. | Next.js web app (runs in browser) |

## Stack Patterns by Variant

**If running fully local (no cloud APIs):**
- Use Ollama as the LLM provider (OpenAI-compatible endpoint at localhost:11434)
- Use nomic-embed-text or all-minilm for local embeddings
- Everything stays on-machine, zero network calls

**If using cloud APIs (OpenAI, Anthropic):**
- Set baseURL to https://api.openai.com/v1 for OpenAI
- For Anthropic: write a thin adapter translating OpenAI tool-call format to Anthropic Messages API
- Store API keys in local config file (not .env committed to git)

**If deploying to self-hosted server:**
- Docker Compose: Go backend + Next.js frontend + headless Chrome (for go-rod)
- SQLite file mounted as Docker volume for persistence
- Reverse proxy (Caddy) for HTTPS

## Version Compatibility

| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| go-zero v1.9.0 | Go 1.26.x | go-zero tracks Go releases closely |
| ncruces/go-sqlite3 v0.30.1 | Go 1.22+ | Uses wazero v1.10.1 internally |
| sqlite-vec-go-bindings/ncruces | ncruces/go-sqlite3 v0.30+ | Must use ncruces variant, NOT the CGO variant |
| Next.js 16.1 | React 19, Node 20+ | Use Node 22 LTS for best compatibility |
| Vercel AI SDK v6 | Next.js 15+, React 19 | Uses Server Actions (no API routes needed) |
| shadcn/ui CLI v4 | Next.js 16, Tailwind v4 | March 2026 release -- scaffolds with Tailwind v4 by default |
| Tailwind CSS v4 | Next.js 16 | CSS-first config, no tailwind.config.js |

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Go 1.26 + go-zero v1.9 | HIGH | Verified via official releases and pkg.go.dev |
| Next.js 16.1 + Tailwind v4 | HIGH | Verified via nextjs.org blog and shadcn changelog |
| ncruces/go-sqlite3 no-CGO | HIGH | Verified via GitHub releases, sqlite-vec docs confirm WASM bindings |
| openai/openai-go | MEDIUM | Official SDK still in beta. API is stable but may have breaking changes. |
| Vercel AI SDK v6 | MEDIUM | v6 confirmed in multiple sources but exact API surface may shift. useChat core is stable. |
| go-rod for browsing | MEDIUM | Well-maintained, good comparison data vs chromedp. Less battle-tested than chromedp in production. |
| sqlite-vec hybrid search | MEDIUM | Approach validated by ZeroClaw. WASM bindings for ncruces are newer, less community usage data. |

## Sources

- [go-zero releases](https://github.com/zeromicro/go-zero/releases) -- v1.9.0 confirmed Feb 2026
- [go-zero SSE docs](https://go-zero.dev/en/docs/tutorials/http/server/sse) -- built-in SSE support
- [Go 1.26 release](https://go.dev/blog/go1.26) -- Feb 10, 2026
- [Next.js 16 blog](https://nextjs.org/blog/next-16) -- stable release
- [shadcn/ui CLI v4 changelog](https://ui.shadcn.com/docs/changelog/2026-03-cli-v4) -- March 2026
- [Tailwind CSS v4](https://tailwindcss.com/blog/tailwindcss-v4) -- Rust engine, CSS-first config
- [openai/openai-go](https://github.com/openai/openai-go) -- official Go SDK
- [any-llm-go](https://blog.mozilla.ai/run-openai-claude-mistral-llamafile-and-more-from-one-interface-now-in-go/) -- evaluated but not recommended
- [ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3) -- v0.30.1 with SQLite 3.51.0
- [sqlite-vec Go bindings](https://alexgarcia.xyz/sqlite-vec/go.html) -- WASM + CGO options
- [ZeroClaw hybrid memory](https://zeroclaws.io/blog/zeroclaw-hybrid-memory-sqlite-vector-fts5/) -- hybrid search validation
- [go-rod](https://github.com/go-rod/rod) -- browser automation
- [go-rod vs chromedp](https://github.com/go-rod/go-rod.github.io/blob/main/why-rod.md) -- comparison
- [Vercel AI SDK](https://ai-sdk.dev/docs/introduction) -- v6 streaming chat
- [chromem-go](https://github.com/philippgille/chromem-go) -- evaluated, prefer sqlite-vec for unified storage

---
*Stack research for: open-jarvis (locally-run personal AI assistant)*
*Researched: 2026-03-11*
