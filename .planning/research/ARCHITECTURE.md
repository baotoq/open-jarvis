# Architecture Patterns

**Domain:** Personal AI assistant (local-first, agent-based)
**Researched:** 2026-03-11

## Recommended Architecture

Open-jarvis follows a **Gateway + Agent Runtime** pattern, drawing from OpenClaw's proven architecture but simplified for single-user, self-hosted deployment. The system is a long-running Go backend (the Gateway) that manages an Agent Loop, tool execution, memory, and exposes APIs consumed by a Next.js dashboard.

```
+-------------------+        HTTP/SSE         +---------------------------+
|                   | <---------------------> |                           |
|   Next.js         |                         |   Go Backend (Gateway)    |
|   Dashboard       |   REST + SSE stream     |                           |
|                   |                         |  +---------------------+  |
+-------------------+                         |  |   Agent Runtime     |  |
                                              |  |                     |  |
                                              |  |  System Prompt      |  |
                                              |  |  Context Assembly   |  |
                                              |  |  Tool Orchestration |  |
                                              |  |  Response Streaming |  |
                                              |  +---------------------+  |
                                              |           |               |
                                              |     +-----+------+       |
                                              |     |            |       |
                                              |  +--v--+   +----v----+  |
                                              |  |Tools|   | Memory  |  |
                                              |  +-----+   +---------+  |
                                              |     |            |       |
                                              +-----|------------|-------+
                                                    |            |
                                              +-----v---+  +----v------+
                                              | Shell,   |  | SQLite    |
                                              | Files,   |  | + sqlite- |
                                              | Web      |  |   vec     |
                                              +---------+  +-----------+
                                                    |
                                              +-----v-----------+
                                              | OpenAI-compat   |
                                              | LLM Provider    |
                                              | (OpenAI/Ollama) |
                                              +-----------------+
```

### Component Boundaries

| Component | Responsibility | Communicates With |
|-----------|---------------|-------------------|
| **Next.js Dashboard** | Chat UI, conversation list, settings, streaming display | Go Backend (HTTP REST + SSE) |
| **Go Backend (Gateway)** | HTTP server, request routing, session management, config | Dashboard, Agent Runtime, Memory Store |
| **Agent Runtime** | Agent loop (prompt assembly, LLM call, tool dispatch, iteration) | LLM Provider, Tool Registry, Memory |
| **Tool Registry** | Discovers, registers, and dispatches tool calls to implementations | Agent Runtime, individual Tool implementations |
| **Tool: Shell** | Executes shell commands on the host machine | Host OS |
| **Tool: File System** | Reads/writes files within allowed paths | Host file system |
| **Tool: Web Fetch** | Fetches and summarizes web pages | Internet |
| **Tool: Web Search** | Searches the web via search API | Search API (SearXNG, Brave, etc.) |
| **Memory Store** | Persists conversations, retrieves relevant context | SQLite + sqlite-vec |
| **LLM Provider Adapter** | Normalizes OpenAI-compatible API calls, handles streaming | Any OpenAI-compatible endpoint |

### Data Flow

**Chat Message Flow (the core loop):**

```
1. User types message in Dashboard
2. Dashboard sends POST /api/chat with message + conversation_id
3. Backend opens SSE stream back to Dashboard
4. Agent Runtime assembles context:
   a. System prompt (base personality + active tool descriptions)
   b. Conversation history (from Memory Store)
   c. Relevant memories (vector search from sqlite-vec)
   d. User message
5. Agent Runtime calls LLM Provider with assembled context
6. LLM responds with either:
   a. Text content --> stream tokens to Dashboard via SSE
   b. Tool call request --> execute tool, append result, goto step 5
7. When LLM returns final text (no more tool calls):
   a. Stream remaining tokens to Dashboard
   b. Persist conversation turn to Memory Store
   c. Close SSE stream
8. Dashboard renders complete response
```

**Key insight:** The agent loop is a **while loop with tool calls** -- the same pattern used by Claude Code, OpenAI Agents SDK, and OpenClaw. Each iteration: assemble context, call LLM, if tool call then execute and loop, if text then finish. This is the canonical agent architecture (HIGH confidence -- multiple authoritative sources agree).

