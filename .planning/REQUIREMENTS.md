# Requirements: open-jarvis

**Defined:** 2026-03-11
**Core Value:** A fast, private, general-purpose AI agent that knows your context, automates tasks, and is actually yours to own and extend.

## v1 Requirements

### Chat

- [x] **CHAT-01**: User can send a message and receive a streaming token-by-token response
- [x] **CHAT-02**: Agent maintains multi-turn conversation context within a session
- [x] **CHAT-03**: Conversations are persisted to SQLite and survive restarts
- [ ] **CHAT-04**: User can configure and switch between OpenAI-compatible model providers (OpenAI, Ollama, Anthropic)

### Agent Tools

- [x] **TOOL-01**: Agent can read and write local files on user's machine
- [x] **TOOL-02**: Agent can execute shell commands (subject to safety controls)
- [ ] **TOOL-03**: Agent can fetch and summarize web pages
- [ ] **TOOL-04**: Agent can search the web and return results

### Safety

- [x] **SAFE-01**: User can configure a command allowlist/denylist for shell tool
- [x] **SAFE-02**: Agent prompts user for approval before executing destructive actions
- [x] **SAFE-03**: Agent loop is bounded by configurable max tool calls and timeout per turn
- [ ] **SAFE-04**: All tool executions are recorded in an audit log

### Web Dashboard

- [ ] **UI-01**: Chat interface shows tool calls and their results inline alongside messages
- [x] **UI-02**: Conversation history sidebar lets user browse and switch past conversations
- [ ] **UI-03**: Settings UI lets user configure model provider, API keys, and preferences

### Memory

- [ ] **MEM-01**: User can search past conversations via full-text keyword search (SQLite FTS5)

## v2 Requirements

### Memory

- **MEM-02**: Semantic memory search via vector embeddings (sqlite-vec)
- **MEM-03**: Automatic context summarization when conversation history exceeds token limit

### Dashboard

- **UI-04**: Workspace file browser in dashboard
- **UI-05**: Memory inspection and editing UI
- **UI-06**: Resource/usage monitoring panel

### Extensibility

- **EXT-01**: MCP (Model Context Protocol) client support for third-party tool integrations
- **EXT-02**: Skill/plugin management UI

### Messaging

- **MSG-01**: Telegram adapter for chat-based interaction
- **MSG-02**: Additional messaging adapter (Discord, WhatsApp, or Signal)

### Advanced Agent

- **AGT-01**: Cron-based scheduled tasks
- **AGT-02**: Multi-agent orchestration (agent spawning sub-agents)

## Out of Scope

| Feature | Reason |
|---------|--------|
| Multi-user / SaaS | Single-user personal assistant by design |
| Mobile app | Web-first; mobile later |
| Voice interaction | Out of scope for v1; high complexity |
| Autonomous loops without approval gates | Safety risk — all tool chains require human-in-the-loop |
| RAG over external document collections | Deferred; FTS5 covers personal assistant needs for v1 |
| Built-in LLM hosting | Users bring their own model (cloud or Ollama) |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| CHAT-01 | Phase 1 | Complete |
| CHAT-02 | Phase 1 | Complete |
| CHAT-03 | Phase 2 | Complete |
| CHAT-04 | Phase 5 | Pending |
| TOOL-01 | Phase 3 | Complete |
| TOOL-02 | Phase 3 | Complete |
| TOOL-03 | Phase 4 | Pending |
| TOOL-04 | Phase 4 | Pending |
| SAFE-01 | Phase 3 | Complete |
| SAFE-02 | Phase 3 | Complete |
| SAFE-03 | Phase 1 | Complete |
| SAFE-04 | Phase 4 | Pending |
| UI-01 | Phase 3 | Pending |
| UI-02 | Phase 2 | Complete |
| UI-03 | Phase 5 | Pending |
| MEM-01 | Phase 5 | Pending |

**Coverage:**
- v1 requirements: 14 total
- Mapped to phases: 14
- Unmapped: 0

---
*Requirements defined: 2026-03-11*
*Last updated: 2026-03-11 after roadmap creation*
