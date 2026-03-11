# Phase 5: Configuration and Search - Research

**Researched:** 2026-03-11
**Domain:** SQLite FTS5 full-text search, runtime config persistence, settings UI
**Confidence:** HIGH

---

## Summary

Phase 5 delivers three capabilities on top of the fully-wired Phase 4 backend: (1) a backend API for reading and writing the model-provider configuration so users never touch `etc/config.yaml`, (2) a settings UI page in Next.js that exposes those fields, and (3) full-text search over past conversations powered by SQLite FTS5.

The config-persistence problem is straightforward: read/write the existing `etc/config.yaml` file at runtime using `gopkg.in/yaml.v3` (already a transitive dependency). The ServiceContext already holds a `config.Config` value; the GET/PUT `/api/config` endpoints expose the `Model` sub-struct. The hard part is thread safety — file writes must be serialised and the in-memory config value on ServiceContext must be updated atomically.

The FTS5 problem is also well-contained. The existing `messages` table is the canonical content source. An FTS5 external-content virtual table shadows it with triggers that keep the index in sync. A new `GET /api/conversations/search?q=<term>` endpoint returns matching conversation IDs with a snippet. The frontend adds a search input to the Sidebar that calls this endpoint and highlights matches.

**Primary recommendation:** Use SQLite FTS5 content= external table pattern with AFTER INSERT/DELETE/UPDATE triggers; persist config changes directly to the YAML file; drive the settings UI with plain React controlled state and the existing shadcn/ui primitives (no React Hook Form / Zod needed for this form complexity).

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| CHAT-04 | User can configure and switch between OpenAI-compatible model providers (OpenAI, Ollama, Anthropic) from the UI | Config GET/PUT API + YAML file write pattern + ServiceContext hot-reload |
| UI-03 | Settings UI lets user configure model provider, API keys, and preferences | Next.js settings page at `/settings`, plain controlled form with shadcn/ui Input/Button |
| MEM-01 | User can search past conversations via full-text keyword search (SQLite FTS5) | FTS5 external content table on `messages`, search endpoint, Sidebar search input |
</phase_requirements>

---

## Standard Stack

### Core (backend)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| modernc.org/sqlite | v1.46.1 (already) | SQLite driver | Already in use; supports FTS5 |
| gopkg.in/yaml.v3 | transitive (already) | Read/write YAML config at runtime | Already in go.sum; canonical Go YAML library |
| database/sql | stdlib | FTS5 queries | Same connection already used by svc package |

### Core (frontend)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Next.js 16 App Router | 16.1.6 (already) | Settings page route at `/settings` | Already installed |
| shadcn/ui primitives | already installed | Input, Button, Label, Switch for settings form | Already in use in Sidebar/ChatArea |
| lucide-react | already installed | Icons in settings page | Already installed |

### No new dependencies needed
Both the backend and frontend have everything required. The only additions are new Go files and a new Next.js page/route.

**Installation:** No new packages.

---

## Architecture Patterns

### Recommended New Structure
```
src/backend/internal/
├── handler/
│   ├── getconfighandler.go         # GET /api/config
│   ├── updateconfighandler.go      # PUT /api/config
│   └── searchconvshandler.go       # GET /api/conversations/search
├── logic/
│   ├── getconfiglogic.go
│   ├── updateconfiglogic.go
│   └── searchconvslogic.go
├── svc/
│   └── configstore.go              # read/write YAML file with mutex
└── types/
    └── types.go                    # add ConfigResponse, UpdateConfigRequest, SearchResult

src/frontend/
├── app/
│   └── settings/
│       └── page.tsx                # /settings route
├── components/
│   ├── SettingsForm.tsx             # controlled form for model provider fields
│   └── Sidebar.tsx                  # add search input + results
└── lib/
    └── api.ts                       # add getConfig(), updateConfig(), searchConversations()
```

### Pattern 1: SQLite FTS5 External Content Table

**What:** A `messages_fts` virtual table that mirrors the `messages` table. FTS5 builds a search index; triggers keep it in sync. Query returns conversation IDs + snippets, joined back to `conversations` for titles.

