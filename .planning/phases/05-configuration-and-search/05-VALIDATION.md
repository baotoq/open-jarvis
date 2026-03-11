---
phase: 5
slug: configuration-and-search
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-11
---

# Phase 5 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — standard Go test files |
| **Quick run command** | `cd src/backend && go test ./internal/...` |
| **Full suite command** | `cd src/backend && go test ./...` |
| **Estimated runtime** | ~5-10 seconds |

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
| 05-01-01 | 01 | 0 | MEM-01 | unit | `cd src/backend && go test ./internal/svc/ -run TestFTS` | ❌ W0 | ⬜ pending |
| 05-01-02 | 01 | 0 | MEM-01 | unit | `cd src/backend && go test ./internal/svc/ -run TestSearch` | ❌ W0 | ⬜ pending |
| 05-01-03 | 01 | 0 | MEM-01 | unit | `cd src/backend && go test ./internal/svc/ -run TestSearchSanitize` | ❌ W0 | ⬜ pending |
| 05-01-04 | 01 | 0 | MEM-01 | unit | `cd src/backend && go test ./internal/logic/ -run TestSearchConvs` | ❌ W0 | ⬜ pending |
| 05-02-01 | 02 | 0 | CHAT-04 | unit | `cd src/backend && go test ./internal/svc/ -run TestConfigStore` | ❌ W0 | ⬜ pending |
| 05-02-02 | 02 | 0 | CHAT-04 | unit | `cd src/backend && go test ./internal/svc/ -run TestConfigStoreYAML` | ❌ W0 | ⬜ pending |
| 05-02-03 | 02 | 0 | CHAT-04 | unit | `cd src/backend && go test ./internal/handler/ -run TestGetConfig` | ❌ W0 | ⬜ pending |
| 05-02-04 | 02 | 0 | CHAT-04 | unit | `cd src/backend && go test ./internal/handler/ -run TestUpdateConfig` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `src/backend/internal/svc/configstore_test.go` — covers CHAT-04 ConfigStore unit tests
- [ ] `src/backend/internal/svc/search_test.go` — covers MEM-01 FTS5 schema + query tests
- [ ] `src/backend/internal/handler/getconfighandler_test.go` — covers CHAT-04 handler
- [ ] `src/backend/internal/handler/updateconfighandler_test.go` — covers CHAT-04 handler
- [ ] `src/backend/internal/handler/searchconvshandler_test.go` — covers MEM-01 handler
- [ ] `src/backend/internal/logic/searchconvslogic_test.go` — covers MEM-01 logic

*Existing `go test` infrastructure covers framework needs. No new test framework needed.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Settings page loads and saves | UI-03 | Browser UI interaction | Open /settings, update a field, reload, verify persisted |
| Search input in sidebar returns results | MEM-01 | Browser UI interaction | Type keyword in sidebar search, verify matching convs appear |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
