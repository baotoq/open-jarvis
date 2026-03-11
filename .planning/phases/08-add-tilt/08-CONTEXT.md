# Phase 8: Add Tilt - Context

**Gathered:** 2026-03-12
**Status:** Ready for planning

<domain>
## Phase Boundary

Add [Tilt](https://tilt.dev/) for local development workflow orchestration. A single `tilt up` command starts both the Go backend and Next.js frontend with live reload. No Kubernetes, no Docker — bare local processes using the host toolchain.

</domain>

<decisions>
## Implementation Decisions

### Deployment target
- Local processes only — no Kubernetes cluster required
- Local dev only — not targeting self-hosted server
- `tilt up` starts both backend and frontend, auto-opens http://localhost:3000 in browser
- `tilt down` stops processes only — SQLite data in `src/backend/data/` and `etc/config.yaml` are preserved

### Docker vs bare process
- Bare local processes via `local_resource()` — no Docker, no image builds
- No Dockerfiles will be created in this phase
- Uses host Go and Node.js toolchains directly

### Go live reload
- Use `air` for Go backend live reload
- Tilt runs `air` as the local_resource command inside `src/backend/`
- `air` watches `.go` files and rebuilds/restarts automatically
- `air` must be installed (prerequisite — document in README or Tiltfile)

### Frontend
- Next.js dev server started via `npm run dev` from `src/frontend/`
- Next.js already provides its own hot reload — no extra tooling needed

### Claude's Discretion
- Exact `air` configuration (`.air.toml` contents — build command, temp dir, watch paths)
- Whether to add a `Makefile` target or just rely on `tilt up`
- Tiltfile resource naming and grouping
- Whether to add a health check / readiness probe before opening browser

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `src/backend/cmd/main.go` — entry point for Go backend; `air` will run `go build -o ./tmp/main ./cmd/main.go && ./tmp/main`
- `src/backend/etc/config.yaml` — runtime config read at startup; preserved by `tilt down`
- `src/frontend/` — Next.js app; `npm run dev` already defined in package.json

### Established Patterns
- Backend commands run from `src/backend/` directory (all CLAUDE.md commands scoped there)
- Frontend commands run from `src/frontend/` directory
- SQLite data lives in `src/backend/data/` — must exist before first run (`mkdir -p`)

### Integration Points
- Tiltfile lives at project root (standard Tilt convention)
- `local_resource("backend", ...)` watches `src/backend/**/*.go`
- `local_resource("frontend", ...)` runs `npm run dev` in `src/frontend/`
- Tilt's `serve_cmd` or `cmd` with a serve process for long-running services

</code_context>

<specifics>
## Specific Ideas

- Single `tilt up` from project root should be the complete dev workflow
- Browser should open automatically to http://localhost:3000 (Next.js frontend)

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 08-add-tilt*
*Context gathered: 2026-03-12*
