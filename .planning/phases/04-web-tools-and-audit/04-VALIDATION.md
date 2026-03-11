---
phase: 4
slug: web-tools-and-audit
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-11
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — standard Go test files |
| **Quick run command** | `cd src/backend && go test ./internal/...` |
| **Full suite command** | `cd src/backend && go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `cd src/backend && go test ./internal/...`
- **After every plan wave:** Run `cd src/backend && go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 04-01-01 | 01 | 1 | TOOL-04 | unit | `cd src/backend && go test ./internal/tools/...` | ❌ W0 | ⬜ pending |
| 04-01-02 | 01 | 1 | TOOL-03 | unit | `cd src/backend && go test ./internal/tools/...` | ❌ W0 | ⬜ pending |
| 04-02-01 | 02 | 1 | SAFE-04 | unit | `cd src/backend && go test ./internal/svc/...` | ❌ W0 | ⬜ pending |
| 04-02-02 | 02 | 2 | SAFE-04 | integration | `cd src/backend && go test ./internal/logic/...` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `src/backend/internal/tools/web_search_test.go` — stubs for TOOL-04
- [ ] `src/backend/internal/tools/web_fetch_test.go` — stubs for TOOL-03
- [ ] `src/backend/internal/svc/audit_store_test.go` — stubs for SAFE-04

*Existing `go test` infrastructure covers framework needs.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Web search returns results from Brave API | TOOL-04 | Requires live API key | Configure BraveSearchAPIKey, ask agent to search, verify results returned |
| Web fetch summarizes real page content | TOOL-03 | Requires live HTTP | Ask agent to fetch a URL, verify readable content returned |
| Audit log viewable in UI | SAFE-04 | UI inspection required | Execute any tool, check audit log entries appear with correct fields |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
