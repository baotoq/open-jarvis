# Phase 6: Add Go Linting - Research

**Researched:** 2026-03-12
**Domain:** Go static analysis / golangci-lint configuration
**Confidence:** HIGH

## Summary

This phase adds golangci-lint to the Go backend (`src/backend/`) to enforce code quality standards consistently. The project skill `go-linting` prescribes the Uber Go Style Guide's minimum set of linters: `errcheck`, `goimports`, `revive`, `govet`, and `staticcheck`. golangci-lint v2 is the standard runner.

A key constraint discovered during research: golangci-lint must be installed from the **official pre-built binary** (install script or GitHub release), not via `go install`. The project's `go.mod` declares `go 1.26` (Go 1.26 was released 2026-02-10, current project uses go1.26.0). Binaries compiled via `go install` inherit the local toolchain (go1.23.4 or go1.25.8 depending on context), which golangci-lint v2 rejects when the module's declared Go version is higher.

Running the official pre-built golangci-lint v2.11.3 binary (built with go1.26.1) against the codebase revealed **50 existing issues** across 5 linter categories. All are real and fixable — none require architecture changes. Two categories need exclusion rules (go-zero json struct tags trigger false-positive staticcheck SA5008; blank sqlite driver import is intentional).

**Primary recommendation:** Install golangci-lint v2.11.3 via the official install script, add `.golangci.yml` to `src/backend/`, fix all 50 existing issues, and add `golangci-lint run` to the CLAUDE.md commands section. No CI infrastructure exists yet — enforcement is local only for this phase.

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| golangci-lint | v2.11.3 | Lint runner — aggregates multiple linters | Project skill prescribes it; Uber Go Style Guide standard |
| errcheck | built-in | Catch unchecked error returns | 21 existing violations; prevents silent failures |
| staticcheck | built-in | SA-series static analysis checks | 6 existing violations (all false positives for go-zero tags — need exclusion) |
| revive | built-in | Style linter (golint successor) | 20 existing violations; enforces naming, comments, unused params |
| ineffassign | built-in | Detect ineffectual assignments | 1 existing violation |
| unused | built-in | Detect unused exported symbols | 2 existing violations |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| govet | built-in (go toolchain) | Compiler-level correctness checks | Already passing; included for completeness |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| golangci-lint | Running each linter separately | golangci-lint parallelizes and caches; much faster |
| revive | golint | golint is deprecated; revive is drop-in replacement with more features |
| official install script | `go install` | `go install` compiles with local toolchain — fails if local Go < go.mod version |

**Installation (from `src/backend/` directory):**

```bash
# Official install script — provides binary built with go1.26.1
curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.11.3

# Verify
golangci-lint --version
# Expected: golangci-lint has version 2.11.3 built with go1.26.1 ...
```

**Why NOT `go install`:**
```bash
# DO NOT USE — installs binary built with local go toolchain (e.g., go1.25.8)
# This fails with: "go language version (go1.25) used to build golangci-lint is lower than targeted Go version (1.26)"
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

## Architecture Patterns

### Config File Location

Place `.golangci.yml` at `src/backend/.golangci.yml` — same directory as `go.mod`. golangci-lint searches upward from the package being linted.

### Recommended Config Structure (golangci-lint v2 format)

```yaml
# src/backend/.golangci.yml
version: "2"

linters:
  default: none
  enable:
    - errcheck
    - ineffassign
    - staticcheck
    - revive
    - unused

  settings:
    goimports:
      local-prefixes: open-jarvis
    revive:
      rules:
        - name: blank-imports
        - name: context-as-argument
        - name: error-return
        - name: error-strings
        - name: exported
        - name: var-naming
        - name: package-comments
        - name: unused-parameter

  exclusions:
    rules:
      # go-zero uses json struct tag extensions (default=, optional) that are
      # unknown to staticcheck SA5008 — these are intentional and correct
      - linters: [staticcheck]
        text: "SA5008"
      # SQLite driver must be registered via blank import; this is intentional
      - linters: [revive]
        text: "blank-imports"
        path: "svc/servicecontext.go"

