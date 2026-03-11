# Roadmap: open-jarvis

## Overview

open-jarvis delivers a locally-run AI assistant in five phases: first a working streaming chat loop with safety guardrails, then conversation persistence, then local tools (file/shell) with approval gates, then web tools with audit logging, and finally model configuration and conversation search. Each phase delivers a complete, testable capability that builds on the previous one.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Streaming Chat Loop** - End-to-end streaming chat with multi-turn context and agent loop guardrails
- [ ] **Phase 2: Conversation Persistence** - SQLite storage for conversations with history sidebar
- [ ] **Phase 3: File and Shell Tools** - Local file and shell tools with approval gates and inline tool display
- [ ] **Phase 4: Web Tools and Audit** - Web search, page fetching, and complete audit logging
- [ ] **Phase 5: Configuration and Search** - Multi-provider model config UI, settings page, and full-text conversation search

## Phase Details

### Phase 1: Streaming Chat Loop
**Goal**: User can have a real-time streaming conversation with any OpenAI-compatible model through the web dashboard
**Depends on**: Nothing (first phase)
**Requirements**: CHAT-01, CHAT-02, SAFE-03
**Success Criteria** (what must be TRUE):
  1. User can type a message and see the response appear token-by-token in real time
  2. User can send follow-up messages and the agent responds with awareness of the full conversation so far
  3. Agent loop terminates automatically when it exceeds a configurable max tool calls or timeout limit
  4. Go backend starts as a single binary and serves both the API and the Next.js frontend assets
**Plans**: 3 plans

Plans:
- [ ] 01-01-PLAN.md — Go backend: config, conversation store, SSE handler, streaming chat logic (CHAT-01, CHAT-02, SAFE-03)
- [ ] 01-02-PLAN.md — Next.js frontend: scaffold, chat components, SSE client, session management (CHAT-01)
- [ ] 01-03-PLAN.md — Integration: embed frontend into Go binary, end-to-end browser verification (CHAT-01, CHAT-02, SAFE-03)

### Phase 2: Conversation Persistence
**Goal**: Conversations survive restarts and users can browse and resume past conversations
**Depends on**: Phase 1
**Requirements**: CHAT-03, UI-02
**Success Criteria** (what must be TRUE):
  1. User can close the browser, reopen it, and see their previous conversation intact
  2. User can see a sidebar listing all past conversations and click one to load it
  3. Starting a new conversation creates a fresh session without losing previous ones
**Plans**: 5 plans

Plans:
- [ ] 02-01-PLAN.md — Backend data layer: ConversationStore interface, SQLiteConvStore, config DBPath, servicecontext wiring (CHAT-03)
- [ ] 02-02-PLAN.md — Frontend scaffold: Next.js 15 app, Tailwind v4, shadcn/ui, cn() utility, API_BASE (UI-02)
- [ ] 02-03-PLAN.md — Backend API endpoints: list/get/delete conversations, SSE done event with session ID, UUID assignment in chatlogic (CHAT-03, UI-02)
- [ ] 02-04-PLAN.md — Frontend UI: useSession hook, Sidebar, ChatArea, page layout (UI-02, CHAT-03)
- [ ] 02-05-PLAN.md — Verification: human browser test — stream, persist, sidebar, reload, delete (CHAT-03, UI-02)

### Phase 3: File and Shell Tools
**Goal**: Agent can read/write files and run shell commands with user approval, and tool actions are visible in the chat
**Depends on**: Phase 2
**Requirements**: TOOL-01, TOOL-02, SAFE-01, SAFE-02, UI-01
**Success Criteria** (what must be TRUE):
  1. User can ask the agent to read a file and see its contents in the conversation
  2. User can ask the agent to create or modify a file and verify the change on disk
  3. User can ask the agent to run a shell command and see the output inline in the chat
  4. Agent prompts user for confirmation before executing shell commands flagged by the allowlist/denylist
  5. Tool calls and their results appear as distinct, inspectable blocks in the chat UI alongside messages
**Plans**: 5 plans

Plans:
- [ ] 03-01-PLAN.md — Tool executor package: Executor interface, FileTool, ShellTool, unit tests (TOOL-01, TOOL-02, SAFE-01)
- [ ] 03-02-PLAN.md — Config + ServiceContext: ShellAllowlist/Denylist/WorkspaceRoot, ApprovalStore, wiring (SAFE-01, SAFE-02)
- [ ] 03-03-PLAN.md — Agentic loop: ChatLogic tool dispatch, approval gate, SSE events, approve handler (TOOL-01, TOOL-02, SAFE-02)
- [ ] 03-04-PLAN.md — Frontend: MessagePart types, ToolCallBlock, ApprovalDialog, extended ChatArea SSE parser (UI-01)
- [ ] 03-05-PLAN.md — Verification: human browser test — file read/write, shell approval, tool blocks in UI (TOOL-01, TOOL-02, SAFE-01, SAFE-02, UI-01)

### Phase 4: Web Tools and Audit
**Goal**: Agent can search the web and fetch pages, and all tool executions are recorded for inspection
**Depends on**: Phase 3
**Requirements**: TOOL-03, TOOL-04, SAFE-04
**Success Criteria** (what must be TRUE):
  1. User can ask the agent to search the web and receive summarized results from multiple sources
  2. User can ask the agent to fetch and summarize a specific web page
  3. Every tool execution (file, shell, web) is recorded in an audit log with timestamp, tool name, parameters, and result
**Plans**: 4 plans

Plans:
- [ ] 04-01-PLAN.md — Web tools: go-readability dep, WebFetchTool, WebSearchTool, config extensions (TOOL-03, TOOL-04)
- [ ] 04-02-PLAN.md — AuditStore: tool_audit_log SQLite table, migrate(), Log() method, unit tests (SAFE-04)
- [ ] 04-03-PLAN.md — Integration: wire web tools + AuditStore into ServiceContext and chatlogic agentic loop (TOOL-03, TOOL-04, SAFE-04)
- [ ] 04-04-PLAN.md — Verification: human browser test — web fetch, web search, audit log inspection (TOOL-03, TOOL-04, SAFE-04)

### Phase 5: Configuration and Search
**Goal**: User can configure model providers through the UI and search across all past conversations
**Depends on**: Phase 4
**Requirements**: CHAT-04, UI-03, MEM-01
**Success Criteria** (what must be TRUE):
  1. User can add, edit, and switch between model providers (OpenAI, Ollama, Anthropic) from a settings page without editing config files
  2. User can configure API keys, model names, and preferences through the settings UI
  3. User can search across all past conversations by keyword and jump to matching results
**Plans**: TBD

Plans:
- [ ] 05-01: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3 -> 4 -> 5

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Streaming Chat Loop | 1/3 | In Progress|  |
| 2. Conversation Persistence | 2/5 | In Progress|  |
| 3. File and Shell Tools | 4/5 | In Progress|  |
| 4. Web Tools and Audit | 3/4 | In Progress|  |
| 5. Configuration and Search | 0/1 | Not started | - |