**Streaming implementation:** Use go-zero's built-in SSE support (`rest.WithSSE()` in go-zero >1.8.1 or `rest.WithTimeout(0)` in <=1.8.1). SSE is simpler than WebSocket and sufficient because the primary data flow is server-to-client (token streaming). The Dashboard sends messages via regular HTTP POST; the backend streams responses back via SSE. No bidirectional WebSocket needed.

## Component Deep Dives

### Agent Runtime

The most critical component. It owns the agent loop and context assembly.

**Context Assembly** (assembled before each LLM call):
1. **System prompt** -- base instructions + dynamically generated tool catalog
2. **Conversation history** -- recent messages from the active conversation
3. **Retrieved memories** -- vector-similarity search for relevant past context
4. **User message** -- the current input

**Tool orchestration pattern:**
```
func (r *Runtime) Run(ctx context.Context, msg Message) <-chan StreamEvent {
    // 1. Assemble context
    // 2. Loop:
    //    a. Call LLM with context
    //    b. If response has tool_calls:
    //       - Execute each tool via ToolRegistry
    //       - Append tool results to context
    //       - Continue loop
    //    c. If response is text:
    //       - Stream tokens
    //       - Break loop
    // 3. Persist conversation
}
```

**Confidence:** HIGH -- this is the standard agent loop pattern documented across OpenClaw, LangChain, Google ADK, and Anthropic's agent frameworks.

### Tool Registry and Skill System

Tools are the extensibility point. Use a **registry pattern** where tools self-register with a name, description (for the LLM), and parameter schema.

**Architecture:**
```go
type Tool interface {
    Name() string
    Description() string        // Injected into system prompt
    Parameters() JSONSchema     // For LLM function calling
    Execute(ctx context.Context, params json.RawMessage) (ToolResult, error)
}

type Registry struct {
    tools map[string]Tool
}

func (r *Registry) Register(t Tool) { ... }
func (r *Registry) Catalog() []ToolDescription { ... }  // For system prompt
func (r *Registry) Execute(name string, params json.RawMessage) (ToolResult, error) { ... }
```

