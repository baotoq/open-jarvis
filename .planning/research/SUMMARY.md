# Project Research Summary

**Project:** open-jarvis
**Domain:** Locally-run personal AI assistant with agent capabilities and web dashboard
**Researched:** 2026-03-11
**Confidence:** MEDIUM-HIGH

## Executive Summary

Open-jarvis is a local-first personal AI assistant built on a Go backend (go-zero framework) and Next.js dashboard, with a self-contained agent loop that can call tools (shell, files, web), remember context, and stream responses in real time. The primary differentiator over the closest competitor (OpenClaw) is the purpose-built web dashboard: OpenClaw routes everything through messaging apps like WhatsApp and Discord, while open-jarvis gives users a dedicated, feature-rich browser interface for interacting with the agent, inspecting its actions, and managing its settings. This web-first approach unlocks visual tool feedback, memory inspection, conversation branching, and configuration UIs that are simply impossible in a chat-app interface.

The recommended technical approach is a single Go binary exposing REST + SSE endpoints consumed by a Next.js frontend. The agent loop is a canonical "while loop with tool calls" pattern: assemble context, call the LLM, if the model requests a tool then execute it and loop, if the model returns text then stream it and terminate. SQLite (via ncruces/go-sqlite3, no CGO) serves as the sole data store for both relational data and vector search (sqlite-vec), eliminating all external infrastructure. LLM providers are abstracted behind the OpenAI-compatible chat completions interface, making Ollama, OpenAI, and any OpenRouter-style proxy interchangeable by changing a base URL config value.

The biggest risks are security-related and must be addressed at project start, not retrofitted later. Unrestricted shell access is the most dangerous pitfall: an LLM hallucinating a destructive command can cause irreversible data loss. Agent runaway loops (infinite tool-call cycles) can burn API credits or block the system indefinitely. Context window mismanagement silently degrades response quality as conversations grow. All three of these must be solved in Phase 1 — the agent loop must ship with guardrails baked in. A secondary risk is that go-zero, chosen as the backend framework, is a microservices framework being used for a single-user app; it must be constrained to its REST-only features to avoid fighting its architecture.

## Key Findings

### Recommended Stack

The stack is a monolithic Go binary (Go 1.26.1, go-zero v1.9.0 for HTTP/SSE) serving a Next.js 16.1 frontend. SQLite handles all persistence via the ncruces/go-sqlite3 driver (no CGO, cross-compiles cleanly), with sqlite-vec loaded as a WASM extension for vector search in the same database file. LLM calls use the official openai/openai-go SDK pointed at a configurable base URL; the Vercel AI SDK v6 (useChat hook) handles streaming token rendering on the frontend. Browser automation for the web-fetch tool uses go-rod (simpler API than chromedp). Frontend components come from shadcn/ui CLI v4 with Tailwind CSS v4.

See `.planning/research/STACK.md` for full version matrix and compatibility notes.

**Core technologies:**
- Go 1.26.1 + go-zero v1.9.0: backend runtime — single-binary deploys, goroutines for concurrent streaming, built-in SSE support
- Next.js 16.1 + TypeScript 5.7+: dashboard frontend — App Router, React Server Components, stable Turbopack
- SQLite via ncruces/go-sqlite3: primary database — no CGO, file-based, embeds in Go binary, FTS5 included
- sqlite-vec (ncruces WASM bindings): vector search — hybrid semantic + keyword recall, same DB file as conversations
- openai/openai-go (beta): LLM client — configurable base URL covers OpenAI, Ollama, LM Studio, OpenRouter
- Vercel AI SDK v6: frontend streaming — useChat hook handles SSE consumption and token-by-token rendering
- go-rod: web browsing skill — better DX than chromedp for agent-style page interaction
- shadcn/ui CLI v4 + Tailwind CSS v4: UI components — CSS-first config, 5x faster Rust engine

### Expected Features

The FEATURES.md research identifies 4 build phases aligned from simplest (core chat) to most complex (advanced intelligence). The web dashboard itself is the primary differentiator, so UI features are not secondary polish — they are core product value.

