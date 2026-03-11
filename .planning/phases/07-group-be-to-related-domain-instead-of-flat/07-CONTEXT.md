# Phase 7: Group BE to Related Domain Instead of Flat - Context

**Gathered:** 2026-03-12
**Status:** Ready for planning

<domain>
## Phase Boundary

Reorganize the Go backend's flat `internal/handler/` and `internal/logic/` packages into a domain-grouped, domain-first subdirectory structure. This is a pure structural refactor — no new capabilities, no behavior changes, no API contract changes.

</domain>

<decisions>
## Implementation Decisions

### Scope
- Reorganize: `internal/handler/` and `internal/logic/` only
- Leave untouched: `internal/svc/`, `internal/types/`, `internal/toolexec/`, `internal/config/`
- `svc/` stays flat — it's the wiring layer and depends on all stores; splitting would require circular imports or a new shared types package
- `types/` stays flat — shared request/response structs used across all domains
- `toolexec/` stays flat — already a coherent domain package (file/shell/web tools)

### Domain grouping
Three domains, mirroring the API URL structure:
- **chat** — `ChatStreamHandler`, `ApproveHandler`, `chatlogic.go` (approval is part of the agentic loop, chatlogic calls waitForApproval)
- **conv** — `ListConversationsHandler`, `GetConversationHandler`, `GetConversationMessagesHandler`, `DeleteConversationHandler`, `SearchConversationsHandler`, and corresponding logic files
- **config** — `GetConfigHandler`, `UpdateConfigHandler`, and corresponding logic files

### Package structure style
**Domain-first layout:**
```
internal/
  chat/
    handler/
      chatstreamhandler.go
      approvehandler.go
    logic/
      chatlogic.go
  conv/
    handler/
      listconvshandler.go
      getconvhandler.go
      getconvmessageshandler.go
      deleteconvhandler.go
      searchconvshandler.go
    logic/
      listconvslogic.go
      getconvlogic.go
      getconvmessageslogic.go
      deleteconvlogic.go
      searchconvslogic.go
  config/
    handler/
      getconfighandler.go
      updateconfighandler.go
    logic/
      getconfiglogic.go
      updateconfiglogic.go
```

### Package naming
- Package names should be short and idiomatic — avoid import aliasing in `main.go`
- Each domain's `handler/` subdirectory: use the domain name as package (`package chat`, `package conv`, `package config` — or `package cfg` if `config` conflicts with `internal/config`)
- Each domain's `logic/` subdirectory: use the domain name as package for consistency

### main.go updates
- Route registrations in `cmd/main.go` must be updated to use the new domain package imports
- All three domain handler packages will be imported; package names drive the call syntax (e.g., `chat.ChatStreamHandler(svcCtx)`)

### Tests
- Test files move with their implementation files — no restructuring of test strategy

### Claude's Discretion
- Exact package names for `logic/` subdirectories (could use domain name or `logic` with aliases)
- Whether to use `package cfg` vs `package config` to avoid collision with `internal/config`
- File naming within new directories (can keep existing names verbatim)

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- All existing handler and logic files move as-is — content unchanged, only import paths update
- `svc.ServiceContext` is the single dependency passed to all handlers/logic — import path unchanged

### Established Patterns
- go-zero Handler→Logic→ServiceContext pattern is preserved — just moves to subdirectory
- Handlers import `internal/svc` and `internal/types` (both stay flat, import paths unchanged)
- Logic files import `internal/svc` and `internal/types` (same)
- `cmd/main.go` imports all handler packages for route registration — needs updating

### Integration Points
- `cmd/main.go` is the only file that imports from `internal/handler/` directly — update needed
- `internal/logic/` files are only imported from `internal/handler/` — both move together
- No cross-domain logic dependencies (chatlogic does not import convlogic, etc.)

</code_context>

<specifics>
## Specific Ideas

- User referenced the flat `src/backend/internal/handler` directory as the motivation — the goal is to reduce the flat file list in each package
- Domain names mirror API URL segments: `/api/chat`, `/api/conversations` → `conv`, `/api/config` → `config`

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 07-group-be-to-related-domain-instead-of-flat*
*Context gathered: 2026-03-12*
