# Phase 8: Add Tilt - Research

**Researched:** 2026-03-12
**Domain:** Developer tooling — Tilt local process orchestration + Air Go live reload
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- Local processes only — no Kubernetes cluster required
- Local dev only — not targeting self-hosted server
- `tilt up` starts both backend and frontend, auto-opens http://localhost:3000 in browser
- `tilt down` stops processes only — SQLite data in `src/backend/data/` and `etc/config.yaml` are preserved
- Bare local processes via `local_resource()` — no Docker, no image builds
- No Dockerfiles will be created in this phase
- Uses host Go and Node.js toolchains directly
- Use `air` for Go backend live reload
- Tilt runs `air` as the local_resource command inside `src/backend/`
- `air` watches `.go` files and rebuilds/restarts automatically
- `air` must be installed (prerequisite — document in README or Tiltfile)
- Next.js dev server started via `npm run dev` from `src/frontend/`
- Next.js already provides its own hot reload — no extra tooling needed

### Claude's Discretion

- Exact `air` configuration (`.air.toml` contents — build command, temp dir, watch paths)
- Whether to add a `Makefile` target or just rely on `tilt up`
- Tiltfile resource naming and grouping
- Whether to add a health check / readiness probe before opening browser

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope.
</user_constraints>

---

## Summary

This phase adds Tilt as a local development orchestrator. A single `tilt up` at the project root starts both the Go backend (via `air` for live reload) and the Next.js frontend (`npm run dev`). No Docker or Kubernetes is involved — everything runs as bare host processes using the `local_resource()` Tilt API.

The Tilt `local_resource()` API supports `serve_dir` for per-resource working directories, which cleanly handles the monorepo layout where backend lives in `src/backend/` and frontend in `src/frontend/`. Air requires a `.air.toml` in `src/backend/` and a `tmp/` directory which must be gitignored. The frontend has no extra configuration — Next.js hot reload works out of the box.

Tilt opens its own web dashboard at `http://localhost:10350` automatically. The `links` parameter on `local_resource()` shows the app URL in that dashboard. There is no native `tilt up` flag to open `http://localhost:3000` automatically in the user's browser — the CONTEXT.md decision to auto-open the browser requires either a `local()` call to `open` / `xdg-open` or documentation directing the user to press space in the Tilt TUI.

**Primary recommendation:** Write a single `Tiltfile` at project root using `local_resource()` with `serve_cmd` and `serve_dir`. Create `src/backend/.air.toml` for Air config. Add `src/backend/tmp/` to gitignore. Document `air` and `tilt` as prerequisites in README.

---

## Standard Stack

### Core

| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| Tilt | latest (v0.33+) | Dev workflow orchestrator — starts/stops/watches all services | Industry standard for multi-service local dev; `local_resource()` runs bare processes |
| air | latest (v1.61+) | Go live reload daemon | De-facto standard Go hot-reload; watches `.go` files, rebuilds binary, restarts process |

### Supporting

| Tool | Version | Purpose | When to Use |
|------|---------|---------|-------------|
| Next.js dev server | built-in | Frontend hot reload | Already provided by `npm run dev` — no extra tooling |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| air | `watchexec` or `gow` | air is Go-specific, has native build integration; alternatives are generic |
| Tilt `local_resource` | `Foreman`/`overmind`/`honcho` | Tilt provides UI dashboard, resource dependencies, readiness probes — richer than Procfile tools |

**Installation:**

```bash
# Tilt
curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash
# or macOS
brew install tilt

# air
go install github.com/air-verse/air@latest
# or macOS
brew install go-air
```

---

## Architecture Patterns

### Recommended Project Structure

```
open-jarvis/           # project root — Tiltfile lives here
├── Tiltfile           # single entry point: tilt up
├── src/
│   ├── backend/
│   │   ├── .air.toml  # air configuration for Go live reload
│   │   ├── tmp/       # air build output — gitignored
│   │   └── ...
│   └── frontend/
│       └── ...        # npm run dev handled directly
└── README.md          # updated with tilt up quickstart
```

### Pattern 1: local_resource with serve_cmd and serve_dir

**What:** Use `serve_cmd` (not `cmd`) for long-running processes. `serve_dir` sets the working directory. `deps` triggers restarts on file changes. `links` exposes the service URL in the Tilt dashboard.

**When to use:** Any persistent process (web server, dev server) that should stay running and restart on changes.

