---
phase: 04-web-tools-and-audit
plan: 03
subsystem: api
tags: [go, go-zero, sqlite, audit-log, web-tools, sse, agentic-loop]

requires:
  - phase: 04-01
    provides: WebFetchTool and WebSearchTool constructors in toolexec package
  - phase: 04-02
    provides: AuditStore with NewAuditStore and Log method backed by SQLite

provides:
  - ServiceContext.AuditStore field wired to same SQLite DB as ConvStore
  - web_fetch and web_search registered in ToolRegistry for all ServiceContext constructors
  - AuditStore.Log called after every Executor.Execute and after shell_run denial
  - chatTools slice now has 5 entries (read_file, write_file, shell_run, web_fetch, web_search)
  - config.yaml documents BraveSearchAPIKey and WebFetchTimeoutSeconds as optional

affects: [05-polish-and-deployment, chatlogic, servicecontext]

tech-stack:
  added: []
  patterns:
    - "AuditStore uses same *sql.DB as SQLiteConvStore — single DB connection for all persistence"
    - "Nil-guard pattern: `if l.svcCtx.AuditStore != nil` — NewServiceContextWithClient leaves it nil for legacy tests"
    - "Audit result truncated at 2000 chars before Log call to cap DB row size"
    - "NewServiceContextForTest wires in-memory AuditStore via `:memory:` SQLite — audit calls safe in tests"

key-files:
  created:
    - src/backend/internal/svc/servicecontext_test.go
    - src/backend/internal/logic/chatlogic_audit_test.go
  modified:
    - src/backend/internal/svc/servicecontext.go
    - src/backend/internal/logic/chatlogic.go
    - src/backend/etc/config.yaml

key-decisions:
  - "AuditStore field added to ServiceContext struct; nil in NewServiceContextWithClient to preserve backward compat with legacy tests"
  - "NewServiceContextForTest creates in-memory AuditStore via sql.Open(sqlite, :memory:) so audit calls never panic in integration tests"

patterns-established:
  - "Audit-after-execute: every Executor.Execute call is followed by a nil-guarded AuditStore.Log"
  - "Denial audit: shell_run denials logged before continue with empty result and error=denial message"

requirements-completed: [TOOL-03, TOOL-04, SAFE-04]

duration: 4min
completed: 2026-03-11
---

# Phase 4 Plan 3: Wire Web Tools and AuditStore Summary

**AuditStore and web tools (web_fetch, web_search) wired into ServiceContext and agentic loop — every tool execution and shell denial is now recorded to SQLite audit log**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-11T15:40:14Z
- **Completed:** 2026-03-11T15:44:00Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- ServiceContext.AuditStore wired using the same *sql.DB as SQLiteConvStore (no extra connection)
- web_fetch and web_search registered in both NewServiceContext and NewServiceContextForTest
- chatTools has 5 entries; LLM can now invoke web_fetch (url param) and web_search (query param)
- AuditStore.Log called after every Executor.Execute with 2000-char result truncation
- Shell denial path also audited with errMsg="error: user denied command execution"
- AuditStore nil-guard prevents panic in legacy NewServiceContextWithClient code path
- config.yaml documents BraveSearchAPIKey and WebFetchTimeoutSeconds as commented optional fields

## Task Commits

Each task was committed atomically:

1. **Task 1: Wire AuditStore and web tools into ServiceContext** - `c478153` (feat)
2. **Task 2: Add web tool definitions to chatTools and audit every tool execution** - `4fbdc4a` (feat)

**Plan metadata:** (docs commit — see final commit)

_Note: TDD tasks had combined test+implementation commits since tests passed immediately after implementation_

## Files Created/Modified
- `src/backend/internal/svc/servicecontext.go` - Added AuditStore field, web tool registration, in-memory AuditStore for tests
- `src/backend/internal/svc/servicecontext_test.go` - Tests: AuditStore non-nil in ForTest, nil in WithClient, web tools registered
- `src/backend/internal/logic/chatlogic.go` - Added web_fetch/web_search to chatTools; audit calls after Execute and after denial
- `src/backend/internal/logic/chatlogic_audit_test.go` - Tests: web tools resolve, audit nil-guard no-panic
- `src/backend/etc/config.yaml` - Added commented BraveSearchAPIKey and WebFetchTimeoutSeconds

## Decisions Made
- AuditStore left nil in NewServiceContextWithClient to preserve backward compatibility with all existing logic tests that use that constructor
- NewServiceContextForTest creates in-memory AuditStore (`:memory:` SQLite) so audit calls in integration tests are safe without a disk DB
- Audit log truncates result at 2000 chars before storing to cap row size

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required for default operation. BraveSearchAPIKey is optional; web_search returns "not configured" error when key is absent.

## Next Phase Readiness
- All Phase 4 backend components complete: web tools, audit store, full integration
- Running system can search the web, fetch pages, and record every tool call to SQLite
- Phase 5 (polish and deployment) can proceed

---
*Phase: 04-web-tools-and-audit*
*Completed: 2026-03-11*
