# Feature Landscape

**Domain:** Personal AI assistant (locally-run, open-source, OpenClaw-inspired)
**Researched:** 2026-03-11

## Table Stakes

Features users expect from a personal AI assistant in 2026. Missing any of these and the product feels incomplete or broken compared to OpenClaw and peers.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **Chat interface with streaming** | Every AI product streams tokens in real-time. Users will not wait for a full response. | Medium | Use SSE or WebSocket from Go backend to Next.js. OpenAI-compatible APIs all support streaming. |
| **Multi-turn conversation** | Users expect context carries across messages within a session. The agent must reference earlier turns. | Low | Maintain conversation history in memory and send as context to the LLM. |
| **Conversation persistence** | Users expect to close the browser, come back, and resume or review past conversations. | Medium | Store conversations in SQLite or on disk. OpenClaw uses timestamped Markdown files. |
| **Conversation list/history** | Users need to browse, search, and resume past conversations from the UI. | Medium | Standard sidebar pattern (ChatGPT-style). Web UI advantage over OpenClaw's messaging-app approach. |
| **File read/write** | Core agent capability. The assistant must read and write files on the local filesystem. | Medium | Go backend exposes file operations as tools. Needs path sandboxing for safety. |
| **Shell command execution** | Users expect the agent to run commands, install packages, run scripts. OpenClaw and Open Interpreter both do this. | Medium | Go backend executes commands and streams stdout/stderr back. Needs timeout and kill controls. |
| **Web search** | The agent must be able to search the web for current information. | Medium | Integrate a search provider (SearXNG for self-hosted, or Brave/Google API). OpenClaw supports configurable providers. |
| **Web page fetching/reading** | Agent must fetch and read web pages, converting HTML to readable text. | Medium | HTTP GET + HTML-to-markdown extraction. OpenClaw calls this `web_fetch`. |
| **Model provider flexibility** | Must support any OpenAI-compatible API: OpenAI, Anthropic (via proxy), Ollama (local), OpenRouter, etc. | Medium | Abstract the LLM client behind a standard OpenAI chat completions interface. Go has good libraries for this. |
| **System prompt / agent personality** | Users expect to customize the agent's behavior, tone, and instructions. | Low | Editable system prompt stored in config. |
| **Markdown rendering in chat** | AI responses contain code blocks, lists, tables. Must render properly. | Low | Standard in every chat UI. Use react-markdown or similar. |
| **Code syntax highlighting** | Developers (primary audience) expect highlighted code blocks. | Low | Use Shiki or Prism.js in the frontend. |
| **Tool call visibility** | Users need to see what the agent is doing -- which tools it called, what arguments, what results. | Medium | Show expandable tool call blocks in the chat UI. Critical for trust and debugging. |
| **Configuration UI** | Users need a settings page to configure model provider, API keys, and behavior without editing config files. | Medium | Web UI advantage. OpenClaw requires YAML config editing. |
| **Local-first / privacy by default** | Data stays on user's machine. No telemetry, no cloud dependency. | Low | Architectural constraint, not a feature to build. Just don't add telemetry. |

## Differentiators