formatters:
  enable:
    - goimports

  settings:
    goimports:
      local-prefixes:
        - open-jarvis

run:
  timeout: 5m
```

**CRITICAL:** golangci-lint v2 changed the config schema from v1:
- Top-level `version: "2"` is required
- `linters-settings` is now `linters.settings` (nested)
- `issues.exclude-rules` is now `linters.exclusions.rules` (nested under linters)
- `goimports` moved to `formatters` section
- `linters.default: none` replaces `linters.disable-all: true`

### Running Lint

```bash
# From src/backend/
golangci-lint run ./...

# Run only fast linters during development
golangci-lint run --fast-only ./...

# Check a single package
golangci-lint run ./internal/svc/...
```

### Anti-Patterns to Avoid

- **Adding `//nolint` comments without justification:** Only suppress with `//nolint:lintername // reason` explaining why.
- **Using `linters.default: standard` or `linters.default: all`:** The project uses explicit `default: none` + `enable` list for predictability.
- **Suppressing entire linter categories:** Fix issues; exclusions should target specific false positives only.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Linting multiple packages | Shell loop over packages | `golangci-lint run ./...` | Built-in parallelism and caching |
| Suppressing false positives | Per-file nolint directives | Config-level `exclusions.rules` | Centralized, auditable, reviewable |
| Enforcing via git hook | Custom pre-commit script | golangci-lint with standard runner | Config-driven; reuses same config as CI |

**Key insight:** The `staticcheck SA5008` false positives for go-zero struct tags are a known cross-tool incompatibility. go-zero extends the `json:` struct tag with `default=` and `optional` options that are meaningless to standard JSON tooling. The correct fix is a targeted exclusion rule, not modifying the struct tags (which would break go-zero config loading).

## Common Pitfalls

### Pitfall 1: Installing via `go install` on a Mixed-Toolchain System

**What goes wrong:** `go install` compiles golangci-lint with whatever Go version ran the install command. If that version is lower than `go.mod`'s declared version, golangci-lint refuses to run with: `"can't load config: the Go language version (goX.Y) used to build golangci-lint is lower than the targeted Go version (Z.W)"`.

**Why it happens:** This project uses `go 1.26` in `go.mod` (required by `modernc.org/sqlite v1.46.1`), but the local `go` binary may be 1.23 or 1.25.

**How to avoid:** Always install golangci-lint from the official install script or GitHub release tarball, which provides a binary built with the newest Go release.

**Warning signs:** `golangci-lint --version` shows `built with go1.25.x` or earlier.

### Pitfall 2: Using golangci-lint v1 Config Format

**What goes wrong:** golangci-lint v2 changed its config schema. v1 config keys like `linters-settings`, `issues.exclude-rules`, and `linters.disable-all` are silently ignored or produce parse errors.

**Why it happens:** Most online examples and Stack Overflow answers still use v1 format.

**How to avoid:** Always start with `version: "2"` at the top of `.golangci.yml`. Use `linters.settings` (not `linters-settings`), `linters.exclusions.rules` (not `issues.exclude-rules`), and `linters.default: none` (not `linters.disable-all: true`).

**Warning signs:** Config validates but linters you didn't enable are running, or exclusions don't work.

### Pitfall 3: Fixing `staticcheck SA5008` by Removing go-zero Struct Tags

**What goes wrong:** Removing `json:",default=..."` struct tags breaks go-zero's config loading. go-zero reads these tags to set field defaults.

**Why it happens:** staticcheck SA5008 reports "invalid appearance of unknown `default` tag option" — the warning is accurate for standard `encoding/json` but wrong for go-zero's custom tag parser.

**How to avoid:** Exclude SA5008 at the config level; do not modify go-zero config struct tags.

**Warning signs:** After removing tags, config fields load as zero values; tests that rely on defaults fail.

### Pitfall 4: Treating `revive` `blank-imports` Warning on SQLite Driver as Real Issue