```python
# Source: https://docs.tilt.dev/api.html#api.local_resource
local_resource(
    'backend',
    serve_cmd='air',
    serve_dir='src/backend',
    deps=['src/backend'],
    ignore=['src/backend/tmp'],
    links=[link('http://localhost:8888', 'Backend API')],
    labels=['services'],
)

local_resource(
    'frontend',
    serve_cmd='npm run dev',
    serve_dir='src/frontend',
    deps=['src/frontend/app', 'src/frontend/components', 'src/frontend/hooks', 'src/frontend/lib'],
    ignore=['src/frontend/.next', 'src/frontend/node_modules'],
    links=[link('http://localhost:3000', 'Frontend')],
    labels=['services'],
)
```

### Pattern 2: Dependency ordering with resource_deps

**What:** `resource_deps` ensures one resource starts only after another is healthy/ready.

**When to use:** When frontend must wait for backend to be reachable before starting.

```python
# Source: https://docs.tilt.dev/api.html#api.local_resource
local_resource(
    'frontend',
    serve_cmd='npm run dev',
    serve_dir='src/frontend',
    resource_deps=['backend'],  # waits for backend to be ready
    ...
)
```

**Note:** `resource_deps` waits for the dependency's readiness probe to pass. Without a `readiness_probe` on backend, Tilt considers it "ready" immediately after `serve_cmd` starts. For this project, a TCP probe on port 8888 is the right approach.

### Pattern 3: Readiness probe for backend

```python
# Source: https://docs.tilt.dev/api.html#api.local_resource
local_resource(
    'backend',
    serve_cmd='air',
    serve_dir='src/backend',
    readiness_probe=probe(
        tcp_socket=tcp_socket_action(port=8888),
        period_secs=2,
        failure_threshold=10,
    ),
    ...
)
```

### Pattern 4: Browser open via local()

Tilt has no native flag to open an application URL on startup. Use a one-shot `local_resource` with `auto_init=True` to open the browser after the frontend is ready:

```python
# Source: https://docs.tilt.dev/api.html#api.local
local_resource(
    'open-browser',
    cmd='open http://localhost:3000',         # macOS
    # cmd='xdg-open http://localhost:3000',   # Linux
    resource_deps=['frontend'],
    auto_init=True,
    labels=['dev'],
)
```

**Caveat (LOW confidence):** Cross-platform `open` command differs between macOS (`open`) and Linux (`xdg-open`). The Tiltfile is Starlark (Python-like) so `os.getenv` or a conditional can detect platform. This is Claude's Discretion territory.

### Pattern 5: .air.toml for src/backend layout

Air runs from `src/backend/` (set via `serve_dir`). The entry point is `cmd/main.go`.

```toml
# src/backend/.air.toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ./cmd/main.go"
bin = "tmp/main"
include_ext = ["go"]
exclude_dir = ["tmp", "data"]
exclude_regex = ["_test\\.go"]
exclude_unchanged = true
delay = 500
stop_on_error = true
send_interrupt = true
kill_delay = 500000000

[misc]
clean_on_exit = true

[screen]
clear_on_rebuild = false
keep_scroll = true
```

### Anti-Patterns to Avoid

- **Using `cmd` instead of `serve_cmd` for the backend/frontend:** `cmd` expects a one-shot command that exits. Long-running servers must use `serve_cmd`.
- **Not setting `serve_dir`:** Without `serve_dir`, commands run from the Tiltfile directory (project root). `air` and `npm run dev` both require their respective package directories.
- **Watching `node_modules` or `tmp/`:** These must be in `ignore` or Tilt will thrash on rebuilds from generated files.
- **Using `go install` for air in go.mod:** Air installed as a `go tool` (Go 1.24 feature) requires `go tool air`. For simplicity and consistency, install air globally via `go install github.com/air-verse/air@latest` or Homebrew.
- **Air proxy enabled in dev:** The air proxy (`[proxy]` section in `.air.toml`) adds a browser live-reload proxy on a separate port. This is not needed since Tilt orchestrates the workflow; leave `enabled = false` (default) to avoid port conflicts.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| File watching + rebuild | Custom `fsnotify` watcher + shell loop | `air` | air handles debounce, interrupt signals, partial rebuilds, tmp cleanup |
| Multi-service startup | Shell script or Makefile with background `&` | Tilt | Tilt provides restart on crash, logs aggregation, dependency ordering, readiness probes |
| Go build binary management | Custom build scripts | air `.air.toml` `cmd`/`bin` fields | Consistent temp dir, exit cleanup, cross-platform |

**Key insight:** A shell script starting two processes with `&` has no restart-on-crash, no dependency ordering, and no live log aggregation. Tilt provides all of this with 20 lines of Starlark.

---

## Common Pitfalls

### Pitfall 1: air binary not found

**What goes wrong:** `tilt up` launches `air` via `serve_cmd` but air is not on PATH, causing the backend resource to fail immediately with "command not found".