Features that set open-jarvis apart. Not expected by default, but create real value -- especially the web UI features that OpenClaw lacks entirely.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Web dashboard (the UI itself)** | OpenClaw has NO web UI -- it uses messaging apps (WhatsApp, Discord, etc.). A purpose-built dashboard is open-jarvis's primary differentiator. | High | This is the product. Everything below enhances it. |
| **Visual tool execution feedback** | Show file diffs, command output, search results, and fetched pages as rich UI cards -- not just text dumps. | High | Where messaging apps are limited to plain text, a web UI can render diffs, tables, images, terminal output with color. |
| **Skill/plugin management UI** | Browse, install, enable/disable skills from the web dashboard. OpenClaw requires CLI or config file editing. | High | OpenClaw has ClawHub registry but no GUI for it. Big UX win. |
| **Agent workspace browser** | Visual file browser showing the agent's workspace, memory files, and logs. | Medium | OpenClaw's workspace is opaque files on disk. A web UI can make them explorable. |
| **Memory inspection and editing** | View what the agent "remembers" (memory files, daily logs) and edit/delete memories through the UI. | Medium | OpenClaw memory is Markdown files. A web UI can present them as a searchable, editable knowledge base. |
| **Persistent memory (semantic search)** | Hybrid vector + BM25 search across past conversations and notes, so the agent recalls relevant context. | High | OpenClaw uses sqlite-vec + FTS5. Open-jarvis should implement similar. This is table stakes for OpenClaw users but differentiating vs simpler assistants. |
| **Conversation branching/forking** | Branch a conversation to explore alternatives without losing the original thread. | Medium | Web UI exclusive. Impossible in messaging apps. |
| **Multi-model switching mid-conversation** | Switch between models (e.g., use GPT-4o for reasoning, a local model for private tasks) within the same session. | Medium | Dropdown in chat UI. OpenClaw supports multiple models but switching is config-based. |
| **Real-time resource monitoring** | Show token usage, cost estimation, model latency, and memory usage in the dashboard. | Medium | Developers care about cost. OpenClaw has no visibility into this. |
| **Proactive/scheduled tasks (cron)** | Agent can run tasks on a schedule (morning briefing, inbox summary, monitoring). | High | OpenClaw's cron system is mature. For open-jarvis, start simple with cron expressions triggering agent prompts. |
| **MCP (Model Context Protocol) support** | Connect to any MCP server to gain tools (GitHub, databases, APIs) without custom code. | High | MCP is the 2026 standard for tool integration. Supporting it makes open-jarvis instantly compatible with thousands of tools. |
| **Multi-agent delegation** | Spawn sub-agents for parallel tasks (research + code + review simultaneously). | Very High | OpenClaw supports this with agent routing. Defer to later phase -- complex orchestration. |
| **Browser automation (CDP)** | Full browser control via Chrome DevTools Protocol for sites requiring JavaScript, auth, or interaction. | High | OpenClaw uses Chromium automation. Valuable but complex. Defer to later phase. |
| **Dark/light theme** | Developer-expected UI polish. | Low | Easy win. Use CSS variables or Tailwind dark mode. |
| **Keyboard shortcuts** | Power-user efficiency (Cmd+K for new chat, Cmd+Enter to send, etc.). | Low | Easy win for developer audience. |
| **Export conversations** | Export chats as Markdown, JSON, or PDF. | Low | Simple feature, good for archival. |

## Anti-Features

Features to explicitly NOT build. Either out of scope, harmful to the product vision, or a common trap.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| **Messaging app integrations (Telegram, Discord, WhatsApp)** | OpenClaw's entire UX is built around messaging apps. Competing there means competing on their turf. The web UI IS the differentiator. | Web UI first. If messaging adapters are added later, treat them as notification channels, not primary interfaces. |
| **Multi-user / SaaS mode** | Adds auth, permissions, billing, tenant isolation. Massive complexity for a personal assistant. | Single-user, local-first. If needed later, add basic auth (one user + password). |
| **Mobile app** | Native mobile development splits focus. The web UI can be made responsive. | Make the Next.js dashboard responsive/PWA-capable. No native app. |
| **Custom model training/fine-tuning** | Users should bring their own model via API. Training is a different product. | Support any OpenAI-compatible API. Let users pick their model. |
| **Built-in RAG over large document corpora** | Full RAG systems (chunk, embed, index thousands of documents) are complex and a separate product category. | Support memory search for agent notes and conversation history. For document Q&A, integrate via MCP or a skill, not core. |
| **Voice interface** | Adds speech-to-text and text-to-speech complexity. Nice-to-have, not core. | Text-first. Voice can be added as a later enhancement. |
| **Image generation** | Scope creep. The agent can call image APIs if needed, but don't build a gallery/editor. | If needed, expose as a tool that returns an image URL. Display in chat. |
| **Autonomous agent loops without approval** | Letting the agent run unbounded tool chains is dangerous (deleting files, running destructive commands). | Implement approval gates: agent proposes actions, user confirms. Allow "auto-approve" for trusted tool categories. |
| **Electron/desktop wrapper** | Adds packaging and update complexity. The Go backend already runs locally. | Users access via browser at localhost. Optionally provide a system tray icon for the Go process. |

