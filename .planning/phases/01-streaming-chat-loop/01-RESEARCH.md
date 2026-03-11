# Phase 1: Streaming Chat Loop - Research

**Researched:** 2026-03-11
**Domain:** Go SSE streaming + OpenAI-compatible client + Next.js chat UI
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Chat UI Style**
- Light, polished visual style — white/light background, rounded message bubbles (ChatGPT-like)
- Assistant responses render markdown: headings, code blocks, bullet lists, bold/italic
- Enter to send, Shift+Enter for newline
- Full-width layout for Phase 1 — no sidebar scaffold yet

**Multi-turn Context Strategy**
- Send the full conversation history with every request — no rolling window or token budget cutoff
- Context growth in long sessions is accepted as a v2 problem
- Conversation resets on browser refresh — in-memory only
- Backend (Go) maintains the in-memory conversation store keyed by session ID — frontend only sends the latest message

**System Prompt**
- Short, practical default: "You are Jarvis, a personal AI assistant. Be concise and helpful."
- System prompt is a configurable field in config.yaml

**Model Configuration**
- go-zero YAML config file (config.yaml) — standard go-zero pattern
- Default: local Ollama at http://localhost:11434, model llama3.2
- Zero API key required with defaults
- Config fields: model.baseURL, model.name, model.apiKey, model.systemPrompt

**Loop Guardrail Defaults (SAFE-03)**
- Default max tool calls per turn: 10
- Default timeout per turn: 60 seconds
- When a limit is hit: return partial response + error message
- Both limits exposed in config.yaml (maxToolCalls, turnTimeoutSeconds)
- Note: tool calls don't exist in Phase 1 but the guardrail infrastructure is built here

### Claude's Discretion
- Streaming protocol: SSE (Server-Sent Events) — locked per CONTEXT.md
- Frontend directory structure: separate frontend/ directory or co-located
- Loading/typing indicator design while streaming
- Error state UI for failed API calls

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| CHAT-01 | User can send a message and receive a streaming token-by-token response | SSE handler in go-zero (rest.WithSSE), go-openai CreateChatCompletionStream, EventSource in frontend |
| CHAT-02 | Agent maintains multi-turn conversation context within a session | In-memory conversation store keyed by session ID (sync.RWMutex map), session ID via query param or header, full history passed on every LLM call |
| SAFE-03 | Agent loop is bounded by configurable max tool calls and timeout per turn | context.WithTimeout wrapping the LLM call, configurable limits in config.yaml, Phase 1 has no tool calls but struct is in place for Phase 3 |
</phase_requirements>

---

## Summary

Phase 1 builds the end-to-end streaming chat loop: a Next.js frontend that renders tokens as they arrive, communicating with a go-zero backend via Server-Sent Events. The backend proxies streaming responses from any OpenAI-compatible provider (defaulting to Ollama at localhost) and maintains an in-memory conversation store keyed by session ID so multi-turn context is preserved within a browser session.

The three major implementation surfaces are: (1) the go-zero SSE handler that opens a long-lived HTTP connection and flushes tokens as they arrive from the upstream LLM stream; (2) the OpenAI-compatible client (`sashabaranov/go-openai`) configured with a custom `BaseURL` to target Ollama or any other provider; and (3) the React frontend using the browser's native `EventSource` API to consume the SSE stream and accumulate tokens into a growing message rendered with `react-markdown`.

The success criteria for SAFE-03 are satisfied by wrapping every LLM call in a `context.WithTimeout`. Since Phase 1 has no actual tool calls, a `currentToolCalls` counter struct is scaffolded now and enforced in Phase 3.

