# Phase 3: File and Shell Tools - Research

**Researched:** 2026-03-11
**Domain:** OpenAI tool calling (function calling), Go file/shell execution, SSE streaming with tool events, React approval UI
**Confidence:** HIGH

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| TOOL-01 | Agent can read and write local files on user's machine | Go `os.ReadFile` / `os.WriteFile` with path validation; tool defined as OpenAI `Tool` struct in ChatLogic |
| TOOL-02 | Agent can execute shell commands (subject to safety controls) | Go `exec.CommandContext` with stdout/stderr capture; allowlist/denylist check before execution |
| SAFE-01 | User can configure a command allowlist/denylist for shell tool | Config struct extension with `ShellAllowlist []string` and `ShellDenylist []string`; checked in tool executor before `exec.Command` |
| SAFE-02 | Agent prompts user for approval before executing destructive actions | SSE `{"type":"approval_request", ...}` event pauses the stream; frontend dialog resolves via new `/api/chat/approve` endpoint |
| UI-01 | Chat interface shows tool calls and their results inline alongside messages | New `MessagePart` union type in frontend replaces flat `Message`; `ToolCallBlock` component renders call + result |
</phase_requirements>

---

## Summary

Phase 3 adds tool-use capability to the agent. The OpenAI Chat Completions API's function-calling protocol (`tools` array, `finish_reason: "tool_calls"`, `role: "tool"` messages) is the integration point. The existing `go-openai` library (`v1.41.2`) already provides the Go types needed: `openai.Tool`, `openai.ToolCall`, `openai.FinishReasonToolCalls`, and `ChatCompletionMessage.ToolCalls`. No new backend library is required.

Tool execution lives in a new `internal/toolexec` package (one file per tool: `filetool.go`, `shelltool.go`). The `ChatLogic.StreamChat` agentic loop wraps the existing streaming path: when the model returns `finish_reason: "tool_calls"`, the loop calls the executor, appends a `role: "tool"` result message, and re-calls the model. Shell commands trigger an approval gate before execution, implemented as a pause-and-poll SSE pattern: the backend emits a typed SSE event, the frontend shows a dialog, and a separate `/api/chat/approve` REST endpoint resumes execution.

On the frontend, the `Message` interface currently holds a flat `content` string. Phase 3 extends this to a `MessagePart[]` union of text, tool-call, and tool-result parts, which are rendered as distinct collapsible blocks inside `ChatArea`. This is a backward-compatible extension — existing text messages remain intact.

**Primary recommendation:** Use the existing `go-openai` tool-calling types without adding any agent framework library. Keep the agentic loop entirely inside `ChatLogic`. Implement tool execution as pure Go stdlib (`os`, `exec`) in a dedicated `toolexec` package. For approval, use a session-scoped in-memory channel and a new approval REST endpoint — no WebSocket needed.

---

## Standard Stack

### Core (no new dependencies needed)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/sashabaranov/go-openai` | v1.41.2 (already in go.mod) | `Tool`, `ToolCall`, `FinishReasonToolCalls`, `ChatCompletionMessage.ToolCalls` types | Already in project; contains full tool-calling support |
| Go stdlib `os` | go 1.26 | `ReadFile`, `WriteFile`, `Stat` | No wrapper needed for file operations |
| Go stdlib `os/exec` | go 1.26 | `CommandContext` for shell execution with timeout and cancellation | Standard; context-aware; streams stdout+stderr |
| Go stdlib `path/filepath` | go 1.26 | `Clean`, `Abs` for path traversal prevention | Essential safety primitive |

### Supporting (frontend, no new deps needed)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `shadcn` CLI | 4.0.5 (already in package.json) | Add `Dialog`, `Badge`, `Collapsible` components via `npx shadcn add` | Wave 0 setup step |
| `lucide-react` | 0.577.0 (already installed) | Icons for tool call blocks (FileText, Terminal, Check, X) | Already available |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| In-process approval channel (stdlib `chan`) | Separate Redis/DB for approval state | Channel is sufficient for single-user local server; Redis adds infra complexity |
| Custom tool executor package | `langchaingo` or `openai-agents-go` | Both add large transitive deps; our needs are two tools, not a framework |
| SSE pause-and-poll for approval | WebSocket bidirectional | SSE is already established in this codebase; WebSocket would require new infrastructure |

