# Phase 4: Web Tools and Audit - Research

**Researched:** 2026-03-11
**Domain:** Go HTTP clients, HTML readability parsing, web search APIs, SQLite audit logging
**Confidence:** HIGH

## Summary

Phase 4 adds two new tools to the existing ToolRegistry (`web_fetch` and `web_search`) and wraps all tool executions with an append-only SQLite audit log. The backend already has the full agentic infrastructure from Phase 3: `ToolRegistry`, `Executor` interface, `chatTools` slice in `chatlogic.go`, `ServiceContext`, and the SSE event protocol. Adding web tools follows the exact same pattern as `FileTool` and `ShellTool` — implement a struct with a `Run(ctx, argsJSON) ToolResult` method, register it in `servicecontext.go`, and add a tool definition to `chatTools`.

For web page fetching, `go-shiori/go-readability` (now deprecated in favour of `codeberg.org/readeck/go-readability/v2`) provides a Mozilla Readability-compatible HTML-to-text extractor. For web search, the DuckDuckGo Instant Answer API is free with no API key, but is limited to instant answers rather than full result lists. The Brave Search API offers real results at $5/1k queries (with $5 free credit monthly) and is the best production option. SerpAPI starts at $25/month and is over-engineered for personal use. The recommended approach is: implement `WebFetchTool` using stdlib `net/http` + `go-readability`, implement `WebSearchTool` backed by Brave Search API (key configured via YAML), and build `AuditStore` as a SQLite table using the existing `*sql.DB` already wired into `ServiceContext`.

The audit log is implemented as a new table in the existing SQLite database — no new dependency needed. Insert one row per tool call: `timestamp`, `tool_name`, `args_json`, `result_content`, `result_error`, `session_id`. The `AuditStore` wraps the same `*sql.DB` and is injected into `ServiceContext`. The logging call is added in `chatlogic.go` immediately after each `l.svcCtx.Executor.Execute(...)` call.

**Primary recommendation:** `go-shiori/go-readability` for web fetch (stable, maintained, in Go package index); Brave Search API for web search (real results, configurable API key); SQLite audit table using existing DB connection for SAFE-04.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| TOOL-03 | Agent can fetch and summarize web pages | `web_fetch` tool using `net/http` + `go-readability`; returns `TextContent` field |
| TOOL-04 | Agent can search the web and return results | `web_search` tool using Brave Search API; returns titles + URLs + descriptions |
| SAFE-04 | All tool executions are recorded in an audit log | `AuditStore` backed by new `tool_audit_log` SQLite table; logged in `chatlogic.go` after every `Executor.Execute` call |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/go-shiori/go-readability` | v0.0.0-20251205110129 | HTML-to-readable-text extraction for `web_fetch` | Implements Mozilla Readability.js algorithm; `FromURL` or `FromReader` API; actively maintained; pkg.go.dev listed |
| `net/http` (stdlib) | Go 1.22 | HTTP GET for web fetch and Brave Search API requests | Already used by go-zero; no new dependency |
| `database/sql` + `modernc.org/sqlite` | already in go.mod | Audit log table in existing SQLite DB | Zero new deps; same patterns as `SQLiteConvStore` |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `encoding/json` (stdlib) | Go 1.22 | Parse Brave Search JSON response | Already used throughout codebase |
| `codeberg.org/readeck/go-readability/v2` | v2 | Successor to go-shiori/go-readability | Use if `go-shiori` is unavailable or v2 API needed |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Brave Search API | DuckDuckGo Instant Answer API | DDG is free/no-key but returns only instant answers (abstracts), not search result lists; not suitable for TOOL-04 |
| Brave Search API | SerpAPI | SerpAPI returns richer results from Google but costs $25/month minimum vs $5/1k requests on Brave |
| go-readability | goquery + net/http manual | Lower-level; requires custom HTML traversal; go-readability handles JS-free article extraction correctly |
| go-readability | colly | Colly is a crawler framework, not a readability extractor; overkill for single-page fetch |

**Installation:**
```bash
cd src/backend
go get github.com/go-shiori/go-readability
go mod tidy
```

## Architecture Patterns

### Recommended Project Structure
```
src/backend/internal/
├── toolexec/
│   ├── executor.go          # existing: ToolRegistry, Executor interface
│   ├── executor_test.go     # existing
│   ├── filetool.go          # existing
│   ├── shelltool.go         # existing
│   ├── webtool.go           # NEW: WebFetchTool + WebSearchTool
│   └── webtool_test.go      # NEW: unit tests
└── svc/
    ├── auditstore.go        # NEW: AuditStore struct + SQLite migration
    ├── auditstore_test.go   # NEW: unit tests
    ├── servicecontext.go    # MODIFY: wire new tools + AuditStore
    └── ...