**Primary recommendation:** Use go-zero's native SSE route option (`rest.WithSSE()`, available since v1.8.2) and `sashabaranov/go-openai` with `DefaultConfig` + custom `BaseURL`. Serve the Next.js static export (`output: 'export'`) via `rest.WithFileServer` from the same Go binary.

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/zeromicro/go-zero | latest (v1.8+) | HTTP server, config, logging | Project's chosen framework; has native SSE support via rest.WithSSE() |
| github.com/sashabaranov/go-openai | latest | OpenAI-compatible streaming client | De-facto Go client; supports custom BaseURL for Ollama, streaming via CreateChatCompletionStream |
| Next.js | 15.x | Frontend framework | Project's chosen framework; App Router, TypeScript strict mode |
| react-markdown | 9.x | Render LLM markdown output | Standard for rendering markdown in React; handles code blocks, headings, lists |
| tailwindcss | 4.x | Styling | Project skill (tailwind-design-system) targets v4; CSS-first @theme configuration |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| remark-gfm | 4.x | GitHub Flavored Markdown plugin | Enables tables, task lists, strikethrough in react-markdown |
| clsx + tailwind-merge | latest | Class name utilities | Project skill pattern (cn() utility from tailwind-design-system skill) |
| class-variance-authority | latest | Component variant management | Used in project's tailwind-design-system skill for CVA components |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| sashabaranov/go-openai | openai-go (official) | official is newer but go-openai has broader OpenAI-compatible API support and more community patterns for Ollama |
| react-markdown | marked + DOMPurify | react-markdown integrates directly in JSX tree; no XSS sanitization step required |
| native EventSource | @microsoft/fetch-event-source | fetch-event-source adds POST support but SSE is GET; native is sufficient for this use case |

**Installation (Go):**
```bash
go get github.com/sashabaranov/go-openai
```

**Installation (Frontend):**
```bash
npm install react-markdown remark-gfm clsx tailwind-merge class-variance-authority
```

---

## Architecture Patterns

### Recommended Project Structure

```
src/                          # Go backend (existing)
├── cmd/main.go               # Entry point — go-zero server init, route registration
├── etc/config.yaml           # go-zero config file (model, guardrails, server)
├── internal/
│   ├── config/config.go      # Config struct embedding rest.RestConf
│   ├── handler/
│   │   └── chathandler.go    # SSE handler (HTTP concerns only)
│   ├── logic/
│   │   └── chatlogic.go      # LLM call, conversation store, guardrail enforcement
│   ├── svc/
│   │   └── servicecontext.go # Dependency injection (OpenAI client, conversation store)
│   └── types/
│       └── types.go          # Request/response types (or from generated .api file)
└── api/
    └── chat.api              # go-zero API spec definition

frontend/                     # Next.js app (to be created)
├── app/
│   ├── layout.tsx
│   ├── page.tsx              # Main chat page (full-width)
│   └── globals.css           # Tailwind v4 @import + @theme tokens
├── components/
│   ├── ChatWindow.tsx        # Message list with auto-scroll
│   ├── MessageBubble.tsx     # Single message with react-markdown
│   ├── ChatInput.tsx         # Textarea, Enter/Shift+Enter handling
│   └── TypingIndicator.tsx   # Animated dots while streaming
├── lib/
│   └── api.ts                # SSE client helper (EventSource wrapper)
├── next.config.ts            # output: 'export', distDir: '../src/frontend/out'
└── tsconfig.json             # strict: true (already configured)
```

### Pattern 1: go-zero SSE Handler

**What:** An HTTP handler that sets SSE headers, creates a per-request channel, and flushes LLM tokens to the client as they arrive.

**When to use:** All streaming chat endpoints.

**Critical config:** Must use `rest.WithSSE()` (go-zero v1.8.2+) or `rest.WithTimeout(0)` (older) to disable the default request timeout that would kill long-lived SSE connections.

```go
// Source: https://go-zero.dev/guides/http/server/sse/
// internal/handler/chathandler.go

func ChatStreamHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/event-stream")
        w.Header().Set("Cache-Control", "no-cache")
        w.Header().Set("Connection", "keep-alive")
        w.Header().Set("Access-Control-Allow-Origin", "*") // dev only

        var req types.ChatRequest
        if err := httpx.Parse(r, &req); err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
            return
        }

        l := logic.NewChatLogic(r.Context(), svcCtx)
        if err := l.StreamChat(&req, w); err != nil {
            // Stream may be partially written; log but don't re-write headers
            logx.WithContext(r.Context()).Errorf("stream chat error: %v", err)
        }
    }
}
```

**Route registration:**
```go
// Source: https://go-zero.dev/guides/http/server/sse/
// cmd/main.go

server.AddRoute(rest.Route{
    Method:  http.MethodPost,
    Path:    "/api/chat/stream",
    Handler: handler.ChatStreamHandler(svcCtx),
}, rest.WithSSE())
```

**Alternative .api spec (for goctl generation):**
```
syntax = "v1"

type ChatRequest {
    SessionId string `json:"sessionId"`
    Message   string `json:"message"`
}

@server (sse: true)
service chat-api {
    @handler chatStream
    post /api/chat/stream (ChatRequest)
}
```