**Installation:** No new `go get` or `npm install` commands needed. Only `npx shadcn add dialog badge collapsible` for UI components.

---

## Architecture Patterns

### Recommended Project Structure

```
src/backend/internal/
├── toolexec/             # NEW: tool execution (no handler/logic imports)
│   ├── executor.go       # ToolExecutor interface + registry
│   ├── filetool.go       # ReadFile, WriteFile implementations
│   ├── shelltool.go      # ShellRun with allowlist/denylist check
│   └── executor_test.go  # table-driven tests with t.TempDir()
├── logic/
│   └── chatlogic.go      # MODIFIED: agentic loop, tool dispatch, approval gate
├── svc/
│   └── servicecontext.go # MODIFIED: add ToolExecutor, ApprovalStore fields
├── config/
│   └── config.go         # MODIFIED: add ShellAllowlist, ShellDenylist []string
└── handler/
    └── approvehandler.go  # NEW: POST /api/chat/approve
```

```
src/frontend/
├── lib/
│   └── api.ts            # MODIFIED: add submitApproval(), MessagePart types
├── components/
│   ├── ChatArea.tsx       # MODIFIED: render MessagePart[], show approval dialog
│   ├── ToolCallBlock.tsx  # NEW: collapsible tool call + result display
│   └── ApprovalDialog.tsx # NEW: approve/deny dialog for shell commands
```

### Pattern 1: OpenAI Tool Definition (Go)

**What:** Declare tools as `openai.Tool` structs with JSON Schema parameters. Pass the `Tools` slice to every `CreateChatCompletionStream` call.

**When to use:** At the start of `StreamChat`, build the tools slice once per call.

```go
// Source: pkg.go.dev/github.com/sashabaranov/go-openai + OpenAI function calling docs
import openai "github.com/sashabaranov/go-openai"

var chatTools = []openai.Tool{
    {
        Type: openai.ToolTypeFunction,
        Function: &openai.FunctionDefinition{
            Name:        "read_file",
            Description: "Read the contents of a file at the given path",
            Parameters: map[string]any{
                "type": "object",
                "properties": map[string]any{
                    "path": map[string]any{
                        "type":        "string",
                        "description": "Absolute or relative path to the file",
                    },
                },
                "required": []string{"path"},
            },
        },
    },
    // write_file, shell_run similarly
}
```

### Pattern 2: Agentic Loop in ChatLogic

**What:** After the initial stream, check `FinishReason`. If `tool_calls`, execute tools and re-call the model. Loop until `FinishReason == "stop"` or `MaxToolCalls` is reached (SAFE-03, already in config).

**When to use:** Replaces the single-shot `StreamChat` streaming loop.

```go
// Source: OpenAI function calling docs + go-openai FinishReason constants
for iteration := 0; iteration < l.svcCtx.Config.MaxToolCalls; iteration++ {
    stream, err := l.svcCtx.AIClient.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
        Model:    l.svcCtx.Config.Model.Name,
        Messages: history,
        Tools:    chatTools,
    })
    // ... stream tokens as before ...

    if finishReason == openai.FinishReasonToolCalls {
        // Execute tool calls, append tool result messages, continue loop
        for _, tc := range pendingToolCalls {
            result := l.svcCtx.Executor.Execute(ctx, tc)
            history = append(history, openai.ChatCompletionMessage{
                Role:       openai.ChatMessageRoleTool,
                Content:    result,
                ToolCallID: tc.ID,
            })
        }
        continue
    }
    break // FinishReason == "stop"
}
```

### Pattern 3: Streaming Tool Call Delta Accumulation

**What:** During streaming, tool call arguments arrive in fragments across multiple `Recv()` calls. The `delta.ToolCalls` field carries partial arguments with an `Index` to identify which tool call the fragment belongs to.

