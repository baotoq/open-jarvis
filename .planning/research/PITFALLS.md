# Pitfalls Research

**Domain:** Local-first personal AI assistant (Go backend, Next.js UI, OpenAI-compatible LLM agents)
**Researched:** 2026-03-11
**Confidence:** HIGH

## Critical Pitfalls

### Pitfall 1: Unrestricted Shell and File System Access

**What goes wrong:**
The agent executes destructive shell commands (`rm -rf`, overwrites config files, deletes user data) either through LLM hallucination, prompt injection, or simply misunderstanding the user's intent. In December 2025, a developer's entire Mac home directory was deleted when an AI agent executed `rm -rf` with an accidental tilde expansion. Shell syntax is too expressive for safe agent execution -- features that make shell powerful for humans become liabilities when an LLM generates commands.

**Why it happens:**
Developers give the agent full shell access because it is the easiest path to "it works." They trust the LLM to be careful, but LLMs are probabilistic systems that will eventually produce dangerous commands. Indirect prompt injection (malicious content in files the agent reads) can also trigger destructive actions.

**How to avoid:**
- Implement a command allowlist/blocklist from day one. Block `rm -rf`, `dd`, `mkfs`, `chmod -R`, pipe to `sh`/`bash`, and similar destructive patterns.
- Require explicit user approval for all shell commands by default. Show the exact command and arguments before execution.
- Run agent commands in a restricted working directory. Never expose `/etc`, `/var`, `$HOME` broadly.
- Implement a "dry run" mode that shows what the command would do without executing.
- Set hard timeouts on all command execution (30s default, user-configurable).

**Warning signs:**
- No approval step exists for shell commands during initial implementation.
- Agent can access paths outside the designated workspace.
- No command blocklist exists in the codebase.

**Phase to address:**
Phase 1 (Core Agent Loop). Shell execution must ship with safety guardrails from the first working version. Never ship an unguarded shell executor, even internally.

---

### Pitfall 2: Agent Runaway Loops and Token Cost Explosion

**What goes wrong:**
The agent enters an infinite or near-infinite tool-calling loop -- retrying failed commands, re-reading the same files, or calling itself recursively. Each loop iteration resends the entire conversation context to the LLM API. A 2-hour runaway loop at GPT-4 rates costs $15-40+. With local models the cost is time and compute, but the loop still blocks the system.

**Why it happens:**
LLMs misinterpret termination signals, generate repetitive actions, or suffer from inconsistent internal state ("Loop Drift"). Weak or absent stopping criteria mean the system has no external mechanism to halt the agent. The agent itself cannot be trusted to terminate -- it is the system running the agent that must guarantee termination.

**How to avoid:**
- Implement deterministic loop guardrails external to the agent: `max_tool_calls_per_turn` (e.g., 25), `max_total_tokens_per_request` (e.g., 100k), and `max_elapsed_time` (e.g., 5 minutes).
- Track tool call history per turn. If the agent calls the same tool with the same arguments 3 times, force-terminate.
- Assign a token budget per request. Once exhausted, the turn dies immediately.
- Log and surface loop metrics in the UI so users can see when the agent is spinning.

**Warning signs:**
- No `max_iterations` or `max_tool_calls` constant exists in the agent loop code.
- Token usage per request is not tracked or logged.
- Agent loop has no timeout mechanism.

**Phase to address:**
Phase 1 (Core Agent Loop). The agent execution loop must have hard limits before any tool use is enabled. Retrofitting limits after shipping is dangerous because users will already be running unguarded agents.

---

### Pitfall 3: Context Window Mismanagement

**What goes wrong:**
Conversation history grows unbounded. Each API call resends the full history, and eventually the context exceeds the model's window. The request either fails silently (truncated), errors out, or the model loses coherence because early context is dropped. Tool outputs (file contents, command results) inflate context rapidly -- a single `cat` of a large file can consume half the context window.

**Why it happens:**
Developers treat the message array as an append-only list and never implement truncation, summarization, or sliding window strategies. Tool outputs are inserted verbatim without size limits. Different models have different context limits (4k to 128k+), but the code assumes one size fits all.