See `.planning/research/FEATURES.md` for full feature dependency graph and anti-feature list.

**Must have (table stakes):**
- Streaming chat interface — users will not wait for full responses; SSE delivery is mandatory
- Multi-turn conversation with history — context must carry across messages within a session
- Conversation persistence and list — save/resume sessions; SQLite storage with sidebar UI
- File read/write and shell execution — core agent tools; require path sandboxing and approval UX
- Web search and web page fetching — current information access; SearXNG or Brave API
- Model provider flexibility — configurable base URL covers OpenAI, Ollama, OpenRouter
- Tool call visibility — users must see what the agent did; expandable tool call blocks in chat
- Markdown rendering and code syntax highlighting — developer audience expects this

**Should have (competitive differentiators):**
- Visual tool execution feedback — rich UI cards for file diffs, command output, search results (web-UI-exclusive)
- Configuration UI — settings page in browser vs. editing YAML files (OpenClaw pain point)
- Agent workspace browser — visual file explorer for agent's files and memory
- Memory inspection and editing — view/delete/edit what the agent "remembers"
- Persistent memory with semantic search — hybrid FTS5 + vector recall for long-term context
- MCP (Model Context Protocol) support — instant compatibility with thousands of third-party tools
- Real-time resource monitoring — token usage and cost tracking in dashboard
- Conversation branching — explore alternatives without losing original thread

**Defer (v2+):**
- Multi-agent delegation — very high complexity, niche use case for personal assistant scope
- Browser automation via CDP — complex; better introduced via MCP skill later
- Proactive/scheduled tasks (cron) — depends on mature memory system; not core to v1 value prop
- Messaging app integrations — contradicts the web-UI differentiator

### Architecture Approach

The system follows a Gateway + Agent Runtime pattern. The Go binary is the gateway: it handles HTTP routing, session management, and configuration, and it hosts the Agent Runtime, Tool Registry, and Memory Store. The Next.js dashboard communicates with the gateway over REST (for CRUD operations) and SSE (for streaming LLM responses). There is no BFF layer — Next.js calls the Go backend directly. The agent loop is a while-loop: assemble context (system prompt + conversation history + vector-retrieved memories + current message), call LLM, if tool_calls execute each via Tool Registry and loop, if text stream tokens to client and persist the turn.

See `.planning/research/ARCHITECTURE.md` for component diagrams, data flow, schema sketch, and anti-patterns.

**Major components:**
1. Next.js Dashboard — chat UI, conversation sidebar, settings page, tool approval dialogs
2. Go Backend (Gateway) — HTTP/SSE server, routing, session management, config (go-zero REST only)
3. Agent Runtime — agent while-loop, context assembly, LLM calls, tool dispatch, response streaming
4. Tool Registry — interface + registration pattern; built-in tools: shell, file, web-fetch, web-search
5. Memory Store — two-tier: SQLite tables for conversation history + sqlite-vec for semantic memory
6. LLM Provider Adapter — thin wrapper over OpenAI-compatible chat completions; configurable base URL

### Critical Pitfalls

All 6 critical pitfalls from PITFALLS.md must inform phase design. Three are Phase 1 requirements.

See `.planning/research/PITFALLS.md` for full analysis, integration gotchas, and recovery strategies.

1. **Unrestricted shell/file access** — ship an allowlist and explicit approval step for shell commands from day one; never allow access outside the designated workspace; a single missing guardrail can cause irreversible data loss
2. **Agent runaway loops** — implement hard limits before enabling any tools: `max_tool_calls_per_turn` (e.g., 25), `max_elapsed_time` (e.g., 5 min), duplicate-call detection; these must exist in the loop structure itself, not as optional config
3. **Context window mismanagement** — track token counts per request from the start; truncate tool outputs to a configurable max; implement sliding window for conversation history; the database stores everything but the LLM only sees a managed window
4. **Prompt injection via tool outputs** — frame all tool outputs with role separators; validate every LLM-generated tool call against the allowlist before execution; sanitize fetched web content before inserting into context
5. **go-zero framework overhead** — use go-zero as a thin HTTP/SSE layer only; do not configure service discovery or RPC; resist over-relying on goctl code generation for business logic; keep the agent runtime framework-agnostic

