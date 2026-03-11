---
phase: 01-streaming-chat-loop
plan: 01
subsystem: api
tags: [go, go-zero, openai, sse, streaming, ollama]

# Dependency graph
requires: []
provides:
  - POST /api/chat/stream endpoint serving SSE tokens from any OpenAI-compatible LLM
  - In-memory ConvStore with sync.RWMutex for concurrent-safe session history
  - Config struct with go-zero RestConf embedding, ModelConfig, MaxToolCalls, TurnTimeoutSeconds
  - ServiceContext dependency container with AIStreamer interface for testability
  - ChatLogic with context.WithTimeout guardrail (SAFE-03) and system prompt injection
  - Runnable Go binary: go run ./cmd/... --f etc/config.yaml
affects:
  - 01-02-frontend
  - future-phases

# Tech tracking
tech-stack:
  added:
    - github.com/zeromicro/go-zero v1.10.0 (HTTP server, config, SSE routing)
    - github.com/sashabaranov/go-openai v1.41.2 (OpenAI-compatible streaming client)
  patterns:
    - go-zero three-layer: Handler (HTTP) -> Logic (business) -> ServiceContext (deps)
    - AIStreamer interface in svc package for mock injection in tests
    - ConvStore.Get returns copy to prevent data races on returned slices
    - rest.WithSSE() on route to disable go-zero timeout middleware for long-lived connections
    - system prompt applied in logic layer when session history is empty

key-files:
  created:
    - src/internal/config/config.go
    - src/internal/svc/convstore.go
    - src/internal/svc/servicecontext.go
    - src/internal/types/types.go
    - src/internal/logic/chatlogic.go
    - src/internal/handler/chathandler.go
    - src/cmd/main.go
    - src/etc/config.yaml
    - src/internal/config/config_test.go
    - src/internal/svc/convstore_test.go
    - src/internal/logic/chatlogic_test.go
    - src/internal/handler/chathandler_test.go
  modified:
    - src/go.mod
    - src/go.sum

key-decisions:
  - "AIStreamer interface defined in svc package (not logic) to avoid import cycles while enabling mock injection in both logic and handler tests"
  - "DefaultSystemPrompt moved to config.DefaultSystemPrompt const — go vet rejects struct tags with spaces in default values, so applied at NewServiceContext init time"
  - "TurnTimeoutSeconds=0 in tests produces immediate timeout via context.WithTimeout(ctx, 0) — valid test strategy confirmed by TestStreamChatTimeout"

patterns-established:
  - "Handler pattern: set SSE headers -> parse request with httpx.Parse -> delegate to logic -> log error (never re-write headers)"
  - "Logic pattern: WithTimeout -> build history -> assert Flusher -> call AIClient -> stream tokens -> persist history"
  - "Test pattern: mockAIClient implements svc.AIStreamer; mockStream implements svc.StreamRecver; httptest.NewRecorder as ResponseWriter"

requirements-completed: [CHAT-01, CHAT-02, SAFE-03]

# Metrics
duration: 6min
completed: 2026-03-11
---

# Phase 1 Plan 1: Go Backend Foundation Summary

**Go backend serving POST /api/chat/stream via SSE from Ollama (or any OpenAI-compatible LLM), with per-session conversation history and context.WithTimeout guardrail**

## Performance

- **Duration:** 6 min
- **Started:** 2026-03-11T12:17:14Z
- **Completed:** 2026-03-11T12:22:39Z
- **Tasks:** 2
- **Files modified:** 14 (12 created, 2 modified)

## Accomplishments
- Runnable Go binary (`go run ./cmd/...`) that serves POST /api/chat/stream with SSE token streaming from any OpenAI-compatible provider (defaults to Ollama at localhost:11434)
- Thread-safe in-memory ConvStore (sync.RWMutex + copy-on-read) maintains full conversation history per session, enabling multi-turn context
- SAFE-03 turn timeout guardrail via context.WithTimeout wrapping every LLM call, with configurable TurnTimeoutSeconds in config.yaml
- Full test suite: 10 tests across config, svc, logic, and handler packages — all passing with -race flag; go vet clean