**How to avoid:**
- Query the model's context limit at connection time (or configure it per model). Track token count for every message added to context.
- Implement a context management strategy from the start: sliding window (drop oldest messages) for MVP, with summarization as an upgrade path.
- Truncate tool outputs to a configurable max size (e.g., 4000 tokens). For large files, return only head/tail with a note about truncation.
- Store full conversation in the database but only send a managed window to the LLM.
- Do NOT rely on tiktoken for non-OpenAI models -- it produces wrong counts. Use a model-agnostic approximation (4 chars per token) or model-specific tokenizers.

**Warning signs:**
- No token counting anywhere in the codebase.
- Tool outputs are passed to the LLM without any size check.
- Conversation history is passed as-is to the API without truncation logic.
- The code uses a hardcoded context limit rather than per-model configuration.

**Phase to address:**
Phase 1 (Core Agent Loop). Context management must be designed into the message handling layer from the start. It is extremely painful to retrofit because it touches every part of the conversation flow.

---

### Pitfall 4: Prompt Injection via Tool Outputs

**What goes wrong:**
The agent reads a file, fetches a web page, or receives command output that contains adversarial instructions. The LLM treats these as legitimate instructions and performs unauthorized actions (exfiltrating data, running commands, changing behavior). An attacker can embed instructions in a README, a web page, or even a filename.

**Why it happens:**
LLMs fundamentally cannot distinguish between user instructions and data content -- to the LLM, everything is just text. There is no input sanitizer that can reliably distinguish legitimate documentation from adversarial instructions at the text level.

**How to avoid:**
- Use structured message formats that clearly delineate system instructions, user messages, and tool outputs. Never concatenate tool output into the system prompt.
- Implement output-side guardrails: after the LLM responds with a tool call, validate it against the allowlist before execution regardless of what the LLM "thinks" it should do.
- For web fetching, strip scripts and suspicious patterns from fetched content before inserting into context.
- Apply the principle of least privilege: the agent should only have access to tools and paths that the user has explicitly enabled.
- Log all tool calls with their provenance (what triggered them) for post-hoc audit.

**Warning signs:**
- Tool outputs are inserted into context without any framing or role separation.
- No validation layer exists between LLM tool-call responses and actual execution.
- Web content is fetched and inserted raw into the conversation.

**Phase to address:**
Phase 2 (Tool Integration). When file reading, web browsing, and shell execution are connected, prompt injection becomes a real attack surface. Build the validation layer as part of tool integration, not after.

---

### Pitfall 5: Overengineering Memory Before Validating the Core Loop

**What goes wrong:**
Developers spend weeks building sophisticated vector-based memory retrieval (embeddings, vector DB, hybrid search, RAG pipelines) before the basic chat-and-execute loop works reliably. The memory system retrieves irrelevant or contradictory context, confusing the agent. The vector database grows indefinitely without pruning, and chunking strategy choices made early prove wrong later.

**Why it happens:**
Memory feels like the "AI" part of the project and is intellectually interesting. Developers copy patterns from enterprise RAG systems that solve different problems. OpenClaw uses hybrid vector + FTS5, which tempts direct replication without understanding why.

**How to avoid:**
- Start with simple conversation persistence: store messages in SQLite, load recent N messages per session. This covers 80% of "memory" needs for a personal assistant.
- Add full-text search (FTS5 in SQLite) as the second step -- it is simple, requires no embeddings infrastructure, and handles keyword-based recall well.
- Only add vector search when you have concrete evidence that FTS5 is insufficient for your use cases. Most personal assistants never reach this point.
- If you do add vector search, plan for memory pruning/forgetting from day one. Without it, the database grows indefinitely and retrieval quality degrades.

**Warning signs:**
- Setting up a vector database before the chat loop works end-to-end.
- Spending time on embedding model selection before having 100 real conversations to test against.
- No simple fallback exists if the vector search is removed.

**Phase to address:**
Phase 3 (Memory/Persistence). Memory should be the third major feature, after chat and tools. Start simple (SQLite + recent messages), add FTS5, then evaluate whether vectors are needed.

---

### Pitfall 6: go-zero Overhead for a Single-User Application

