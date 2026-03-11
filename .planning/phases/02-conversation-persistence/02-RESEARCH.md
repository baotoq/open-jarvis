# Phase 2: Conversation Persistence - Research

**Researched:** 2026-03-11
**Domain:** SQLite persistence (Go backend) + conversation sidebar UI (Next.js frontend)
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **Storage:** SQLite-backed persistence for conversations (CHAT-03)
- **Sidebar design:** Each entry shows title + relative date (e.g. "Fix my Go code · 2 hours ago"); fixed, always visible, ~240-260px wide; flat list newest-first; no date grouping
- **Delete:** Hover/context menu to delete; no rename action in this phase
- **Session identity:** Active session ID stored in `localStorage`; on load, if stored ID no longer exists in DB, silently auto-create new conversation and clear stale ID
- **Backend generates session IDs (UUID):** Frontend receives ID back in response and stores in localStorage
- **Conversation title:** Auto-titled from first user message, truncated to ~50 chars; no AI-generated title; no rename action

### Claude's Discretion
- Exact SQLite schema (columns, indexes, migrations)
- ConvStore replacement strategy (interface swap vs new implementation)
- Exact sidebar CSS/Tailwind styling (colors, spacing, hover states)
- Loading state while fetching conversation list on load
- Error state if SQLite connection fails at startup

### Deferred Ideas (OUT OF SCOPE)
- Conversation rename (Phase 5 or standalone micro-phase)
- Sidebar collapse/toggle (fixed sidebar only)
- Date-grouped sidebar headers (Today / Yesterday / Older)
- Conversation search (Phase 5 — MEM-01)
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| CHAT-03 | Conversations are persisted to SQLite and survive restarts | SQLite store, ConvStore interface, schema design, migration strategy |
| UI-02 | Conversation history sidebar lets user browse and switch past conversations | Next.js sidebar component, localStorage session management, relative date formatting, new/delete conversation API |
</phase_requirements>

---

## Summary

Phase 2 has two parallel workstreams: a Go backend SQLite persistence layer replacing the in-memory `ConvStore`, and a Next.js frontend with the first real UI (sidebar + chat area). On the backend, the key pattern is extracting a `ConvStore` interface from the existing concrete struct so that both in-memory (tests) and SQLite (production) implementations satisfy it — a clean drop-in swap in `ServiceContext`. Three new HTTP endpoints follow the established go-zero handler/logic pattern: list conversations, get conversation messages, and delete conversation.

On the frontend, this phase bootstraps the entire Next.js application from scratch. The critical complexity is `localStorage` session management: read session ID at app init (client component), validate it against the backend, create a new one if missing, then store the returned ID. The sidebar is a `'use client'` component that fetches the conversation list on mount, showing title + relative date for each entry. `date-fns` provides `formatDistanceToNow` for relative timestamps, and `shadcn/ui` provides the base component primitives.

The SQLite driver choice is `modernc.org/sqlite` (v1.46.1, no CGO) — it enables clean cross-compilation of the Go binary without a C toolchain, which matters for a personal-assistant project that users may build themselves. WAL mode is enabled via DSN pragma for concurrent read performance.

**Primary recommendation:** Define `ConvStore` as an interface in `svc/`, implement `SQLiteConvStore` satisfying it, wire via `NewServiceContext` based on `DBPath` config presence, and use `t.TempDir()` with in-memory SQLite for store tests.

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| modernc.org/sqlite | v1.46.1 | SQLite driver (no CGO) | Pure Go, cross-compiles, stable 2+ years in production, fits personal-tool deployment model |
| github.com/google/uuid | v1.6.0 | UUID generation for session IDs | Already in go.mod as transitive dep; backend-authoritative ID generation |
| date-fns | v4.x | Relative time formatting in frontend | Tree-shakeable, TypeScript-first, `formatDistanceToNow` covers the exact "2 hours ago" pattern needed |
| shadcn/ui | current | Sidebar, button, context menu primitives | Already referenced in project VSCode config; Tailwind v4 + React 19 compatible |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| database/sql (stdlib) | Go stdlib | SQL abstraction layer | Use via modernc.org/sqlite driver — no ORM needed for simple schema |
| clsx + tailwind-merge | current | Tailwind class merging (`cn` util) | Required by shadcn/ui pattern, already common in skills/tailwind-design-system |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| modernc.org/sqlite | mattn/go-sqlite3 | mattn is ~2x faster but requires CGO and a C toolchain — worse DX for self-hosted project |
| modernc.org/sqlite | github.com/glebarez/sqlite | Lesser-known wrapper around modernc; no advantage over modernc directly |
| date-fns | dayjs | dayjs is smaller but date-fns has better TypeScript types and tree-shaking for monorepos |
| shadcn/ui | custom Tailwind components | shadcn/ui is already in the project skill set; faster to compose than hand-roll |

