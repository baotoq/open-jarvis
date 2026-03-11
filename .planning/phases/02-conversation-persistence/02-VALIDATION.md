---
phase: 2
slug: conversation-persistence
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-11
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go built-in `testing` + `github.com/stretchr/testify` v1.11.1 |
| **Config file** | none — standard `go test` discovery |
| **Quick run command** | `cd src && go test ./internal/svc/... ./internal/logic/... -count=1` |
| **Full suite command** | `cd src && go test ./... -count=1` |
| **Estimated runtime** | ~10 seconds |

---

## Sampling Rate

- **After every task commit:** Run `cd src && go test ./internal/svc/... ./internal/logic/... -count=1`
- **After every plan wave:** Run `cd src && go test ./... -count=1`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** ~10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 2-01-01 | 01 | 0 | CHAT-03 | unit | `cd src && go test ./internal/svc/... -run TestSQLite -v` | ❌ W0 | ⬜ pending |
| 2-01-02 | 01 | 0 | CHAT-03 | unit | `cd src && go test ./internal/svc/... -run TestSQLiteList -v` | ❌ W0 | ⬜ pending |
| 2-01-03 | 01 | 0 | CHAT-03 | unit | `cd src && go test ./internal/svc/... -run TestSQLiteDelete -v` | ❌ W0 | ⬜ pending |
| 2-01-04 | 01 | 0 | CHAT-03 | compile | `cd src && go build ./...` | ❌ W0 | ⬜ pending |
| 2-01-05 | 01 | 0 | CHAT-03 | unit | `cd src && go test ./internal/logic/... -run TestStreamChatNewSession -v` | ❌ W0 | ⬜ pending |
| 2-01-06 | 01 | 1 | CHAT-03 | unit | `cd src && go test ./internal/svc/... -run TestConvStore -v` | ✅ | ⬜ pending |
| 2-02-01 | 02 | 1 | UI-02 | integration | `cd src && go test ./internal/handler/... -run TestListConversations -v` | ❌ W0 | ⬜ pending |
| 2-02-02 | 02 | 1 | UI-02 | integration | `cd src && go test ./internal/handler/... -run TestDeleteConversation -v` | ❌ W0 | ⬜ pending |
| 2-02-03 | 02 | 2 | UI-02 | manual | Browser inspection | manual-only | ⬜ pending |
| 2-02-04 | 02 | 2 | UI-02 | manual | Browser tab close + reopen | manual-only | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `src/internal/svc/sqlitestore_test.go` — stubs for CHAT-03 SQLite store behavior (Get/Set/List/Delete)
- [ ] `src/internal/logic/chatlogic_test.go` — add `TestStreamChatNewSession` for UUID assignment
- [ ] `src/internal/handler/listconvshandler_test.go` — stubs for UI-02 list endpoint
- [ ] `src/internal/handler/deleteconvhandler_test.go` — stubs for UI-02 delete endpoint

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Frontend sidebar renders conversation entries with title + relative date | UI-02 | Frontend rendering requires browser | Open app, send a message, verify sidebar shows title + relative date |
| localStorage session persists across browser reload | UI-02 | Requires browser session management | Close tab, reopen app, verify same conversation loads |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