**When to use:** When the canonical data already lives in a regular table (messages). External content avoids duplicating data — the FTS shadow tables only store the inverted index.

**Schema (added to SQLiteConvStore.migrate()):**
```sql
-- Source: https://www.sqlite.org/fts5.html §4.4 External Content Tables
CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
    content,
    content='messages',
    content_rowid='id'
);

-- Rebuild triggers
CREATE TRIGGER IF NOT EXISTS messages_ai AFTER INSERT ON messages BEGIN
    INSERT INTO messages_fts(rowid, content) VALUES (new.id, new.content);
END;

CREATE TRIGGER IF NOT EXISTS messages_ad AFTER DELETE ON messages BEGIN
    INSERT INTO messages_fts(messages_fts, rowid, content)
    VALUES ('delete', old.id, old.content);
END;

CREATE TRIGGER IF NOT EXISTS messages_au AFTER UPDATE ON messages BEGIN
    INSERT INTO messages_fts(messages_fts, rowid, content)
    VALUES ('delete', old.id, old.content);
    INSERT INTO messages_fts(rowid, content) VALUES (new.id, new.content);
END;
```

**Query pattern:**
```go
// Source: https://www.sqlite.org/fts5.html §2.1 MATCH operator
const searchSQL = `
SELECT DISTINCT m.conversation_id,
       c.title,
       c.updated_at,
       snippet(messages_fts, 0, '<b>', '</b>', '...', 20) AS snippet
FROM messages_fts
JOIN messages m ON messages_fts.rowid = m.id
JOIN conversations c ON m.conversation_id = c.id
WHERE messages_fts MATCH ?
ORDER BY rank
LIMIT 20
`
```

**Pitfall — existing data on migration:** When `messages_fts` is first created, existing rows in `messages` are NOT automatically indexed. Must run a one-time populate after `CREATE VIRTUAL TABLE`:
```sql
INSERT INTO messages_fts(rowid, content)
SELECT id, content FROM messages;
```

**Pitfall — FTS5 query syntax:** The user's raw input must be sanitised before passing to MATCH. A bare apostrophe or special FTS5 token (`AND`, `OR`, `NOT`, `*`, `"`) causes parse errors. Safest approach: wrap the user input in double-quotes for a literal phrase match, or use the `fts5_tokenize` approach. Simple fix:
```go
// Escape double-quotes in the raw query, then wrap in quotes for phrase search.
// Alternatively: split on whitespace and join with AND for keyword search.
func sanitizeFTSQuery(q string) string {
    q = strings.TrimSpace(q)
    // Replace double-quotes to avoid breaking FTS5 phrase syntax
    q = strings.ReplaceAll(q, `"`, `""`)
    return `"` + q + `"`  // phrase match
}
```

### Pattern 2: Runtime Config Persistence

**What:** A thin `ConfigStore` in the `svc` package that holds a `sync.RWMutex`-protected copy of the mutable `ModelConfig` fields and a path to `etc/config.yaml`. GET reads the mutex-protected copy; PUT writes back to disk then updates the in-memory copy atomically.

**Why not just re-read the file on every request:** The file is the source of truth for restarts, but in-memory reads are faster and avoid file contention.

**Why not update go-zero's `conf` reload mechanism:** go-zero's `conf.MustLoad` is startup-only. There is no hot-reload built in. Writing the file + updating the in-memory struct is the correct pattern.

```go
// svc/configstore.go
type ConfigStore struct {
    mu      sync.RWMutex
    cfg     config.ModelConfig
    cfgPath string   // path to etc/config.yaml for persistence
}

func (cs *ConfigStore) Get() config.ModelConfig {
    cs.mu.RLock()
    defer cs.mu.RUnlock()
    return cs.cfg
}

func (cs *ConfigStore) Update(updated config.ModelConfig) error {
    cs.mu.Lock()
    defer cs.mu.Unlock()
    // Write to file, then update in-memory
    if err := cs.writeYAML(updated); err != nil {
        return err
    }
    cs.cfg = updated
    return nil
}
```

**YAML write approach** using `gopkg.in/yaml.v3`:
```go
// Read full file first, update only Model section, write back.
// This preserves non-Model fields (Host, Port, DBPath, etc.).
import "gopkg.in/yaml.v3"