**When to use:** Inside the stream receive loop when `delta.ToolCalls` is non-empty.

```go
// Source: go-openai ChatCompletionStreamChoiceDelta.ToolCalls, OpenAI streaming docs
toolCallAccum := map[int]*openai.ToolCall{}  // keyed by Index

for {
    resp, err := stream.Recv()
    if errors.Is(err, io.EOF) { break }
    // ...
    delta := resp.Choices[0].Delta
    // Accumulate text content
    if delta.Content != "" {
        fullResponse.WriteString(delta.Content)
        fmt.Fprintf(w, "data: %s\n\n", delta.Content)
        flusher.Flush()
    }
    // Accumulate tool call fragments
    for _, tc := range delta.ToolCalls {
        idx := tc.Index  // *int in go-openai
        if idx == nil { continue }
        if _, ok := toolCallAccum[*idx]; !ok {
            toolCallAccum[*idx] = &openai.ToolCall{ID: tc.ID, Type: tc.Type}
            toolCallAccum[*idx].Function.Name = tc.Function.Name
        }
        toolCallAccum[*idx].Function.Arguments += tc.Function.Arguments
    }

    finishReason = resp.Choices[0].FinishReason
}
// After loop: convert toolCallAccum map to ordered []ToolCall
```

**Important:** `delta.ToolCalls[i].Index` is `*int` in go-openai v1.41. Dereference with nil check.

### Pattern 4: SSE Approval Gate

**What:** When a shell command requires approval, emit a typed SSE event, then block on a per-request channel. The frontend shows a dialog; the user's response is POSTed to `/api/chat/approve`, which sends the decision into the channel. The blocked goroutine resumes.

**When to use:** Before executing any shell command not auto-approved by the allowlist check.

```go
// Backend: emit approval request event, block on channel
approvalID := uuid.New().String()
ch := make(chan bool, 1)
l.svcCtx.ApprovalStore.Register(approvalID, ch)
defer l.svcCtx.ApprovalStore.Delete(approvalID)

// Emit SSE event to frontend
event := map[string]any{
    "type":       "approval_request",
    "approvalId": approvalID,
    "tool":       "shell_run",
    "command":    command,
}
data, _ := json.Marshal(event)
fmt.Fprintf(w, "data: %s\n\n", data)
flusher.Flush()

// Block until approval or timeout
select {
case approved := <-ch:
    if !approved {
        return "", errors.New("user denied command execution")
    }
case <-ctx.Done():
    return "", ctx.Err()
}
```

```go
// Approve handler: POST /api/chat/approve
// Body: { "approvalId": "...", "approved": true/false }
ch := svcCtx.ApprovalStore.Get(req.ApprovalID)
if ch == nil {
    httpx.Error(w, errors.New("unknown approval ID"))
    return
}
ch <- req.Approved
w.WriteHeader(http.StatusNoContent)
```

### Pattern 5: Tool Call SSE Events for UI

**What:** Emit typed SSE JSON events for tool call start, tool result, and approval requests. The frontend parses these alongside text tokens and builds the `MessagePart[]` array.

**When to use:** Immediately before and after each tool execution.

```
// Tool call started
data: {"type":"tool_call","id":"call_abc","name":"read_file","args":"{\"path\":\"/tmp/x\"}"}

// Tool result
data: {"type":"tool_result","id":"call_abc","content":"file contents here..."}

// Approval request (shell only)
data: {"type":"approval_request","approvalId":"uuid","tool":"shell_run","command":"ls -la"}
```

Text tokens remain as bare `data: token\n\n` (no JSON). The frontend tries `JSON.parse(data)` — success means typed event, catch means text token. This preserves backward compatibility with the existing SSE parser in `ChatArea.tsx`.

### Pattern 6: Shell Safety (SAFE-01)

**What:** Before executing a shell command, check it against configurable allowlist/denylist patterns. If neither list is configured, require user approval for every command.

**When to use:** In `shelltool.go` before calling `exec.CommandContext`.

