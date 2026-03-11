# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

open-jarvis is a personal AI assistant inspired by OpenClaw, built as separate Go (backend) and TypeScript (frontend) services.

## Architecture

- **Backend**: Go service using the [go-zero](https://go-zero.dev) framework — handles API, logic, and AI integrations
- **Frontend**: Next.js (TypeScript) application — the user-facing UI (not yet implemented)

The two services are **decoupled** and communicate over HTTP/RPC. Changes to one do not require changes to the other unless the API contract changes.

## Commands

### Go (Backend)
> Run all Go commands from the `src/` directory.

```bash
cd src
go build ./...                     # build
go run ./cmd/main.go               # run server (reads etc/config.yaml)
go test ./...                      # test all
go test -v -run TestName ./pkg/    # run specific test
go test -cover ./...               # with coverage
go vet ./...                       # static analysis
go mod tidy                        # clean dependencies
```

### TypeScript / Next.js (Frontend)
> Frontend not yet implemented.

## Conventions

- Go: `gofmt` style, explicit error returns, go-zero Handler→Logic→ServiceContext pattern
- Tests: use `github.com/stretchr/testify/assert` and `require` (not raw `t.Errorf`)
- TypeScript (future): strict mode, Next.js App Router, PascalCase components
- Package manager for TypeScript: `npm`

## Structure (Backend)

```
src/
├── cmd/main.go                 # entry point
├── etc/config.yaml             # runtime config (model URL, system prompt, timeouts)
├── internal/
│   ├── config/                 # Config struct with defaults
│   ├── handler/                # HTTP handlers (parse request, call logic)
│   ├── logic/                  # Business logic (StreamChat SSE loop)
│   ├── svc/                    # ServiceContext, ConvStore, AI client interface
│   └── types/                  # Shared request/response types
└── go.mod
```

## Gotchas

- SSE streaming: handlers write `data: <token>\n\n` directly to `http.ResponseWriter`; no framework abstraction
- go-zero layer rule: handlers must not contain logic; logic must not import handler types
- Config defaults are set via struct tags (`default:"value"`), not code

## Codebase Docs

Refer to `.planning/codebase` for more info:
`ARCHITECTURE.md` `CONCERNS.md` `CONVENTIONS.md` `INTEGRATIONS.md` `STACK.md` `STRUCTURE.md` `TESTING.md`