## Feature Dependencies

```
Conversation persistence --> Conversation list/history
Conversation persistence --> Conversation branching
File read/write --> Agent workspace browser
File read/write --> Memory file storage
Shell command execution --> Tool call visibility
Web search --> Web page fetching (search results need to be readable)
Model provider flexibility --> Multi-model switching
Memory file storage --> Memory inspection UI
Memory file storage + Embeddings --> Persistent memory (semantic search)
Persistent memory --> Proactive/scheduled tasks (need context recall)
Tool call visibility --> Visual tool execution feedback (rich rendering)
Streaming chat --> Multi-turn conversation
Configuration UI --> Model provider flexibility (configure providers in UI)
MCP support --> Skill/plugin management UI (MCP servers as "skills")
```

## MVP Recommendation

Build in this order to deliver a usable product fastest:

### Phase 1: Core Chat Loop
1. **Chat interface with streaming** -- the foundation
2. **Multi-turn conversation** -- basic usability
3. **Model provider flexibility** -- connect to any LLM
4. **Markdown rendering + code highlighting** -- polish the output
5. **System prompt configuration** -- basic customization
6. **Dark/light theme** -- quick visual polish

### Phase 2: Agent Capabilities
1. **File read/write** -- first real tool
2. **Shell command execution** -- second real tool
3. **Tool call visibility** -- users must see what the agent does
4. **Web search + web fetch** -- third tool category
5. **Conversation persistence** -- save and resume

### Phase 3: Dashboard Differentiation
1. **Conversation list/history** -- browse past sessions
2. **Configuration UI** -- manage settings in browser
3. **Agent workspace browser** -- see agent's files
4. **Memory inspection and editing** -- view/edit what agent remembers
5. **Real-time resource monitoring** -- token/cost tracking

### Phase 4: Advanced Intelligence
1. **Persistent memory (semantic search)** -- long-term recall
2. **MCP support** -- extensible tool ecosystem
3. **Proactive/scheduled tasks** -- cron-based automation
4. **Conversation branching** -- explore alternatives
5. **Multi-model switching** -- per-task model selection

### Defer Indefinitely
- Multi-agent delegation (very high complexity, niche use case for personal assistant)
- Browser automation (high complexity, can be added via MCP/skill later)
- Messaging app integrations (contradicts core differentiator)

## Sources

- [OpenClaw Official Docs - Skills](https://docs.openclaw.ai/tools/skills)
- [OpenClaw Official Docs - Memory](https://docs.openclaw.ai/concepts/memory)
- [OpenClaw Official Docs - Web Tools](https://docs.openclaw.ai/tools/web)
- [OpenClaw Official Docs - Cron Jobs](https://docs.openclaw.ai/automation/cron-jobs)
- [OpenClaw Official Docs - Multi-Agent](https://docs.openclaw.ai/concepts/multi-agent)
- [What is OpenClaw - DigitalOcean](https://www.digitalocean.com/resources/articles/what-is-openclaw)
- [OpenClaw Review 2026 - CyberNews](https://cybernews.com/ai-tools/openclaw-review/)
- [OpenClaw Memory System Deep Dive](https://snowan.gitbook.io/study-notes/ai-blogs/openclaw-memory-system-deep-dive)
- [Memsearch - Extracted OpenClaw Memory (Milvus)](https://milvus.io/blog/we-extracted-openclaws-memory-system-and-opensourced-it-memsearch.md)
- [Open Interpreter](https://www.openinterpreter.com/)
- [Model Context Protocol](https://modelcontextprotocol.io/)
- [MCP Complete Guide 2026 - Calmops](https://calmops.com/ai/model-context-protocol-mcp-2026-complete-guide/)
- [Assistant UI React Library](https://www.assistant-ui.com/)
- [Tool Calling Guide 2026 - Composio](https://composio.dev/blog/ai-agent-tool-calling-guide)