### Pattern 2: OpenAI-Compatible Client with Ollama

**What:** Configure `sashabaranov/go-openai` to point at Ollama by overriding `BaseURL`. Same code works for OpenAI, Anthropic-compatible, or any v1 endpoint.

```go
// Source: https://pkg.go.dev/github.com/sashabaranov/go-openai#ClientConfig
// internal/svc/servicecontext.go

func buildOpenAIClient(cfg config.ModelConfig) *openai.Client {
    c := openai.DefaultConfig(cfg.APIKey) // empty string is valid for Ollama
    c.BaseURL = cfg.BaseURL               // e.g. "http://localhost:11434/v1"
    return openai.NewClientWithConfig(c)
}
```

**Streaming call in logic layer:**
```go
// Source: https://pkg.go.dev/github.com/sashabaranov/go-openai
// internal/logic/chatlogic.go

func (l *ChatLogic) StreamChat(req *types.ChatRequest, w http.ResponseWriter) error {
    // Apply turn timeout (SAFE-03)
    ctx, cancel := context.WithTimeout(l.ctx, time.Duration(l.svcCtx.Config.TurnTimeoutSeconds)*time.Second)
    defer cancel()

    history := l.svcCtx.ConvStore.Get(req.SessionId)
    history = append(history, openai.ChatCompletionMessage{
        Role:    openai.ChatMessageRoleUser,
        Content: req.Message,
    })

    stream, err := l.svcCtx.AIClient.CreateChatCompletionStream(ctx,
        openai.ChatCompletionRequest{
            Model:    l.svcCtx.Config.Model.Name,
            Messages: history,
            Stream:   true,
        })
    if err != nil {
        return fmt.Errorf("create stream: %w", err)
    }
    defer stream.Close()

    var fullResponse strings.Builder
    flusher := w.(http.Flusher)

    for {
        resp, err := stream.Recv()
        if errors.Is(err, io.EOF) {
            break
        }
        if err != nil {
            return fmt.Errorf("stream recv: %w", err)
        }
        token := resp.Choices[0].Delta.Content
        fullResponse.WriteString(token)
        fmt.Fprintf(w, "data: %s\n\n", token)
        flusher.Flush()
    }

    // Persist assistant turn to conversation store
    history = append(history, openai.ChatCompletionMessage{
        Role:    openai.ChatMessageRoleAssistant,
        Content: fullResponse.String(),
    })
    l.svcCtx.ConvStore.Set(req.SessionId, history)
    return nil
}
```

### Pattern 3: In-Memory Conversation Store

**What:** A thread-safe store keyed by session ID holding the full message history. Session ID is a UUID generated by the frontend and sent with every request.

```go
// internal/svc/convstore.go

type ConvStore struct {
    mu   sync.RWMutex
    data map[string][]openai.ChatCompletionMessage
}

func NewConvStore() *ConvStore {
    return &ConvStore{data: make(map[string][]openai.ChatCompletionMessage)}
}

func (s *ConvStore) Get(sessionID string) []openai.ChatCompletionMessage {
    s.mu.RLock()
    defer s.mu.RUnlock()
    msgs := s.data[sessionID]
    if msgs == nil {
        return nil
    }
    // Return a copy to avoid data races on the slice header
    out := make([]openai.ChatCompletionMessage, len(msgs))
    copy(out, msgs)
    return out
}

func (s *ConvStore) Set(sessionID string, msgs []openai.ChatCompletionMessage) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data[sessionID] = msgs
}
```

### Pattern 4: go-zero Config for Model + Guardrails

```go
// internal/config/config.go

type ModelConfig struct {
    BaseURL      string `json:",default=http://localhost:11434/v1"`
    Name         string `json:",default=llama3.2"`
    APIKey       string `json:",optional"`
    SystemPrompt string `json:",default=You are Jarvis, a personal AI assistant. Be concise and helpful."`
}

type Config struct {
    rest.RestConf                           // Always embed for REST services
    Model               ModelConfig
    MaxToolCalls        int `json:",default=10"`
    TurnTimeoutSeconds  int `json:",default=60"`
}
```