**What goes wrong:**
go-zero is a microservices framework designed for large-scale distributed systems. Using it for a single-user personal assistant means fighting the framework: service discovery, RPC infrastructure, distributed tracing, and configuration centers add complexity without benefit. Code generation produces "magical" boilerplate that is hard to debug. The steep learning curve slows development significantly.

**Why it happens:**
go-zero is a legitimate choice for Go microservices, and the decision may have been made with future scaling in mind. But a personal assistant running on one machine for one user is the opposite of go-zero's design target.

**How to avoid:**
- Evaluate whether go-zero's microservice features (service discovery, RPC, distributed tracing) are actually needed. For a single-user local app, they are not.
- Consider using go-zero's REST module only (ignoring RPC/microservice features) to keep things simpler, or switch to a lighter framework like Chi, Echo, or Fiber that match the project's actual scope.
- If staying with go-zero, resist the urge to use its code generation for everything. Write the agent logic manually for full control and debuggability.
- Structure the code so the framework is a thin HTTP layer, not deeply coupled to business logic. This allows swapping frameworks later if go-zero proves too heavy.

**Warning signs:**
- Setting up service discovery or RPC for a single-process application.
- Debugging issues caused by generated code you do not understand.
- go-zero's conventions fighting the application's natural architecture.

**Phase to address:**
Phase 1 (Project Setup). This is a foundational decision that affects every subsequent phase. Evaluate early and honestly whether go-zero's weight is justified.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Hardcoded model config (context size, API URL) | Ship faster | Cannot switch models without code changes; breaks multi-model support | Never -- use config from day one |
| No streaming (wait for full response) | Simpler implementation | Terrible UX for long responses (10-60s blank screen) | First 2 days of prototyping only |
| Storing API keys in config files | Quick setup | Security risk if the machine is shared or repo is pushed | MVP only, with `.gitignore` protection. Move to OS keychain or env vars before any "release" |
| Monolithic agent prompt (one giant system prompt) | Easy to iterate | Impossible to test, version, or swap components independently | First prototype only. Split into composable prompt sections early |
| Synchronous tool execution | Simpler control flow | Agent blocks on slow tools (web fetch, large file reads). UI freezes | Never for user-facing tools. Use goroutines with context cancellation from the start |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| OpenAI-compatible APIs (Ollama, Anthropic adapters) | Assuming all providers implement the spec identically. Temperature `0` with `omitempty` in Go structs sends no temperature field, causing provider to use default (1.0). `response_format` for JSON mode is not universally supported. | Test against at least 2 providers (OpenAI + Ollama) from the start. Use pointer types for optional numeric fields in Go (`*float32` for temperature). Feature-detect capabilities per provider. |
| SSE streaming to Next.js frontend | Using WebSockets for one-way token streaming (unnecessary complexity). Nginx/reverse proxy buffering kills streaming. Browser SSE limit of 6 connections per domain on HTTP/1.1. | Use SSE (not WebSocket) for token streaming -- it is the industry standard for LLM responses. Disable response buffering on streaming routes. Use HTTP/2 to avoid the 6-connection limit. |
| Ollama (local models) | Assuming Ollama is always running and responsive. Not handling model download/loading time. Treating local model capabilities as identical to GPT-4. | Health-check Ollama on startup. Show model loading state in UI. Set appropriate expectations per model (tool calling support varies wildly across local models). |
| File system access | Using absolute paths from user input without validation. Following symlinks outside the workspace. Not handling permission errors gracefully. | Resolve and validate all paths against an allowed root directory. Do not follow symlinks outside the workspace. Return clear error messages for permission failures. |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Resending full conversation history on every API call | Increasing latency per message, growing API costs | Implement sliding window or summarization. Track token count per request. | After ~20 messages in a conversation (or fewer with tool outputs) |
| Unbounded tool output in context | Single tool call consumes most of the context window | Truncate tool outputs to configurable max. Summarize large outputs. | First time agent reads a file >100 lines or fetches a long web page |
| Synchronous web page fetching | Agent blocks for 5-30s while fetching and parsing a page | Fetch asynchronously with timeout. Stream partial results. Cache fetched pages. | Any web fetch on a slow or large site |
| SQLite write contention | Database locks under concurrent writes (agent + UI + memory) | Use WAL mode. Serialize writes through a single goroutine or channel. | When agent runs tools while user sends new messages simultaneously |
| No response caching for identical queries | Same question hits the LLM API every time | Cache responses keyed on (model, messages hash) with TTL | During development/testing when the same prompts are sent repeatedly |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Agent runs with user's full OS permissions | Destructive commands affect the entire system. Agent can read SSH keys, credentials, browser data. | Run agent commands in a restricted sandbox (chroot, container, or at minimum restricted PATH and HOME). Apply allowlists for accessible directories. |
| API keys stored in plaintext config | Keys leaked if repo is pushed, machine is compromised, or backups are exposed | Use environment variables or OS keychain. Never commit config files with keys. Add `.env` to `.gitignore` before first commit. |
| No rate limiting on the local API | If exposed on network (even LAN), any device can send requests and trigger shell commands | Bind to localhost only by default. Require authentication token for non-localhost access. Implement rate limiting even locally. |
| Web fetch without content sanitization | Fetched pages inject prompt attacks, or contain malicious scripts that affect rendering | Strip `<script>` tags, sanitize HTML to markdown, truncate content, frame as tool output (not system instructions) |
| No audit log for agent actions | Cannot determine what the agent did after something goes wrong | Log every tool call (command, args, output, timestamp) to a persistent, append-only log. Surface in UI. |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| No streaming -- blank screen while LLM generates | User thinks the app is frozen. Abandons after 5-10 seconds. | Stream tokens via SSE. Show typing indicator immediately. Display partial responses as they arrive. |
| No visibility into agent "thinking" | User cannot tell if the agent is reading files, running commands, or stuck | Show real-time tool call status: "Reading file X...", "Running command...", "Searching web...". Display tool results inline. |
| Error messages from LLM API shown raw | User sees `{"error": {"code": 429, "message": "Rate limit exceeded"}}` | Translate API errors to human-readable messages. Implement retry with backoff for transient errors. Show "Model is busy, retrying..." not the raw JSON. |
| No way to cancel a running agent action | Agent is stuck in a loop or running a slow command, user has no recourse | Implement cancel button that sends context cancellation through the entire tool chain. Kill running subprocesses on cancel. |
| Conversation list without search or organization | After 50+ conversations, users cannot find anything | Implement conversation search (title + content) and basic organization (folders or tags) before conversation count grows. |