## Implications for Roadmap

Research across all 4 files converges on the same 4-5 phase structure. Architecture and Features both independently recommend the same ordering: streaming chat loop first, then tools, then memory, then advanced intelligence. Pitfalls add a strong requirement that safety guardrails (shell approval, loop limits, context management) live in Phase 1, not Phase 2.

### Phase 1: Foundation — Streaming Chat Loop

**Rationale:** Nothing else can be built or tested without LLM connectivity and a working chat UI. The agent loop architecture (while-loop with tool calls) must be in place before adding individual tools. Safety guardrails must be present before any tool use is enabled.
**Delivers:** Working end-to-end chat with streaming, multi-turn context, model provider configuration, and the agent loop skeleton with guardrails wired in.
**Addresses:** Streaming chat interface, multi-turn conversation, model provider flexibility, markdown rendering, system prompt configuration, dark/light theme.
**Avoids:**
- Agent runaway loops (implement max_tool_calls, max_elapsed_time, token budget from the start)
- Context window mismanagement (token counting and sliding window baked into message handling)
- go-zero overhead (establish it as HTTP/SSE layer only; keep agent runtime framework-agnostic)
- API key security (no committed keys; .env in .gitignore before first commit)

### Phase 2: Agent Tools

**Rationale:** Conversation persistence is needed first because tools generate messages that must be stored. Tool integration immediately exposes prompt injection and requires a validation layer, which is why that pitfall lands in Phase 2. Tool call visibility in the UI is non-negotiable — users must see what the agent is doing.
**Delivers:** File read/write, shell execution (with approval gates), web search, web page fetching, conversation persistence with list/history UI, and tool call display in chat.
**Addresses:** File read/write, shell command execution, web search, web page fetching, conversation persistence, conversation list, tool call visibility.
**Uses:** Tool Registry interface + registration pattern, go-rod for web browsing, os/exec with timeout for shell, SQLite for conversation storage.
**Avoids:**
- Unrestricted shell/file access (allowlist, approval step, path sandboxing in this phase)
- Prompt injection via tool outputs (validation layer between LLM tool_calls and execution; content sanitization for web fetch)
- No cancellation mechanism (cancel button wired to context cancellation in tool execution chain)

### Phase 3: Dashboard Differentiation

**Rationale:** Once the agent works, the web dashboard's competitive advantages over OpenClaw can be built out. These are UI-heavy features that require the agent's core capabilities to be stable first. Configuration UI unlocks proper multi-provider and multi-model support.
**Delivers:** Settings/configuration UI, agent workspace browser, memory inspection/editing UI, real-time resource monitoring (token and cost tracking), conversation search.
**Addresses:** Configuration UI, agent workspace browser, memory inspection and editing, real-time resource monitoring, export conversations.
**Avoids:**
- Overengineered memory before validating core loop (this phase uses simple SQLite + FTS5; no vectors yet)

### Phase 4: Persistent Memory and Intelligence

**Rationale:** Memory must come after the core loop is stable and there are real conversations to test retrieval against. The pitfall research explicitly warns against building vector search before you have evidence that FTS5 is insufficient. This phase adds sqlite-vec, MCP support, and semantic recall.
**Delivers:** Hybrid FTS5 + vector semantic memory, context window summarization for long conversations, MCP protocol support (connects to third-party tool servers), multi-model switching, conversation branching.
**Addresses:** Persistent memory with semantic search, MCP support, context window budget management, multi-model switching mid-conversation, conversation branching.
**Implements:** sqlite-vec setup in same SQLite DB, LLM-based conversation summarization, MCP client in Go, memory pruning strategy.
**Avoids:**
- Overengineered memory (validate FTS5 before adding vectors; plan pruning from day one)
- Unbounded context (this is the phase where summarization strategy is implemented properly)