```yaml
# etc/config.yaml
Name: open-jarvis
Host: 0.0.0.0
Port: 8888

Model:
  BaseURL: http://localhost:11434/v1
  Name: llama3.2
  APIKey: ""
  SystemPrompt: "You are Jarvis, a personal AI assistant. Be concise and helpful."

MaxToolCalls: 10
TurnTimeoutSeconds: 60
```

### Pattern 5: Serving Next.js Static Export from Go Binary

**What:** Build Next.js with `output: 'export'`, embed the `out/` directory into the Go binary with `//go:embed`, serve with `rest.WithFileServer`.

```go
// cmd/main.go

//go:embed frontend/out
var frontendFS embed.FS

// In main(), after API routes:
subFS, _ := fs.Sub(frontendFS, "frontend/out")
server.AddRoute(rest.Route{
    Method:  http.MethodGet,
    Path:    "/*",
    Handler: http.FileServer(http.FS(subFS)).ServeHTTP,
})
```

```js
// frontend/next.config.ts
const nextConfig = {
  output: 'export',
  distDir: 'out',
}
export default nextConfig
```

**Build sequence:** `npm run build` (in frontend/) then `go build ./cmd/...`.

Note: `next export` was removed in Next.js 14. Use `output: 'export'` in `next.config.ts` and run `next build`.

### Pattern 6: Frontend SSE Consumer

**What:** The browser's native `EventSource` only supports GET. Since the chat API uses POST, use `fetch` with `ReadableStream` instead.

```typescript
// Source: MDN Streams API / Fetch API
// frontend/lib/api.ts

export async function streamChat(
  sessionId: string,
  message: string,
  onToken: (token: string) => void,
  onDone: () => void,
  onError: (err: Error) => void,
): Promise<void> {
  const res = await fetch('/api/chat/stream', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ sessionId, message }),
  })

  if (!res.ok || !res.body) {
    onError(new Error(`HTTP ${res.status}`))
    return
  }

  const reader = res.body.getReader()
  const decoder = new TextDecoder()

  while (true) {
    const { done, value } = await reader.read()
    if (done) { onDone(); return }
    const chunk = decoder.decode(value, { stream: true })
    // Parse SSE: "data: <token>\n\n"
    const lines = chunk.split('\n')
    for (const line of lines) {
      if (line.startsWith('data: ')) {
        onToken(line.slice(6))
      }
    }
  }
}
```

**Why not native EventSource:** The browser's `EventSource` only supports GET requests. Since the chat endpoint needs a POST body (message + sessionId), use `fetch` with `ReadableStream` to consume the SSE-formatted byte stream.

### Anti-Patterns to Avoid

- **Business logic in SSE handler:** Handler must only parse request and delegate to logic layer. LLM call, context management, and guardrail checks belong in `ChatLogic`.
- **Modifying go-zero generated code:** Customize via the logic layer; generated handlers/routes are overwritten by re-running `goctl`.
- **Storing context in struct:** Pass `ctx context.Context` as the first parameter through all layers, not in a field.
- **Skipping `rest.WithSSE()`:** Without it, go-zero's default timeout middleware terminates the connection before streaming completes.
- **Using `sync.Map` for the conversation store:** `sync.RWMutex` + regular map is preferred when you need to copy slices out safely; `sync.Map` has a less ergonomic API for this pattern.
- **Buffering the full response before flushing:** Flush after every token to deliver the streaming experience.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| OpenAI-compatible HTTP streaming | Custom HTTP client + SSE parser | `sashabaranov/go-openai` | Handles reconnects, error codes, done sentinel, delta parsing |
| Markdown rendering with syntax highlighting | Custom parser | `react-markdown` + `remark-gfm` | XSS safe, extensible, handles all CommonMark edge cases |
| Class name merging | String concatenation | `clsx` + `tailwind-merge` (cn utility) | Handles conditional classes, Tailwind class deduplication |
| Config loading with defaults/validation | Manual YAML parse | `conf.MustLoad` (go-zero built-in) | Handles env overrides, required/optional/default tags, exits on error |

**Key insight:** The LLM streaming protocol has subtle edge cases (done sentinel `[DONE]`, empty delta choices, error events mid-stream). Using `go-openai` abstracts all of these.

---

## Common Pitfalls

### Pitfall 1: SSE Connection Killed by go-zero Timeout Middleware

**What goes wrong:** The go-zero server terminates the HTTP connection after the default 5-second timeout, cutting off the stream mid-response.

