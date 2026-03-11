---
phase: 07-group-be-to-related-domain-instead-of-flat
plan: 02
subsystem: api
tags: [go, go-zero, handler, domain-grouping, refactor]

# Dependency graph
requires:
  - phase: 07-01
    provides: "Domain logic packages in internal/chat/logic, internal/conv/logic, internal/config/logic"
provides:
  - "17 handler files (including tests) in domain subdirectories: chat/handler, conv/handler, config/handler"
  - "All three domain handler packages compile independently"
affects: ["07-03", "cmd/main.go update phase"]

# Tech tracking
tech-stack:
  added: []
  patterns: [domain-grouped handlers, import alias to avoid same-package-name collision]

key-files:
  created:
    - src/backend/internal/chat/handler/chatstreamhandler.go
    - src/backend/internal/chat/handler/approvehandler.go
    - src/backend/internal/chat/handler/chathandler_test.go
    - src/backend/internal/chat/handler/approvehandler_test.go
    - src/backend/internal/conv/handler/listconvshandler.go
    - src/backend/internal/conv/handler/getconvhandler.go
    - src/backend/internal/conv/handler/getconvmessageshandler.go
    - src/backend/internal/conv/handler/deleteconvhandler.go
    - src/backend/internal/conv/handler/searchconvshandler.go
    - src/backend/internal/conv/handler/deleteconvhandler_test.go
    - src/backend/internal/conv/handler/listconvshandler_test.go
    - src/backend/internal/conv/handler/getconvmessageshandler_test.go
    - src/backend/internal/conv/handler/searchconvshandler_test.go
    - src/backend/internal/config/handler/getconfighandler.go
    - src/backend/internal/config/handler/updateconfighandler.go
    - src/backend/internal/config/handler/getconfighandler_test.go
    - src/backend/internal/config/handler/updateconfighandler_test.go
  modified: []

key-decisions:
  - "chatlogic/convlogic/cfglogic import aliases used in handler files because handler package name (chat/conv/cfg) matches the imported logic package name, requiring explicit alias to disambiguate"
  - "Test files use handler alias (handler 'open-jarvis/internal/chat/handler') so tests can call handler.ChatStreamHandler rather than chat.ChatStreamHandler"
  - "Original internal/handler/ files left intact — cleanup deferred to Plan 03 as specified"

patterns-established:
  - "Import alias pattern: when handler and logic share a package name, use {domain}logic alias for logic import"
  - "Test package convention: package {domain}_test with explicit handler alias for imported package"

requirements-completed: []

# Metrics
duration: 8min
completed: 2026-03-12
---

# Phase 07 Plan 02: Move Handler Files to Domain Subdirectories Summary

**17 handler files (including tests) moved to domain subdirectories (chat/handler, conv/handler, config/handler) with updated package declarations and logic import paths; all three domain packages compile and tests pass**

## Performance

- **Duration:** 8 min
- **Started:** 2026-03-12T00:00:00Z
- **Completed:** 2026-03-12T00:08:00Z
- **Tasks:** 2
- **Files modified:** 17 created

## Accomplishments
- Created internal/chat/handler/ with ChatStreamHandler and ApproveHandler (package chat)
- Created internal/conv/handler/ with 5 conversation handlers + 4 test files (package conv)
- Created internal/config/handler/ with GetConfigHandler and UpdateConfigHandler (package cfg)
- All three domain packages build and pass tests; original internal/handler/ left intact for Plan 03

## Task Commits

Each task was committed atomically:

1. **Task 1: Move chat handler files** - `fa69651` (feat)
2. **Task 2: Move conv and config handler files** - `5cadf42` (feat)

**Plan metadata:** (docs: complete plan — pending)

## Files Created/Modified
- `src/backend/internal/chat/handler/chatstreamhandler.go` - Chat SSE stream handler, package chat, imports chatlogic alias
- `src/backend/internal/chat/handler/approvehandler.go` - Tool approval handler, package chat
- `src/backend/internal/chat/handler/chathandler_test.go` - Chat stream handler test, package chat_test
- `src/backend/internal/chat/handler/approvehandler_test.go` - Approve handler test, package chat_test
- `src/backend/internal/conv/handler/listconvshandler.go` - List conversations handler, package conv
- `src/backend/internal/conv/handler/getconvhandler.go` - Get conversation handler, package conv
- `src/backend/internal/conv/handler/getconvmessageshandler.go` - Get messages handler, package conv
- `src/backend/internal/conv/handler/deleteconvhandler.go` - Delete conversation handler, package conv
- `src/backend/internal/conv/handler/searchconvshandler.go` - Search conversations handler, package conv
- `src/backend/internal/conv/handler/listconvshandler_test.go` - List convs test with mockStore, package conv_test
- `src/backend/internal/conv/handler/deleteconvhandler_test.go` - Delete conv test, package conv_test
- `src/backend/internal/conv/handler/getconvmessageshandler_test.go` - Get messages test, package conv_test
- `src/backend/internal/conv/handler/searchconvshandler_test.go` - Search handler test with SQLite, package conv_test
- `src/backend/internal/config/handler/getconfighandler.go` - Get config handler, package cfg
- `src/backend/internal/config/handler/updateconfighandler.go` - Update config handler, package cfg
- `src/backend/internal/config/handler/getconfighandler_test.go` - Get config handler test, package cfg_test
- `src/backend/internal/config/handler/updateconfighandler_test.go` - Update config handler test, package cfg_test

## Decisions Made

- Used `chatlogic`, `convlogic`, and `cfglogic` import aliases for logic packages because handler package names (`chat`, `conv`, `cfg`) match the imported logic package names, creating a naming collision that Go requires an explicit alias to resolve.
- Test files use `handler` as the import alias for the domain handler package so test code stays readable (`handler.ChatStreamHandler(...)` vs `chat.ChatStreamHandler(...)`).
- Original `internal/handler/` left untouched per plan — cleanup is Plan 03's responsibility.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Added import aliases to avoid package name collision**
- **Found during:** Task 1 (Move chat handler files)
- **Issue:** The plan noted that logic package qualifier changes to `chat.NewChatLogic`, but the handler package itself is also named `chat`, creating an ambiguous reference. Similarly for conv (both `conv`) and config logic (package `cfg` = same as handler package).
- **Fix:** Used `chatlogic`, `convlogic`, `cfglogic` import aliases in handler files; used `handler` alias in `_test` files to reference the handler package under test.
- **Files modified:** All 17 handler and test files
- **Verification:** `go build ./internal/chat/... && go build ./internal/conv/... && go build ./internal/config/...` pass; all tests pass
- **Committed in:** fa69651, 5cadf42

---

**Total deviations:** 1 auto-fixed (Rule 1 - bug: import alias required due to package name collision)
**Impact on plan:** Required fix for correct compilation. No functional change, no scope creep.

## Issues Encountered
- Package name collision between handler and logic packages in the same domain required import aliases — identified and resolved before first build attempt.

## Next Phase Readiness
- All domain handler packages compile independently
- Plan 03 can proceed to update cmd/main.go to reference the new domain handler packages and delete internal/handler/
- Original internal/handler/ still exists and still compiles — no breakage to existing code

---
*Phase: 07-group-be-to-related-domain-instead-of-flat*
*Completed: 2026-03-12*