```go
// Source: Go stdlib os/exec; pattern from SAFE-01 requirement
func (s *ShellTool) requiresApproval(command string) bool {
    for _, pattern := range s.denylist {
        if matched, _ := filepath.Match(pattern, command); matched {
            return true // Always block (denylist overrides allowlist)
        }
    }
    for _, pattern := range s.allowlist {
        if matched, _ := filepath.Match(pattern, command); matched {
            return false // Auto-approved
        }
    }
    return true // Default: require approval
}
```

Config additions:
```yaml
# etc/config.yaml
ShellAllowlist:
  - "ls *"
  - "cat *"
ShellDenylist:
  - "rm *"
  - "sudo *"
  - "curl *"
```

### Pattern 7: File Tool Path Safety (TOOL-01)

**What:** Prevent path traversal attacks. Resolve to absolute path and validate it does not escape a configured workspace root. Fall back to a default safe root if none configured.

**When to use:** In `filetool.go` before any file operation.

```go
// Source: Go stdlib path/filepath
func safePath(root, userPath string) (string, error) {
    abs, err := filepath.Abs(filepath.Join(root, userPath))
    if err != nil {
        return "", fmt.Errorf("resolve path: %w", err)
    }
    if !strings.HasPrefix(abs, root) {
        return "", errors.New("path escapes workspace root")
    }
    return abs, nil
}
```

### Anti-Patterns to Avoid

- **Running `exec.Command` without `exec.CommandContext`:** No cancellation on timeout; goroutine leaks if parent context is cancelled.
- **Passing shell string to `sh -c` for simple commands:** Enables injection attacks. Use `exec.CommandContext(ctx, parts[0], parts[1:]...)` after splitting.
- **Blocking approval channel without `ctx.Done()`:** SSE connection can hold open indefinitely if user closes browser. Always select on both `ch` and `ctx.Done()`.
- **Trusting user-supplied paths directly:** Always `filepath.Abs` + prefix-check against workspace root.
- **Storing tool results only in SSE:** Tool results must also be appended to `history` as `role: "tool"` messages before re-calling the model.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| OpenAI tool schema validation | Custom JSON Schema builder | `map[string]any` or inline struct — OpenAI accepts raw `interface{}` | go-openai `FunctionDefinition.Parameters` is `interface{}`, no schema library needed |
| Shell argument parsing/splitting | Custom tokenizer | `strings.Fields(command)` for simple cases; document limitation | Full POSIX shell parsing is complex; for v1 simple splitting is sufficient |
| Process sandboxing | Custom seccomp rules | Allowlist/denylist + user approval (SAFE-01/SAFE-02) | Sandboxing requires root/capabilities; approval gate is the v1 safety model |
| WebSocket for approval | Custom bidirectional protocol | Per-request `chan bool` + REST endpoint | Simpler, consistent with existing HTTP-only backend |
| Agent framework | `langchaingo`, `openai-agents-go` | Custom `toolexec` package | Two tools, bounded loop; frameworks add 50+ transitive deps for no gain |

**Key insight:** The tool-calling protocol is HTTP message format, not a framework. Two tools and a bounded loop are easier to own than a framework abstraction.

---

## Common Pitfalls

### Pitfall 1: Tool Call Arguments Arrive Fragmented in Streaming

**What goes wrong:** Reading `delta.ToolCalls[0].Function.Arguments` in the first chunk yields a partial JSON string. Trying to unmarshal it immediately causes a JSON parse error.

**Why it happens:** The model streams argument JSON character-by-character just like text tokens.

**How to avoid:** Accumulate all argument fragments into a string buffer keyed by `tc.Index`. Only unmarshal after `finish_reason: "tool_calls"` is received.

**Warning signs:** JSON unmarshal errors on tool arguments in streaming path.

### Pitfall 2: Missing `rest.WithSSE()` on New SSE Routes

**What goes wrong:** go-zero's default timeout middleware terminates long-running SSE connections after the configured timeout (default 60s), even mid-tool-execution.

**Why it happens:** Documented in this codebase's CLAUDE.md and in previous phase decisions.

**How to avoid:** All SSE routes MUST use `rest.WithSSE()`. The approval endpoint is non-SSE (POST, returns 204).

