# Phase 1: Streaming Chat Loop - Context

**Gathered:** 2026-03-11
**Status:** Ready for planning

<domain>
## Phase Boundary

End-to-end streaming chat: user types a message, the response appears token-by-token in real time, multi-turn conversation context is maintained within a session, and the agent loop has configurable safety limits. No tool execution yet (that's Phase 3). No persistence yet (that's Phase 2).

</domain>

<decisions>
## Implementation Decisions

### Chat UI Style
- Light, polished visual style — white/light background, rounded message bubbles (ChatGPT-like)
- Assistant responses render markdown: headings, code blocks, bullet lists, bold/italic
- Enter to send, Shift+Enter for newline
- Full-width layout for Phase 1 — no sidebar scaffold yet (Phase 2 adds the sidebar when needed)

### Multi-turn Context Strategy
- Send the full conversation history with every request — no rolling window or token budget cutoff
- Context growth in long sessions is accepted as a v2 problem (REQUIREMENTS.md MEM-03 covers auto-summarization)
- Conversation resets on browser refresh — in-memory only; Phase 2 handles persistence
- Backend (Go) maintains the in-memory conversation store keyed by session ID — frontend only sends the latest message. This makes the Phase 2 migration to SQLite a store swap, not a protocol change.

### System Prompt
- Short, practical default: e.g. "You are Jarvis, a personal AI assistant. Be concise and helpful."
- System prompt is a configurable field in config.yaml — user can customize the assistant's persona without code changes

### Model Configuration
- go-zero YAML config file (config.yaml) — standard go-zero pattern
- Default out-of-the-box target: local Ollama at http://localhost:11434, model llama3.2
- Zero API key required with defaults — works immediately if Ollama is installed (aligns with local-first project value)
- Config fields: model.baseURL, model.name, model.apiKey, model.systemPrompt

### Loop Guardrail Defaults (SAFE-03)
- Default max tool calls per turn: 10
- Default timeout per turn: 60 seconds
- When a limit is hit: return partial response + clear error message ("I hit my tool call limit for this turn. Try a more specific request.")
- Both limits exposed in config.yaml (maxToolCalls, turnTimeoutSeconds) — user can tune without code changes
- Note: tool calls don't exist in Phase 1, but the guardrail infrastructure is built here so Phase 3 inherits it

### Claude's Discretion
- Streaming protocol: SSE (Server-Sent Events) for unidirectional token streaming — clear technical choice
- Frontend directory structure: separate frontend/ directory or co-located — standard Next.js placement
- Loading/typing indicator design while streaming
- Error state UI for failed API calls

</decisions>

<specifics>
## Specific Ideas

- "Light, polished (ChatGPT-like)" — rounded bubbles, white/light background, consumer-app feel
- Default to Ollama localhost — prioritizes the "private, local-first" project value; no credentials needed to get started

</specifics>

<code_context>
## Existing Code Insights

### Reusable Assets
- None yet — greenfield project. cmd/main.go is empty.

### Established Patterns
- go-zero framework: API definitions go in api/*.api files, handlers in internal/handler/, service logic in internal/service/
- TypeScript strict mode enabled (tsconfig.json) — all new frontend code must compile under strict
- npm as package manager for frontend

### Integration Points
- cmd/main.go: backend entry point — all server initialization, go-zero setup, and route registration start here
- Frontend: not yet created — will be a Next.js app (location to be decided by implementor; frontend/ dir is conventional)
- config.yaml: new file to create — go-zero reads it at startup and maps to a Go struct

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 01-streaming-chat-loop*
*Context gathered: 2026-03-11*