**Why it happens:** go-zero applies a global request timeout middleware to all routes. SSE connections are long-lived by design.

**How to avoid:** Register the SSE route with `rest.WithSSE()` (v1.8.2+) or `rest.WithTimeout(0)` (older). Required — no workaround.

**Warning signs:** Stream cuts off cleanly after ~5 seconds regardless of response length.

### Pitfall 2: ResponseWriter Flusher Not Asserted

**What goes wrong:** Tokens are buffered in the HTTP response buffer and delivered in one batch at the end, not token-by-token.

**Why it happens:** `http.ResponseWriter` does not automatically flush; you must cast to `http.Flusher` and call `Flush()` after each `fmt.Fprintf`.

**How to avoid:**
```go
flusher, ok := w.(http.Flusher)
if !ok {
    http.Error(w, "streaming not supported", http.StatusInternalServerError)
    return
}
// ...
flusher.Flush() // after every token write
```

**Warning signs:** All tokens appear at once after the full response completes.

### Pitfall 3: Data Race on Conversation Store Slice

**What goes wrong:** Concurrent requests for the same session ID corrupt the conversation history or cause a panic.

**Why it happens:** Go slices are not safe for concurrent modification. Appending to a slice returned from the store without a copy causes races.

**How to avoid:** Always copy the slice out of the store before appending (see ConvStore.Get pattern above). Use `sync.RWMutex` for read-heavy workloads.

**Warning signs:** `go test -race` reports a race condition on the conversation data.

### Pitfall 4: Next.js API Routes Conflict with Go Backend

**What goes wrong:** `next build` creates API route handlers (e.g. `app/api/`) that conflict with the Go backend's `/api/*` routes.

**Why it happens:** With `output: 'export'`, Next.js cannot serve dynamic API routes — it only produces static HTML/JS/CSS. Any `route.ts` files in `app/api/` will cause the build to fail.

**How to avoid:** Do not create any `app/api/` route handlers in the Next.js app. All API calls go directly to the Go backend. The frontend is a pure static export served from Go.

**Warning signs:** `next build` fails with "API Routes cannot be used with next export".

### Pitfall 5: Session ID Not Persisted Across Page Loads

**What goes wrong:** Every browser navigation generates a new session ID, so conversation history is lost even within a single tab session.

**Why it happens:** Generating the session ID in component state (via `useState`) resets on navigation.

**How to avoid:** Generate the session ID once on app load and store it in `sessionStorage` (clears on tab close, survives navigation). This matches the "resets on browser refresh" contract.

```typescript
// frontend/lib/session.ts
export function getOrCreateSessionId(): string {
  const key = 'jarvis-session-id'
  let id = sessionStorage.getItem(key)
  if (!id) {
    id = crypto.randomUUID()
    sessionStorage.setItem(key, id)
  }
  return id
}
```

### Pitfall 6: react-markdown Re-renders on Every Token

**What goes wrong:** The chat interface becomes slow/janky as the response grows because `react-markdown` re-parses and re-renders the entire accumulated string on each new token.

**Why it happens:** The streaming token is appended to state on every update, triggering a full re-render of the Markdown component tree.

**How to avoid:** Memoize the rendered content. Render accumulating text as plain text while streaming; convert to Markdown only when the `done` event arrives. Or use `useMemo` on the markdown component with a debounced string.

**Warning signs:** CPU usage spikes during streaming; UI response becomes sluggish for long responses.

---

## Code Examples

### go-zero Config Struct with Model and Guardrails

```go
// Source: https://go-zero.dev/en/docs/tutorials/go-zero/configuration/overview
// internal/config/config.go

package config

import "github.com/zeromicro/go-zero/rest"

type ModelConfig struct {
    BaseURL      string `json:",default=http://localhost:11434/v1"`
    Name         string `json:",default=llama3.2"`
    APIKey       string `json:",optional"`
    SystemPrompt string `json:",default=You are Jarvis, a personal AI assistant. Be concise and helpful."`
}

type Config struct {
    rest.RestConf
    Model              ModelConfig
    MaxToolCalls       int `json:",default=10"`
    TurnTimeoutSeconds int `json:",default=60"`
}
```

### Streaming Chat Logic (Complete)

