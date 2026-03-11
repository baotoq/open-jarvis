---
phase: 03-file-and-shell-tools
plan: "02"
subsystem: api
tags: [go, go-zero, toolexec, config, servicecontext, approvalstore, sync]

# Dependency graph
requires:
  - phase: 03-file-and-shell-tools
    plan: "01"
    provides: "toolexec package with Executor interface, ToolRegistry, FileTool, ShellTool"
provides:
  - "Config extended with ShellAllowlist, ShellDenylist, WorkspaceRoot fields"
  - "Thread-safe ApprovalStore with Register, Resolve, Delete (sync.Mutex)"
  - "ServiceContext wired with Executor, ApprovalStore, ShellTool fields"
  - "NewServiceContextForTest constructor for test injection"
affects: [03-03-chat-logic-tool-integration, 03-04, 03-05]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "ApprovalStore: pending map[string]chan bool with sync.Mutex, unlock before channel send to avoid deadlock"
    - "ServiceContext wires tool registry in NewServiceContext; NewServiceContextForTest mirrors it for tests"

key-files:
  created:
    - src/backend/internal/svc/approvalstore.go
    - src/backend/internal/svc/approvalstore_test.go
  modified:
    - src/backend/internal/config/config.go
    - src/backend/internal/config/config_test.go
    - src/backend/internal/svc/servicecontext.go

key-decisions:
  - "ApprovalStore.Resolve unlocks mutex before sending to channel to avoid deadlock when channel is unbuffered"
  - "ShellTool stored as *toolexec.ShellTool field (not behind interface) so ChatLogic can call RequiresApproval directly"
  - "NewServiceContextForTest added alongside existing NewServiceContextWithClient for backward compatibility"

patterns-established:
  - "Approval gate pattern: Register(id, ch) → emit SSE → block on ch → Delete(id) via defer"

requirements-completed: [SAFE-01, SAFE-02]

# Metrics
duration: 3min
completed: 2026-03-11
---

# Phase 03 Plan 02: Config and ServiceContext Tool Infrastructure Summary

**Config extended with shell safety lists and workspace root; thread-safe ApprovalStore added; ServiceContext wired with Executor, ApprovalStore, and ShellTool**

## Performance

- **Duration:** ~3 min
- **Started:** 2026-03-11T14:31:52Z
- **Completed:** 2026-03-11T14:34:00Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Extended `Config` struct with `ShellAllowlist []string`, `ShellDenylist []string`, `WorkspaceRoot string` fields using go-zero struct tag defaults
- Created thread-safe `ApprovalStore` in `svc` package with `Register`, `Resolve`, `Delete` methods (sync.Mutex, unlock before channel send to prevent deadlock)
- Wired `Executor`, `ApprovalStore`, and `ShellTool` into `ServiceContext`; `NewServiceContext` builds a `ToolRegistry` registering `read_file`, `write_file`, `shell_run`
- Added `NewServiceContextForTest` for test-time injection while all existing tests remain passing

## Task Commits

Each task was committed atomically:

1. **Task 1: Extend Config with tool safety fields and add ApprovalStore** - `25d1b31` (feat)
2. **Task 2: Wire Executor and ApprovalStore into ServiceContext** - `185c7d2` (feat)

**Plan metadata:** (see below — docs commit)

## Files Created/Modified

- `src/backend/internal/config/config.go` - Added ShellAllowlist, ShellDenylist, WorkspaceRoot fields
- `src/backend/internal/config/config_test.go` - Added TestConfig_Defaults asserting WorkspaceRoot="." and nil lists
- `src/backend/internal/svc/approvalstore.go` - New: thread-safe ApprovalStore
- `src/backend/internal/svc/approvalstore_test.go` - New: 4 tests including concurrent goroutine test
- `src/backend/internal/svc/servicecontext.go` - Added Executor, ApprovalStore, ShellTool fields; wired in NewServiceContext; added NewServiceContextForTest

## Decisions Made

- `ApprovalStore.Resolve` unlocks the mutex before sending to the channel. If the lock were held during the send, a goroutine calling `Delete` while `Resolve` is blocked on an unbuffered channel would deadlock.
- `ShellTool` is stored as `*toolexec.ShellTool` (concrete type) on `ServiceContext` so `ChatLogic` can call `RequiresApproval` without requiring an extended interface.
- `NewServiceContextForTest` added as a separate constructor rather than modifying `NewServiceContextWithClient` to preserve backward compatibility with existing tests.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- ServiceContext has all tool infrastructure Plan 03 (ChatLogic tool integration) needs
- `svcCtx.Executor`, `svcCtx.ApprovalStore`, and `svcCtx.ShellTool` are ready for ChatLogic to use
- No blockers

---
*Phase: 03-file-and-shell-tools*
*Completed: 2026-03-11*