### Phase 5: Polish and Extensibility

**Rationale:** Quality-of-life improvements and ecosystem integration that require stable underlying features. Skill/plugin management UI depends on MCP support from Phase 4. Safety tools (dry-run mode, audit logs surfaced in UI) complete the trust model.
**Delivers:** Skill/plugin management UI, tool safety level controls (auto-approve vs. confirm per tool category), enhanced error handling with human-readable messages, keyboard shortcuts, audit log UI, performance tuning.
**Addresses:** Skill/plugin management UI, proactive/scheduled tasks (basic cron), visual tool execution feedback (rich cards), keyboard shortcuts, dark/light theme polish.

### Phase Ordering Rationale

- Phase 1 before everything: the agent loop and streaming infrastructure are the foundation all other components build on; cannot test tools, memory, or dashboard features without them
- Phase 2 after Phase 1: tool execution requires conversation persistence (tool messages must be stored); safety guardrails in Phase 1 make it safe to connect real tools in Phase 2
- Phase 3 after Phase 2: dashboard differentiation features require stable agent capabilities; configuration UI requires tools to exist to configure
- Phase 4 after Phase 3: semantic memory needs real conversations to test against; adding vector search before validating FTS5 sufficiency is the most common pitfall in this domain
- Phase 5 after Phase 4: plugin management UI depends on MCP support; audit log UI depends on logging infrastructure from earlier phases

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 2:** Shell command approval UX design needs thought — what exactly triggers confirmation, how does the UI present the command, how does auto-approve work per tool category
- **Phase 4:** MCP client implementation in Go — the MCP spec is young (2026 standard); implementation details for Go client may require API-level research during planning; also, embedding model selection (OpenAI text-embedding-3-small vs. Ollama nomic-embed-text) may need benchmarking
- **Phase 4:** Context window summarization strategy — multiple valid approaches exist (rolling summary, selective retention, importance scoring); needs a concrete decision before implementation

Phases with standard patterns (can skip research-phase):
- **Phase 1:** SSE streaming chat loop is extremely well-documented; go-zero SSE support is verified; agent while-loop pattern is canonical and documented across multiple authoritative sources
- **Phase 2:** File and shell tools use stdlib; web fetch pattern (HTTP GET + HTML-to-markdown) is standard; SQLite schema for conversation history is straightforward
- **Phase 3:** Dashboard UI patterns (sidebar, settings page, file browser) are standard Next.js/React patterns; no novel research needed
- **Phase 5:** Keyboard shortcuts, theming, error messages — standard frontend patterns

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | MEDIUM-HIGH | Core stack (Go, go-zero, Next.js, SQLite) verified via official releases. openai/openai-go and Vercel AI SDK v6 are in beta — API surface is stable but may shift. sqlite-vec WASM bindings for ncruces are newer with less community usage data. |
| Features | HIGH | Features derived from direct analysis of OpenClaw docs and peer products (Open Interpreter, assistant-ui). Feature gap analysis (web UI vs. messaging apps) is factual, not speculative. |
| Architecture | HIGH | Agent while-loop pattern is canonical and confirmed across OpenClaw, Anthropic, Google ADK, and LangChain independently. SQLite + sqlite-vec approach validated by ZeroClaw production use. Anti-patterns are documented and corroborated. |
| Pitfalls | HIGH | Pitfalls sourced from real incidents (December 2025 rm -rf case), security research (NVIDIA, MIT), and production experience (Cursor sandboxing blog). Highly specific and actionable. |

**Overall confidence:** MEDIUM-HIGH

### Gaps to Address