```

### Pattern 1: Tool Implementation (same as FileTool/ShellTool)
**What:** Struct with a `Run(ctx context.Context, argsJSON string) ToolResult` method that JSON-decodes args, does the work, and returns `ToolResult{Content: ...}` or `ToolResult{Error: ...}`.
**When to use:** Every new tool in this project follows this signature.
**Example:**
```go
// Source: existing toolexec/filetool.go and toolexec/shelltool.go patterns
type WebFetchTool struct {
    timeout time.Duration
}

func NewWebFetchTool(timeoutSeconds int) *WebFetchTool {
    if timeoutSeconds <= 0 {
        timeoutSeconds = 30
    }
    return &WebFetchTool{timeout: time.Duration(timeoutSeconds) * time.Second}
}

func (w *WebFetchTool) Fetch(ctx context.Context, argsJSON string) toolexec.ToolResult {
    var args struct{ URL string }
    if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
        return toolexec.ToolResult{Error: "invalid args: " + err.Error()}
    }
    article, err := readability.FromURL(args.URL, w.timeout)
    if err != nil {
        return toolexec.ToolResult{Error: "fetch failed: " + err.Error()}
    }
    return toolexec.ToolResult{Content: article.Title + "\n\n" + article.TextContent}
}
```

### Pattern 2: Tool Registration (same as existing servicecontext.go)
**What:** Instantiate tool in `NewServiceContext`, call `registry.Register("tool_name", tool.Method)`.
**Example:**
```go
// Source: existing svc/servicecontext.go pattern
webFetchTool := toolexec.NewWebFetchTool(c.WebFetchTimeoutSeconds)
webSearchTool := toolexec.NewWebSearchTool(c.BraveSearchAPIKey)
registry.Register("web_fetch", webFetchTool.Fetch)
registry.Register("web_search", webSearchTool.Search)
```

### Pattern 3: chatTools Slice Extension
**What:** Append new `openai.Tool` entries to the `chatTools` var in `chatlogic.go`.
**Example:**
```go
// Source: existing logic/chatlogic.go chatTools var pattern
{Type: openai.ToolTypeFunction, Function: &openai.FunctionDefinition{
    Name:        "web_fetch",
    Description: "Fetch and extract readable text content from a web page URL",
    Parameters: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "url": map[string]any{"type": "string", "description": "Full URL to fetch"},
        },
        "required": []string{"url"},
    },
}},
{Type: openai.ToolTypeFunction, Function: &openai.FunctionDefinition{
    Name:        "web_search",
    Description: "Search the web and return result titles, URLs, and descriptions",
    Parameters: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "query": map[string]any{"type": "string", "description": "Search query"},
        },
        "required": []string{"query"},
    },
}},
```

### Pattern 4: AuditStore with SQLite (append-only table)
**What:** New table `tool_audit_log` in the same SQLite database, write one row per tool invocation. `AuditStore` wraps `*sql.DB`, follows the same `migrate()` + struct pattern as `SQLiteConvStore`.
**Example:**
```go
// Source: existing svc/sqlitestore.go migration pattern
const auditSchema = `
CREATE TABLE IF NOT EXISTS tool_audit_log (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp   INTEGER NOT NULL,
    session_id  TEXT NOT NULL DEFAULT '',
    tool_name   TEXT NOT NULL,
    args_json   TEXT NOT NULL DEFAULT '',
    result      TEXT NOT NULL DEFAULT '',
    error       TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_audit_session ON tool_audit_log(session_id);
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON tool_audit_log(timestamp);
`

type AuditStore struct {
    db *sql.DB
}