**What goes wrong:** Removing `_ "modernc.org/sqlite"` from `servicecontext.go` causes a runtime panic — the SQLite driver never registers with `database/sql`.

**Why it happens:** `revive` warns that blank imports should only appear in `main` or test packages, which is a valid style rule. But SQLite driver registration is a known Go idiom that requires it.

**How to avoid:** Exclude this specific case via `linters.exclusions.rules` targeting `servicecontext.go`.

### Pitfall 5: Fixing `SessionId` Naming Requires Coordinated Frontend Change

**What goes wrong:** Renaming `SessionId` to `SessionID` in `internal/types/types.go` changes the struct field name but NOT the JSON key (which is `json:"sessionId"`). The frontend reads `sessionId` from JSON — no breakage.

**Why it happens:** Developers assume JSON field names change with struct field names.

**How to avoid:** Confirm the `json:"sessionId"` tag stays unchanged. The rename is safe.

## Code Examples

### errcheck: Fixing `defer rows.Close()`

```go
// BEFORE (triggers errcheck)
defer rows.Close()

// AFTER
defer func() {
    if err := rows.Close(); err != nil {
        log.Printf("rows.Close: %v", err)
    }
}()

// Alternative for test code where errors are ignorable
defer rows.Close() //nolint:errcheck // cleanup in defer; error logged by sql driver
```

### errcheck: Fixing `fmt.Fprintf` in SSE handler

```go
// BEFORE — SSE streaming in chatlogic.go (w is http.ResponseWriter)
fmt.Fprintf(w, "data: %s\n\n", delta.Content)

// AFTER — write errors in SSE are not actionable (connection already broken)
if _, err := fmt.Fprintf(w, "data: %s\n\n", delta.Content); err != nil {
    return // connection closed; stop streaming
}
```

### revive: Fixing `unused-parameter` with underscore

```go
// BEFORE
func (w *WebFetchTool) Fetch(ctx context.Context, argsJSON string) ToolResult {

// AFTER — if ctx is genuinely unused in implementation
func (w *WebFetchTool) Fetch(_ context.Context, argsJSON string) ToolResult {
```

### revive: Adding package comment

```go
// BEFORE — no package comment
package toolexec

// AFTER
// Package toolexec provides tool execution primitives for the agentic loop.
package toolexec
```

### ineffassign: Fix unused append in test

```go
// BEFORE — internal/svc/convstore_test.go line 61
got = append(got, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: "extra"})

// AFTER — either remove the line or use the result
_ = append(got, openai.ChatCompletionMessage{...}) // if testing append doesn't panic
// OR remove the line entirely if the test doesn't need it
```

### unused: Fix unused mock in servicecontext_test.go

```go
// BEFORE — mockAIStreamer is defined but never used in tests
type mockAIStreamer struct{}
func (m *mockAIStreamer) CreateChatCompletionStream(...) (interface{}, error) { return nil, nil }

// AFTER — remove the type and method entirely, or use it in a test
```

## Existing Issues: Complete Inventory

**Total: 50 issues across 5 files-of-concern**

| Category | Count | Files Affected | Fix Approach |
|----------|-------|----------------|--------------|
| errcheck | 21 | chatlogic.go, sqlitestore.go, webtool.go, several _test.go | Handle or explicitly ignore with comment |
| revive | 20 | config.go, convstore.go, types.go, several _test.go, servicecontext.go | Fix naming, add comments, use `_` for unused params |
| staticcheck SA5008 | 6 | config.go | Config-level exclusion (go-zero tags — false positive) |
| unused | 2 | servicecontext_test.go | Remove unused mockAIStreamer |
| ineffassign | 1 | convstore_test.go | Remove or use the appended value |

**False positives (exclude in config, do NOT fix in code):**
- staticcheck SA5008 on config.go — go-zero struct tag extensions
- revive blank-imports on servicecontext.go — SQLite driver registration

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| golint | revive | ~2020 | golint is deprecated; revive is drop-in with more rules |
| golangci-lint v1 config (`linters-settings`) | golangci-lint v2 config (`linters.settings`) | v2.0 (2024) | Config schema changed; v1 configs silently malfunction |
| `go install golangci-lint` | Official install script | Ongoing | `go install` can't cross-compile; binary install is version-stable |
| `linters.disable-all: true` | `linters.default: none` | v2.0 | Schema rename |

