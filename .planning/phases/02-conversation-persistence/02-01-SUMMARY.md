---
phase: 02-conversation-persistence
plan: "01"
subsystem: database
tags: [sqlite, modernc, go, convstore, interface, persistence]

# Dependency graph
requires:
  - phase: 01-streaming-chat-loop
    provides: ConvStore in-memory store + ServiceContext + ChatLogic used by this plan
provides:
  - ConversationStore interface (Get/Set/List/GetConversation/Delete/Create)
  - SQLiteConvStore backed by modernc.org/sqlite (no CGO)
  - Conversation struct with ID/Title/CreatedAt/UpdatedAt
  - DBPath config field with default data/conversations.db
  - ServiceContext.Store field replacing ConvStore field
affects:
  - 02-conversation-persistence (all subsequent plans depend on this interface)

# Tech tracking
tech-stack:
  added: [modernc.org/sqlite v1.46.1, modernc.org/libc, modernc.org/memory, modernc.org/mathutil]
  patterns:
    - ConversationStore interface for storage abstraction (in-memory vs SQLite)
    - TDD RED/GREEN cycle for persistence layer
    - SQLite WAL mode + foreign keys + busy_timeout via DSN pragmas
    - rowid DESC tie-breaking for deterministic ordering when updated_at equals

key-files:
  created:
    - src/internal/svc/sqlitestore.go
    - src/internal/svc/sqlitestore_test.go
  modified:
    - src/internal/svc/convstore.go
    - src/internal/svc/servicecontext.go
    - src/internal/config/config.go
    - src/internal/logic/chatlogic.go
    - src/internal/logic/chatlogic_test.go
    - src/go.mod
    - src/go.sum

key-decisions:
  - "rowid DESC used as secondary sort in ListConversations to ensure deterministic ordering when updated_at timestamps are equal (common in tests)"
  - "ConvStore stub methods return nil/empty to satisfy ConversationStore interface without breaking in-memory path"
  - "ServiceContext.Store replaces ServiceContext.ConvStore field to reflect interface generalization"
  - "DBPath defaulting to data/conversations.db via struct tag; empty string means in-memory fallback"

patterns-established:
  - "Storage abstraction: ConversationStore interface with in-memory (ConvStore) and persistent (SQLiteConvStore) implementations"
  - "SQLite migration pattern: NewSQLiteConvStore runs migrate() on init, CREATE TABLE IF NOT EXISTS"
  - "Transactional Set: upsert conversation + delete+reinsert messages in a single tx"

requirements-completed: [CHAT-03]

# Metrics
duration: 3min
completed: 2026-03-11
---

# Phase 2 Plan 01: ConversationStore Interface + SQLiteConvStore Summary

**ConversationStore interface extracted from ConvStore, with SQLiteConvStore backed by modernc.org/sqlite (no CGO) persisting conversations and messages across process restarts**

## Performance

- **Duration:** ~3 min
- **Started:** 2026-03-11T13:39:30Z
- **Completed:** 2026-03-11T13:42:09Z
- **Tasks:** 2 (TDD RED + GREEN)
- **Files modified:** 9

## Accomplishments

- Extracted `ConversationStore` interface supporting Get/Set/List/GetConversation/Delete/Create methods
- Implemented `SQLiteConvStore` with schema (conversations + messages tables), WAL mode, CASCADE delete, and transactional Set
- Wired SQLite into `ServiceContext.Store` when `DBPath != ""`, falling back to in-memory `ConvStore`
- All 5 new SQLite tests pass; all 4 existing ConvStore tests remain green; full suite green

## Task Commits

1. **RED: failing SQLite tests + sqlite dependency** - `1ad19b6` (test)
2. **GREEN: full implementation** - `dc7a311` (feat)

## Files Created/Modified

- `src/internal/svc/sqlitestore.go` - SQLiteConvStore implementing ConversationStore with schema migration, WAL, transactional Set
- `src/internal/svc/sqlitestore_test.go` - TDD tests: GetEmpty, SetGet, Persists, List (newest first), Delete
- `src/internal/svc/convstore.go` - Added ConversationStore interface + Conversation struct + stub methods on ConvStore
- `src/internal/svc/servicecontext.go` - Store field (ConversationStore), SQLite wiring when DBPath set, interface param on constructor
- `src/internal/config/config.go` - Added DBPath field with default data/conversations.db
- `src/internal/logic/chatlogic.go` - ConvStore.Get/Set -> Store.Get/Set
- `src/internal/logic/chatlogic_test.go` - svcCtx.ConvStore -> svcCtx.Store
- `src/go.mod` + `src/go.sum` - Added modernc.org/sqlite and transitive dependencies

## Decisions Made

- **rowid DESC tie-breaking:** `ListConversations` uses `ORDER BY updated_at DESC, rowid DESC` to guarantee deterministic ordering when timestamps are equal (common in same-second test execution).
- **ConvStore stubs:** In-memory `ConvStore` satisfies `ConversationStore` with no-op stubs for List/GetConversation/Delete/Create — callers won't invoke these in-memory paths during Phase 2 (those endpoints require SQLite).
- **ServiceContext.Store:** Renamed from `ConvStore` to `Store` to reflect that the field is now an abstraction, not a concrete type.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Added rowid DESC secondary sort in ListConversations**
- **Found during:** GREEN phase — TestSQLiteList
- **Issue:** Two conversations created in the same second had equal `updated_at`, causing non-deterministic ordering; test expected `conv-2` first but got `conv-1`
- **Fix:** Added `rowid DESC` as secondary sort key in the `ORDER BY` clause
- **Files modified:** `src/internal/svc/sqlitestore.go`
- **Verification:** `TestSQLiteList` passes consistently
- **Committed in:** `dc7a311` (GREEN commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Minimal fix for deterministic ordering. No scope creep.

## Issues Encountered

None beyond the ordering bug above.

## User Setup Required

None - SQLite is embedded (no CGO, no system library). Production `DBPath` defaults to `data/conversations.db` relative to working directory.

## Next Phase Readiness

- `ConversationStore` interface and `SQLiteConvStore` are the foundation for Phase 2 subsequent plans (conversation listing endpoints, session wiring, etc.)
- No blockers; in-memory `ConvStore` still works for tests that don't need persistence

---
*Phase: 02-conversation-persistence*
*Completed: 2026-03-11*