**Why it happens:** air is installed separately from the project (`go install` puts it in `$GOPATH/bin` or `~/go/bin`). If that directory is not on PATH in the shell Tilt uses, the command fails.

**How to avoid:** Document `air` as a prerequisite in README. Use a `local()` guard in Tiltfile to check for air before starting:
```python
local('which air || (echo "ERROR: air not installed. Run: go install github.com/air-verse/air@latest" && exit 1)')
```

**Warning signs:** Backend resource goes red immediately on first `tilt up`.

### Pitfall 2: serve_dir not set — wrong working directory

**What goes wrong:** `air` runs from project root instead of `src/backend/`, fails to find `go.mod` or `cmd/main.go`.

**Why it happens:** `serve_dir` defaults to the Tiltfile directory (project root). The `air` tool expects to run from the Go module root.

**How to avoid:** Always set `serve_dir='src/backend'` on the backend resource and `serve_dir='src/frontend'` on the frontend resource.

### Pitfall 3: deps includes node_modules or .next

**What goes wrong:** Tilt detects thousands of file changes inside `node_modules` or `.next` after `npm run dev` starts, triggering continuous restart loops.

**Why it happens:** `deps` recursively watches directories. Generated directories must be excluded via `ignore`.

**How to avoid:**
```python
local_resource(
    'frontend',
    deps=['src/frontend'],
    ignore=['src/frontend/.next', 'src/frontend/node_modules'],
    ...
)
```

### Pitfall 4: src/backend/tmp/ not gitignored

**What goes wrong:** Air's compiled binary (`tmp/main`) gets committed to git.

**Why it happens:** `tmp/` is created by air on first build but `src/backend/.gitignore` doesn't exclude it.

**How to avoid:** Add `tmp/` to `src/backend/.gitignore`.

### Pitfall 5: resource_deps without readiness_probe causes premature frontend start

**What goes wrong:** Frontend starts before backend is listening on port 8888. SSE connections fail on initial page load.

