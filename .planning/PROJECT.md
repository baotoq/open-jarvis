# open-jarvis

## What This Is

open-jarvis is a locally-run, open-source personal AI assistant inspired by OpenClaw. It connects any OpenAI-compatible language model to your files, shell, and the web through a custom Next.js dashboard — with a Go backend built for performance and low memory usage. Designed to run on your local machine or a self-hosted server.

## Core Value

A fast, private, general-purpose AI agent that knows your context, automates tasks, and is actually yours to own and extend.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] User can chat with an AI agent through a web UI
- [ ] Agent can read and write local files
- [ ] Agent can run shell commands
- [ ] Agent can browse and summarize web pages
- [ ] Agent can search the web
- [ ] Any OpenAI-compatible model can be used (OpenAI, Ollama, Anthropic, etc.)
- [ ] Conversation memory persists across sessions
- [ ] Runs locally on developer's machine and on a self-hosted server

### Out of Scope

- Messaging app adapters (Telegram, Discord, etc.) — web UI first; add later
- Mobile app — web-first
- Multi-user / SaaS — single-user personal assistant for now

## Context

- Inspired by OpenClaw (GitHub: openclaw/openclaw), which uses Node/Python and messaging apps as UI
- open-jarvis differentiates with: Go backend (lower memory, faster response), custom web dashboard (OpenClaw has none), and a simpler deployment model
- OpenClaw uses hybrid vector + FTS5 memory — memory approach TBD, to be informed by research
- Tech stack decided: Go (go-zero) backend, TypeScript (Next.js) frontend, npm

## Constraints

- **Tech stack**: Go (go-zero) backend, Next.js frontend — already decided
- **Privacy**: Data stays local by default; no telemetry
- **Model**: Must support any OpenAI-compatible API, including local models (Ollama)

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go backend | Performance — lower memory, faster than OpenClaw's Node/Python | — Pending |
| Next.js web UI | Custom dashboard OpenClaw lacks; full control over UX | — Pending |
| Web UI before messaging adapters | Validate core agent loop before adding channel complexity | — Pending |
| Any OpenAI-compatible API | Model-agnostic like OpenClaw; supports local + cloud | — Pending |

---
*Last updated: 2026-03-11 after initialization*
