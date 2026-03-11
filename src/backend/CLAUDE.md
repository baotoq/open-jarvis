# CLAUDE.md — Backend

Go service for open-jarvis. Built with [go-zero](https://go-zero.dev).

## Commands

> Run all commands from `src/backend/`.

```bash
go build ./...                          # build
go run ./cmd/main.go                    # run server (http://localhost:8888)
go test ./...                           # test all
go test -v -run TestName ./internal/... # run specific test
go test -cover ./...                    # with coverage
go vet ./...                            # static analysis
go mod tidy                             # clean dependencies
```

## Architecture

**go-zero Handler → Logic → ServiceContext** — strict layer rule:
- `handler/` — parse HTTP request, call logic, write response. No business logic.
- `logic/` — all business logic. No handler imports.
- `svc/` — shared dependencies (config, AI client, conversation store).
- `types/` — shared request/response structs.
- `config/` — Config struct; defaults via go-zero struct tags.

## Key Patterns

**Error handling:** Return errors, don't panic. Wrap with `%w` for caller inspection, `%v` to hide internals.

**Interfaces belong to consumers:** Define interfaces in the package that uses them (`svc.ConversationStore` is used by logic, defined in svc).

**Receivers:** Single-letter abbreviation, consistent across all methods. Never `this` or `self`.

**Context first:** `ctx context.Context` is always the first parameter.

**Tests:** Table-driven with `testify/assert` and `require`. Never raw `t.Errorf`. Use `t.TempDir()` for temp files.

## Conventions

- `gofmt` formatting (tabs, not spaces)
- Explicit error returns everywhere
- Initialisms: `sessionID` not `sessionId`, `apiURL` not `apiUrl`
- Short variable names in local scope (`r`, `w`, `ctx`); descriptive for struct fields
- No `init()` functions — explicit initialization in `NewServiceContext`

## SSE Streaming

Handlers write `data: <token>\n\n` directly to `http.ResponseWriter`. No framework abstraction. Routes must use `rest.WithSSE()` to disable go-zero's default timeout middleware.

Final event format: `data: {"done":true,"sessionId":"<uuid>"}\n\n`

## Config

go-zero YAML config at `etc/config.yaml`. Defaults via struct tags:

```go
type ModelConfig struct {
    BaseURL      string `json:",default=http://localhost:11434/v1"`
    Name         string `json:",default=llama3.2"`
    APIKey       string `json:",optional"`
    SystemPrompt string `json:",optional"`
}
```

Default target: local Ollama. Zero config required if Ollama is running.

## SQLite

Store at `data/conversations.db` (gitignored). WAL mode enabled via DSN pragma. `ConversationStore` interface allows in-memory swap for tests — pass a `*ConvStore` to `NewServiceContextWithClient`.

## Gotchas

- `DefaultSystemPrompt` is a `const`, not a struct tag default (go vet rejects struct tag defaults with spaces)
- `rest.WithSSE()` required on SSE route or go-zero timeout middleware kills the stream
- Handler layer must not import logic types; logic layer must not import handler types
