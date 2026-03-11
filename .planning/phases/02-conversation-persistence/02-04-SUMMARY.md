---
phase: 02-conversation-persistence
plan: "04"
subsystem: ui
tags: [next.js, react, typescript, sse, localStorage, tailwind-v4, shadcn-ui, date-fns]

# Dependency graph
requires:
  - phase: 02-conversation-persistence
    plan: "02"
    provides: Next.js 15 scaffold, Tailwind v4, shadcn/ui, cn() utility, API_BASE constant
  - phase: 02-conversation-persistence
    plan: "03"
    provides: Conversation API endpoints (list/get/delete/messages), SSE done event with sessionId
provides:
  - useSession hook — localStorage session ID management with stale-ID validation
  - Sidebar component — conversation list with title, relative date, delete-on-hover, re-fetch on activeId change
  - ChatArea component — SSE streaming client, message history loading on session switch, auto-scroll
  - page.tsx two-panel layout composing Sidebar and ChatArea
  - api.ts typed fetch wrappers for all backend endpoints
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns:
    - useSession hook pattern for localStorage with backend validation (stale-ID handled gracefully)
    - SSE streaming with ReadableStream reader and TextDecoder for token-by-token display
    - effectiveConvId derived from activeConvId ?? sessionId to sync localStorage on initial render without extra effect

key-files:
  created:
    - frontend/hooks/useSession.ts
    - frontend/components/Sidebar.tsx
    - frontend/components/ChatArea.tsx
  modified:
    - frontend/lib/api.ts
    - frontend/app/page.tsx
    - frontend/app/layout.tsx

key-decisions:
  - "effectiveConvId = activeConvId ?? sessionId avoids extra useEffect for syncing localStorage session on mount"
  - "Sidebar re-fetches conversation list when activeId changes to surface newly created conversations after first message"
  - "SSE parsing tries JSON.parse(data) — if it succeeds and has done:true, extract sessionId; otherwise treat as text token"

patterns-established:
  - "SSE client pattern: ReadableStream reader + TextDecoder, split on newlines, extract 'data: ' prefix, parse JSON for done event"
  - "useSession: read localStorage on mount, validate against GET /api/conversations/:id, clear stale IDs silently"

requirements-completed: [UI-02, CHAT-03]

# Metrics
duration: 2min
completed: 2026-03-11
---

# Phase 02 Plan 04: Frontend Chat UI Summary

**Two-panel Next.js UI with SSE streaming chat, localStorage session persistence, sidebar conversation browser, and message history loading**

## Performance

- **Duration:** ~2 min
- **Started:** 2026-03-11T13:52:21Z
- **Completed:** 2026-03-11T13:53:55Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- Extended `api.ts` with typed fetch wrappers for all four conversation endpoints plus `Conversation` and `Message` interfaces
- Created `useSession` hook that reads `jarvis-session-id` from localStorage, validates against backend, and silently clears stale IDs
- Created `Sidebar` component listing conversations with title + relative date (date-fns), delete button on hover, re-fetch on session change
- Created `ChatArea` component with SSE streaming (ReadableStream reader), history loading on session switch, auto-scroll to bottom
- Updated `page.tsx` to compose two-panel layout with `effectiveConvId` sync pattern (avoids extra useEffect)

## Task Commits

1. **Task 1: API layer and useSession hook** - `f941698` (feat)
2. **Task 2: Sidebar, ChatArea, and page layout** - `cf976fe` (feat)

## Files Created/Modified

- `frontend/lib/api.ts` — Added Conversation/Message interfaces and listConversations, getConversation, getConversationMessages, deleteConversation typed fetch wrappers
- `frontend/hooks/useSession.ts` — localStorage session ID management with backend validation and stale-ID clearing
- `frontend/components/Sidebar.tsx` — Conversation list with title, relative date, delete-on-hover, loading/error states
- `frontend/components/ChatArea.tsx` — SSE streaming chat, history loading on sessionId change, auto-scroll
- `frontend/app/page.tsx` — Two-panel flex layout composing Sidebar and ChatArea via useSession
- `frontend/app/layout.tsx` — Added bg-background text-foreground to body classes

## Decisions Made

1. **effectiveConvId pattern** — `const effectiveConvId = activeConvId ?? sessionId` in page.tsx avoids a separate `useEffect` to sync the localStorage-resolved sessionId into activeConvId state. On initial render, before the user clicks any conversation, the ChatArea and Sidebar both receive the localStorage-persisted session automatically.

2. **Sidebar re-fetches on activeId change** — The Sidebar's `useEffect` has `[activeId]` in its dependency array so it re-fetches the conversation list whenever the active session changes. This surfaces newly created conversations (from first-message SSE done event) without requiring a manual refresh.

3. **SSE parsing strategy** — For each `data: ` line, attempt `JSON.parse(data)`. If it parses and has `done: true`, extract `sessionId` from it. If JSON.parse throws (plain text token), append the raw data to the assistant message. This matches the backend's pattern of sending plain tokens then a final `{"done":true,"sessionId":"..."}` event.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Full conversation UI is live: sidebar listing, session persistence, message history loading, SSE streaming
- Phase 2 (conversation persistence) is complete end-to-end (backend + frontend)
- Phase 3 (local tools) can proceed

---
*Phase: 02-conversation-persistence*
*Completed: 2026-03-11*

## Self-Check: PASSED

- frontend/lib/api.ts — FOUND
- frontend/hooks/useSession.ts — FOUND
- frontend/components/Sidebar.tsx — FOUND
- frontend/components/ChatArea.tsx — FOUND
- frontend/app/page.tsx — FOUND
- Commit f941698 — FOUND
- Commit cf976fe — FOUND
