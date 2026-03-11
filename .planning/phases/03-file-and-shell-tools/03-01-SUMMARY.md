---
phase: 03-file-and-shell-tools
plan: 01
subsystem: api
tags: [go, toolexec, file-tool, shell-tool, executor, tdd]

# Dependency graph
requires:
  - phase: 02-conversation-persistence
    provides: ServiceContext and ConversationStore patterns that toolexec builds upon

provides:
  - toolexec package with Executor interface and ToolRegistry
  - FileTool with read_file, write_file, and path traversal protection (safePath)
  - ShellTool with Run (exec.CommandContext) and RequiresApproval (glob allowlist/denylist)
  - Complete table-driven unit tests for all tool behaviors

affects:
  - 03-02-PLAN.md (config extension: ShellAllowlist, ShellDenylist, WorkspaceRoot)
  - 03-03-PLAN.md (ChatLogic tool-call integration consumes Executor interface)
  - 03-05-PLAN.md (approval flow uses RequiresApproval)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Executor interface with ToolRegistry dispatcher (dispatch by name string)
    - safePath guard using filepath.Abs + strings.HasPrefix with trailing separator
    - ShellTool allowlist/denylist with filepath.Match glob patterns, denylist takes priority

key-files:
  created:
    - src/backend/internal/toolexec/executor.go
    - src/backend/internal/toolexec/filetool.go
    - src/backend/internal/toolexec/shelltool.go
    - src/backend/internal/toolexec/executor_test.go
  modified: []

key-decisions:
  - "ToolRegistry dispatches by name string map; unknown tool returns ToolResult{Error: 'unknown tool: <name>'} rather than panicking"
  - "safePath adds trailing separator to root before prefix check to prevent /tmp matching /tmpother"
  - "ShellTool.Run returns partial output in Content even on command failure, so callers can surface stderr"
  - "RequiresApproval uses filepath.Match (not regexp); denylist evaluated before allowlist; default is true (requires approval)"

patterns-established:
  - "Executor pattern: interface + registry separate from tool implementations, enabling mock injection in tests"
  - "Tool functions always accept (context.Context, argsJSON string) ToolResult — uniform dispatch signature"
  - "Path traversal guard: filepath.Abs(filepath.Join(root, userPath)) + strings.HasPrefix check with trailing sep"

requirements-completed:
  - TOOL-01
  - TOOL-02
  - SAFE-01

# Metrics
duration: 2min
completed: 2026-03-11
---

# Phase 3 Plan 01: toolexec Package Summary

**Go toolexec package with Executor interface, ToolRegistry dispatcher, FileTool (read/write with path-traversal guard), and ShellTool (exec.CommandContext with allowlist/denylist approval logic)**

## Performance

- **Duration:** ~2 min
- **Started:** 2026-03-11T14:31:28Z
- **Completed:** 2026-03-11T14:33:12Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- `Executor` interface and `ToolRegistry` with name-based dispatch and unknown-tool error handling
- `FileTool` with `ReadFile` and `WriteFile` confined to workspace root via `safePath` traversal guard
- `ShellTool` with `Run` using `exec.CommandContext` (context-cancellable) and `RequiresApproval` glob matching
- All 12 test cases pass across `TestToolRegistry`, `TestFileTool`, and `TestShellTool` suites; `go vet` clean

## Task Commits

Each task was committed atomically:

1. **Task 1: Define Executor interface and ToolRegistry** - `f7a466f` (feat)
2. **Task 2: Implement FileTool and ShellTool with tests** - `a97be1e` (feat)

_Note: TDD tasks combined test writing and implementation into single commits per task (RED included in Task 1 commit, GREEN in Task 2 commit)._

## Files Created/Modified

- `src/backend/internal/toolexec/executor.go` - `ToolResult`, `Executor` interface, `ToolRegistry`, `NewRegistry`
- `src/backend/internal/toolexec/filetool.go` - `FileTool` with `ReadFile`, `WriteFile`, `safePath` traversal guard
- `src/backend/internal/toolexec/shelltool.go` - `ShellTool` with `Run` and `RequiresApproval`
- `src/backend/internal/toolexec/executor_test.go` - Table-driven tests for all tools (12 test cases)

## Decisions Made

- `ToolRegistry` dispatches by name via `map[string]func(...)` — unknown name returns error, no panic
- `safePath` appends trailing separator to root before `strings.HasPrefix` to prevent `/tmp` matching `/tmpother`
- `ShellTool.Run` includes partial stdout/stderr in `ToolResult.Content` even on failure, enabling callers to surface command output
- `RequiresApproval` uses `filepath.Match` glob patterns; denylist checked first (overrides allowlist); default `true` (requires approval)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `toolexec` package fully implements the contracts expected by plans 03-02 through 03-05
- Plan 03-02 can extend `Config` to add `ShellAllowlist`, `ShellDenylist`, `WorkspaceRoot` and wire `NewFileTool`/`NewShellTool` into `ServiceContext`
- Plan 03-03 can import `toolexec.Executor` directly into `ChatLogic` for tool-call dispatch

---
*Phase: 03-file-and-shell-tools*
*Completed: 2026-03-11*
