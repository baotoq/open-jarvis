---
phase: 07-group-be-to-related-domain-instead-of-flat
plan: 01
subsystem: api
tags: [go, refactoring, domain-grouping, package-structure]

# Dependency graph
requires:
  - phase: 06-add-go-linting
    provides: Clean, lint-passing logic files ready to reorganize
provides:
  - 11 logic files (including tests) copied into domain subdirectories internal/chat/logic, internal/conv/logic, internal/config/logic
  - New domain packages compile independently (chat, conv, cfg)
affects: [07-02-handler-imports, 07-03-cleanup]

# Tech tracking
tech-stack:
  added: []
  patterns: [domain-grouped package structure with chat/conv/cfg package names]

key-files:
  created:
    - src/backend/internal/chat/logic/chatlogic.go
    - src/backend/internal/chat/logic/chatlogic_test.go
    - src/backend/internal/chat/logic/chatlogic_audit_test.go
    - src/backend/internal/conv/logic/listconvslogic.go
    - src/backend/internal/conv/logic/getconvlogic.go
    - src/backend/internal/conv/logic/getconvmessageslogic.go
    - src/backend/internal/conv/logic/deleteconvlogic.go
    - src/backend/internal/conv/logic/searchconvslogic.go
    - src/backend/internal/conv/logic/searchconvslogic_test.go
    - src/backend/internal/config/logic/getconfiglogic.go
    - src/backend/internal/config/logic/updateconfiglogic.go
  modified: []

key-decisions:
  - "Chat domain package named 'chat' (package chat) — matches domain name"
  - "Conv domain package named 'conv' (package conv) — matches domain name"
  - "Config logic package named 'cfg' (package cfg) — avoids collision with internal/config package"
  - "Originals in internal/logic/ preserved — cleanup deferred to Plan 03 after handler imports updated in Plan 02"
  - "searchconvslogic_test.go in conv_test package has standalone mock types (not shared from chat_test) to keep packages independent"

patterns-established:
  - "Domain logic packages reside in internal/{domain}/logic/ with package name matching domain (or shortened alias for cfg)"
  - "Test files use package {domain}_test for black-box testing"

requirements-completed: []

# Metrics
duration: 5min
completed: 2026-03-11
---

# Phase 07 Plan 01: Group Backend Logic into Domain Subdirectories Summary

**11 logic files (including tests) reorganized from flat internal/logic/ into chat/conv/cfg domain packages, all compiling independently**

## Performance

- **Duration:** ~5 min
- **Started:** 2026-03-11T19:10:11Z
- **Completed:** 2026-03-11T19:15:00Z
- **Tasks:** 2
- **Files modified:** 11 (all created)

## Accomplishments
- Chat domain logic (chatlogic.go + 2 test files) copied to internal/chat/logic/ with package chat
- Conv domain logic (5 logic files + 1 test file) copied to internal/conv/logic/ with package conv
- Config domain logic (2 files) copied to internal/config/logic/ with package cfg
- All three new domain packages compile with go build ./internal/chat/... ./internal/conv/... ./internal/config/logic/...
- Original internal/logic/ files untouched; full go build ./... still passes

## Task Commits

Each task was committed atomically:

1. **Task 1: Move chat domain logic files** - `3ac78b1` (feat)
2. **Task 2: Move conv and config domain logic files** - `47cfed5` (feat)

**Plan metadata:** (docs commit follows)

## Files Created/Modified
- `src/backend/internal/chat/logic/chatlogic.go` - Chat streaming logic in package chat
- `src/backend/internal/chat/logic/chatlogic_test.go` - Chat logic tests in package chat_test
- `src/backend/internal/chat/logic/chatlogic_audit_test.go` - Audit/web tool tests in package chat_test
- `src/backend/internal/conv/logic/listconvslogic.go` - List conversations logic in package conv
- `src/backend/internal/conv/logic/getconvlogic.go` - Get conversation logic in package conv
- `src/backend/internal/conv/logic/getconvmessageslogic.go` - Get messages logic in package conv
- `src/backend/internal/conv/logic/deleteconvlogic.go` - Delete conversation logic in package conv
- `src/backend/internal/conv/logic/searchconvslogic.go` - Search conversations logic in package conv
- `src/backend/internal/conv/logic/searchconvslogic_test.go` - Search logic tests in package conv_test
- `src/backend/internal/config/logic/getconfiglogic.go` - Get config logic in package cfg
- `src/backend/internal/config/logic/updateconfiglogic.go` - Update config logic in package cfg

## Decisions Made
- Config logic package uses alias `cfg` (not `config`) to avoid naming collision with the existing `internal/config` package that defines Config structs
- The searchconvslogic_test.go in conv_test redefines mockAIClient locally (minimal stub) rather than importing from chat_test, keeping packages independent
- Originals in internal/logic/ are preserved intact; Plan 02 will update handler imports and Plan 03 will delete the originals

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- All 11 domain logic files exist at new paths with correct package declarations
- New packages compile independently
- Original internal/logic/ still intact so handlers continue to compile
- Plan 02 (handler import updates) can proceed immediately

---
*Phase: 07-group-be-to-related-domain-instead-of-flat*
*Completed: 2026-03-11*
