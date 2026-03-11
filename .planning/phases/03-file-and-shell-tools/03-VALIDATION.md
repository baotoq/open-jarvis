---
phase: 3
slug: file-and-shell-tools
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-11
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test / jest |
| **Config file** | src/backend/go.mod / src/frontend/package.json |
| **Quick run command** | `cd src/backend && go test ./...` |
| **Full suite command** | `cd src/backend && go test ./... && cd ../frontend && npm run lint` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `cd src/backend && go test ./...`
- **After every plan wave:** Run `cd src/backend && go test ./... && cd ../frontend && npm run lint`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 3-01-01 | 01 | 1 | TOOL-01 | unit | `cd src/backend && go test ./internal/logic/...` | ❌ W0 | ⬜ pending |
| 3-01-02 | 01 | 1 | TOOL-01 | unit | `cd src/backend && go test ./internal/logic/...` | ❌ W0 | ⬜ pending |
| 3-01-03 | 01 | 1 | TOOL-02 | unit | `cd src/backend && go test ./internal/logic/...` | ❌ W0 | ⬜ pending |
| 3-02-01 | 02 | 2 | SAFE-01 | unit | `cd src/backend && go test ./internal/svc/...` | ❌ W0 | ⬜ pending |
| 3-02-02 | 02 | 2 | SAFE-02 | unit | `cd src/backend && go test ./internal/svc/...` | ❌ W0 | ⬜ pending |
| 3-03-01 | 03 | 3 | UI-01 | manual | N/A — frontend visual | N/A | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `src/backend/internal/logic/tools_test.go` — stubs for TOOL-01, TOOL-02
- [ ] `src/backend/internal/svc/approval_test.go` — stubs for SAFE-01, SAFE-02

*Existing test infrastructure (go test + jest) covers the framework requirements.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Tool call blocks appear in chat UI | UI-01 | Frontend visual rendering | Start dev server, send a file read request, verify tool call blocks render inline |
| Approval dialog appears for shell commands | SAFE-01 | Frontend interaction flow | Ask agent to run a shell command, verify dialog appears before execution |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