## "Looks Done But Isn't" Checklist

- [ ] **Chat works:** Often missing proper error handling for API failures (network timeout, invalid API key, model not found). Verify the chat gracefully handles all LLM API error codes.
- [ ] **Shell execution works:** Often missing timeout handling, output size limits, and stderr capture. Verify a command that runs for 60s is killed, and a command that outputs 10MB does not crash the agent.
- [ ] **File reading works:** Often missing binary file detection, encoding handling, and size limits. Verify the agent does not try to read a 500MB video file into context.
- [ ] **Web browsing works:** Often missing JavaScript-rendered page handling, redirect following, and timeout. Verify it handles sites that require JS, return 403, or take 30s to respond.
- [ ] **Conversation persistence works:** Often missing conversation title generation, deletion, and export. Verify conversations survive a server restart and can be deleted without corrupting the database.
- [ ] **Multi-model support works:** Often missing per-model configuration (context size, temperature defaults, tool calling support). Verify switching from GPT-4 to Ollama llama3 does not break tool calling or exceed context limits.
- [ ] **Streaming works:** Often missing reconnection on SSE disconnect and proper cleanup on page navigation. Verify streaming survives a brief network interruption and does not leak connections when the user navigates away.

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Destructive shell command executed | HIGH | Restore from backup (if exists). Implement allowlist/approval. Cannot undo deleted files without backups. This is why prevention is critical. |
| Runaway loop burned API credits | LOW | Kill the process. Add loop guardrails. Credits are gone but the system is undamaged. |
| Context window overflow corrupting responses | LOW | Implement truncation strategy. Re-architecture message handling. Old conversations are fine in DB. |
| Prompt injection caused data exfiltration | HIGH | Audit logs to determine what was accessed. Rotate any exposed credentials. Implement tool output framing and execution validation. Damage depends on what the agent had access to. |
| Vector memory returning irrelevant context | MEDIUM | Fall back to FTS5 or recent-messages-only. Re-index with better chunking strategy. May need to re-evaluate embedding model. |
| go-zero framework fighting application design | MEDIUM | Extract business logic from framework-coupled code. Swap HTTP layer to lighter framework. Cost depends on how deeply coupled the code became. |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Unrestricted shell/file access | Phase 1: Core Agent Loop | Command allowlist exists. Approval step works. Path validation rejects out-of-scope access. |
| Agent runaway loops | Phase 1: Core Agent Loop | Max iterations, token budget, and timeout constants exist. A test triggers each limit successfully. |
| Context window mismanagement | Phase 1: Core Agent Loop | Token counting is logged per request. Truncation activates before hitting model limit. |
| Prompt injection via tools | Phase 2: Tool Integration | Tool outputs are framed with role markers. Validation layer exists between LLM response and execution. |
| Overengineered memory | Phase 3: Memory/Persistence | SQLite + FTS5 works. Vector search is not introduced before 100+ real conversations demonstrate the need. |
| go-zero overhead | Phase 1: Project Setup | Framework is used as thin HTTP layer only. No service discovery or RPC configured for single-user mode. |
| No streaming UX | Phase 1: Core Agent Loop | SSE streaming delivers tokens to the frontend. Typing indicator shows immediately. |
| No audit logging | Phase 1: Core Agent Loop | Every tool call is logged with timestamp, command, args, and result. Logs are queryable. |
| API key security | Phase 1: Project Setup | No API keys in committed files. `.env` in `.gitignore`. Environment variable or keychain loading works. |
| No cancellation mechanism | Phase 2: Tool Integration | Cancel button in UI sends context cancellation. Running subprocess is killed within 1 second. |

