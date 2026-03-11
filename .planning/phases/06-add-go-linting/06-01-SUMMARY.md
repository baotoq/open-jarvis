---
phase: 06-add-go-linting
plan: 01
subsystem: infra
tags: [golangci-lint, linting, go, staticcheck, errcheck, revive, ineffassign, unused, goimports]

# Dependency graph
requires: []
provides:
  - golangci-lint v2.11.3 binary installed on PATH (built with go1.26.1)
  - src/backend/.golangci.yml with v2 schema and five linters configured
  - Two targeted exclusions for known false positives (SA5008, blank-imports)
  - goimports formatter with open-jarvis local prefix
  - CLAUDE.md Go commands updated with golangci-lint run ./...
affects: [06-add-go-linting Wave 2 (lint fix plans)]

# Tech tracking
tech-stack:
  added: [golangci-lint v2.11.3]
  patterns:
    - golangci-lint v2 config schema — version: "2", linters.settings (nested), linters.exclusions.rules, linters.default: none
    - Config-level exclusion for false positives rather than per-file nolint directives

key-files:
  created: [src/backend/.golangci.yml]
  modified: [CLAUDE.md]

key-decisions:
  - "golangci-lint installed via official install script (not go install) — go install compiles with local toolchain which may be older than go.mod's declared go 1.26, causing golangci-lint to refuse to run"
  - "staticcheck SA5008 excluded at config level for go-zero struct tags (default=, optional extensions are false positives not code bugs)"
  - "revive blank-imports excluded only for svc/servicecontext.go — SQLite driver registration is an intentional blank import idiom"

patterns-established:
  - "golangci-lint v2 config: use linters.default: none + explicit enable list for predictability"
  - "Exclusions target specific false positives only — do not suppress entire linter categories"

requirements-completed: [lint-install, lint-config]

# Metrics
duration: 2min
completed: 2026-03-11
---

# Phase 6 Plan 01: Install golangci-lint and Write Config Summary

**golangci-lint v2.11.3 (built with go1.26.1) installed via official script with .golangci.yml enabling errcheck, ineffassign, staticcheck, revive, and unused linters with two targeted false-positive exclusions**

## Performance

- **Duration:** ~2 min
- **Started:** 2026-03-11T18:27:12Z
- **Completed:** 2026-03-11T18:28:59Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- Installed golangci-lint v2.11.3 pre-built binary (go1.26.1) from official install script — replaces previously installed v2.11.3 built with go1.25.8 which would fail against the project's go 1.26 go.mod
- Created src/backend/.golangci.yml using v2 schema with five linters: errcheck, ineffassign, staticcheck, revive, unused
- Added two config-level exclusion rules to suppress known false positives (SA5008 on go-zero struct tags, blank-imports on servicecontext.go)
- Updated CLAUDE.md Go commands section with `golangci-lint run ./...`
- Confirmed linter runs and reports ~46 real issues (Wave 2 will fix them); config verify exits 0

## Task Commits

Each task was committed atomically:

1. **Task 1: Install golangci-lint v2.11.3** — no repo file changes (binary installed to GOPATH/bin)
2. **Task 2: Write .golangci.yml and update CLAUDE.md** - `0ec30da` (chore)

**Plan metadata:** (docs commit — see below)

## Files Created/Modified

- `src/backend/.golangci.yml` — golangci-lint v2 configuration with five linters, two exclusions, goimports formatter
- `CLAUDE.md` — Added `golangci-lint run ./...` to Go (Backend) commands section

## Decisions Made

- Installed pre-built binary from official install script — `go install` would compile with local toolchain (go1.25.8 or go1.23.4) which golangci-lint v2 rejects when go.mod declares `go 1.26`
- Used v2 config schema — `linters.settings`, `linters.exclusions.rules`, `linters.default: none` (not v1 keys)
- Excluded SA5008 at config level, not per-file — go-zero struct tag extensions (`default=`, `optional`) are globally intentional
- Excluded blank-imports only for `svc/servicecontext.go` path — SQLite driver registration is a known Go idiom requiring blank import

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

None. The install script fetched the darwin-arm64 binary built with go1.26.1. Config verify passed on first attempt. Linter runs and reports 46 real issues as expected.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- golangci-lint foundation is in place; Wave 2 plans can now fix the 46 real issues across errcheck (21), revive (19), unused (2), ineffassign (1), and goimports (3) categories
- All exclusion rules are verified working — SA5008 and blank-imports on servicecontext.go are suppressed correctly

## Self-Check: PASSED

All files found on disk. Commit 0ec30da verified in git log.

---
*Phase: 06-add-go-linting*
*Completed: 2026-03-11*
