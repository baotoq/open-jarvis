---
phase: 8
slug: add-tilt
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-12
---

# Phase 8 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go standard testing + testify (existing) |
| **Config file** | `src/backend/.golangci.yml` (linting) — no new test config |
| **Quick run command** | `cd src/backend && go test ./... -count=1` |
| **Full suite command** | `cd src/backend && go test ./... -cover -count=1` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `cd src/backend && go test ./... -count=1`
- **After every plan wave:** Run `cd src/backend && go test ./... -cover -count=1`
- **Before `/gsd:verify-work`:** Full suite must be green + manual smoke test
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 8-01-01 | 01 | 1 | DEV-01 | smoke/manual | `which air` | N/A | ⬜ pending |
| 8-01-02 | 01 | 1 | DEV-01/02 | smoke/manual | `tilt up` (manual) | N/A | ⬜ pending |
| 8-01-03 | 01 | 1 | DEV-03 | smoke/manual | manual observation | N/A | ⬜ pending |

---

## Wave 0 Requirements

None — Existing infrastructure covers all phase requirements. This phase adds configuration files only (Tiltfile, .air.toml). No new test files required.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| `tilt up` starts both services | DEV-01, DEV-02 | Requires running process orchestrator | Run `tilt up`, verify both resources green in Tilt dashboard |
| Go file change triggers rebuild | DEV-03 | Requires live observation | Edit any `.go` file, verify air detects change and rebuilds |
| Backend reachable | DEV-04 | Requires running server | `curl -sf http://localhost:8888` returns response |
| Frontend reachable | DEV-05 | Requires running server | `curl -sf http://localhost:3000` returns response |
| `tilt down` preserves SQLite data | DEV-06 | Requires running + stopping | `ls src/backend/data/*.db` after `tilt down` |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