**Why it happens:** Without a `readiness_probe`, Tilt marks a resource ready as soon as `serve_cmd` is launched (not when it's listening).

**How to avoid:** Add a TCP readiness probe on port 8888 to the backend resource. Frontend uses `resource_deps=['backend']`.

---

## Code Examples

Verified patterns from official sources:

### Complete Tiltfile

```python
# Source: https://docs.tilt.dev/api.html#api.local_resource

# Prerequisite guard: fail fast if air is not installed
local('which air || (echo "ERROR: air not installed. Run: go install github.com/air-verse/air@latest" && exit 1)', quiet=True)

local_resource(
    'backend',
    serve_cmd='air',
    serve_dir='src/backend',
    deps=['src/backend'],
    ignore=['src/backend/tmp', 'src/backend/data'],
    readiness_probe=probe(
        tcp_socket=tcp_socket_action(port=8888),
        period_secs=2,
        failure_threshold=15,
    ),
    links=[link('http://localhost:8888', 'Backend API')],
    labels=['services'],
)

local_resource(
    'frontend',
    serve_cmd='npm run dev',
    serve_dir='src/frontend',
    deps=['src/frontend'],
    ignore=['src/frontend/.next', 'src/frontend/node_modules'],
    resource_deps=['backend'],
    links=[link('http://localhost:3000', 'Frontend')],
    labels=['services'],
)
```

### Minimal .air.toml for src/backend

```toml
# Source: https://github.com/air-verse/air (air_example.toml)
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ./cmd/main.go"
bin = "tmp/main"
include_ext = ["go"]
exclude_dir = ["tmp", "data"]
exclude_regex = ["_test\\.go"]
exclude_unchanged = true
delay = 500
stop_on_error = true
send_interrupt = true
kill_delay = 500000000

[misc]
clean_on_exit = true
```

### link() helper

```python
# Source: https://docs.tilt.dev/accessing_resource_endpoints.html
links=[
    link('http://localhost:3000', 'Frontend'),
    link('http://localhost:8888', 'Backend API'),
]
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `air` run manually per terminal | `tilt up` orchestrates air + npm | Tilt `local_resource` v0.20+ | Single command for full stack |
| `air` installed via `go install` only | Also available as `go tool` (Go 1.24+) or `brew` | 2024 | `go tool air` is project-scoped, avoids PATH issues |
| Tilt required Kubernetes | Tilt works fully with `local_resource` only (no K8s) | Tilt v0.17+ | No cluster needed for local dev |

**Deprecated/outdated:**

- **air proxy** (`[proxy]` in .air.toml): The browser live-reload proxy is not needed when using Tilt — Tilt already orchestrates restarts. Leave proxy disabled.
- **air `full_bin`**: Superseded by `entrypoint` field. Using `bin` + standard invocation is fine for this use case.

---

## Open Questions

1. **Browser auto-open cross-platform**
   - What we know: Tilt has no native `--open-browser <url>` for app URLs; `tilt up` opens its own dashboard (port 10350). macOS uses `open`, Linux uses `xdg-open`.
   - What's unclear: Whether the Tiltfile should implement platform detection or simply document "visit http://localhost:3000".
   - Recommendation: As Claude's Discretion, implement a simple `local_resource` with `cmd='open http://localhost:3000'` (macOS-first) and note in README that Linux users may need to change to `xdg-open`. Alternatively, rely on the Tilt dashboard link — simpler and more portable.

2. **air prerequisite enforcement**
   - What we know: `which air` will fail gracefully if air is absent; Tilt surfaces the error in the resource log.
   - What's unclear: Whether to check for `tilt` itself in a Makefile target.
   - Recommendation: Add a `which air` guard in the Tiltfile; document installation in README. A Makefile target (`make dev`) is Claude's Discretion — low value given `tilt up` is already one command.

---

## Validation Architecture

> nyquist_validation is enabled (config.json does not set it to false).

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go standard testing + testify (already installed) |
| Config file | `src/backend/.golangci.yml` (linting), no separate test config |
| Quick run command | `cd src/backend && go test ./... -count=1` |
| Full suite command | `cd src/backend && go test ./... -cover -count=1` |

### Phase Requirements → Test Map

This phase is pure developer tooling with no v1 application requirements. There are no automated tests to write for Tiltfile or `.air.toml`. Validation is manual/smoke.

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| DEV-01 | `tilt up` starts backend (air) | smoke/manual | `which air && cd src/backend && air -v` | N/A |
| DEV-02 | `tilt up` starts frontend (npm run dev) | smoke/manual | `cd src/frontend && npm run dev` | N/A |
| DEV-03 | Go file change triggers rebuild | smoke/manual | manual observation | N/A |
| DEV-04 | Backend reachable at localhost:8888 | smoke/manual | `curl -sf http://localhost:8888/api/health || curl -sf http://localhost:8888` | N/A |
| DEV-05 | Frontend reachable at localhost:3000 | smoke/manual | `curl -sf http://localhost:3000` | N/A |
| DEV-06 | `tilt down` preserves SQLite data | smoke/manual | `ls src/backend/data/*.db` | N/A |

### Sampling Rate

- **Per task commit:** No automated test suite changes — existing `go test ./...` still applies to backend code
- **Per wave merge:** `cd src/backend && go test ./... -cover` (existing suite)
- **Phase gate:** Manual smoke test: `tilt up`, verify both services green in dashboard, change a `.go` file, verify rebuild

### Wave 0 Gaps

None — no new test files required. This phase adds configuration files only (Tiltfile, .air.toml). Existing test infrastructure is unaffected.

---

## Sources

### Primary (HIGH confidence)

- https://docs.tilt.dev/api.html#api.local_resource — complete `local_resource` parameter list including `serve_cmd`, `serve_dir`, `dir`, `deps`, `ignore`, `links`, `readiness_probe`, `resource_deps`
- https://docs.tilt.dev/local_resource.html — local resource guide with serve_cmd patterns
- https://docs.tilt.dev/accessing_resource_endpoints.html — `link()` helper usage
- https://github.com/air-verse/air (air_example.toml) — canonical `.air.toml` configuration
- https://docs.tilt.dev/cli/tilt_up.html — tilt up CLI flags

### Secondary (MEDIUM confidence)

- https://github.com/tilt-dev/tilt/issues/4549 — working directory for local_resource confirmed via `dir`/`serve_dir` parameters
- https://github.com/tilt-dev/tilt/pull/3603 — `serve_cmd` workdir fix confirmed merged

### Tertiary (LOW confidence)

- Cross-platform browser open via `local_resource` — pattern inferred from Tilt `local()` API; no official example found for this specific pattern

---

## Metadata

**Confidence breakdown:**

- Standard stack: HIGH — Tilt and air are well-documented; API verified from official docs
- Architecture: HIGH — `local_resource` API parameters confirmed from https://docs.tilt.dev/api.html
- Pitfalls: HIGH — derived from API behavior (serve_dir default, ignore semantics) plus common Go tooling patterns
- Browser auto-open: LOW — no official Tilt mechanism; pattern inferred

**Research date:** 2026-03-12
**Valid until:** 2026-06-12 (Tilt API is stable; air config format is stable)