func (cs *ConfigStore) writeYAML(m config.ModelConfig) error {
    // Read current full config as raw map to preserve all fields
    data, err := os.ReadFile(cs.cfgPath)
    if err != nil {
        return fmt.Errorf("read config: %w", err)
    }
    var raw map[string]any
    if err := yaml.Unmarshal(data, &raw); err != nil {
        return fmt.Errorf("parse config: %w", err)
    }
    raw["Model"] = map[string]any{
        "BaseURL":      m.BaseURL,
        "Name":         m.Name,
        "APIKey":       m.APIKey,
        "SystemPrompt": m.SystemPrompt,
    }
    out, err := yaml.Marshal(raw)
    if err != nil {
        return fmt.Errorf("marshal config: %w", err)
    }
    return os.WriteFile(cs.cfgPath, out, 0644)
}
```

**ServiceContext change:** Add `ConfigStore *ConfigStore` field. Initialise in `NewServiceContext` with the config path. `UpdateConfigLogic` calls `ConfigStore.Update(m)` and then rebuilds the `AIClient` with new credentials.

**AIClient rebuild on config change:** When the user changes `BaseURL` or `APIKey`, the existing `AIClient` must be replaced. Since `ServiceContext.AIClient` is an interface, the update logic can replace it:
```go
func (l *UpdateConfigLogic) UpdateConfig(req *types.UpdateConfigRequest) error {
    updated := config.ModelConfig{
        BaseURL:      req.BaseURL,
        Name:         req.Name,
        APIKey:       req.APIKey,
        SystemPrompt: req.SystemPrompt,
    }
    if err := l.svcCtx.ConfigStore.Update(updated); err != nil {
        return err
    }
    // Rebuild AI client with new credentials
    cfg := openai.DefaultConfig(updated.APIKey)
    cfg.BaseURL = updated.BaseURL
    l.svcCtx.AIClient = &realClient{client: openai.NewClientWithConfig(cfg)}
    return nil
}
```
**Thread safety caveat:** The AIClient replacement is a data race if concurrent streams are in flight. For a single-user personal assistant this is acceptable — document it in code comments. If needed, wrap `AIClient` in an `atomic.Pointer` or a separate `sync.RWMutex`; defer that complexity.

### Pattern 3: Settings Page (Frontend)

**What:** A new `/settings` page (`app/settings/page.tsx`) with a controlled form (plain React `useState`). On mount, calls `GET /api/config` to populate fields. On submit, calls `PUT /api/config`. Uses existing shadcn/ui `Input`, `Button`, `Label` primitives.

**No React Hook Form / Zod needed:** Four text fields with simple required/URL validation does not justify adding two new dependencies. Plain `useState` + `onSubmit` handler is sufficient.

**Navigation:** Add a settings gear icon link in the Sidebar header area. Use Next.js `<Link>` (no `router.push`) for navigation.

```tsx
// app/settings/page.tsx — simplified structure
'use client'
import { useEffect, useState } from 'react'
import { getConfig, updateConfig } from '@/lib/api'

