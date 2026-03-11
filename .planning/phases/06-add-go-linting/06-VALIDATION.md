---
phase: 6
slug: add-go-linting
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-12
---

# Phase 6 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) + testify v1.11.1 |
| **Config file** | none (runs via `go test ./...`) |
| **Quick run command** | `cd src/backend && go test ./... -count=1` |
| **Full suite command** | `cd src/backend && go test -cover ./...` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `cd src/backend && go test ./... -count=1`
- **After every plan wave:** Run `cd src/backend && golangci-lint run ./... && go test -cover ./...`
- **Before `/gsd:verify-work`:** `golangci-lint run ./...` must exit 0
- **Max feedback latency:** ~30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 6-01-01 | 01 | 1 | lint-config | smoke | `cd src/backend && golangci-lint config verify` | ✅ | ⬜ pending |
| 6-01-02 | 01 | 1 | lint-config | regression | `cd src/backend && go test ./...` | ✅ | ⬜ pending |
| 6-02-01 | 02 | 2 | errcheck-fixes | regression | `cd src/backend && go test ./...` | ✅ | ⬜ pending |
| 6-02-02 | 02 | 2 | errcheck-fixes | smoke | `cd src/backend && golangci-lint run ./...` | ✅ | ⬜ pending |
| 6-03-01 | 03 | 2 | revive-fixes | regression | `cd src/backend && go test ./...` | ✅ | ⬜ pending |
| 6-03-02 | 03 | 2 | revive-fixes | smoke | `cd src/backend && golangci-lint run ./...` | ✅ | ⬜ pending |
| 6-04-01 | 04 | 3 | final-clean | smoke | `cd src/backend && golangci-lint run ./... && go test -cover ./...` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements. No new test stubs are needed — the linter itself is the primary test artifact, and the existing Go test suite provides regression coverage.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| golangci-lint installation succeeds | install | Binary install from script; no automated verify | Run `golangci-lint --version` and confirm output shows `built with go1.26.x` |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