func (a *AuditStore) Log(sessionID, toolName, argsJSON, result, errMsg string) error {
    _, err := a.db.Exec(
        `INSERT INTO tool_audit_log(timestamp, session_id, tool_name, args_json, result, error)
         VALUES(?, ?, ?, ?, ?, ?)`,
        time.Now().Unix(), sessionID, toolName, argsJSON, result, errMsg,
    )
    return err
}
```

### Pattern 5: Audit Logging Call Site in chatlogic.go
**What:** After `Executor.Execute(...)` returns, call `AuditStore.Log(...)`. The session ID is available in scope. This captures all tools — file, shell, and web.
**Example:**
```go
// Source: chatlogic.go tool dispatch loop
result := l.svcCtx.Executor.Execute(ctx, tc.Function.Name, tc.Function.Arguments)
// Log the execution to audit trail
if l.svcCtx.AuditStore != nil {
    _ = l.svcCtx.AuditStore.Log(req.SessionId, tc.Function.Name, tc.Function.Arguments, result.Content, result.Error)
}
```

### Pattern 6: Brave Search API HTTP Request
**What:** Simple `net/http` GET with `X-Subscription-Token` header; parse JSON response.
**Example:**
```go
// Source: https://brave.com/search/api/ documentation
type WebSearchTool struct {
    apiKey string
    client *http.Client
}

func (w *WebSearchTool) Search(ctx context.Context, argsJSON string) toolexec.ToolResult {
    var args struct{ Query string }
    if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
        return toolexec.ToolResult{Error: "invalid args: " + err.Error()}
    }
    req, _ := http.NewRequestWithContext(ctx, "GET",
        "https://api.search.brave.com/res/v1/web/search?q="+url.QueryEscape(args.Query)+"&count=5",
        nil)
    req.Header.Set("X-Subscription-Token", w.apiKey)
    req.Header.Set("Accept", "application/json")
    // ... parse response
}
```

### Anti-Patterns to Avoid
- **Returning Go errors from tool Execute:** Tool errors must go in `ToolResult.Error`, not as Go `error` return values. The agentic loop must continue running even when a tool fails. This is the existing convention confirmed in STATE.md decisions.
- **Adding AuditStore to the in-memory ConvStore stub:** The in-memory `ConvStore` used in tests doesn't need audit logging. Use nil-guard `if l.svcCtx.AuditStore != nil`.
- **Importing `chatlogic` types from `toolexec`:** go-zero layer rule — `toolexec` must never import `logic` or `svc` packages; data flows one way.
- **Using colly for single-URL fetch:** Colly is a crawler framework with callbacks and concurrency machinery; `go-readability.FromURL` is a single function call.
- **Using unofficial DuckDuckGo scraper for web search:** Unofficial scrapers get IP-banned; Brave Search API is the appropriate stable alternative.
- **Storing full web page HTML in audit log:** Store only result summary or truncated content. Full HTML can be megabytes. Truncate to ~2000 chars in the audit `result` column.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| HTML-to-text extraction | Custom HTML parser with `golang.org/x/net/html` | `go-shiori/go-readability` | Readability extraction involves scoring paragraphs, removing boilerplate (nav, ads, scripts), and normalising HTML — this is a solved problem |
| Web search results | Scraping Google or Bing HTML | Brave Search API | Scraping search engines violates TOS and will get blocked; Brave provides structured JSON |
| Audit log ID generation | UUID per log entry | `INTEGER PRIMARY KEY AUTOINCREMENT` | SQLite autoincrement is sufficient for local append-only log; UUIDs add no value here |
| HTTP client timeout | Custom `context.WithTimeout` in tool | `http.Client{Timeout: d}` | Standard pattern; also tools already receive a `ctx` from the agentic loop which enforces `TurnTimeoutSeconds` |

**Key insight:** The Go standard library HTTP client is sufficient for both web fetch and Brave Search API calls. The only external dependency needed is `go-readability` for HTML text extraction — everything else reuses existing patterns.

## Common Pitfalls

### Pitfall 1: web_fetch Returns Raw HTML Instead of Text
**What goes wrong:** Returning `article.Content` (HTML) instead of `article.TextContent` (plain text). The LLM receives thousands of tokens of HTML tags.
**Why it happens:** `go-readability` Article struct has both `Content` (HTML) and `TextContent` (plain text) fields.
**How to avoid:** Always use `article.TextContent` in the tool result. Optionally prepend `article.Title`.
**Warning signs:** LLM responses contain HTML tag names or attributes.

### Pitfall 2: Missing BraveSearchAPIKey in Config — Tool Silently Fails
**What goes wrong:** `WebSearchTool` is registered in `ServiceContext` with empty API key; requests return 401; `ToolResult.Error` is set but LLM receives no useful context.
**Why it happens:** Config field is optional; key not set in `etc/config.yaml`.
**How to avoid:** At `NewServiceContext` time, if `BraveSearchAPIKey` is empty, log a warning. Consider returning `ToolResult{Error: "web_search is not configured: set BraveSearchAPIKey in config"}` from the tool itself.
**Warning signs:** All web_search results return errors.

### Pitfall 3: AuditStore Nil Dereference When Using In-Memory Store in Tests
**What goes wrong:** `l.svcCtx.AuditStore` is nil in tests that use `NewServiceContextWithClient` (which does not wire AuditStore).
**Why it happens:** `NewServiceContextWithClient` is a legacy test constructor that doesn't wire AuditStore; `NewServiceContextForTest` does wire it but uses a real tool registry.
**How to avoid:** Guard every AuditStore call: `if l.svcCtx.AuditStore != nil { _ = l.svcCtx.AuditStore.Log(...) }`. Update `NewServiceContextForTest` to also wire a real `AuditStore` or pass a nil-safe stub.
**Warning signs:** Nil pointer panic in `chatlogic_test.go`.

### Pitfall 4: web_fetch Hangs on Slow or Unresponsive URLs
**What goes wrong:** `readability.FromURL` uses the provided timeout, but the agentic loop's `TurnTimeoutSeconds` context may cancel first, producing confusing errors.
**Why it happens:** Two competing timeouts — the per-turn context cancellation and the per-request HTTP timeout.
**How to avoid:** Use `readability.FromReader` with a manually constructed `http.Client` that respects the passed `ctx` via `http.NewRequestWithContext`. Set HTTP client timeout to `min(WebFetchTimeoutSeconds, remaining turn time)` or simply pass the ctx and let it cancel.
**Warning signs:** Tool returns context deadline errors rather than HTTP timeout errors.

### Pitfall 5: Result Truncation Breaks Audit Log Queries
**What goes wrong:** Storing entire web page text content in `tool_audit_log.result` causes the SQLite database to grow rapidly. A single `web_fetch` call can return 50KB+ of text.
**Why it happens:** `article.TextContent` is unbounded.
**How to avoid:** Truncate result stored in audit log to a fixed limit (e.g. 2000 chars). The full content is what the LLM receives; the audit log stores a summary for inspection.
**Warning signs:** `data/conversations.db` grows by megabytes per web_fetch call.

### Pitfall 6: Approval Gate Logic Only Checks shell_run
**What goes wrong:** The approval gate in `chatlogic.go` has a hardcoded `if tc.Function.Name == "shell_run"` check. Web tools bypass the check entirely, which is correct — but if you add the audit call after the shell_run approval handling, you may miss logging denied commands.
**Why it happens:** The audit log call site is after `Executor.Execute`, but denied shell commands never reach `Executor.Execute`.
**How to avoid:** Add a separate audit log entry for denied approval requests too. Pass `result_content = ""` and `error = "user denied"`.
**Warning signs:** Audit log has no entries for denied shell commands.

## Code Examples

### web_fetch Tool (go-readability FromURL)
```go
// Source: https://pkg.go.dev/github.com/go-shiori/go-readability
import readability "github.com/go-shiori/go-readability"