## Open Questions

1. **CI integration**
   - What we know: No `.github/workflows/` directory exists; this is a personal project
   - What's unclear: Is CI planned? GitHub Actions would be the natural next step
   - Recommendation: Document `golangci-lint run ./...` in CLAUDE.md; defer CI until there's a stated need. The phase scope is local enforcement only.

2. **Makefile vs README commands**
   - What we know: The project has no Makefile; commands are documented in CLAUDE.md
   - What's unclear: Whether a `make lint` target is preferred
   - Recommendation: Add `golangci-lint run ./...` to the Go commands section in `src/backend/CLAUDE.md` (alongside `go vet ./...`). No Makefile needed.

3. **`fmt.Fprintf` in SSE error paths**
   - What we know: SSE writes to `http.ResponseWriter`; write errors mean the connection closed
   - What's unclear: Whether logging or returning is preferred for each error site
   - Recommendation: For error SSE frames, log and return. For data SSE frames, check error and return to stop streaming. Specific decisions per call site.

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | go test (stdlib) + testify v1.11.1 |
| Config file | none (runs via `go test ./...`) |
| Quick run command | `cd src/backend && go test ./... -count=1` |
| Full suite command | `cd src/backend && go test -cover ./...` |

### Phase Requirements to Test Map

| Behavior | Test Type | Automated Command | Notes |
|----------|-----------|-------------------|-------|
| golangci-lint runs without errors on codebase | smoke | `cd src/backend && golangci-lint run ./...` | Exit 0 = pass |
| All existing Go tests still pass after fixes | regression | `cd src/backend && go test ./...` | Must not regress |
| Config file is valid (no parse errors) | smoke | `cd src/backend && golangci-lint config verify` | If `verify` subcommand exists |

### Sampling Rate

- **Per task commit:** `cd src/backend && go test ./...` (quick regression check)
- **Per wave merge:** `cd src/backend && golangci-lint run ./... && go test -cover ./...`
- **Phase gate:** `golangci-lint run ./...` exits 0 before `/gsd:verify-work`

### Wave 0 Gaps

None — no new test files are needed. The linter is the test. The regression suite already exists.

## Sources

### Primary (HIGH confidence)

- golangci-lint v2.11.3 binary (built with go1.26.1) — ran directly against codebase; all 50 issues confirmed
- Project skill `go-linting` SKILL.md — prescribes minimum linter set and golangci-lint runner
- `src/backend/go.mod` — confirmed `go 1.26`, `modernc.org/sqlite v1.46.1` as root of version constraint
- `src/backend/CLAUDE.md` — confirmed conventions: initialisims (`sessionID` not `sessionId`), explicit error returns

### Secondary (MEDIUM confidence)

- [golangci-lint.run/docs/configuration/file/](https://golangci-lint.run/docs/configuration/file/) — v2 config schema structure (verified via live binary testing)
- [golangci-lint.run/docs/welcome/install/local/](https://golangci-lint.run/docs/welcome/install/local/) — binary install recommendation
- GitHub API: golangci-lint v2.11.3 darwin-arm64 release asset — confirmed `built with go1.26.1`

### Tertiary (LOW confidence)

- WebSearch: Go 1.26 release date (February 10, 2026) — consistent with go.dev/blog/go1.26

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — verified by running binary, reading skill
- Architecture: HIGH — verified by running golangci-lint with proposed config against live codebase
- Pitfalls: HIGH — all pitfalls discovered empirically during research, not theorized
- Existing issues inventory: HIGH — output directly from golangci-lint v2.11.3 on the actual codebase

**Research date:** 2026-03-12
**Valid until:** 2026-06-12 (golangci-lint releases; go-zero struct tag behavior is stable)
