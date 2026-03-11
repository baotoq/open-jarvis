---
phase: 03-file-and-shell-tools
plan: 03
subsystem: api
tags: [go, go-zero, openai, sse, agentic-loop, tool-dispatch, approval-gate]

# Dependency graph
requires:
  - phase: 03-file-and-shell-tools plan 01
    provides: toolexec.Executor, ToolRegistry, FileTool, ShellTool with RequiresApproval
  - phase: 03-file-and-shell-tools plan 02
    provides: ServiceContext with Executor/ApprovalStore/ShellTool, NewServiceContextForTest

provides:
  - Agentic loop in ChatLogic.StreamChat with tool dispatch and approval gate
  - chatTools package-level var (read_file, write_file, shell_run OpenAI tool definitions)
  - tool_call and tool_result SSE events emitted during tool dispatch
  - approval_request SSE event blocking on shell commands requiring approval
  - POST /api/chat/approve endpoint via ApproveHandler
  - ApproveRequest type (ApprovalID, Approved)

affects:
  - frontend — needs to handle tool_call, tool_result, approval_request SSE events
  - phase-04-web-search-tools — same agentic loop pattern, add more tools to chatTools

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Agentic loop: bounded for-loop, accumulate tool call fragments by Index, dispatch after FinishReasonToolCalls
    - Tool call fragment accumulation: map[int]*openai.ToolCall keyed by delta.ToolCalls[i].Index
    - Approval gate: register channel, emit SSE, select on ch and ctx.Done()
    - TDD: RED (failing tests) committed before GREEN (implementation)

key-files:
  created:
    - src/backend/internal/handler/approvehandler.go
    - src/backend/internal/handler/approvehandler_test.go
  modified:
    - src/backend/internal/logic/chatlogic.go
    - src/backend/internal/logic/chatlogic_test.go
    - src/backend/internal/types/types.go
    - src/backend/cmd/main.go
    - src/backend/etc/config.yaml

key-decisions:
  - "waitForApproval extracted as helper method on ChatLogic to keep StreamChat readable"
  - "Tool result error (denial/execution failure) surfaced in content field of tool result message, not as Go error — keeps loop running"
  - "lastAssistantContent captures only final text response; intermediate tool-call assistant messages are stored with ToolCalls field populated"
  - "maxIter defaults to 10 when Config.MaxToolCalls <= 0 to prevent infinite loops on zero config"

patterns-established:
  - "Tool call accumulation: initialize entry on first fragment (non-nil Index), append Arguments on subsequent fragments"
  - "Always append assistant message with ToolCalls BEFORE tool result messages in history"
  - "Approval gate: buffered channel (size 1) prevents deadlock when Resolve is called before select"

requirements-completed: [TOOL-01, TOOL-02, SAFE-02]

# Metrics
duration: 6min
completed: 2026-03-11
---

# Phase 3 Plan 03: Agentic Loop with Tool Dispatch and Approval Gate Summary

**Single-shot StreamChat transformed into a bounded agentic loop dispatching read_file/write_file/shell_run tools via OpenAI function calling, with shell approval gate blocking on user decision via SSE**

## Performance

- **Duration:** ~6 min
- **Started:** 2026-03-11T14:36:27Z
- **Completed:** 2026-03-11T14:42:27Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments
- StreamChat rewrites into a MaxToolCalls-bounded agentic loop with FinishReasonToolCalls detection
- Tool call fragments accumulated by Index into full ToolCall objects, dispatched via Executor
- Shell approval gate: registers channel, emits approval_request SSE, blocks until user resolves or ctx cancels
- POST /api/chat/approve endpoint wired via ApproveHandler calling ApprovalStore.Resolve
- 3 new TDD tests: TestChatLogic_ToolLoop, TestChatLogic_ApprovalGate, TestChatLogic_ApprovalDenied
- All 9 chatlogic tests and all handler tests pass

## Task Commits

Each task was committed atomically:

1. **Task 1: Rewrite StreamChat as agentic loop with tool support** - `3dddeff` (feat + test TDD)
2. **Task 2: Add approve handler, types, and wire route** - `8c459e0` (feat)

**Plan metadata:** (docs commit below)

_Note: Task 1 used TDD — tests written first to confirm RED, then implementation for GREEN_

## Files Created/Modified
- `src/backend/internal/logic/chatlogic.go` - Agentic loop with tool dispatch, approval gate, chatTools var
- `src/backend/internal/logic/chatlogic_test.go` - Added ToolLoop, ApprovalGate, ApprovalDenied tests; new mock types
- `src/backend/internal/handler/approvehandler.go` - POST /api/chat/approve handler
- `src/backend/internal/handler/approvehandler_test.go` - Approved and UnknownID tests
- `src/backend/internal/types/types.go` - Added ApproveRequest type
- `src/backend/cmd/main.go` - Registered /api/chat/approve route
- `src/backend/etc/config.yaml` - Documented WorkspaceRoot, ShellAllowlist, ShellDenylist config fields

## Decisions Made
- `waitForApproval` extracted as a helper method to keep the agentic loop body readable
- Tool denial and execution errors are surfaced in the tool result message `content` field, not returned as Go errors — this keeps the loop running so the model can respond to the failure
- `lastAssistantContent` captures only the final text response; intermediate assistant messages with ToolCalls are appended to history inline during the loop
- maxIter defaults to 10 when `Config.MaxToolCalls <= 0` as a safety net

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Full tool-capable agentic backend is complete through Phase 3
- Frontend needs to be updated to handle the new SSE event types: `tool_call`, `tool_result`, `approval_request`
- Phase 4 (web search tools) can extend `chatTools` var and register new tools in `NewServiceContext`

---
*Phase: 03-file-and-shell-tools*
*Completed: 2026-03-11*

## Self-Check: PASSED

Files verified:
- FOUND: src/backend/internal/logic/chatlogic.go
- FOUND: src/backend/internal/logic/chatlogic_test.go
- FOUND: src/backend/internal/handler/approvehandler.go
- FOUND: src/backend/internal/handler/approvehandler_test.go
- FOUND: src/backend/internal/types/types.go
- FOUND: src/backend/cmd/main.go

Commits verified:
- FOUND: 3dddeff
- FOUND: 8c459e0