export default function SettingsPage() {
  const [form, setForm] = useState({ baseURL: '', name: '', apiKey: '', systemPrompt: '' })
  const [status, setStatus] = useState<'idle' | 'saving' | 'saved' | 'error'>('idle')

  useEffect(() => {
    getConfig().then(setForm)
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setStatus('saving')
    try {
      await updateConfig(form)
      setStatus('saved')
    } catch {
      setStatus('error')
    }
  }
  // ... render Input fields for each form property
}
```

### Pattern 4: Conversation Search (Frontend)

**What:** A search input in the Sidebar. As the user types (debounced ~300ms), call `GET /api/conversations/search?q=<term>`. Show matching conversations instead of the full list. Clicking a result navigates to that conversation.

**Debounce:** Use a simple `useRef` + `setTimeout` debounce — no external library needed.

```tsx
const timerRef = useRef<ReturnType<typeof setTimeout>>()
const handleSearch = (q: string) => {
  setQuery(q)
  clearTimeout(timerRef.current)
  if (!q.trim()) { setResults(null); return }
  timerRef.current = setTimeout(async () => {
    const res = await searchConversations(q)
    setResults(res)
  }, 300)
}
```

### Anti-Patterns to Avoid

- **FTS5 without initial populate:** Creating the virtual table on an existing database without `INSERT INTO messages_fts SELECT...` leaves existing messages unsearchable. MUST populate in migrate().
- **Passing raw user input to MATCH:** Special FTS5 characters (`"`, `*`, `AND`, `OR`, `NOT`) cause `sql: no rows` or panics. Always sanitise/escape.
- **Rebuilding the FTS table on every search:** The virtual table is persistent; no rebuild needed per query.
- **Writing full config struct blindly:** Marshalling the entire `config.Config` struct may strip go-zero struct tag defaults. Safer to read the raw YAML map, update only the `Model` key, and write back.
- **Mutating `svc.Config.Model` directly without mutex:** `Config` is a value type on `ServiceContext`; concurrent access from multiple goroutines requires synchronisation.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Full-text search tokenisation | Custom string matching on content | SQLite FTS5 | FTS5 handles stemming, stopwords, ranking, prefix queries; custom matching is O(n) per query |
| Search result ranking | Count keyword frequency manually | FTS5 `rank` column + `ORDER BY rank` | FTS5 BM25-like ranking is built in |
| Search result highlighting | String replace in Go | SQLite `snippet()` function | Handles edge cases (overlapping matches, multi-token matches) |
| YAML manipulation | Regex replace in config file | `gopkg.in/yaml.v3` marshal/unmarshal | Correct round-trip, handles quoting, preserves structure |
| Debounce | `setTimeout` reimplementation | Plain `useRef` + `setTimeout` | Sufficient for this use case |

**Key insight:** SQLite FTS5 is a first-class extension in SQLite, not an add-on. It is already available in `modernc.org/sqlite` without any additional build flags or CGO.

---

## Common Pitfalls

### Pitfall 1: FTS5 not populated for existing rows
**What goes wrong:** After adding `CREATE VIRTUAL TABLE messages_fts ...` to `migrate()`, a fresh database works but an existing database (with prior messages) returns zero search results.
**Why it happens:** FTS5 external content tables do not auto-index pre-existing rows — only rows inserted after trigger creation are indexed.
**How to avoid:** After `CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts`, run a conditional populate:
```sql
INSERT INTO messages_fts(rowid, content)
SELECT id, content FROM messages
WHERE id NOT IN (SELECT rowid FROM messages_fts);
```
**Warning signs:** `MATCH` returns 0 rows on a database that clearly has matching messages.

### Pitfall 2: FTS5 MATCH syntax errors on user input
**What goes wrong:** User types `hello "world` (unclosed quote) or `NOT` (reserved word) — SQLite returns an error that propagates as a 500.
**Why it happens:** FTS5 MATCH has its own query language. Raw user strings are not safe to pass directly.
**How to avoid:** Wrap user input in double-quotes (escaping internal double-quotes) for phrase-mode search.
**Warning signs:** `sql: parse error` in logs when searching for specific strings.

### Pitfall 3: Config file path not known to ConfigStore
**What goes wrong:** At startup, `main.go` passes `*configFile` flag to `conf.MustLoad` but does not pass it to `NewServiceContext`. ConfigStore has no path to write back.
**Why it happens:** The config path is only used in `main.go` today; it does not flow into ServiceContext.
**How to avoid:** Pass the config file path to `NewServiceContext` (add a `configPath string` parameter) so it can initialise `ConfigStore`.
**Warning signs:** Config PUT succeeds in-memory but changes are lost on restart.

### Pitfall 4: Race condition on AIClient replacement
**What goes wrong:** A streaming response is in flight when the user saves new config; the new `AIClient` reference replaces the old one mid-stream.
**Why it happens:** `ServiceContext.AIClient` is a bare interface field with no synchronisation.
**How to avoid:** For a single-user assistant, document the limitation; tell users to wait for the stream to finish before changing config. Optionally use `atomic.Pointer[svc.AIStreamer]` for correctness.
**Warning signs:** Panic or corrupted stream after config save.

### Pitfall 5: modernc.org/sqlite and FTS5 support
**What goes wrong:** Assume FTS5 is disabled or requires a build tag.
**Why it is NOT a problem:** `modernc.org/sqlite` compiles the full SQLite amalgamation as pure Go. FTS5 is enabled by default (SQLITE_ENABLE_FTS5). No build tags or CGO required.
**How to confirm:** `SELECT fts5_version()` returns a version string without error.

### Pitfall 6: go-zero struct tag defaults stripped on YAML round-trip
**What goes wrong:** Reading the config struct, marshalling it with `gopkg.in/yaml.v3`, and writing back strips go-zero `json:",default=..."` metadata and results in empty values for optional fields.
**Why it happens:** `yaml.Marshal` of a Go struct only serialises current field values, not tag defaults. Fields with empty-string values will write as `""`.
**How to avoid:** Use the raw-map approach: `yaml.Unmarshal` to `map[string]any`, update only the `Model` key, `yaml.Marshal` back. This preserves all existing non-Model fields verbatim.

---

## Code Examples

### FTS5 Migration (full schema addition)
```go
// Source: https://www.sqlite.org/fts5.html §4.4
const ftsSchema = `
CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
    content,
    content='messages',
    content_rowid='id'
);
CREATE TRIGGER IF NOT EXISTS messages_ai AFTER INSERT ON messages BEGIN
    INSERT INTO messages_fts(rowid, content) VALUES (new.id, new.content);
