# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

open-jarvis is a personal AI assistant inspired by OpenClaw, built as separate Go (backend) and TypeScript (frontend) services.

## Architecture

- **Backend**: Go service using the [go-zero](https://go-zero.dev) framework — handles API, logic, and AI integrations
- **Frontend**: Next.js (TypeScript) application — the user-facing UI

The two services are **decoupled** and communicate over HTTP/RPC. Changes to one do not require changes to the other unless the API contract changes.

## Commands

### Go (Backend)
```bash
go build ./...        # build
go run ./cmd/...      # run
go test ./...         # test all
go test ./path/...    # test a specific package
go vet ./...          # static analysis
go mod tidy           # clean dependencies
```

### TypeScript / Next.js (Frontend)
```bash
npm install           # install dependencies
npm run dev           # dev server
npm run build         # production build
npm run lint          # lint
```

## Conventions

- Go: standard `gofmt` style, explicit error returns, go-zero patterns for API/RPC definitions
- TypeScript: strict mode enabled, Next.js App Router conventions, PascalCase for components/types
- Package manager for TypeScript: `npm`

## Codebase Docs
- refer to `.planning/codebase` for more infor
  `ARCHITECTURE.md`
  `CONCERNS.md`
  `CONVENTIONS.md`
  `INTEGRATIONS.md`
  `STACK.md`
  `STRUCTURE.md`
  `TESTING.md`
