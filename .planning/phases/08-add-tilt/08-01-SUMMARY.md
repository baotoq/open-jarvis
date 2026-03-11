---
phase: 08-add-tilt
plan: 01
subsystem: infra
tags: [tilt, air, go, nextjs, live-reload, developer-experience]

# Dependency graph
requires:
  - phase: 06-add-go-linting
    provides: clean Go codebase that air can build and reload cleanly
provides:
  - Tilt single-command dev environment startup (tilt up)
  - Air live-reload for Go backend on .go file saves
  - Frontend npm run dev started after backend is ready (readiness probe)
affects: [07-be-reorganization, all future development phases]

# Tech tracking
tech-stack:
  added: [tilt, air]
  patterns: [local_resource with serve_cmd for long-running dev processes, readiness_probe TCP for ordered startup, resource_deps for dependency ordering]

key-files:
  created:
    - Tiltfile
    - src/backend/.air.toml
  modified:
    - src/backend/.gitignore
    - README.md

key-decisions:
  - "serve_cmd (not cmd) used in Tiltfile for long-running processes — cmd exits immediately after process spawned"
  - "readiness_probe tcp_socket_action(port=8888) ensures frontend resource_deps waits for backend to actually listen"
  - "ignore=['src/backend/tmp', 'src/backend/data'] prevents Tilt thrashing on air build output and SQLite writes"
  - "open-browser uses macOS open; Linux users change to xdg-open per README note"
  - "air exclude_regex=[_test\\.go] builds production binary only, not test files"

patterns-established:
  - "Tilt local_resource pattern: serve_cmd + serve_dir + deps + ignore + readiness_probe + resource_deps"

requirements-completed: [DEV-01, DEV-02, DEV-03, DEV-04, DEV-05, DEV-06]

# Metrics
duration: 5min
completed: 2026-03-12
---

# Phase 8 Plan 01: Add Tilt Summary

**Tiltfile with air live-reload for Go backend and npm run dev for Next.js, single `tilt up` starts both services with TCP readiness probe ordering**

## Performance

- **Duration:** ~5 min
- **Started:** 2026-03-12T08:06:09Z
- **Completed:** 2026-03-12T08:11:00Z
- **Tasks:** 1 of 2 (checkpoint:human-verify pending)
- **Files modified:** 4

## Accomplishments
- Tiltfile at project root with backend (air) and frontend (npm run dev) as local_resource processes
- TCP readiness probe on port 8888 ensures frontend waits for backend before starting
- air .air.toml configured for Go backend: builds to tmp/main, excludes test files, graceful shutdown
- src/backend/.gitignore updated with tmp/ to exclude air build output from git
- README.md updated with Tilt prerequisites, quickstart, stop instructions, and manual fallback

## Task Commits

Each task was committed atomically:

1. **Task 1: Create Tiltfile, .air.toml, update .gitignore and README** - `a9e4f70` (feat)

## Files Created/Modified
- `Tiltfile` - Tilt orchestration: backend (air) and frontend (npm run dev) as local_resource processes with readiness probe and resource_deps ordering
- `src/backend/.air.toml` - Air config: go build to tmp/main, exclude test files and tmp/data, graceful shutdown via send_interrupt
- `src/backend/.gitignore` - Added tmp/ entry to exclude air build output
- `README.md` - Added Development section with prerequisites (tilt, air), tilt up quickstart, tilt down, and manual fallback commands

## Decisions Made
- `serve_cmd` (not `cmd`) mandatory for long-running Tilt processes — `cmd` spawns and exits immediately
- `readiness_probe` with TCP on port 8888 chosen over HTTP probe — go-zero server may not have a health endpoint; TCP is sufficient to confirm it's listening
- `ignore` list for tmp/ and data/ prevents Tilt from re-triggering on air build artifacts and SQLite WAL writes
- `open-browser` resource uses macOS `open`; README notes Linux users should change to `xdg-open`

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

Manual verification required before plan is fully complete:

1. Ensure `tilt` and `air` are installed (see README prerequisites)
2. Run `mkdir -p src/backend/data`
3. Run `tilt up` from project root
4. Verify backend and frontend go green in Tilt dashboard (http://localhost:10350)
5. Test live reload by editing a .go file
6. Run `tilt down` to verify clean shutdown

## Next Phase Readiness
- Tilt workflow complete once human verification passes
- Phase 7 (BE reorganization) can proceed independently of this phase

## Self-Check

- [x] Tiltfile exists at `/Users/baotoq/Work/open-jarvis/Tiltfile`
- [x] src/backend/.air.toml exists
- [x] src/backend/.gitignore contains `tmp/`
- [x] README.md contains `tilt up`
- [x] Tiltfile contains `serve_cmd`, `readiness_probe`, `resource_deps`
- [x] Go tests pass: `go test ./... -count=1` — all ok
- [x] Commit a9e4f70 exists

---
*Phase: 08-add-tilt*
*Completed: 2026-03-12*