END;
CREATE TRIGGER IF NOT EXISTS messages_ad AFTER DELETE ON messages BEGIN
    INSERT INTO messages_fts(messages_fts, rowid, content)
        VALUES ('delete', old.id, old.content);
END;
CREATE TRIGGER IF NOT EXISTS messages_au AFTER UPDATE ON messages BEGIN
    INSERT INTO messages_fts(messages_fts, rowid, content)
        VALUES ('delete', old.id, old.content);
    INSERT INTO messages_fts(rowid, content) VALUES (new.id, new.content);
END;
-- Populate existing rows (idempotent)
INSERT INTO messages_fts(rowid, content)
SELECT id, content FROM messages
WHERE id NOT IN (SELECT rowid FROM messages_fts);
`
```

### Search Query
```go
// Source: https://www.sqlite.org/fts5.html §2.1, §4.4
const searchSQL = `
SELECT DISTINCT
    m.conversation_id,
    c.title,
    c.updated_at,
    snippet(messages_fts, 0, '<b>', '</b>', '...', 20) AS snippet
FROM messages_fts
JOIN messages m ON messages_fts.rowid = m.id
JOIN conversations c ON m.conversation_id = c.id
WHERE messages_fts MATCH ?
ORDER BY rank
LIMIT 20
`

func sanitizeFTSQuery(q string) string {
    q = strings.TrimSpace(q)
    if q == "" {
        return ""
    }
    q = strings.ReplaceAll(q, `"`, `""`)
    return `"` + q + `"`
}
```

### go-zero GET handler with query param
```go
// handler/searchconvshandler.go
// Source: existing project pattern (listconvshandler.go) + go-zero httpx.ParseForm
type SearchRequest struct {
    Query string `form:"q"`
}

func SearchConversationsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.SearchRequest
        if err := httpx.ParseForm(r, &req); err != nil {
            http.Error(w, "bad request", http.StatusBadRequest)
            return
        }
        l := logic.NewSearchConvsLogic(r.Context(), svcCtx)
        results, err := l.Search(req.Query)
        if err != nil {
            http.Error(w, "internal server error", http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(results)
    }
}
```

### API types additions
```go
// types/types.go additions