func (w *WebFetchTool) Fetch(ctx context.Context, argsJSON string) ToolResult {
    var args struct{ URL string }
    if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
        return ToolResult{Error: "invalid args: " + err.Error()}
    }
    if args.URL == "" {
        return ToolResult{Error: "url is required"}
    }
    article, err := readability.FromURL(args.URL, w.timeout)
    if err != nil {
        return ToolResult{Error: "fetch failed: " + err.Error()}
    }
    content := article.TextContent
    if len(content) > 8000 {
        content = content[:8000] + "\n[truncated]"
    }
    result := article.Title
    if result != "" {
        result += "\n\n"
    }
    result += content
    return ToolResult{Content: result}
}
```

### web_search Tool (Brave Search API)
```go
// Source: https://brave.com/search/api/ and https://api.search.brave.com/app/documentation/web-search/query
type braveSearchResponse struct {
    Web struct {
        Results []struct {
            Title       string `json:"title"`
            URL         string `json:"url"`
            Description string `json:"description"`
        } `json:"results"`
    } `json:"web"`
}

func (w *WebSearchTool) Search(ctx context.Context, argsJSON string) ToolResult {
    var args struct{ Query string }
    if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
        return ToolResult{Error: "invalid args: " + err.Error()}
    }
    if w.apiKey == "" {
        return ToolResult{Error: "web_search not configured: set BraveSearchAPIKey in config"}
    }
    endpoint := "https://api.search.brave.com/res/v1/web/search?q=" +
        url.QueryEscape(args.Query) + "&count=5"
    req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
    if err != nil {
        return ToolResult{Error: "request error: " + err.Error()}
    }
    req.Header.Set("X-Subscription-Token", w.apiKey)
    req.Header.Set("Accept", "application/json")
    resp, err := w.client.Do(req)
    if err != nil {
        return ToolResult{Error: "search failed: " + err.Error()}
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return ToolResult{Error: fmt.Sprintf("search API error: %d", resp.StatusCode)}
    }
    var result braveSearchResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return ToolResult{Error: "decode error: " + err.Error()}
    }
    var sb strings.Builder
    for i, r := range result.Web.Results {
        fmt.Fprintf(&sb, "%d. %s\n   %s\n   %s\n\n", i+1, r.Title, r.URL, r.Description)
    }
    return ToolResult{Content: sb.String()}
}
```

### AuditStore Migration (adds table to existing DB)
```go
// Source: existing svc/sqlitestore.go migrate() pattern
func (a *AuditStore) migrate() error {
    _, err := a.db.Exec(`
CREATE TABLE IF NOT EXISTS tool_audit_log (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp   INTEGER NOT NULL,
    session_id  TEXT NOT NULL DEFAULT '',
    tool_name   TEXT NOT NULL,
    args_json   TEXT NOT NULL DEFAULT '',
    result      TEXT NOT NULL DEFAULT '',
    error       TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_audit_session   ON tool_audit_log(session_id);
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON tool_audit_log(timestamp);
`)
    return err
}
```

### Config Additions
```go
// Extending internal/config/config.go
type Config struct {
    rest.RestConf
    Model                  ModelConfig
    MaxToolCalls           int      `json:",default=10"`
    TurnTimeoutSeconds     int      `json:",default=60"`
    DBPath                 string   `json:",default=data/conversations.db"`
    ShellAllowlist         []string `json:",optional"`
    ShellDenylist          []string `json:",optional"`
    WorkspaceRoot          string   `json:",default=."`
    BraveSearchAPIKey      string   `json:",optional"`
    WebFetchTimeoutSeconds int      `json:",default=30"`
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `go-shiori/go-readability` | `codeberg.org/readeck/go-readability/v2` | Dec 2025 | go-shiori is deprecated but still works and is on pkg.go.dev; v2 has API-breaking changes; both are viable for Phase 4 |
| Custom audit file logging | SQLite append-only table | Established pattern | Same DB connection reuse; indexed queries; no extra file handles |
| Brave Search API free tier (5000/month) | Credit-based $5/1k (with $5 monthly credit) | Feb 2026 | ~1000 free queries/month for new users; sufficient for personal use |

**Deprecated/outdated:**
- `github.com/gocolly/colly`: Crawler framework — not appropriate for single-URL text extraction; use `go-readability` instead.
- Unofficial DuckDuckGo scraper: Returns only instant answers (abstracts), not full search results; IP banning risk.

## Open Questions

1. **go-readability vs go-readability/v2**
   - What we know: `go-shiori/go-readability` is deprecated (Dec 2025) but still published on pkg.go.dev and functional. `codeberg.org/readeck/go-readability/v2` is the successor with API-breaking changes (Article fields became methods).
   - What's unclear: Whether v2 is indexed by pkg.go.dev and whether `go get` works cleanly from Codeberg.
   - Recommendation: Use `go-shiori/go-readability` for Phase 4 (stable, pkg.go.dev listed, same API patterns). Note deprecation in code comment. Migrate to v2 in a future phase if needed.

2. **Web search when BraveSearchAPIKey is not configured**
   - What we know: Users need to register at api-dashboard.search.brave.com and generate a key.
   - What's unclear: Whether a DuckDuckGo fallback is wanted for local/offline use.
   - Recommendation: Return a clear error from the tool ("web_search not configured") rather than silently failing or falling back. Document key setup in config.yaml comments.

3. **Audit log result size limit**
   - What we know: `web_fetch` can return 8KB+ of text; storing in SQLite is fine for personal use.
   - What's unclear: Whether users want full content or summaries in the audit log.
   - Recommendation: Truncate audit log `result` column to 2000 chars. The LLM still receives the full content; audit log is for inspection.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go built-in `testing` + `github.com/stretchr/testify` v1.11.1 |
| Config file | none (standard `go test`) |
| Quick run command | `cd src/backend && go test ./internal/toolexec/... ./internal/svc/... -run TestWeb -v` |
| Full suite command | `cd src/backend && go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| TOOL-03 | WebFetchTool.Fetch returns title + text from URL | unit (mock server) | `go test ./internal/toolexec/... -run TestWebFetchTool -v` | Wave 0 |
| TOOL-03 | WebFetchTool.Fetch returns error for invalid URL | unit | `go test ./internal/toolexec/... -run TestWebFetchTool_InvalidURL -v` | Wave 0 |
| TOOL-04 | WebSearchTool.Search returns formatted results | unit (mock HTTP) | `go test ./internal/toolexec/... -run TestWebSearchTool -v` | Wave 0 |
| TOOL-04 | WebSearchTool.Search returns error when API key empty | unit | `go test ./internal/toolexec/... -run TestWebSearchTool_NoKey -v` | Wave 0 |
| SAFE-04 | AuditStore.Log inserts row with correct fields | unit | `go test ./internal/svc/... -run TestAuditStore -v` | Wave 0 |
| SAFE-04 | AuditStore.migrate creates table idempotently | unit | `go test ./internal/svc/... -run TestAuditStore_Migrate -v` | Wave 0 |
| SAFE-04 | chatlogic logs every tool execution to audit log | unit (table-driven) | `go test ./internal/logic/... -run TestStreamChat_AuditLog -v` | Wave 0 |

### Sampling Rate
- **Per task commit:** `cd src/backend && go test ./internal/toolexec/... ./internal/svc/... -v`
- **Per wave merge:** `cd src/backend && go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `src/backend/internal/toolexec/webtool_test.go` — covers TOOL-03, TOOL-04 (use `httptest.NewServer` to avoid real HTTP)
- [ ] `src/backend/internal/svc/auditstore_test.go` — covers SAFE-04 (use `t.TempDir()` for SQLite file)
- [ ] `src/backend/go.mod` — add `github.com/go-shiori/go-readability` after `go get`

## Sources

### Primary (HIGH confidence)
- [go-shiori/go-readability pkg.go.dev](https://pkg.go.dev/github.com/go-shiori/go-readability) — API functions `FromURL`, `FromReader`, `Article` struct fields, version v0.0.0-20251205110129
- [Brave Search API docs](https://brave.com/search/api/) — endpoint, authentication header `X-Subscription-Token`, response format
- Existing codebase: `src/backend/internal/toolexec/executor.go`, `shelltool.go`, `filetool.go` — confirmed ToolResult, Executor, Registry patterns
- Existing codebase: `src/backend/internal/svc/sqlitestore.go` — confirmed SQLite migration and transaction patterns
- Existing codebase: `src/backend/internal/logic/chatlogic.go` — confirmed chatTools slice, tool dispatch loop, SSE protocol

### Secondary (MEDIUM confidence)
- [Brave Search API pricing](https://www.implicator.ai/brave-drops-free-search-api-tier-puts-all-developers-on-metered-billing/) — $5/1k requests, $5 monthly credit for new users (as of Feb 2026)
- [go-readability deprecation notice](https://github.com/go-shiori/go-readability) — confirmed deprecated in favour of codeberg.org/readeck/go-readability/v2
- [Golang Web Scraping 2025 guide](https://www.zyte.com/learn/golang-web-scraping-in-2025-tools-techniques-and-best-practices/) — confirmed go-readability + net/http as standard pattern

### Tertiary (LOW confidence)
- DuckDuckGo Instant Answer API (free, no key) — suitable only for instant answers, not full search results; not recommended for TOOL-04

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — go-readability verified on pkg.go.dev; Brave API documented officially; SQLite reuses existing patterns
- Architecture: HIGH — patterns copied from existing Phase 3 code with direct evidence
- Pitfalls: MEDIUM — based on code analysis and general Go patterns; web-specific pitfalls verified against library docs
- Audit log design: HIGH — follows exact same pattern as SQLiteConvStore in codebase

**Research date:** 2026-03-11
**Valid until:** 2026-04-10 (go-readability deprecation status may change; Brave API pricing confirmed stable for 30 days)