## Sources

- [MIT Technology Review: Is a secure AI assistant possible?](https://www.technologyreview.com/2026/02/11/1132768/is-a-secure-ai-assistant-possible/) - Prompt injection fundamentals
- [NVIDIA: Practical Security Guidance for Sandboxing Agentic Workflows](https://developer.nvidia.com/blog/practical-security-guidance-for-sandboxing-agentic-workflows-and-managing-execution-risk/) - Sandbox design patterns
- [Securing Shell Execution Agents](https://yortuc.com/posts/securing-shell-execution-agents/) - Shell safety patterns
- [DEV Community: Rate Limiting Your Own AI Agent](https://dev.to/askpatrick/rate-limiting-your-own-ai-agent-the-runaway-loop-problem-nobody-talks-about-3dh2) - Runaway loop problem
- [DEV Community: The Token Budget Pattern](https://dev.to/askpatrick/the-token-budget-pattern-how-to-stop-ai-agent-cost-surprises-before-they-happen-5hb3) - Token budget guardrails
- [Procedure Tech: SSE Still Wins for LLM Streaming in 2026](https://procedure.tech/blogs/the-streaming-backbone-of-llms-why-server-sent-events-(sse)-still-wins-in-2025) - SSE vs WebSocket for streaming
- [Cursor Blog: Implementing a Secure Sandbox for Local Agents](https://cursor.com/blog/agent-sandboxing) - Local agent sandboxing
- [LangChain: LangMem Conceptual Guide](https://langchain-ai.github.io/langmem/concepts/conceptual_guide/) - Memory architecture patterns
- [Three Dots Labs: When You Shouldn't Use Frameworks in Go](https://threedots.tech/episode/when-you-should-not-use-frameworks/) - Go framework overuse
- [Medium: go-zero framework evaluation](https://medium.com/@g.zhufuyi/ive-tried-many-go-frameworks-here-s-why-i-finally-chose-this-one-a73ad2636a50) - go-zero limitations
- [OpenAI Go Library](https://github.com/openai/openai-go) - Official Go client
- [sashabaranov/go-openai](https://pkg.go.dev/github.com/sashabaranov/go-openai) - Go OpenAI client pitfalls (temperature omitempty)

---
*Pitfalls research for: Local-first personal AI assistant (Go + Next.js)*
*Researched: 2026-03-11*