- **go-zero REST-only usage patterns:** go-zero's documentation focuses on its full microservices stack; specific guidance for using only its REST/HTTP module as a thin layer is sparse. During Phase 1 setup, validate that go-zero's REST module is genuinely usable without its service discovery and RPC infrastructure before committing deeply. If it proves too opinionated, Chi or Echo are documented alternatives that can be swapped with minimal business logic impact.
- **openai/openai-go beta API surface:** The official Go SDK is still in beta. Before Phase 1, pin to a specific commit or version and read the changelog carefully. Streaming tool call accumulation (partial JSON reassembly from streaming chunks) is the most likely area to have rough edges.
- **Embedding model selection for local use:** nomic-embed-text (Ollama) and text-embedding-3-small (OpenAI) are recommended but not benchmarked against each other for this use case. Address during Phase 4 planning with real conversation data.
- **MCP Go client maturity:** MCP is the 2026 standard but Go client libraries are newer than Python equivalents. Evaluate available Go MCP client libraries during Phase 4 planning to confirm one is production-ready.

## Sources

### Primary (HIGH confidence)
- [go-zero releases](https://github.com/zeromicro/go-zero/releases) — v1.9.0 confirmed Feb 2026
- [Go 1.26 release](https://go.dev/blog/go1.26) — Feb 10, 2026
- [Next.js 16 blog](https://nextjs.org/blog/next-16) — stable release confirmed
- [shadcn/ui CLI v4 changelog](https://ui.shadcn.com/docs/changelog/2026-03-cli-v4) — March 2026
- [ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3) — v0.30.1 with SQLite 3.51.0
- [sqlite-vec Go bindings](https://alexgarcia.xyz/sqlite-vec/go.html) — WASM + CGO options documented
- [OpenClaw architecture documentation](https://github.com/centminmod/explain-openclaw) — HIGH confidence
- [OpenClaw GitHub repository](https://github.com/openclaw/openclaw) — feature comparison source
- [Braintrust: The canonical agent architecture is a while loop with tools](https://www.braintrust.dev/blog/agent-while-loop) — agent loop pattern
- [go-zero SSE documentation](https://go-zero.dev/guides/http/server/sse/) — built-in SSE confirmed
- [MIT Technology Review: Is a secure AI assistant possible?](https://www.technologyreview.com/2026/02/11/1132768/is-a-secure-ai-assistant-possible/) — prompt injection
- [NVIDIA: Practical Security Guidance for Sandboxing Agentic Workflows](https://developer.nvidia.com/blog/practical-security-guidance-for-sandboxing-agentic-workflows-and-managing-execution-risk/) — sandbox patterns

### Secondary (MEDIUM confidence)
- [ZeroClaw hybrid memory (SQLite + sqlite-vec + FTS5)](https://zeroclaws.io/blog/zeroclaw-hybrid-memory-sqlite-vector-fts5/) — production validation of storage approach
- [go-rod vs chromedp comparison](https://github.com/go-rod/go-rod.github.io/blob/main/why-rod.md) — browser automation choice
- [Vercel AI SDK](https://ai-sdk.dev/docs/introduction) — v6 streaming chat, useChat hook
- [Model Context Protocol](https://modelcontextprotocol.io/) + [MCP Complete Guide 2026](https://calmops.com/ai/model-context-protocol-mcp-2026-complete-guide/) — MCP standard
- [DEV Community: Rate Limiting Your Own AI Agent](https://dev.to/askpatrick/rate-limiting-your-own-ai-agent-the-runaway-loop-problem-nobody-talks-about-3dh2) — runaway loop patterns
- [DEV Community: The Token Budget Pattern](https://dev.to/askpatrick/the-token-budget-pattern-how-to-stop-ai-agent-cost-surprises-before-they-happen-5hb3) — token budget guardrails
- [Three Dots Labs: When You Shouldn't Use Frameworks in Go](https://threedots.tech/episode/when-you-should-not-use-frameworks/) — go-zero risk assessment
- [Cursor Blog: Implementing a Secure Sandbox for Local Agents](https://cursor.com/blog/agent-sandboxing) — sandboxing patterns

### Tertiary (LOW confidence / needs validation)
- openai/openai-go beta API surface — SDK is official but in beta; streaming tool call accumulation needs hands-on validation
- MCP Go client library maturity — standard is 2026; Go ecosystem support less documented than Python

---
*Research completed: 2026-03-11*
*Ready for roadmap: yes*