**Installation (Go):**
```bash
cd src
go get modernc.org/sqlite
go mod tidy
```

**Installation (Frontend — when Next.js app is initialized):**
```bash
npm install date-fns
npx shadcn@latest init
npx shadcn@latest add button
```

---

## Architecture Patterns

### Recommended Project Structure

**Backend additions:**
```
src/internal/
├── svc/
│   ├── convstore.go          # ConvStore interface + in-memory impl (existing, refactored)
│   ├── convstore_test.go     # existing tests — must stay green
│   ├── sqlitestore.go        # SQLiteConvStore implementing ConvStore
│   ├── sqlitestore_test.go   # tests using in-memory SQLite (:memory:)
│   └── servicecontext.go     # wires SQLiteConvStore when DBPath is set
├── handler/
│   ├── chathandler.go        # existing
│   ├── listconvshandler.go   # GET /api/conversations
│   ├── getconvhandler.go     # GET /api/conversations/:id
│   └── deleteconvhandler.go  # DELETE /api/conversations/:id
└── logic/
    ├── chatlogic.go          # existing — also persists title on first message
    ├── listconvslogic.go
    ├── getconvlogic.go
    └── deleteconvlogic.go
```

**Frontend (new Next.js app):**
```
frontend/
├── app/
│   ├── layout.tsx            # root layout — wraps with ConversationProvider
│   ├── page.tsx              # main chat page (server component shell)
│   └── globals.css           # Tailwind v4 @import
├── components/
│   ├── Sidebar.tsx           # 'use client' — conversation list
│   ├── ChatArea.tsx          # 'use client' — message display + input
│   └── ui/                   # shadcn/ui generated components
├── lib/
│   ├── api.ts                # typed fetch wrappers for backend endpoints
│   └── utils.ts              # cn() helper
└── hooks/
    └── useSession.ts         # localStorage session ID management
```

### Pattern 1: ConvStore Interface Extraction

**What:** Extract an interface from the existing `*ConvStore` concrete type so `ServiceContext` holds the interface, not the concrete type. Both in-memory and SQLite stores satisfy it.

**When to use:** Any time you have a concrete dependency that needs swapping — Go's standard approach.

**Example:**
```go
// Source: project pattern (go interface segregation)
// svc/convstore.go

// ConversationStore is the interface both in-memory and SQLite stores satisfy.
type ConversationStore interface {
    Get(sessionID string) []openai.ChatCompletionMessage
    Set(sessionID string, msgs []openai.ChatCompletionMessage)
    // Phase 2 additions:
    ListConversations() ([]Conversation, error)
    GetConversation(id string) (*Conversation, error)
    DeleteConversation(id string) error
    CreateConversation(id, title string) error
}

// ServiceContext holds the interface, not the concrete type
type ServiceContext struct {
    Config    config.Config
    AIClient  AIStreamer
    Store     ConversationStore   // replaces *ConvStore
}
```

### Pattern 2: SQLite Store with In-Memory Test DB

**What:** Use `"file::memory:?cache=shared"` or a temp file for isolated per-test DBs. Use `t.TempDir()` to get a unique temp directory per test.

**When to use:** All SQLiteConvStore tests. Avoids test pollution and cleans up automatically.

**Example:**
```go
// Source: Go testing skill + modernc.org/sqlite docs
// svc/sqlitestore_test.go

func newTestStore(t *testing.T) *SQLiteConvStore {
    t.Helper()
    db, err := sql.Open("sqlite", ":memory:")
    require.NoError(t, err)
    t.Cleanup(func() { db.Close() })
    store, err := NewSQLiteConvStore(db)
    require.NoError(t, err)
    return store
}
```