// ConfigResponse is the response for GET /api/config.
type ConfigResponse struct {
    BaseURL      string `json:"baseURL"`
    Name         string `json:"name"`
    APIKey       string `json:"apiKey"`
    SystemPrompt string `json:"systemPrompt"`
}

// UpdateConfigRequest is the request body for PUT /api/config.
type UpdateConfigRequest struct {
    BaseURL      string `json:"baseURL"`
    Name         string `json:"name"`
    APIKey       string `json:"apiKey"`
    SystemPrompt string `json:"systemPrompt"`
}

// SearchRequest for GET /api/conversations/search?q=<term>
type SearchRequest struct {
    Query string `form:"q"`
}

// SearchResult is one matching conversation in a search response.
type SearchResult struct {
    ID        string `json:"id"`
    Title     string `json:"title"`
    UpdatedAt int64  `json:"updatedAt"`
    Snippet   string `json:"snippet"`
}
```

### Frontend api.ts additions
```typescript
export interface ModelConfig {
  baseURL: string
  name: string
  apiKey: string
  systemPrompt: string
}

export interface SearchResult {
  id: string
  title: string
  updatedAt: number
  snippet: string
}

export async function getConfig(): Promise<ModelConfig> {
  const res = await fetch(`${API_BASE}/api/config`)
  if (!res.ok) throw new Error(`getConfig failed: ${res.status}`)
  return res.json()
}

export async function updateConfig(cfg: ModelConfig): Promise<void> {
  const res = await fetch(`${API_BASE}/api/config`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(cfg),
  })
  if (!res.ok) throw new Error(`updateConfig failed: ${res.status}`)
}

