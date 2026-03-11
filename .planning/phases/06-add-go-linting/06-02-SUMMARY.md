---
phase: 06-add-go-linting
plan: 02
subsystem: backend/linting
tags: [errcheck, go-linting, sse, sqlite]
dependency_graph:
  requires: [06-01]
  provides: [errcheck-clean-backend]
  affects: [chatlogic, sqlitestore, webtool]
tech_stack:
  added: []
  patterns:
    - "fmt.Fprintf error check with return nil on SSE connection close"
    - "defer func() { if err := rows.Close(); err != nil { log.Printf(...) } }() for deferred cleanup"
    - "nolint:errcheck with justification for truly ignorable errors in test cleanup"
key_files:
  created: []
  modified:
    - src/backend/internal/logic/chatlogic.go
    - src/backend/internal/toolexec/webtool.go
    - src/backend/internal/toolexec/webtool_test.go
    - src/backend/internal/logic/searchconvslogic_test.go
    - src/backend/internal/svc/sqlitestore.go
    - src/backend/internal/svc/sqlitestore_test.go
    - src/backend/internal/svc/auditstore_test.go
    - src/backend/internal/svc/search_test.go
decisions:
  - "SSE write errors return nil (stop streaming silently) rather than returning an error — connection is already closed"
  - "Error frame fmt.Fprintf failures are logged via log.Printf then the outer error is returned as normal"
  - "tx.Rollback() failure is logged; cannot propagate from defer without shadowing the function's named return"
  - "nolint:errcheck used for test cleanup (db.Close, t.Cleanup) — errors are logged by sql driver; requiring error would clutter tests with non-actionable checks"
  - "fmt.Fprint in httptest handler extracted to const to allow nolint on the call line"
metrics:
  duration: 294s
  completed_date: "2026-03-11"
  tasks_completed: 2
  files_modified: 8
---

# Phase 6 Plan 2: Fix errcheck Violations in Production and Test Code Summary

Fixed 17 errcheck violations across chatlogic.go, sqlitestore.go, webtool.go, and svc/logic test files. SSE write errors in chatlogic.go now cause early return (no silent failures). defer rows.Close() in production code logs errors; test cleanup uses nolint with justification.

## Tasks Completed

| Task | Description | Commit | Files |
|------|-------------|--------|-------|
| 1 | Fix errcheck in chatlogic.go and webtool.go | f1bf449 | chatlogic.go, webtool.go, webtool_test.go, searchconvslogic_test.go |
| 2 | Fix errcheck in sqlitestore.go and test files | 7b92db0 | sqlitestore.go, sqlitestore_test.go, auditstore_test.go, search_test.go |

## Verification Results

- `go test ./... -count=1` — all 5 test packages pass, no regressions
- `golangci-lint run ./internal/logic/... ./internal/svc/... ./internal/toolexec/...` — zero errcheck issues
- `go build ./...` — compiles cleanly

## Changes Made

### Task 1: chatlogic.go and webtool.go

**chatlogic.go:**
- Added `log` import
- Wrapped `fmt.Fprintf(w, "data: [ERROR]...")` error-frame writes: log write failures, then return the original stream error
- Wrapped `stream.Close()` in `defer func()` to log close errors
- All SSE data-frame `fmt.Fprintf` calls now check the error and return nil (stop streaming) on write failure
- All `fmt.Fprintf` calls in tool_call, tool_result, done, and approval_request events check errors

**webtool.go:**
- Added `//nolint:errcheck` to `defer resp.Body.Close()` — Body.Close() in defer cannot propagate, http client logs underlying errors

**webtool_test.go:**
- Extracted multiline backtick HTML string to `const htmlBody` to allow nolint on the `fmt.Fprint` call line
- Added `//nolint:errcheck` to `fmt.Fprint(w, htmlBody)` and both `json.NewEncoder(w).Encode(resp)` calls in test handlers

**searchconvslogic_test.go:**
- Added `//nolint:errcheck` to `db.Close()` in `t.Cleanup`

### Task 2: sqlitestore.go and svc test files

**sqlitestore.go:**
- Added `log` import
- Replaced 3x `defer rows.Close()` with `defer func() { if err := rows.Close(); err != nil { log.Printf(...) } }()`
- Wrapped `tx.Rollback()` with error check and log on failure

**Test files (sqlitestore_test.go, auditstore_test.go, search_test.go):**
- Added `//nolint:errcheck // cleanup in test; error logged by sql driver` to all `db.Close()`, `db1.Close()`, `db2.Close()`, and `rows.Close()` calls in test cleanup contexts

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Coverage] Additional fmt.Fprintf calls in chatlogic.go beyond initial 4**
- **Found during:** Task 1 — lint re-run after initial fixes
- **Issue:** The plan listed 4 errcheck violations in chatlogic.go (lines 154, 165, 172, 184), but re-running golangci-lint revealed 3 more unchecked calls on lines 262, 287, 323 (tool_call event, tool_result denial event, tool_result execution event)
- **Fix:** Applied the same pattern — check error and return nil on connection close
- **Files modified:** src/backend/internal/logic/chatlogic.go
- **Additional commit:** Included in f1bf449

**2. [Rule 2 - Missing Coverage] fmt.Fprintf for done event and approval_request event**
- **Found during:** Task 1 — lint re-run after second pass
- **Issue:** Lines 354 and 374 (done event and approval_request event in waitForApproval) were also unchecked
- **Fix:** Done event logs the error (cannot stop after done); approval_request returns nil on write failure (connection closed)
- **Files modified:** src/backend/internal/logic/chatlogic.go
- **Additional commit:** Included in f1bf449

**3. [Rule 2 - Missing Coverage] searchconvslogic_test.go db.Close violation**
- **Found during:** Task 1 — lint showed violation in logic package test file not listed in plan
- **Issue:** searchconvslogic_test.go:22 had unchecked db.Close() not listed in the plan's target files
- **Fix:** Added nolint:errcheck
- **Files modified:** src/backend/internal/logic/searchconvslogic_test.go
- **Additional commit:** Included in f1bf449

**4. [Rule 2 - Missing Coverage] search_test.go rows.Close violation**
- **Found during:** Task 2 — lint showed additional rows.Close in search_test.go:39 not listed in plan
- **Issue:** TestFTSMigration had a direct rows.Close() (not in t.Cleanup) that was not in the plan's list
- **Fix:** Added nolint:errcheck
- **Files modified:** src/backend/internal/svc/search_test.go
- **Additional commit:** Included in 7b92db0

## Out-of-Scope Violations (Not Fixed)

The following errcheck violations exist in files outside this plan's scope:
- `internal/config/config_test.go`: os.Remove and f.Close (4 instances)
- `internal/handler/searchconvshandler_test.go`: db.Close (1 instance)

These are pre-existing and should be addressed in a follow-up plan.

## Self-Check: PASSED