**Warning signs:** SSE stream drops silently after 60s; connection terminates without `done` event.

### Pitfall 3: Tool Result Not Added to History Before Re-Call

**What goes wrong:** The model keeps requesting the same tool because it never received the result. Loop exhausts `MaxToolCalls`.

**Why it happens:** Easy to emit the SSE result event but forget to append the `role: "tool"` message to `history` before the next `CreateChatCompletionStream` call.

**How to avoid:** Immediately after executing a tool, append `openai.ChatCompletionMessage{Role: "tool", Content: result, ToolCallID: tc.ID}` to `history`. The assistant's message (with `ToolCalls` populated) must also be in history.

**Warning signs:** Agent loops to `MaxToolCalls`; model repeats identical tool calls.

### Pitfall 4: Approval Channel Orphan After Browser Close

**What goes wrong:** Browser closes SSE connection; context is cancelled; but the approval `chan bool` remains in the `ApprovalStore`. A subsequent request with the same approvalID (unlikely but possible) could unblock the wrong goroutine.

**Why it happens:** The `defer l.svcCtx.ApprovalStore.Delete(approvalID)` prevents this if the goroutine reaches it, but context cancellation via `ctx.Done()` in the select exits cleanly and the defer fires.

**How to avoid:** Always `defer ApprovalStore.Delete(approvalID)` immediately after `Register`. Select on both `ch` and `ctx.Done()`.

**Warning signs:** Memory leak in `ApprovalStore` map visible over many requests.

### Pitfall 5: Path Traversal in File Tool

**What goes wrong:** User asks agent to read `../../../../etc/passwd`. File tool returns system file contents.

**Why it happens:** No path sanitization.

**How to avoid:** `filepath.Abs` + `strings.HasPrefix(abs, workspaceRoot)` check in every file operation. Log and return error if check fails.

**Warning signs:** Files outside the expected directory being read.

### Pitfall 6: Frontend Message Type Mismatch After Extending MessagePart

**What goes wrong:** Existing conversation history loaded from backend returns flat `{role, content}` objects. New frontend code expects `MessagePart[]`. Runtime error when rendering.

**Why it happens:** `getConversationMessages` still returns `MessageResponse{Role, Content}` — flat format.

**How to avoid:** Keep the existing `Message` type for history loading. Introduce `ChatMessage` union type only for the live-streaming accumulator. Convert history `Message[]` to `ChatMessage[]` (single text part) on load. The backend `MessageResponse` struct does NOT need to change in this phase.

**Warning signs:** TypeScript type error or blank messages when loading a past conversation.

---

## Code Examples

### Go: Tool Executor Interface

```go
// Source: Go interface design patterns (golang-patterns skill)
// internal/toolexec/executor.go
package toolexec

import "context"

// ToolResult holds the output of a tool call.
type ToolResult struct {
    Content string
    Error   string // non-empty on failure
}

// Executor runs a named tool with JSON-encoded arguments.
type Executor interface {
    Execute(ctx context.Context, name, argsJSON string) ToolResult
}
```

### Go: Shell Execution with Context

```go
// Source: Go stdlib os/exec
// internal/toolexec/shelltool.go
func (s *ShellTool) Run(ctx context.Context, command string) (string, error) {
    parts := strings.Fields(command)
    if len(parts) == 0 {
        return "", errors.New("empty command")
    }
    cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
    var out bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &out
    if err := cmd.Run(); err != nil {
        return out.String(), fmt.Errorf("command failed: %w", err)
    }
    return out.String(), nil
}
```

### Go: ApprovalStore (in-memory, thread-safe)

```go
// Source: Go sync patterns (golang-patterns skill)
// internal/svc/approvalstore.go
type ApprovalStore struct {
    mu       sync.Mutex
    pending  map[string]chan bool
}

func NewApprovalStore() *ApprovalStore {
    return &ApprovalStore{pending: make(map[string]chan bool)}
}

func (a *ApprovalStore) Register(id string, ch chan bool) {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.pending[id] = ch
}

func (a *ApprovalStore) Resolve(id string, approved bool) bool {
    a.mu.Lock()
    ch, ok := a.pending[id]
    a.mu.Unlock()
    if !ok { return false }
    ch <- approved
    return true
}

func (a *ApprovalStore) Delete(id string) {
    a.mu.Lock()
    defer a.mu.Unlock()
    delete(a.pending, id)
}
```