```go
// Source: https://pkg.go.dev/github.com/sashabaranov/go-openai
// internal/logic/chatlogic.go

func (l *ChatLogic) StreamChat(req *types.ChatRequest, w http.ResponseWriter) error {
    ctx, cancel := context.WithTimeout(l.ctx,
        time.Duration(l.svcCtx.Config.TurnTimeoutSeconds)*time.Second)
    defer cancel()

    // Build message history with system prompt prepended
    history := l.svcCtx.ConvStore.Get(req.SessionId)
    if len(history) == 0 {
        history = append(history, openai.ChatCompletionMessage{
            Role:    openai.ChatMessageRoleSystem,
            Content: l.svcCtx.Config.Model.SystemPrompt,
        })
    }
    history = append(history, openai.ChatCompletionMessage{
        Role:    openai.ChatMessageRoleUser,
        Content: req.Message,
    })

    stream, err := l.svcCtx.AIClient.CreateChatCompletionStream(ctx,
        openai.ChatCompletionRequest{
            Model:    l.svcCtx.Config.Model.Name,
            Messages: history,
        })
    if err != nil {
        return fmt.Errorf("create stream: %w", err)
    }
    defer stream.Close()

    flusher := w.(http.Flusher)
    var fullResponse strings.Builder

    for {
        resp, err := stream.Recv()
        if errors.Is(err, io.EOF) {
            break
        }
        if err != nil {
            // Timeout from context will surface here as context.DeadlineExceeded
            fmt.Fprintf(w, "data: [ERROR] %s\n\n", err.Error())
            flusher.Flush()
            return fmt.Errorf("stream recv: %w", err)
        }
        token := resp.Choices[0].Delta.Content
        if token == "" {
            continue
        }
        fullResponse.WriteString(token)
        fmt.Fprintf(w, "data: %s\n\n", token)
        flusher.Flush()
    }

    history = append(history, openai.ChatCompletionMessage{
        Role:    openai.ChatMessageRoleAssistant,
        Content: fullResponse.String(),
    })
    l.svcCtx.ConvStore.Set(req.SessionId, history)
    return nil
}
```

### Chat Input with Enter/Shift+Enter

