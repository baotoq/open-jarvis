---
phase: 04-web-tools-and-audit
plan: 01
subsystem: api
tags: [go, go-readability, brave-search, web-fetch, toolexec, http]

# Dependency graph
requires:
  - phase: 03-file-and-shell-tools
    provides: "toolexec package with ToolResult, ToolRegistry, FileTool, ShellTool patterns"
provides:
  - "WebFetchTool: fetches URL and returns readable article title + plain text via go-readability"
  - "WebSearchTool: queries Brave Search API and returns numbered results (title + URL + description)"
  - "Config extended with BraveSearchAPIKey (optional) and WebFetchTimeoutSeconds (default=30)"
  - "go-readability v0.0.0-20251205110129 added as dependency"
affects: [04-02-wire-web-tools, 04-03-audit]

# Tech tracking
tech-stack:
  added:
    - "github.com/go-shiori/go-readability v0.0.0-20251205110129-5db1dc9836f0"
    - "github.com/go-shiori/dom v0.0.0-20230515143342-73569d674e1c"
    - "github.com/araddon/dateparse v0.0.0-20210429162001-6b43995a97de"
    - "github.com/andybalholm/cascadia v1.3.3"
    - "github.com/gogs/chardet v0.0.0-20211120154057-b7413eaefb8f"
  patterns:
    - "Tool struct with injected baseURL field for httptest-based testing without real network"
    - "go-readability.FromURL for HTML-to-plaintext extraction (strips tags, returns Title + TextContent)"
    - "Brave Search API via http.NewRequestWithContext with X-Subscription-Token header"

key-files:
  created:
    - src/backend/internal/toolexec/webtool.go
    - src/backend/internal/toolexec/webtool_test.go
  modified:
    - src/backend/internal/config/config.go
    - src/backend/go.mod
    - src/backend/go.sum

key-decisions:
  - "go-shiori/go-readability used despite deprecation notice; it's functional and on pkg.go.dev — migrate to codeberg.org/readeck/go-readability/v2 in a future phase if needed"
  - "WebSearchTool.baseURL exposed as struct field (not constructor param) to allow httptest override in tests without breaking the public NewWebSearchTool(apiKey) API"
  - "go mod tidy removes unused dependencies; webtool.go must exist (importing go-readability) before tidy is run or the dependency is pruned"

patterns-established:
  - "Tool testability pattern: export baseURL field on HTTP-calling tools so tests can redirect to httptest.NewServer"
  - "Brave API response decoded via braveSearchResponse struct matching the /res/v1/web/search JSON shape"

requirements-completed: [TOOL-03, TOOL-04]

# Metrics
duration: 3min
completed: 2026-03-11
---

# Phase 4 Plan 01: WebFetchTool and WebSearchTool Summary

**go-readability HTML-to-text web fetch tool and Brave Search API integration in toolexec package, with httptest-based unit tests and Config extended for web tool settings**

## Performance

- **Duration:** ~3 min
- **Started:** 2026-03-11T15:11:56Z
- **Completed:** 2026-03-11T15:14:51Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- WebFetchTool.Fetch: fetches any URL via go-readability, returns article title + plain text (no raw HTML tags), truncates to 8000 chars
- WebSearchTool.Search: queries Brave Search API, returns numbered list of results, gracefully errors on empty API key or non-200 response
- Config extended with BraveSearchAPIKey (optional) and WebFetchTimeoutSeconds (default=30) — ready for ServiceContext wiring in Plan 03

## Task Commits

Each task was committed atomically:

1. **Task 1: Fetch go-readability dependency and extend Config** - `0175d45` (chore)
2. **Task 2: Implement WebFetchTool and WebSearchTool with tests** - `b51eec3` (feat)

**Plan metadata:** (docs commit follows)

_Note: TDD tasks — tests written first (RED), then implementation (GREEN)._

## Files Created/Modified

- `src/backend/internal/toolexec/webtool.go` - WebFetchTool and WebSearchTool structs with Fetch/Search methods matching ToolResult interface
- `src/backend/internal/toolexec/webtool_test.go` - 10 unit tests using httptest.NewServer (no real HTTP); tests for empty URL, invalid args, readable text extraction, API errors, empty results
- `src/backend/internal/config/config.go` - Added BraveSearchAPIKey (optional) and WebFetchTimeoutSeconds (default=30) fields after WorkspaceRoot
- `src/backend/go.mod` - Added go-readability and transitive dependencies
- `src/backend/go.sum` - Updated checksums

## Decisions Made

- **go-shiori/go-readability kept despite deprecation:** The library is functional and well-indexed on pkg.go.dev. Migration to codeberg.org/readeck/go-readability/v2 deferred to a future phase.
- **baseURL field on WebSearchTool:** Exposing baseURL as a struct field (not constructor param) lets tests override to httptest.NewServer without changing the public API shape `NewWebSearchTool(apiKey string)`.
- **Dependency ordering:** `go mod tidy` removes unreferenced deps — webtool.go must be created before tidy runs to preserve go-readability in go.mod.

## Deviations from Plan

None - plan executed exactly as written.

The only implementation detail that differed from the plan skeleton was adding the `baseURL` field to `WebSearchTool` to support test redirection via `httptest.NewServer`. This is a testability requirement that the plan implicitly required (tests must not make real HTTP calls) and does not change the public API.

## Issues Encountered

None.

## User Setup Required

To use WebSearchTool, add `BraveSearchAPIKey` to `src/backend/etc/config.yaml`:

```yaml
BraveSearchAPIKey: "your-brave-api-key-here"
```

Get a free key at https://brave.com/search/api/. Without the key, WebSearchTool returns a descriptive error and the agentic loop continues unblocked.

## Next Phase Readiness

- WebFetchTool and WebSearchTool are ready to be registered in ServiceContext (Plan 03)
- Config fields are in place; wiring requires adding tool instantiation in svc/servicecontext.go
- No architectural blockers

---
*Phase: 04-web-tools-and-audit*
*Completed: 2026-03-11*
