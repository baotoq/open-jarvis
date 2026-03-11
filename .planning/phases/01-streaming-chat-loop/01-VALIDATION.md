---
phase: 1
slug: streaming-chat-loop
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-11
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go built-in `testing` package (backend); no frontend test framework in Phase 1 |
| **Config file** | None yet — Wave 0 creates test files |
| **Quick run command** | `go test ./src/...` |
| **Full suite command** | `go test -race ./src/...` |
| **Estimated runtime** | ~10 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./src/...`
- **After every plan wave:** Run `go test -race ./src/...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** ~10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 1-01-01 | 01 | 0 | CHAT-01 | unit | `go test ./src/internal/logic/... -run TestStreamChat -v` | ❌ W0 | ⬜ pending |
| 1-01-02 | 01 | 0 | CHAT-01 | unit | `go test ./src/internal/handler/... -run TestChatStreamHandler -v` | ❌ W0 | ⬜ pending |
| 1-01-03 | 01 | 0 | CHAT-02 | unit | `go test ./src/internal/svc/... -run TestConvStore -v` | ❌ W0 | ⬜ pending |
| 1-01-04 | 01 | 0 | CHAT-02 | unit | `go test -race ./src/internal/svc/... -run TestConvStoreConcurrent` | ❌ W0 | ⬜ pending |
| 1-01-05 | 01 | 0 | SAFE-03 | unit | `go test ./src/internal/logic/... -run TestStreamChatTimeout -v` | ❌ W0 | ⬜ pending |
| 1-01-06 | 01 | 0 | SAFE-03 | unit | `go test ./src/internal/config/... -run TestConfigDefaults -v` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `src/internal/logic/chatlogic_test.go` — stubs for CHAT-01, SAFE-03
- [ ] `src/internal/handler/chathandler_test.go` — covers CHAT-01 headers
- [ ] `src/internal/svc/convstore_test.go` — covers CHAT-02
- [ ] `src/internal/config/config_test.go` — covers SAFE-03 defaults

*Wave 0 must create all test stubs before implementation begins.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Tokens appear token-by-token in browser | CHAT-01 | Requires live browser observation of SSE stream rendering | Start Go server, open browser, send message, observe streaming response |
| Markdown renders correctly (headings, code blocks, lists) | CHAT-01 | Visual rendering quality is subjective | Ask model for a response with markdown; verify headings, code blocks, bullets render |
| Enter sends message, Shift+Enter adds newline | CHAT-01 | Keyboard interaction requires browser | Type in chat input, test both key combinations |
| Session resets on browser refresh | CHAT-02 | Requires browser navigation | Chat, refresh, verify new conversation starts |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