```typescript
// Source: MDN KeyboardEvent
// frontend/components/ChatInput.tsx
'use client'

import { useState, useRef } from 'react'

interface ChatInputProps {
  onSend: (message: string) => void
  disabled: boolean
}

export function ChatInput({ onSend, disabled }: ChatInputProps) {
  const [value, setValue] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      if (value.trim() && !disabled) {
        onSend(value.trim())
        setValue('')
      }
    }
  }

  return (
    <textarea
      ref={textareaRef}
      value={value}
      onChange={(e) => setValue(e.target.value)}
      onKeyDown={handleKeyDown}
      disabled={disabled}
      placeholder="Message Jarvis..."
      rows={3}
      className="w-full resize-none rounded-xl border border-border bg-background px-4 py-3 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
    />
  )
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `next export` CLI command | `output: 'export'` in next.config.ts + `next build` | Next.js 14 | Must use config flag; export command no longer exists |
| `tailwind.config.ts` | `@theme {}` in CSS file | Tailwind v4 | No JS config file; CSS-first configuration |
| `forwardRef()` wrapper | `ref` as regular prop | React 19 | Simpler component signatures; no forwardRef boilerplate |
| `rest.WithTimeout(0)` for SSE | `rest.WithSSE()` | go-zero v1.8.2 | Cleaner intent; use WithSSE() on any version >= 1.8.2 |
| goctl `@server (sse: true)` | Manual route + `rest.WithSSE()` | go-zero v1.8+ | For SSE, manual AddRoute gives more control; both work |

**Deprecated/outdated:**
- `next export` command: removed in Next.js 14; use `output: 'export'` config key
- `tailwind.config.ts` for v4 projects: replaced by CSS `@theme` block
- `React.forwardRef()`: unnecessary in React 19 projects

---

## Open Questions

1. **go-zero version in use**
   - What we know: `go.mod` has `module open-jarvis` with `go 1.26` but no go-zero dependency yet
   - What's unclear: Which go-zero version will be added; rest.WithSSE() requires v1.8.2+
   - Recommendation: Add go-zero latest (v1.8+) to get rest.WithSSE(); verify with `goctl --version` before generation

2. **Next.js static export SPA fallback**
   - What we know: `output: 'export'` produces static files; Go serves them via http.FileServer
   - What's unclear: If Next.js uses client-side routing, the Go server must serve `index.html` for all non-API paths
   - Recommendation: Add a catch-all route in Go that serves `index.html` for any path not matching `/api/*`

3. **CORS in development**
   - What we know: Frontend dev server runs on port 3000, Go backend on 8888
   - What's unclear: Whether to configure CORS middleware in go-zero for dev, or use Next.js rewrites
   - Recommendation: Add a simple CORS middleware to go-zero for development (restrict in production); or use Next.js `rewrites` in `next.config.ts` to proxy API calls

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go built-in `testing` package (backend); no frontend test framework selected yet |
| Config file | None yet — Wave 0 creates test files |
| Quick run command | `go test ./src/...` |
| Full suite command | `go test -race ./src/...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CHAT-01 | StreamChat writes SSE tokens to ResponseWriter | unit | `go test ./src/internal/logic/... -run TestStreamChat -v` | Wave 0 |
| CHAT-01 | Handler sets correct SSE headers | unit | `go test ./src/internal/handler/... -run TestChatStreamHandler -v` | Wave 0 |
| CHAT-02 | ConvStore.Get returns copy of history | unit | `go test ./src/internal/svc/... -run TestConvStore -v` | Wave 0 |
| CHAT-02 | ConvStore is safe for concurrent access | unit | `go test -race ./src/internal/svc/... -run TestConvStoreConcurrent` | Wave 0 |
| SAFE-03 | Stream call is cancelled when timeout exceeded | unit | `go test ./src/internal/logic/... -run TestStreamChatTimeout -v` | Wave 0 |
| SAFE-03 | Config defaults load correctly | unit | `go test ./src/internal/config/... -run TestConfigDefaults -v` | Wave 0 |

### Sampling Rate

- **Per task commit:** `go test ./src/...`
- **Per wave merge:** `go test -race ./src/...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `src/internal/logic/chatlogic_test.go` — covers CHAT-01, SAFE-03
- [ ] `src/internal/handler/chathandler_test.go` — covers CHAT-01 headers
- [ ] `src/internal/svc/convstore_test.go` — covers CHAT-02
- [ ] `src/internal/config/config_test.go` — covers SAFE-03 defaults

---

## Sources

### Primary (HIGH confidence)

- [go-zero SSE guide](https://go-zero.dev/guides/http/server/sse/) — rest.WithSSE(), route registration, handler pattern
- [sashabaranov/go-openai pkg.go.dev](https://pkg.go.dev/github.com/sashabaranov/go-openai) — ClientConfig, DefaultConfig, CreateChatCompletionStream
- [go-zero configuration docs](https://go-zero.dev/en/docs/tutorials/go-zero/configuration/overview) — conf.MustLoad, config struct tags
- zero-skills skill (`/.claude/skills/zero-skills/references/rest-api-patterns.md`) — Handler/Logic/ServiceContext three-layer pattern
- tailwind-design-system skill (`/.agents/skills/tailwind-design-system/SKILL.md`) — Tailwind v4 CSS-first, cn utility, CVA patterns
- next-best-practices skill (`/.agents/skills/next-best-practices/SKILL.md`) — App Router conventions, RSC boundaries, route handlers

### Secondary (MEDIUM confidence)

- [Ollama OpenAI compatibility docs](https://docs.ollama.com/api/openai-compatibility) — BaseURL format `http://localhost:11434/v1`
- [Next.js discussions: output export](https://github.com/vercel/next.js/discussions/58790) — confirms `next export` removed in Next.js 14, use `output: 'export'`
- go-zero SSE WebFetch — confirmed rest.WithSSE() syntax and the `@server (sse: true)` .api spec syntax

### Tertiary (LOW confidence — needs validation)

- go-zero WithFileServer for embed.FS: found in search result summaries but official doc returned 404; pattern is standard Go `http.FileServer(http.FS(subFS))` which is verified HIGH confidence independently

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — verified via official docs and pkg.go.dev
- Architecture: HIGH — go-zero three-layer pattern from zero-skills skill; SSE pattern from official go-zero docs
- Pitfalls: HIGH — SSE timeout and Flusher issues are well-documented; race condition is standard Go; Next.js export restriction is official
- Frontend patterns: HIGH — from project skills (tailwind-design-system, next-best-practices)

**Research date:** 2026-03-11
**Valid until:** 2026-04-11 (stable APIs; go-zero and go-openai change slowly)