**Why a registry, not a plugin system:** For v1, built-in tools (shell, files, web fetch, web search) compiled into the binary is simpler and safer. A dynamic plugin system (like OpenClaw's skills) adds significant complexity (loading, sandboxing, versioning) that is not justified for a single-user assistant. The Tool interface allows adding new tools easily without architectural changes.

**Build toward skills later:** The Tool interface is the foundation. If you later want loadable skill packs, you add a SkillLoader that reads skill definitions from disk and registers them. The runtime does not change.

### Memory Store

**Two-tier memory architecture:**

| Tier | What | Storage | Access Pattern |
|------|------|---------|----------------|
| **Conversation history** | Full message log per conversation | SQLite table (conversations, messages) | Sequential read by conversation_id, paginated |
| **Semantic memory** | Embedded summaries of past interactions, user facts | SQLite + sqlite-vec | Vector similarity search on user query |

**Why SQLite + sqlite-vec:**
- Single file, zero ops, perfect for local-first (HIGH confidence -- sqlite-vec has Go bindings, OpenClaw uses similar SQLite-based approach)
- sqlite-vec provides vector search without running a separate vector DB
- chromem-go is an alternative (pure Go, no CGo) but sqlite-vec keeps everything in one database file
- Conversation history is relational (foreign keys, ordering) -- SQLite is ideal
- Vector search is for retrieval-augmented context -- sqlite-vec handles this in the same DB

**Schema sketch:**
```sql
-- Conversations
CREATE TABLE conversations (
    id TEXT PRIMARY KEY,
    title TEXT,
    created_at DATETIME,
    updated_at DATETIME
);

-- Messages (conversation history)
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    conversation_id TEXT REFERENCES conversations(id),
    role TEXT,        -- 'user', 'assistant', 'tool'
    content TEXT,
    tool_calls JSON,  -- If role='assistant' and tool calls made
    tool_call_id TEXT, -- If role='tool', references the call
    created_at DATETIME
);

-- Semantic memories (vector-searchable)
CREATE TABLE memories (
    id TEXT PRIMARY KEY,
    content TEXT,
    embedding BLOB,   -- sqlite-vec vector
    source TEXT,       -- 'conversation_summary', 'user_fact', etc.
    created_at DATETIME
);
```

### LLM Provider Adapter

**Pattern:** Adapter over the OpenAI chat completions API (`/v1/chat/completions`).

Since the project requirement is "any OpenAI-compatible API," the adapter targets the OpenAI API shape. Ollama, LM Studio, vLLM, and many others expose this same interface.

```go
type LLMProvider interface {
    ChatCompletion(ctx context.Context, req ChatRequest) (<-chan ChatChunk, error)
}

type OpenAIProvider struct {
    BaseURL string   // e.g., "https://api.openai.com/v1" or "http://localhost:11434/v1"
    APIKey  string
    Model   string
}
```

**Key details:**
- Always use streaming (`stream: true`) for responsive UX
- Handle tool_calls in the streaming response (partial JSON accumulation)
- The adapter is thin -- most logic lives in the Agent Runtime

### Next.js Dashboard

**Responsibilities:**
- Chat interface with streaming text display
- Conversation list/history sidebar
- Settings page (model config, API keys, allowed paths)
- Tool approval UI (optional -- for confirming dangerous operations)

**Communication with backend:**
- `POST /api/chat` -- send message, receive SSE stream
- `GET /api/conversations` -- list conversations
- `GET /api/conversations/:id/messages` -- load conversation history
- `POST /api/settings` -- update configuration
- `GET /api/models` -- list available models from provider

**No BFF pattern needed.** The Go backend IS the API server. Next.js runs as a static/SSR frontend that calls the Go backend directly. No Next.js API routes proxying to Go -- that adds unnecessary latency and complexity.

## Patterns to Follow

### Pattern 1: Streaming-First Agent Loop
**What:** Every LLM call streams tokens. The agent loop streams partial results to the client as they arrive, even during multi-tool-call iterations.
**When:** Always -- this is not optional for good UX.
**Why:** Users must see the assistant "thinking" in real time. Waiting for a full response (especially with tool calls that may take seconds) feels broken.

### Pattern 2: Context Window Budget Management
**What:** Before each LLM call, calculate token usage and trim/summarize older messages to stay within the model's context window.
**When:** Conversations exceed ~50% of context window capacity.
**Why:** Sending too many tokens causes API errors or degraded quality. Silently truncating without summarization loses important context.
**Implementation:** Count tokens (approximate with `len(text)/4` for English), keep recent N messages in full, summarize older messages into a condensed block.

### Pattern 3: Tool Safety Boundaries
**What:** Categorize tools by risk level. Low-risk tools (web fetch, file read) execute automatically. High-risk tools (shell execute, file write/delete) require explicit user confirmation via the Dashboard.
**When:** Any tool that modifies state or executes arbitrary code.
**Why:** An LLM hallucinating a destructive `rm -rf` command should not execute silently.

### Pattern 4: Configuration as Code with Runtime Override
**What:** Default configuration in a YAML/TOML file, overridable via environment variables, overridable via Dashboard settings UI.
**When:** Model selection, API keys, allowed file paths, tool permissions.
**Why:** Supports both "edit a config file" developers and "use the UI" users.

## Anti-Patterns to Avoid

### Anti-Pattern 1: Microservices for a Single-User App
**What:** Splitting the backend into multiple services (separate tool service, memory service, agent service).
**Why bad:** Adds deployment complexity, inter-service communication overhead, and operational burden for zero benefit at single-user scale.
**Instead:** Single Go binary with well-separated internal packages. The go-zero framework supports this with its service group pattern.

### Anti-Pattern 2: Proxying Through Next.js API Routes
**What:** Using Next.js API routes as a BFF that forwards requests to the Go backend.
**Why bad:** Adds an extra network hop, complicates SSE streaming, creates two servers to manage.
**Instead:** Next.js serves the frontend; Go serves the API. Frontend calls Go directly. Use CORS configuration for local development.

### Anti-Pattern 3: Storing Embeddings in a Separate Vector DB
**What:** Running Chroma, Qdrant, or Pinecone alongside the app for vector search.
**Why bad:** Massive operational overhead for a personal assistant. Another service to run, configure, back up.
**Instead:** sqlite-vec in the same SQLite database. One file, zero infrastructure.

### Anti-Pattern 4: Loading All Tools Into Every Prompt
**What:** Including every tool's description in the system prompt regardless of relevance.
**Why bad:** Wastes context window tokens. With 10+ tools, descriptions alone can consume 2000+ tokens.
**Instead:** Include core tools always (chat response); activate specialized tools based on conversation context or explicit user request.

### Anti-Pattern 5: Unbounded Conversation Context
**What:** Sending the entire conversation history to the LLM on every turn.
**Why bad:** Hits context window limits quickly, increases cost, degrades response quality.
**Instead:** Sliding window of recent messages + summarized older context + vector-retrieved relevant memories.

## Scalability Considerations

| Concern | Single User (target) | Small Team (future) | Notes |
|---------|---------------------|--------------------|----|
| Concurrent requests | 1-3 at a time | 10-20 | go-zero handles this well out of the box |
| Memory storage | SQLite single file | SQLite still fine up to ~100GB | No need to plan for Postgres yet |
| LLM calls | Sequential per conversation | Queue per user | Rate limiting at provider level |
| Tool execution | Sequential within agent loop | Same | Tools within one turn are sequential by design |
| File storage | Local filesystem | Shared/network filesystem | Cross-user isolation needed later |

**Bottom line:** Do not over-engineer for scale. SQLite, single binary, local filesystem. These choices are correct for the target use case and can be revisited if the project scope changes.

## Suggested Build Order

Based on component dependencies:

```
Phase 1: Foundation
  LLM Provider Adapter --> can call any OpenAI-compatible model
  Basic Agent Runtime  --> simple prompt-in, response-out (no tools yet)
  SSE Streaming        --> go-zero SSE endpoint
  Next.js Chat UI      --> send message, display streamed response

Phase 2: Memory
  SQLite schema        --> conversations + messages tables
  Conversation CRUD    --> create, list, load, delete conversations
  Context assembly     --> inject conversation history into prompt
  Dashboard history    --> conversation list sidebar

Phase 3: Tools
  Tool Registry        --> interface + registration
  Shell tool           --> execute commands
  File read/write tool --> read and write files
  Tool calling in agent loop --> parse tool_calls, execute, loop
  Dashboard tool UI    --> show tool usage, confirmation dialogs

Phase 4: Intelligence
  sqlite-vec setup     --> vector search in same DB
  Semantic memory      --> embed + store conversation summaries
  Context window mgmt  --> token counting, sliding window, summarization
  Web fetch tool       --> fetch and summarize pages
  Web search tool      --> search via API

Phase 5: Polish
  Settings UI          --> model config, API keys, permissions
  Tool safety levels   --> auto-approve vs. confirm
  Error handling       --> graceful failures, retry logic
  Multi-model support  --> switch models mid-conversation
```

**Ordering rationale:**
- Phase 1 must come first: you cannot test anything without LLM connectivity and a UI
- Phase 2 before Phase 3: conversation persistence is needed before tools (tools generate messages that need storing)
- Phase 3 before Phase 4: basic tools validate the agent loop; semantic memory is an optimization
- Phase 4 before Phase 5: intelligence features add real value; settings/polish are quality-of-life

## Sources

- [OpenClaw architecture documentation (explain-openclaw)](https://github.com/centminmod/explain-openclaw) -- HIGH confidence
- [OpenClaw GitHub repository](https://github.com/openclaw/openclaw) -- HIGH confidence
- [Braintrust: The canonical agent architecture is a while loop with tools](https://www.braintrust.dev/blog/agent-while-loop) -- HIGH confidence
- [go-zero SSE documentation](https://go-zero.dev/guides/http/server/sse/) -- HIGH confidence
- [sqlite-vec Go bindings](https://alexgarcia.xyz/sqlite-vec/go.html) -- HIGH confidence
- [chromem-go: embeddable vector DB for Go](https://github.com/philippgille/chromem-go) -- MEDIUM confidence
- [AI Agent Architecture guide (orq.ai)](https://orq.ai/blog/ai-agent-architecture) -- MEDIUM confidence
- [Google ADK for Go](https://developers.googleblog.com/en/announcing-the-agent-development-kit-for-go-build-powerful-ai-agents-with-your-favorite-languages/) -- MEDIUM confidence
- [Microsoft Agent Skills SDK](https://techcommunity.microsoft.com/blog/azuredevcommunityblog/giving-your-ai-agents-reliable-skills-with-the-agent-skills-sdk/4497074) -- MEDIUM confidence
- [Stevens: AI assistant with one SQLite table](https://www.geoffreylitt.com/2025/04/12/how-i-made-a-useful-ai-assistant-with-one-sqlite-table-and-a-handful-of-cron-jobs) -- MEDIUM confidence
