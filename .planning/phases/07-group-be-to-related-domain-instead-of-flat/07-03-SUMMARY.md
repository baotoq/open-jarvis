---
phase: 07-group-be-to-related-domain-instead-of-flat
plan: 03
subsystem: api
tags: [go, go-zero, refactor, domain-grouping]

# Dependency graph
requires:
  - phase: 07-02
    provides: domain handler packages (chat, conv, cfg) that cmd/main.go now imports

provides:
  - cmd/main.go wired to domain handler packages (chat, conv, cfg qualifiers)
  - Flat internal/handler/ and internal/logic/ directories removed
  - Domain-grouped structure is the sole structure: chat/, conv/, config/ each with handler/ and logic/

affects:
  - Any future phase that adds routes or handlers (must use domain packages)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Domain handler packages (chat, conv, cfg) used directly in cmd/main.go route registration"
    - "Go package name (cfg) differs from directory name (config/handler) — no import alias needed"

key-files:
  created: []
  modified:
    - src/backend/cmd/main.go
    - src/backend/internal/chat/logic/chatlogic_test.go
    - src/backend/internal/chat/logic/chatlogic_audit_test.go
    - src/backend/internal/conv/logic/searchconvslogic_test.go

key-decisions:
  - "No import aliases needed in cmd/main.go because chat/conv/cfg package names are already distinct"
  - "Test files for domain logic packages use explicit 'logic' import alias to disambiguate package name from import path segment"

patterns-established:
  - "Route registration in cmd/main.go uses domain qualifiers: chat.Handler, conv.Handler, cfg.Handler"

requirements-completed: []

# Metrics
duration: 5min
completed: 2026-03-12
---

# Phase 07 Plan 03: Wire Domain Handlers and Remove Flat Directories Summary

**cmd/main.go re-wired to import chat/conv/cfg domain handler packages; flat internal/handler/ and internal/logic/ directories deleted; full build and test suite passes.**

## Performance

- **Duration:** ~5 min
- **Started:** 2026-03-12T00:00:00Z
- **Completed:** 2026-03-12T00:05:00Z
- **Tasks:** 1
- **Files modified:** 4 modified, 29 deleted

## Accomplishments

- Updated `cmd/main.go` to import three domain-specific handler packages (`internal/chat/handler`, `internal/conv/handler`, `internal/config/handler`) replacing the single flat `internal/handler` import
- Updated all route handler call sites to use domain qualifiers: `chat.`, `conv.`, `cfg.` (not `config.`)
- Deleted `internal/handler/` (11 files) and `internal/logic/` (11 files) flat directories
- `go build ./...`, `go test ./...`, and `go vet ./...` all pass cleanly

## Task Commits

1. **Task 1: Update cmd/main.go and delete old directories** - `6253324` (feat)

**Plan metadata:** (docs commit to follow)

## Files Created/Modified

- `src/backend/cmd/main.go` - Replaced `internal/handler` import with domain-specific imports; updated all handler call sites to use `chat.`, `conv.`, `cfg.` qualifiers
- `src/backend/internal/chat/logic/chatlogic_test.go` - Added explicit `logic` import alias (carried from 07-02)
- `src/backend/internal/chat/logic/chatlogic_audit_test.go` - Added explicit `logic` import alias (carried from 07-02)
- `src/backend/internal/conv/logic/searchconvslogic_test.go` - Added explicit `logic` import alias (carried from 07-02)
- Deleted: `src/backend/internal/handler/` (11 source files)
- Deleted: `src/backend/internal/logic/` (11 source files)

## Decisions Made

- No import aliases needed in `cmd/main.go` because the declared package names (`chat`, `conv`, `cfg`) are already distinct — Go uses declared names as qualifiers, not directory names
- Test files used explicit `logic` import alias to disambiguate the package name `chatlogic`/`convlogic` from the import path segment; this was a pre-existing uncommitted change from 07-02 that was staged in this commit

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 07 is now complete. The backend codebase has a clean domain-grouped structure: `internal/chat/`, `internal/conv/`, `internal/config/` each with `handler/` and `logic/` subdirectories; no flat `handler/` or `logic/` remnants.
- Ready to proceed with Phase 08 (Tilt) or any subsequent phase.

---
*Phase: 07-group-be-to-related-domain-instead-of-flat*
*Completed: 2026-03-12*