export async function searchConversations(q: string): Promise<SearchResult[]> {
  const res = await fetch(`${API_BASE}/api/conversations/search?q=${encodeURIComponent(q)}`)
  if (!res.ok) throw new Error(`search failed: ${res.status}`)
  return res.json()
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| FTS3/FTS4 virtual tables | FTS5 (SQLite >=3.9, 2015) | SQLite 3.9.0 | Better ranking, faster, external content tables, snippet() |
| Reload process for config changes | In-memory update + file write | Pattern established in this codebase | No restart needed for provider switch |
| React Hook Form + Zod for all forms | Plain React useState for simple forms | Ongoing | Avoids 2 new npm deps for 4-field form |

---

## Open Questions

1. **API key display security**
   - What we know: The settings UI will show `APIKey` from the backend response.
   - What's unclear: Whether to mask the key in the GET response (return `""` or `"***"`) vs return the real value for editing.
   - Recommendation: Return the real value — this is a single-user local app with no auth. Document the tradeoff.

2. **Config file path in ServiceContext**
   - What we know: `main.go` has the `configFile` flag string. `NewServiceContext` currently takes only `config.Config`.
   - What's unclear: Whether to add `configPath string` to `NewServiceContext` or use a separate `WithConfigPath` option.
   - Recommendation: Add `configPath string` as a second parameter to `NewServiceContext`. Update `main.go` to pass it. Keep `NewServiceContextForTest` and `NewServiceContextWithClient` unchanged (they don't need file write-back).

3. **FTS5 snippet HTML escaping**
   - What we know: `snippet()` returns `<b>...</b>` markers as raw HTML.
   - What's unclear: The frontend must render this HTML safely.
   - Recommendation: Use `dangerouslySetInnerHTML` with the snippet value — it originates from the local SQLite database (user's own messages), so XSS risk is self-inflicted. Alternatively strip to plain text.

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | go test |
| Config file | none — standard Go test files |
| Quick run command | `cd src/backend && go test ./internal/...` |
| Full suite command | `cd src/backend && go test ./...` |
| Estimated runtime | ~5-10 seconds |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| MEM-01 | FTS5 schema migration and populate | unit | `cd src/backend && go test ./internal/svc/ -run TestFTS` | ❌ Wave 0 |
| MEM-01 | Search returns matching conversation IDs + snippets | unit | `cd src/backend && go test ./internal/svc/ -run TestSearch` | ❌ Wave 0 |
| MEM-01 | Search sanitises FTS5 special chars without error | unit | `cd src/backend && go test ./internal/svc/ -run TestSearchSanitize` | ❌ Wave 0 |
| MEM-01 | Search logic integration | unit | `cd src/backend && go test ./internal/logic/ -run TestSearchConvs` | ❌ Wave 0 |
| CHAT-04 | ConfigStore Get/Update round-trip | unit | `cd src/backend && go test ./internal/svc/ -run TestConfigStore` | ❌ Wave 0 |
| CHAT-04 | ConfigStore writes YAML and reads back correctly | unit | `cd src/backend && go test ./internal/svc/ -run TestConfigStoreYAML` | ❌ Wave 0 |
| CHAT-04 | GetConfig handler returns 200 with config fields | unit | `cd src/backend && go test ./internal/handler/ -run TestGetConfig` | ❌ Wave 0 |
| CHAT-04 | UpdateConfig handler returns 204 and updates store | unit | `cd src/backend && go test ./internal/handler/ -run TestUpdateConfig` | ❌ Wave 0 |
| UI-03 | Settings page loads and saves — manual browser test | manual-only | n/a | n/a |
| MEM-01 | Search input in sidebar returns results — manual browser test | manual-only | n/a | n/a |

### Sampling Rate
- **Per task commit:** `cd src/backend && go test ./internal/...`
- **Per wave merge:** `cd src/backend && go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `src/backend/internal/svc/configstore_test.go` — covers CHAT-04 ConfigStore unit tests
- [ ] `src/backend/internal/svc/search_test.go` — covers MEM-01 FTS5 schema + query tests
- [ ] `src/backend/internal/handler/getconfighandler_test.go` — covers CHAT-04 handler
- [ ] `src/backend/internal/handler/updateconfighandler_test.go` — covers CHAT-04 handler
- [ ] `src/backend/internal/handler/searchconvshandler_test.go` — covers MEM-01 handler
- [ ] `src/backend/internal/logic/searchconvslogic_test.go` — covers MEM-01 logic

*Existing `go test` infrastructure covers framework needs. No new test framework needed.*

---

## Sources

### Primary (HIGH confidence)
- https://www.sqlite.org/fts5.html — FTS5 documentation: external content tables, triggers, MATCH syntax, snippet(), official spec
- `modernc.org/sqlite v1.46.1` in go.sum — confirmed driver supports FTS5 (pure Go SQLite amalgamation)
- Existing codebase: `src/backend/internal/svc/sqlitestore.go` — migration pattern, `*sql.DB` usage
- Existing codebase: `src/backend/internal/config/config.go` — `ModelConfig` struct fields
- Existing codebase: `src/backend/cmd/main.go` — config path flow, `NewServiceContext` call
- `gopkg.in/yaml.v3` in go.sum (transitive) — confirmed available for YAML read/write

### Secondary (MEDIUM confidence)
- https://pkg.go.dev/github.com/zeromicro/go-zero/rest/httpx — `ParseForm` for query string params, `form:` struct tag
- https://ui.shadcn.com/docs/forms/react-hook-form — confirms React Hook Form + shadcn/ui pattern (used to decide AGAINST it for simple 4-field form)
- https://www.sqlitetutorial.net/sqlite-full-text-search/ — FTS5 CREATE VIRTUAL TABLE examples

### Tertiary (LOW confidence)
- WebSearch: go-zero runtime config write — no official example found; raw-map YAML approach derived from first principles

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all libraries already in the project; FTS5 confirmed available in modernc.org/sqlite
- Architecture patterns: HIGH — FTS5 external content pattern from official SQLite docs; config persistence pattern derived from existing codebase conventions
- Pitfalls: HIGH — FTS5 initial-populate pitfall documented in official SQLite forum; MATCH sanitisation from SQLite FTS5 spec; config-path pitfall from direct code inspection
- Frontend patterns: HIGH — existing shadcn/ui and Next.js App Router already established in project

**Research date:** 2026-03-11
**Valid until:** 2026-09-11 (SQLite FTS5 spec is stable; go-zero patterns stable; 6-month estimate)
