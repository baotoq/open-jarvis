---
phase: 03-file-and-shell-tools
plan: 04
subsystem: ui
tags: [react, typescript, nextjs, shadcn, sse, tool-calls, approval-dialog]

# Dependency graph
requires:
  - phase: 03-file-and-shell-tools
    provides: SSE event formats for tool_call, tool_result, approval_request from backend
  - phase: 02-conversation-persistence
    provides: ChatArea SSE loop, getConversationMessages, Message type
provides:
  - MessagePart union type (TextPart, ToolCallPart, ToolResultPart, ApprovalRequestPart)
  - ChatMessage interface with parts array
  - submitApproval() POST /api/chat/approve
  - ToolCallBlock collapsible component
  - ApprovalDialog modal component
  - Extended ChatArea SSE parser handling all tool events
affects: [03-05-PLAN.md, frontend UI work]

# Tech tracking
tech-stack:
  added: [shadcn/ui badge, shadcn/ui collapsible, shadcn/ui dialog]
  patterns:
    - MessagePart union type for typed SSE event accumulation
    - ChatMessage replaces flat Message for live streaming state
    - Historical messages loaded as Message[] converted to ChatMessage on mount
    - ToolCallBlock renders paired tool_call + tool_result by matching id

key-files:
  created:
    - src/frontend/components/ToolCallBlock.tsx
    - src/frontend/components/ApprovalDialog.tsx
    - src/frontend/components/ui/badge.tsx
    - src/frontend/components/ui/collapsible.tsx
    - src/frontend/components/ui/dialog.tsx
  modified:
    - src/frontend/lib/api.ts
    - src/frontend/components/ChatArea.tsx

key-decisions:
  - "ChatMessage (parts[]) used for live streaming state; historical Message[] converted to ChatMessage on load via single TextPart wrapping"
  - "ToolResultPart skipped in standalone render loop — rendered via ToolCallBlock by pairing with matching ToolCallPart by id"
  - "Pre-existing lint errors in Sidebar.tsx and useSession.ts deferred — out of scope for this plan"

patterns-established:
  - "MessagePart union: use parsed.type switch to route SSE events into typed parts array"
  - "ToolCallBlock: pairs call and result by id, collapsible shadcn pattern"
  - "ApprovalDialog: controlled open prop, parent manages close via setPendingApproval(null)"

requirements-completed:
  - UI-01

# Metrics
duration: 2min
completed: 2026-03-11
---

# Phase 03 Plan 04: Frontend Tool Call Display and Approval Dialog Summary

**Inline tool call display and shell approval dialog for the chat UI using shadcn Collapsible and Dialog, with ChatArea SSE parser extended to handle tool_call, tool_result, and approval_request events**

## Performance

- **Duration:** ~2 min
- **Started:** 2026-03-11T14:36:24Z
- **Completed:** 2026-03-11T14:38:44Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments
- Added MessagePart union types and ChatMessage interface to api.ts, enabling typed accumulation of tool events during SSE streaming
- Created ToolCallBlock component using shadcn Collapsible showing tool name (badge), args, and result with error/success indicators
- Created ApprovalDialog component using shadcn Dialog displaying shell command with Approve/Deny buttons
- Extended ChatArea SSE parser to handle tool_call, tool_result, approval_request events; replaced flat Message[] state with ChatMessage[] parts accumulation
- Production build passes with zero TypeScript errors

## Task Commits

Each task was committed atomically:

1. **Task 1: Add MessagePart types and submitApproval to api.ts, install shadcn components** - `5575e93` (feat)
2. **Task 2: Build ToolCallBlock and ApprovalDialog components, extend ChatArea** - `fd15731` (feat)

**Plan metadata:** (see final commit below)

## Files Created/Modified
- `src/frontend/lib/api.ts` - Added TextPart, ToolCallPart, ToolResultPart, ApprovalRequestPart, MessagePart, ChatMessage, submitApproval()
- `src/frontend/components/ToolCallBlock.tsx` - Collapsible tool call + result display using shadcn Collapsible and Badge
- `src/frontend/components/ApprovalDialog.tsx` - Modal approval dialog using shadcn Dialog with approve/deny buttons
- `src/frontend/components/ChatArea.tsx` - Extended SSE parser, ChatMessage[] state, renders all part types
- `src/frontend/components/ui/badge.tsx` - shadcn Badge (new)
- `src/frontend/components/ui/collapsible.tsx` - shadcn Collapsible (new)
- `src/frontend/components/ui/dialog.tsx` - shadcn Dialog (new)

## Decisions Made
- ChatMessage (parts[]) used for live streaming state; historical Message[] converted to ChatMessage on load by wrapping in a single TextPart — preserves backward compatibility with existing backend API
- ToolResultPart is skipped as a standalone render element — rendered via ToolCallBlock by pairing the matching ToolCallPart by id
- Pre-existing lint errors in Sidebar.tsx and useSession.ts are deferred — these are in files outside the scope of this plan

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered
- Pre-existing ESLint `react-hooks/set-state-in-effect` errors in Sidebar.tsx and useSession.ts (confirmed by git stash check). Deferred to future cleanup. Logged in `deferred-items.md`.

## User Setup Required
None — no external service configuration required.

## Next Phase Readiness
- Frontend is fully wired to display tool activity and handle shell approval dialogs
- ChatArea handles all backend SSE event types: text tokens, tool_call, tool_result, approval_request, and done
- Ready for Phase 03 Plan 05 or end-to-end integration testing

---
*Phase: 03-file-and-shell-tools*
*Completed: 2026-03-11*
