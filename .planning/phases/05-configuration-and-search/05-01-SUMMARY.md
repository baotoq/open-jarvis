---
phase: 05-configuration-and-search
plan: 01
subsystem: database
tags: [sqlite, fts5, full-text-search, migration, go]

requires:
  - phase: 02-conversation-persistence
    provides: SQLiteConvStore with conversations and messages tables

provides:
  - FTS5 virtual table messages_fts with trigger-based sync
  - SearchConversations method on SQLiteConvStore returning ranked results with snippets
  - SanitizeFTSQuery helper for safe FTS5 MATCH input handling

affects: [05-configuration-and-search plan 03 (search endpoint), any plan adding conversation search UI]

tech-stack:
  added: []
  patterns:
    - "FTS5 content table with content_rowid pointing to messages.id; triggers maintain index on INSERT/DELETE/UPDATE"
    - "migrate() splits FTS DDL into separate db.Exec calls to handle trigger BEGIN...END syntax"
    - "FTS5 'rebuild' command used for idempotent initial populate (NOT IN guard doesn't work on content tables)"
    - "SanitizeFTSQuery wraps input in double-quotes and doubles internal quotes for phrase-match safety"

key-files:
  created:
    - src/backend/internal/svc/search_test.go
  modified:
    - src/backend/internal/svc/sqlitestore.go

key-decisions:
  - "FTS5 'rebuild' command used for initial populate instead of WHERE id NOT IN guard: content tables reflect underlying table rowids even before explicit indexing, making the guard always return 0 rows"
  - "SanitizeFTSQuery exported (not unexported) to allow black-box test coverage from package svc_test"
  - "FTS DDL split into separate db.Exec calls: trigger BEGIN...END syntax unreliable in single multi-statement Exec"
  - "fts5_version() scalar not available in modernc.org/sqlite; FTS5 functionality verified via MATCH query in TestFTSMigration"

patterns-established:
  - "FTS5 migration appended to existing migrate() by extending ftsStatements slice with separate Exec per statement"

requirements-completed: [MEM-01]

duration: 4min
completed: 2026-03-11
---

# Phase 5 Plan 01: FTS5 Full-Text Search Summary

**SQLite FTS5 virtual table with trigger-based sync and phrase-safe SearchConversations method for ranking past conversations by relevance**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-11T15:26:29Z
- **Completed:** 2026-03-11T15:31:21Z
- **Tasks:** 2 (TDD: 1 RED commit + 1 GREEN commit)
- **Files modified:** 2

## Accomplishments

- FTS5 virtual table `messages_fts` created in `migrate()` with INSERT/DELETE/UPDATE triggers keeping index in sync
- `SearchConversations` returns up to 20 ranked results with HTML snippets via FTS5 `snippet()` function
- `SanitizeFTSQuery` prevents FTS5 syntax errors from raw user input by phrase-quoting and escaping
- 7 new tests covering migration, initial populate, search, no-match, sanitizer edge cases, and special chars end-to-end

## Task Commits

Each task was committed atomically:

1. **Task 1: Write failing tests for FTS5 migration and search** - `a84692a` (test)
2. **Task 2: Implement FTS5 migration and SearchConversations** - `abae7a4` (feat)

## Files Created/Modified

- `src/backend/internal/svc/search_test.go` - 7 tests: TestFTSMigration, TestFTSMigration_ExistingRows, TestSearchConversations, TestSearchConversations_NoMatch, TestSearchSanitize, TestSearchSanitize_Empty, TestSearchConversations_SpecialChars
- `src/backend/internal/svc/sqlitestore.go` - Added SearchResult struct, FTS5 migrate block, SanitizeFTSQuery, SearchConversations method

## Decisions Made

- **FTS5 'rebuild' for initial populate:** The plan specified `WHERE id NOT IN (SELECT rowid FROM messages_fts)` as the idempotency guard, but FTS5 content tables mirror the underlying table's rowids even before explicit indexing — so the guard always returns 0 rows. Switched to `INSERT INTO messages_fts(messages_fts) VALUES('rebuild')` which is the correct idempotent approach for content tables.
- **SanitizeFTSQuery exported:** Plan specified unexported `sanitizeFTSQuery`, but tests are in `package svc_test` (black-box), requiring it to be exported as `SanitizeFTSQuery`. This does not change the API contract — logic layer will use it directly.
- **fts5_version() not available:** modernc.org/sqlite doesn't expose the `fts5_version()` scalar function. TestFTSMigration was updated to verify FTS5 functionality via a `MATCH` query instead.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] FTS5 initial populate using wrong idempotency guard**
- **Found during:** Task 2 (Implement FTS5 migration)
- **Issue:** Plan's `WHERE id NOT IN (SELECT rowid FROM messages_fts)` guard doesn't work for FTS5 content tables because the virtual table reflects the source table's rowids via the content= option even when no rows are indexed
- **Fix:** Replaced with `INSERT INTO messages_fts(messages_fts) VALUES('rebuild')` — the correct FTS5 command for rebuilding content table indexes
- **Files modified:** src/backend/internal/svc/sqlitestore.go
- **Verification:** TestFTSMigration_ExistingRows passes (pre-existing rows found via search after migrate)
- **Committed in:** abae7a4 (Task 2 commit)

**2. [Rule 1 - Bug] TestFTSMigration used fts5_version() not available in modernc.org/sqlite**
- **Found during:** Task 2 (TDD GREEN verification)
- **Issue:** `fts5_version()` scalar function doesn't exist in modernc.org/sqlite, causing test failure
- **Fix:** Replaced with a `MATCH` query to verify FTS5 is functional; kept `sqlite_master` check for table existence
- **Files modified:** src/backend/internal/svc/search_test.go
- **Verification:** TestFTSMigration passes; FTS5 functionality confirmed by other passing tests
- **Committed in:** abae7a4 (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (2 Rule 1 bugs in plan's assumptions about modernc.org/sqlite behavior)
**Impact on plan:** Both fixes necessary for correctness. No scope creep.

## Issues Encountered

None beyond the two auto-fixed deviations above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `SearchConversations` and `SanitizeFTSQuery` ready for use by the search endpoint (Plan 03)
- FTS index automatically maintained by triggers on all INSERT/DELETE/UPDATE to messages table
- `SearchResult` struct in svc package; logic layer will convert to types.SearchResult for HTTP responses

---
*Phase: 05-configuration-and-search*
*Completed: 2026-03-11*