### TypeScript: MessagePart Union Type

```typescript
// Source: TypeScript discriminated union pattern (typescript-expert skill)
// lib/api.ts additions

export interface TextPart {
  type: 'text'
  content: string
}

export interface ToolCallPart {
  type: 'tool_call'
  id: string
  name: string
  args: string  // JSON string
}

export interface ToolResultPart {
  type: 'tool_result'
  id: string    // matches tool_call id
  content: string
  error?: string
}

export interface ApprovalRequestPart {
  type: 'approval_request'
  approvalId: string
  tool: string
  command: string
}

export type MessagePart = TextPart | ToolCallPart | ToolResultPart | ApprovalRequestPart

// ChatMessage extends Message for live streaming display
export interface ChatMessage {
  role: 'user' | 'assistant' | 'system'
  parts: MessagePart[]
}
```

### TypeScript: SSE Event Parser Extension

```typescript
// ChatArea.tsx - extend existing SSE parser
// Existing: text tokens fall through to catch block (unchanged)
// New: typed events handled in try block

try {
  const parsed = JSON.parse(data)
  if (parsed.done === true && parsed.sessionId) {
    onSessionCreated(parsed.sessionId)
    break
  }
  if (parsed.type === 'tool_call') {
    // Append ToolCallPart to current assistant message
    appendPart({ type: 'tool_call', id: parsed.id, name: parsed.name, args: parsed.args })
  } else if (parsed.type === 'tool_result') {
    appendPart({ type: 'tool_result', id: parsed.id, content: parsed.content, error: parsed.error })
  } else if (parsed.type === 'approval_request') {
    appendPart({ type: 'approval_request', approvalId: parsed.approvalId, tool: parsed.tool, command: parsed.command })
    setPendingApproval(parsed)
  }
} catch {
  // Not JSON — text token (existing behavior preserved)
  assistantContent += data
  appendTextContent(data)
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `function_call` field (deprecated) | `tools` array + `tool_calls` in response | OpenAI mid-2023 | Use `tools`, not `functions` |
| `role: "function"` for results | `role: "tool"` with `tool_call_id` | OpenAI mid-2023 | go-openai `ChatMessageRoleTool` constant |
| Agent frameworks (LangChain) | Minimal custom loop | 2024 trend | Lighter deps, more control |

**Deprecated/outdated:**
- `openai.FunctionDefinition` in `ChatCompletionRequest.Functions`: Deprecated; use `Tools` instead. The `functions` field still works but `tools` is the current API.
- `ChatMessageRoleFunction`: Deprecated; use `ChatMessageRoleTool`.

---

## Open Questions

1. **Workspace root for file tool**
   - What we know: File tool needs a root to prevent path traversal
   - What's unclear: Should this be configurable in `config.yaml` or default to the process's working directory?
   - Recommendation: Add `WorkspaceRoot string` to `config.Config` with a default of `.` (process cwd); document in `etc/config.yaml`

2. **Shell command splitting for quoted arguments**
   - What we know: `strings.Fields("ls -la")` works for simple commands; fails for `echo "hello world"` (splits on space inside quotes)
   - What's unclear: How complex shell commands will the agent generate?
   - Recommendation: Use `strings.Fields` for v1; document the limitation in tool description so the model knows not to use quoted arguments. Full shell parsing can be added in a follow-up.

3. **Approval timeout behavior**
   - What we know: Context timeout (60s default) will cancel the approval wait
   - What's unclear: Should the frontend show a timeout warning if no response in N seconds?
   - Recommendation: Frontend can set a local 30s countdown on the approval dialog; on timeout, auto-deny and close.

---

## Validation Architecture

nyquist_validation is enabled in `.planning/config.json`.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + `github.com/stretchr/testify` v1.11.1 |
| Config file | none (Go test runner, `go test ./...`) |
| Quick run command | `cd src/backend && go test ./internal/toolexec/... ./internal/logic/... -count=1` |
| Full suite command | `cd src/backend && go test ./... -cover` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| TOOL-01 | `ReadFile` returns file contents; `WriteFile` writes to disk | unit | `go test ./internal/toolexec/... -run TestFileTool -v` | Wave 0 |
| TOOL-01 | Path traversal rejected | unit | `go test ./internal/toolexec/... -run TestFileTool_PathTraversal -v` | Wave 0 |
| TOOL-02 | `ShellRun` executes command, returns stdout | unit | `go test ./internal/toolexec/... -run TestShellTool -v` | Wave 0 |
| TOOL-02 | Agentic loop executes tool and feeds result back to model | unit (mock) | `go test ./internal/logic/... -run TestChatLogic_ToolLoop -v` | Wave 0 |
| SAFE-01 | Denylist command returns approval=required; allowlist auto-approves | unit | `go test ./internal/toolexec/... -run TestShellTool_Allowlist -v` | Wave 0 |
| SAFE-02 | Approval gate blocks until decision received | unit | `go test ./internal/logic/... -run TestChatLogic_ApprovalGate -v` | Wave 0 |
| UI-01 | Tool call and result blocks render in chat | manual | npm run dev + browser inspection | manual only |

### Sampling Rate

- **Per task commit:** `cd src/backend && go test ./internal/toolexec/... ./internal/logic/... -count=1`
- **Per wave merge:** `cd src/backend && go test ./... -cover`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `src/backend/internal/toolexec/executor_test.go` — covers TOOL-01, TOOL-02, SAFE-01
- [ ] `src/backend/internal/logic/chatlogic_test.go` — MODIFY to add tool loop tests (TOOL-02, SAFE-02) — file exists, needs new cases
- [ ] `src/backend/internal/svc/approvalstore.go` + `approvalstore_test.go` — covers SAFE-02 approval channel

---

## Sources

### Primary (HIGH confidence)

- `pkg.go.dev/github.com/sashabaranov/go-openai` — Tool, ToolCall, FinishReasonToolCalls, ChatCompletionStreamChoiceDelta.ToolCalls types verified
- Project source code — `src/backend/internal/{logic,svc,handler,config}/` — existing patterns, ServiceContext structure, SSE streaming approach
- Project CLAUDE.md files — SSE gotchas, go-zero layer rules, config defaults pattern
- Go stdlib docs — `os.ReadFile`, `os.WriteFile`, `exec.CommandContext`, `path/filepath.Clean`

### Secondary (MEDIUM confidence)

- OpenAI function calling docs (WebSearch verified) — `role: "tool"`, `tool_call_id`, `finish_reason: "tool_calls"`, `tools` array format
- `github.com/sashabaranov/go-openai/discussions/453` — streaming tool call accumulation pattern; confirmed `ToolCalls` field exists in `ChatCompletionStreamChoiceDelta`
- `ai-sdk.dev/cookbook/next/human-in-the-loop` — approval gate UX pattern (pause SSE, show dialog, resume)
- `ai-sdk.dev/docs/agents/loop-control` — bounded agent loop pattern

### Tertiary (LOW confidence)

- `strings.Fields` for shell argument splitting — community pattern; lacks POSIX quote handling

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all libraries already in project; go-openai tool types verified against pkg.go.dev
- Architecture: HIGH — based on existing codebase patterns, go-zero handler→logic→svc layer rule, SSE streaming precedent
- Tool calling protocol: HIGH — OpenAI API format verified via multiple sources
- Pitfalls: HIGH — drawn from existing codebase decisions log and known go-zero SSE gotchas
- Approval gate pattern: MEDIUM — architecture is sound but specific implementation details (channel vs store) are design choices

**Research date:** 2026-03-11
**Valid until:** 2026-06-11 (stable OpenAI protocol; go-openai v1 API stable)
