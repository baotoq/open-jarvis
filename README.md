# open-jarvis

A personal AI assistant with a streaming chat backend, inspired by OpenClaw. Built with Go and Next.js (frontend coming soon).

## Architecture

- **Backend** (`src/`) — Go service using [go-zero](https://go-zero.dev). Handles HTTP, streaming SSE responses, and AI model integration via OpenAI-compatible API.
- **Frontend** — Next.js/TypeScript (not yet implemented).

The two services are decoupled and communicate over HTTP.

## Quick Start

### Requirements

- Go 1.26+
- An OpenAI-compatible model server (e.g. [Ollama](https://ollama.com))

### Run

```bash
# Start Ollama (or any OpenAI-compatible server)
ollama serve

cd src
go run ./cmd/main.go
```

The server starts on `http://localhost:8888`.

### Configuration

Edit `src/etc/config.yaml`:

```yaml
Name: open-jarvis
Host: 0.0.0.0
Port: 8888
Model:
  BaseURL: http://localhost:11434/v1   # OpenAI-compatible endpoint
  Name: llama3.2                       # Model name
  APIKey: ""                           # Required for hosted APIs (OpenAI, Anthropic, etc.)
  SystemPrompt: "You are Jarvis, a personal AI assistant. Be concise and helpful."
MaxToolCalls: 10
TurnTimeoutSeconds: 60
```

## API

### `POST /api/chat/stream`

Streams a chat response as [Server-Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events).

**Request:**
```json
{ "sessionId": "abc123", "message": "Hello!" }
```

**Response** (SSE stream):
```
data: Hello

data: !

data: How can I help?

data: [DONE]
```

Each session maintains conversation history server-side, identified by `sessionId`.

## Development

### Prerequisites

- [Go](https://golang.org) 1.21+
- [Node.js](https://nodejs.org) 18+
- [Tilt](https://tilt.dev) — `brew install tilt` (macOS) or see https://docs.tilt.dev/install.html
- [air](https://github.com/air-verse/air) — `go install github.com/air-verse/air@latest`

### Start the dev environment

```bash
tilt up
```

This starts:
- Go backend at http://localhost:8888 with live reload (via air)
- Next.js frontend at http://localhost:3000 with hot reload

The Tilt dashboard is at http://localhost:10350.

On macOS, the browser opens automatically to http://localhost:3000. On Linux, visit the URL manually or change `open` to `xdg-open` in the Tiltfile.

### Stop

```bash
tilt down
```

Data in `src/backend/data/` and `src/backend/etc/config.yaml` are preserved.

### Manual (without Tilt)

```bash
# Backend
cd src/backend && air        # or: go run ./cmd/main.go

# Frontend
cd src/frontend && npm run dev
```

### Tests and static analysis

```bash
cd src/backend
go test ./...          # run tests
go test -cover ./...   # with coverage
go vet ./...           # static analysis
golangci-lint run ./...  # lint
go mod tidy            # clean dependencies
```

## Project Structure

```
src/
├── cmd/main.go              # entry point
├── etc/config.yaml          # runtime config
└── internal/
    ├── config/              # Config struct with defaults
    ├── handler/             # HTTP handlers
    ├── logic/               # Business logic (streaming chat loop)
    ├── svc/                 # ServiceContext, ConvStore, AI client
    └── types/               # Request/response types
```
