---
phase: 05-configuration-and-search
plan: 03
subsystem: api
tags: [go, go-zero, sqlite, fts5, config, search, http]

requires:
  - phase: 05-01
    provides: SQLiteConvStore.SearchConversations FTS5 method and SearchResult struct
  - phase: 05-02
    provides: ConfigStore with Get/Update and ServiceContext.ConfigStore field

provides:
  - GET /api/config handler returning ConfigResponse JSON (200 or 503 if not configured)
  - PUT /api/config handler persisting UpdateConfigRequest and rebuilding AIClient (204/400/500)
  - GET /api/conversations/search?q= handler returning JSON array of SearchResult (200)
  - GetConfigLogic and UpdateConfigLogic in logic package
  - SearchConvsLogic with ConvSearcher consumer-defined interface
  - RebuildAIClient method on ServiceContext

affects:
  - 05-04 (frontend settings page and sidebar search consume these endpoints)

tech-stack:
  added: []
  patterns:
    - Handler nil-guard for optional ServiceContext fields (ConfigStore nil -> 503)
    - Consumer-defined interface (ConvSearcher in logic package) for type-safe store abstraction
    - Search handler normalises nil results to empty JSON array (never null in response)
    - RebuildAIClient on ServiceContext encapsulates unexported realClient instantiation

key-files:
  created:
    - src/backend/internal/logic/getconfiglogic.go
    - src/backend/internal/logic/updateconfiglogic.go
    - src/backend/internal/logic/searchconvslogic.go
    - src/backend/internal/logic/searchconvslogic_test.go
    - src/backend/internal/handler/getconfighandler.go
    - src/backend/internal/handler/getconfighandler_test.go
    - src/backend/internal/handler/updateconfighandler.go
    - src/backend/internal/handler/updateconfighandler_test.go
    - src/backend/internal/handler/searchconvshandler.go
    - src/backend/internal/handler/searchconvshandler_test.go
  modified:
    - src/backend/internal/types/types.go
    - src/backend/internal/svc/servicecontext.go
    - src/backend/cmd/main.go

key-decisions:
  - "RebuildAIClient added to ServiceContext to encapsulate unexported realClient struct; UpdateConfigLogic calls svcCtx.RebuildAIClient rather than constructing the client directly"
  - "ConvSearcher interface defined in logic package (consumer) not svc package (provider), following interfaces-belong-to-consumers rule from CLAUDE.md"
  - "Search handler returns empty JSON array [] not null when query is empty or no results; nil guard in handler after logic returns nil"
  - "searchconvslogic.go nil-guards results from SQLiteConvStore so empty query passthrough (nil,nil) returns nil,nil rather than empty slice"

patterns-established:
  - "Handler nil-guard pattern: if svcCtx.OptionalField == nil return 503 before calling logic"
  - "Search normalization: handler converts nil results -> [] for consistent JSON response"

requirements-completed: [CHAT-04, MEM-01]

duration: 12min
completed: 2026-03-11
---

# Phase 5 Plan 3: Config and Search HTTP API Endpoints Summary

**Three new REST endpoints wiring ConfigStore and FTS5 SQLiteConvStore to HTTP: GET/PUT /api/config and GET /api/conversations/search with handler+logic+tests and routes registered in main.go**

## Performance

- **Duration:** ~12 min
- **Started:** 2026-03-11T16:00:00Z
- **Completed:** 2026-03-11T16:12:00Z
- **Tasks:** 2
- **Files modified:** 13

## Accomplishments
- GET /api/config returns ConfigResponse JSON from ConfigStore (200); returns 503 when ConfigStore is nil
- PUT /api/config validates JSON body, persists via ConfigStore.Update, rebuilds AIClient via ServiceContext.RebuildAIClient, returns 204
- GET /api/conversations/search?q=term delegates to SQLiteConvStore.SearchConversations via ConvSearcher interface; always returns JSON array (never null)
- Added RebuildAIClient method to ServiceContext to encapsulate unexported realClient construction
- All handler and logic tests pass; full backend suite green, go vet clean, binary builds

## Task Commits

Each task was committed atomically:

1. **Task 1: Add types and implement config handlers + logic** - `100ffb1` (feat)
2. **Task 2: Implement search handler + logic and register all three routes** - `6523d14` (feat)

**Plan metadata:** (docs commit follows)

## Files Created/Modified
- `src/backend/internal/types/types.go` - Added ConfigResponse, UpdateConfigRequest, SearchConvsRequest, SearchResult
- `src/backend/internal/svc/servicecontext.go` - Added RebuildAIClient method
- `src/backend/internal/logic/getconfiglogic.go` - GetConfigLogic delegating to ConfigStore.Get
- `src/backend/internal/logic/updateconfiglogic.go` - UpdateConfigLogic calling ConfigStore.Update + RebuildAIClient
- `src/backend/internal/logic/searchconvslogic.go` - SearchConvsLogic with ConvSearcher interface and svc.SearchResult mapping
- `src/backend/internal/logic/searchconvslogic_test.go` - Tests with real in-memory SQLiteConvStore
- `src/backend/internal/handler/getconfighandler.go` - GET /api/config handler
- `src/backend/internal/handler/getconfighandler_test.go` - 200/503 status tests
- `src/backend/internal/handler/updateconfighandler.go` - PUT /api/config handler
- `src/backend/internal/handler/updateconfighandler_test.go` - 204/400 status tests
- `src/backend/internal/handler/searchconvshandler.go` - GET /api/conversations/search handler
- `src/backend/internal/handler/searchconvshandler_test.go` - 200 with results + empty query tests
- `src/backend/cmd/main.go` - Registered GET /api/config, PUT /api/config, GET /api/conversations/search routes

## Decisions Made
- RebuildAIClient added to ServiceContext to encapsulate the unexported realClient struct; UpdateConfigLogic calls svcCtx.RebuildAIClient(apiKey, baseURL) rather than constructing the client directly in logic
- ConvSearcher interface defined in the logic package (consumer) not svc (provider), per the "interfaces belong to consumers" rule from CLAUDE.md
- Search handler normalises nil logic results to empty JSON array so frontend always receives an array
- searchconvslogic.go nil-guards results before the make() call so empty query passthrough remains nil,nil

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed empty-query nil handling producing empty slice instead of nil**
- **Found during:** Task 2 (SearchConvsLogic tests)
- **Issue:** `make([]types.SearchResult, 0)` created a non-nil empty slice when SQLiteConvStore returned nil; test expected nil,nil for empty query
- **Fix:** Added nil guard `if results == nil { return nil, nil }` before make() in searchconvslogic.go
- **Files modified:** src/backend/internal/logic/searchconvslogic.go
- **Verification:** TestSearchConvsLogic_Empty passes
- **Committed in:** 6523d14 (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 bug fix)
**Impact on plan:** Necessary for correct nil passthrough semantics. No scope creep.

## Issues Encountered
None beyond the nil-slice deviation handled automatically.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All three endpoints are live in main.go and tested
- Frontend (Phase 5 Plan 4) can now consume GET /api/config, PUT /api/config, and GET /api/conversations/search
- No blockers

---
*Phase: 05-configuration-and-search*
*Completed: 2026-03-11*