## Task Commits

Each task was committed atomically:

1. **Task 1: Go backend foundation — config, deps, store, and tests** - `efba8a9` (feat)
2. **Task 2: SSE handler and streaming chat logic** - `7c0295c` (feat)

## Files Created/Modified
- `src/internal/config/config.go` - Config struct embedding rest.RestConf with ModelConfig, MaxToolCalls, TurnTimeoutSeconds; DefaultSystemPrompt const
- `src/internal/svc/convstore.go` - Thread-safe conversation store with sync.RWMutex, copy-on-read isolation
- `src/internal/svc/servicecontext.go` - Dependency container; AIStreamer/StreamRecver interfaces; realClient wrapping *openai.Client
- `src/internal/types/types.go` - ChatRequest type (SessionId, Message)
- `src/internal/logic/chatlogic.go` - ChatLogic.StreamChat: timeout, system prompt, history, SSE flush loop, history persist
- `src/internal/handler/chathandler.go` - ChatStreamHandler: SSE headers, httpx.Parse, logic delegation
- `src/cmd/main.go` - Server entry point with rest.WithSSE() route registration
- `src/etc/config.yaml` - Default config: Ollama localhost:11434, model llama3.2, MaxToolCalls 10, TurnTimeoutSeconds 60
- `src/go.mod` / `src/go.sum` - go-zero v1.10.0, go-openai v1.41.2 dependencies

## Decisions Made
- **AIStreamer interface in svc package:** Avoids import cycles (logic imports svc; if svc imported logic for the interface there would be a cycle). Tests in both packages implement svc.AIStreamer and svc.StreamRecver.
- **DefaultSystemPrompt as a const, not struct tag default:** go vet flags struct tag defaults containing spaces ("suspicious space in struct tag value"). Moved to `config.DefaultSystemPrompt` const, applied in `NewServiceContext` when field is empty.
- **TurnTimeoutSeconds=0 for timeout test:** `context.WithTimeout(ctx, 0)` creates an already-expired context, making the mock stream return `context.DeadlineExceeded` on first call — clean test approach without goroutine coordination.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed go vet struct tag violation for SystemPrompt default**
- **Found during:** Task 2 verification (`go vet ./...`)
- **Issue:** `json:",default=You are Jarvis, a personal AI assistant. Be concise and helpful."` — go vet reports "suspicious space in struct tag value"
- **Fix:** Changed SystemPrompt field to `json:",optional"`, added `DefaultSystemPrompt` const, applied default in `NewServiceContext` when field is empty
- **Files modified:** `src/internal/config/config.go`, `src/internal/svc/servicecontext.go`
- **Verification:** `go vet ./...` exits 0; `go test -race ./...` still green
- **Committed in:** `7c0295c` (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (Rule 1 - Bug)
**Impact on plan:** Necessary for go vet compliance. SystemPrompt default value is preserved, just applied at init time rather than via struct tag.

## Issues Encountered
- go-zero's `conf.MustLoad` returns an error about missing go.sum entries when dependencies first added — resolved with `go mod tidy`.

## User Setup Required
None - no external service configuration required. Ollama at localhost:11434 is the default and runs locally.

## Next Phase Readiness
- Backend is ready for consumption by the frontend (Plan 02)
- POST /api/chat/stream accepts `{"sessionId": "uuid", "message": "string"}` and streams `data: TOKEN\n\n` SSE events
- Server starts with `go run ./cmd/... -f etc/config.yaml` from the `src/` directory
- To use OpenAI instead of Ollama: set `Model.BaseURL: https://api.openai.com/v1` and `Model.APIKey: sk-...` in config.yaml

---
*Phase: 01-streaming-chat-loop*
*Completed: 2026-03-11*
