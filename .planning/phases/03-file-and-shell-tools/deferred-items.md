# Deferred Items — Phase 03

## Pre-existing Lint Errors (Out of Scope)

Discovered during 03-04 execution. These errors exist in files not modified by this plan.

### react-hooks/set-state-in-effect

**File:** `src/frontend/components/Sidebar.tsx` line 58
`fetchConversations()` called synchronously inside `useEffect`

**File:** `src/frontend/hooks/useSession.ts` line 14
`setIsLoading(false)` called synchronously inside `useEffect`

Both are false positives from the lint rule — the functions called are async and initiate async operations, they don't directly call setState synchronously. However the lint rule triggers because `fetchConversations` and the effect body contain setState calls.

**Action needed:** Fix or suppress (eslint-disable comment) in a future cleanup plan.
