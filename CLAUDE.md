# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

open-jarvis is a personal AI assistant inspired by OpenClaw, built as separate Go (backend) and TypeScript (frontend) services.

## Architecture

- **Backend**: Go service using the [go-zero](https://go-zero.dev) framework — handles API, logic, and AI integrations
- **Frontend**: Next.js (TypeScript) application — the user-facing UI (`src/frontend/`)

The two services are **decoupled** and communicate over HTTP/RPC. Changes to one do not require changes to the other unless the API contract changes.

## Commands

### Go (Backend)
> Run all Go commands from the `src/backend/` directory.

```bash
cd src/backend
go build ./...                     # build
go run ./cmd/main.go               # run server (reads etc/config.yaml)
go test ./...                      # test all
go test -v -run TestName ./pkg/    # run specific test
go test -cover ./...               # with coverage
go vet ./...                       # static analysis
go mod tidy                        # clean dependencies
```

### TypeScript / Next.js (Frontend)
> Run all frontend commands from the `src/frontend/` directory.

```bash
cd src/frontend
npm run dev      # dev server (http://localhost:3000)
npm run build    # production build
npm run lint     # lint
```

## Conventions

- Go: `gofmt` style, explicit error returns, go-zero Handler→Logic→ServiceContext pattern
- Tests: use `github.com/stretchr/testify/assert` and `require` (not raw `t.Errorf`)
- TypeScript: strict mode, Next.js App Router, PascalCase components, Tailwind v4 CSS-first (`@import "tailwindcss"` in globals.css, no tailwind.config.ts)
- Package manager for TypeScript: `npm`

## Structure

```
src/
├── backend/                    # Go service (go-zero)
│   ├── cmd/main.go             # entry point
│   ├── etc/config.yaml         # runtime config (model URL, system prompt, timeouts)
│   ├── internal/
│   │   ├── config/             # Config struct with defaults
│   │   ├── handler/            # HTTP handlers (parse request, call logic)
│   │   ├── logic/              # Business logic (StreamChat SSE loop)
│   │   ├── svc/                # ServiceContext, ConversationStore interface, SQLiteConvStore
│   │   └── types/              # Shared request/response types
│   └── go.mod
└── frontend/                   # Next.js 15 frontend (npm)
    ├── app/                    # App Router pages and layouts
    ├── components/             # Sidebar, ChatArea, shadcn/ui
    ├── hooks/                  # useSession (localStorage session management)
    └── lib/                    # api.ts (typed backend wrappers), utils.ts
```

## Gotchas

- SSE streaming: handlers write `data: <token>\n\n` directly to `http.ResponseWriter`; no framework abstraction
- go-zero layer rule: handlers must not contain logic; logic must not import handler types
- Config defaults are set via struct tags (`default:"value"`), not code

## Codebase Docs

Refer to `.planning/codebase` for more info:
`ARCHITECTURE.md` `CONCERNS.md` `CONVENTIONS.md` `INTEGRATIONS.md` `STACK.md` `STRUCTURE.md` `TESTING.md`