### Pattern 3: SQLite DSN with WAL Mode

**What:** Enable WAL (Write-Ahead Logging) via DSN pragma for concurrent read access (readers don't block writers).

**When to use:** Production SQLite file — single instance, but SSE streaming and list requests may run concurrently.

**Example:**
```go
// Source: https://pkg.go.dev/modernc.org/sqlite (v1.46.1 docs)
dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)", c.DBPath)
db, err := sql.Open("sqlite", dsn)
```

### Pattern 4: go-zero AddRoutes for New Endpoints

**What:** Use `server.AddRoutes([]rest.Route{...})` for the three new conversation endpoints. Regular HTTP (no SSE option needed — only the stream endpoint needs `rest.WithSSE()`).

**When to use:** Adding multiple routes at once in main.go.

**Example:**
```go
// Source: go-zero rest package docs + existing main.go pattern
server.AddRoutes([]rest.Route{
    {Method: http.MethodGet,    Path: "/api/conversations",     Handler: handler.ListConversationsHandler(svcCtx)},
    {Method: http.MethodGet,    Path: "/api/conversations/:id", Handler: handler.GetConversationHandler(svcCtx)},
    {Method: http.MethodDelete, Path: "/api/conversations/:id", Handler: handler.DeleteConversationHandler(svcCtx)},
})
```

### Pattern 5: localStorage Session Init (Next.js Client Component)

**What:** A `useSession` hook reads/writes session ID in localStorage. On mount, validate the ID against the backend; create new conversation if stale or absent.

**When to use:** App initialization — called once in the root client component.

**Example:**
```typescript
// 'use client' — localStorage is only available in browser
// hooks/useSession.ts
export function useSession() {
  const [sessionId, setSessionId] = useState<string | null>(null)

  useEffect(() => {
    const stored = localStorage.getItem('jarvis-session-id')
    if (stored) {
      // Validate against backend — if 404, treat as stale
      fetch(`/api/conversations/${stored}`)
        .then(r => {
          if (r.ok) {
            setSessionId(stored)
          } else {
            // Stale ID: clear and create new session on first message
            localStorage.removeItem('jarvis-session-id')
            setSessionId(null)
          }
        })
    }
    // null sessionId = new session, backend will assign ID on first message
  }, [])

  const persistSessionId = (id: string) => {
    localStorage.setItem('jarvis-session-id', id)
    setSessionId(id)
  }

  return { sessionId, persistSessionId }
}
```

### Pattern 6: Relative Dates with date-fns

**What:** `formatDistanceToNow` from `date-fns` produces "2 hours ago" style strings. Must be rendered client-side to avoid SSR/hydration mismatch on time values.

**When to use:** Sidebar conversation entry date display.

**Example:**
```typescript
// 'use client'
import { formatDistanceToNow } from 'date-fns'

// In sidebar entry component:
const relativeDate = formatDistanceToNow(new Date(conv.updatedAt), { addSuffix: true })
// → "2 hours ago"
```

### Anti-Patterns to Avoid

- **Holding `*ConvStore` (concrete) in ServiceContext:** Prevents interface swap. Change to `ConversationStore` interface immediately.
- **Using `database/sql` without WAL mode:** Default journal mode causes write contention when SSE streaming and list endpoint run concurrently.
- **Generating session IDs on the frontend:** CONTEXT.md locks backend UUID generation. Frontend receives ID in response body.
- **Using `formatDistanceToNow` in a Server Component:** Will cause hydration mismatch because server and client render at different times. Mark sidebar as `'use client'`.
- **Using `cache=shared` in-memory SQLite across parallel test runs:** Can cause cross-test pollution. Use `:memory:` per connection or unique file per test via `t.TempDir()`.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| SQLite driver | CGO sqlite wrapper | `modernc.org/sqlite` | Pure Go, no C toolchain, already battle-tested |
| UUID generation | Custom ID scheme | `github.com/google/uuid` (already in go.mod) | RFC 4122 compliant, collision-free, already available |
| Relative timestamps | Custom time diff logic | `date-fns` `formatDistanceToNow` | Edge cases (DST, pluralization, locale) are solved |
| Sidebar list UI | Custom scroll + hover CSS | `shadcn/ui` primitives | Already in project skills; context menu (for delete) is non-trivial to implement accessibly |
| SQL migration management | Custom migration runner | Inline `CREATE TABLE IF NOT EXISTS` on startup | Single-file schema; no multi-version migrations needed in Phase 2 |

**Key insight:** SQLite schema in this phase is a single table with no migrations — `CREATE TABLE IF NOT EXISTS` in store initialization is sufficient and avoids over-engineering.

---

## Common Pitfalls

### Pitfall 1: SSE Endpoint Returns Session ID

**What goes wrong:** The current `StreamChat` handler writes SSE tokens but never sends the backend-assigned session ID back to the frontend. The frontend cannot store what it never receives.

**Why it happens:** Phase 1 didn't need persistence; session ID was client-supplied and ephemeral.

**How to avoid:** After the stream completes, send a final SSE event with the session ID: `data: {"event":"session","id":"<uuid>"}\n\n`. Or use a separate response header before streaming begins. The simpler approach is a final event at stream end (no header-writing timing issues).

**Warning signs:** Frontend `localStorage` never gets populated; session ID is null on reload.

### Pitfall 2: Multiple database/sql Connections Without WAL

**What goes wrong:** Concurrent SQLite access under default journal mode causes `SQLITE_BUSY` errors — reads and writes block each other.

**Why it happens:** SSE streaming writes the conversation at end of turn; GET /api/conversations may run concurrently.

**How to avoid:** Enable WAL mode in DSN: `_pragma=journal_mode(WAL)`. Also set `busy_timeout`: `_pragma=busy_timeout(5000)`.

**Warning signs:** `database is locked` errors in go-zero logs under concurrent load.

### Pitfall 3: ConvStore Interface Breaking Existing Tests

**What goes wrong:** Changing `ServiceContext.ConvStore *ConvStore` to `Store ConversationStore` breaks `NewServiceContextWithClient` in tests, which constructs `ServiceContext` directly.

**Why it happens:** The field name and type both change.

**How to avoid:** Update `NewServiceContextWithClient` signature to accept `ConversationStore` interface. The existing `*ConvStore` satisfies the interface, so test construction just needs the parameter type updated.

**Warning signs:** Compile errors in `chatlogic_test.go` and `chathandler_test.go`.

### Pitfall 4: Hydration Mismatch from Relative Dates

**What goes wrong:** `formatDistanceToNow` produces different strings on server (build time or request time) vs. client (actual browser time). React throws hydration mismatch error.

**Why it happens:** Time-dependent rendering differs between SSR and client render.

**How to avoid:** Mark the sidebar component as `'use client'`. Do not call `formatDistanceToNow` in any Server Component.

**Warning signs:** Console error: "Text content does not match server-rendered HTML."

### Pitfall 5: Title Truncation Character Boundary

**What goes wrong:** Truncating to 50 bytes instead of 50 runes breaks multi-byte UTF-8 characters (emoji, CJK text) in conversation titles.

**Why it happens:** `msg[:50]` in Go operates on bytes, not characters.

**How to avoid:** Use `[]rune(msg)` conversion: `string([]rune(msg)[:min(50, len([]rune(msg)))])`.

**Warning signs:** Garbled or missing characters at end of titles for non-ASCII messages.

---

## Code Examples

Verified patterns from official sources:

### SQLite Store Initialization

```go
// Source: pkg.go.dev/modernc.org/sqlite v1.46.1
import (
    "database/sql"
    _ "modernc.org/sqlite"
)

func NewSQLiteConvStore(path string) (*SQLiteConvStore, error) {
    dsn := fmt.Sprintf(
        "file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)",
        path,
    )
    db, err := sql.Open("sqlite", dsn)
    if err != nil {
        return nil, fmt.Errorf("open sqlite: %w", err)
    }
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("ping sqlite: %w", err)
    }
    s := &SQLiteConvStore{db: db}
    if err := s.migrate(); err != nil {
        return nil, fmt.Errorf("migrate: %w", err)
    }
    return s, nil
}
```

### Schema (CREATE TABLE IF NOT EXISTS)

```sql
-- conversations table
CREATE TABLE IF NOT EXISTS conversations (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL DEFAULT '',
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL
);

-- messages table (one row per message)
CREATE TABLE IF NOT EXISTS messages (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role            TEXT NOT NULL,
    content         TEXT NOT NULL,
    position        INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_messages_conv_id ON messages(conversation_id, position);
```

### go-zero AddRoutes (multiple routes)

```go
// Source: pkg.go.dev/github.com/zeromicro/go-zero/rest
server.AddRoutes([]rest.Route{
    {Method: http.MethodGet,    Path: "/api/conversations",     Handler: handler.ListConversationsHandler(svcCtx)},
    {Method: http.MethodGet,    Path: "/api/conversations/:id", Handler: handler.GetConversationHandler(svcCtx)},
    {Method: http.MethodDelete, Path: "/api/conversations/:id", Handler: handler.DeleteConversationHandler(svcCtx)},
})
```

### date-fns Relative Time (Client Component)

```typescript
// Source: https://date-fns.org/ — formatDistanceToNow
// Must be in a 'use client' component
import { formatDistanceToNow } from 'date-fns'

function ConvEntry({ title, updatedAt }: { title: string; updatedAt: string }) {
  const relative = formatDistanceToNow(new Date(updatedAt), { addSuffix: true })
  return (
    <div className="flex flex-col px-3 py-2 hover:bg-muted rounded-md cursor-pointer">
      <span className="text-sm font-medium truncate">{title}</span>
      <span className="text-xs text-muted-foreground">{relative}</span>
    </div>
  )
}
```

### Config Addition for DBPath

```go
// src/internal/config/config.go addition
type Config struct {
    rest.RestConf
    Model              ModelConfig
    MaxToolCalls       int    `json:",default=10"`
    TurnTimeoutSeconds int    `json:",default=60"`
    DBPath             string `json:",default=data/conversations.db"`
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| mattn/go-sqlite3 (CGO) | modernc.org/sqlite (pure Go) | ~2022, mainstream 2024 | No C toolchain required; cleaner CI/CD |
| `tailwind.config.ts` | `@theme {}` in CSS (v4) | Tailwind v4 (2024) | Project uses Tailwind v4 per skills directory |
| `forwardRef` in React | `ref` as regular prop (React 19) | React 19 (2024) | shadcn/ui components already use React 19 pattern |
| `@tailwind` directives | `@import "tailwindcss"` | Tailwind v4 | Project must use v4 pattern per tailwind-design-system skill |

**Deprecated/outdated:**
- `tailwind.config.ts`: Do not create — use CSS `@theme` block per tailwind-design-system skill
- `forwardRef`: Not needed in React 19; skills docs explicitly note this

---

## Open Questions

1. **Frontend API URL configuration**
   - What we know: Backend runs on port 8888; frontend is a separate Next.js app
   - What's unclear: No `NEXT_PUBLIC_API_URL` env var convention established yet; no `.env` files exist
   - Recommendation: Use `NEXT_PUBLIC_API_URL=http://localhost:8888` as the convention, defaulting to `http://localhost:8888` in code if unset

2. **Session ID returned from stream endpoint**
   - What we know: The SSE stream writes `data: <token>\n\n` tokens; a final event must carry the session ID for new conversations
   - What's unclear: Exact event format (plain token vs JSON envelope)
   - Recommendation: Send final event as `data: {"done":true,"sessionId":"<uuid>"}\n\n` — minimal protocol extension; frontend must detect `done:true` to extract session ID

3. **Next.js app location in repo**
   - What we know: CLAUDE.md says "frontend not yet implemented"; no `frontend/` directory exists
   - What's unclear: Desired directory name (`frontend/`, `web/`, root-level)
   - Recommendation: Use `frontend/` for clarity, matching the backend's `src/` sibling structure

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go built-in `testing` + `github.com/stretchr/testify` v1.11.1 |
| Config file | none — standard `go test` discovery |
| Quick run command | `cd src && go test ./internal/svc/... ./internal/logic/... -count=1` |
| Full suite command | `cd src && go test ./... -count=1` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CHAT-03 | SQLiteConvStore.Get/Set persists messages across reopen | unit | `cd src && go test ./internal/svc/... -run TestSQLite -v` | Wave 0 |
| CHAT-03 | SQLiteConvStore.ListConversations returns newest first | unit | `cd src && go test ./internal/svc/... -run TestSQLiteList -v` | Wave 0 |
| CHAT-03 | SQLiteConvStore.DeleteConversation removes conversation and messages | unit | `cd src && go test ./internal/svc/... -run TestSQLiteDelete -v` | Wave 0 |
| CHAT-03 | ConvStore interface satisfied by both in-memory and SQLite stores | compile | `cd src && go build ./...` | Wave 0 |
| CHAT-03 | StreamChat assigns UUID session ID when req.SessionId is empty | unit | `cd src && go test ./internal/logic/... -run TestStreamChatNewSession -v` | Wave 0 |
| CHAT-03 | Existing in-memory ConvStore tests remain green after interface refactor | unit | `cd src && go test ./internal/svc/... -run TestConvStore -v` | ✅ (convstore_test.go) |
| UI-02 | GET /api/conversations returns conversation list JSON | integration | `cd src && go test ./internal/handler/... -run TestListConversations -v` | Wave 0 |
| UI-02 | DELETE /api/conversations/:id returns 204 and removes entry | integration | `cd src && go test ./internal/handler/... -run TestDeleteConversation -v` | Wave 0 |
| UI-02 | Frontend sidebar renders conversation entries with title + relative date | manual | Browser inspection | manual-only |
| UI-02 | localStorage session persists across browser reload | manual | Browser tab close + reopen | manual-only |

### Sampling Rate
- **Per task commit:** `cd src && go test ./internal/svc/... ./internal/logic/... -count=1`
- **Per wave merge:** `cd src && go test ./... -count=1`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `src/internal/svc/sqlitestore_test.go` — covers CHAT-03 SQLite store behavior
- [ ] `src/internal/logic/chatlogic_test.go` update — add `TestStreamChatNewSession` for UUID assignment
- [ ] `src/internal/handler/listconvshandler_test.go` — covers UI-02 list endpoint
- [ ] `src/internal/handler/deleteconvhandler_test.go` — covers UI-02 delete endpoint

---

## Sources

### Primary (HIGH confidence)
- `pkg.go.dev/modernc.org/sqlite` — driver registration, DSN options, WAL pragma, v1.46.1 docs
- `pkg.go.dev/github.com/zeromicro/go-zero/rest` — `AddRoutes` method signature, `Route` struct
- `date-fns.org` — `formatDistanceToNow` API, `addSuffix` option
- Project source files (`convstore.go`, `servicecontext.go`, `chatlogic.go`, `main.go`, `config.go`) — direct inspection

### Secondary (MEDIUM confidence)
- `.agents/skills/tailwind-design-system/SKILL.md` — Tailwind v4 CSS-first config, CVA pattern, React 19 ref-as-prop
- `.agents/skills/golang-testing/SKILL.md` — `t.TempDir()`, `t.Cleanup()`, in-memory DB test helper pattern
- `.agents/skills/next-best-practices/SKILL.md` — `'use client'` boundary rules, localStorage client-only access
- WebSearch: modernc.org/sqlite vs mattn/go-sqlite3 tradeoff (multiple sources agree on CGO/performance tradeoff)
- WebSearch: `formatDistanceToNow` hydration mismatch in Next.js SSR (multiple community sources)

### Tertiary (LOW confidence)
- WebSearch: Next.js 15 localStorage session init patterns — principles confirmed but no canonical official example found

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — modernc.org/sqlite version confirmed from pkg.go.dev; uuid already in go.mod; date-fns confirmed from official site
- Architecture: HIGH — patterns derived from existing codebase + official go-zero and modernc docs
- Pitfalls: HIGH — SQLite WAL and hydration mismatches are well-documented; SSE session ID gap identified from direct code inspection

**Research date:** 2026-03-11
**Valid until:** 2026-06-11 (90 days — modernc.org/sqlite and go-zero are stable; Next.js 15 patterns are settled)
