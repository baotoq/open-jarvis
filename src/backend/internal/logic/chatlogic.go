package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	openai "github.com/sashabaranov/go-openai"
	"open-jarvis/internal/svc"
	"open-jarvis/internal/types"
)

// StreamRecver abstracts the go-openai streaming response for testability.
// This mirrors svc.StreamRecver to allow mock injection in tests without import cycles.
type StreamRecver interface {
	Recv() (openai.ChatCompletionStreamResponse, error)
	Close() error
}

// chatTools defines the tools available to the LLM in every chat completion request.
var chatTools = []openai.Tool{
	{Type: openai.ToolTypeFunction, Function: &openai.FunctionDefinition{
		Name:        "read_file",
		Description: "Read contents of a file",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string", "description": "File path (relative to workspace root)"},
			},
			"required": []string{"path"},
		},
	}},
	{Type: openai.ToolTypeFunction, Function: &openai.FunctionDefinition{
		Name:        "write_file",
		Description: "Write content to a file",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":    map[string]any{"type": "string", "description": "File path (relative to workspace root)"},
				"content": map[string]any{"type": "string", "description": "Content to write"},
			},
			"required": []string{"path", "content"},
		},
	}},
	{Type: openai.ToolTypeFunction, Function: &openai.FunctionDefinition{
		Name:        "shell_run",
		Description: "Run a shell command. Use simple commands only (no quotes around arguments). Example: ls -la /tmp",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{"type": "string", "description": "Command to run (no quoted args)"},
			},
			"required": []string{"command"},
		},
	}},
	{Type: openai.ToolTypeFunction, Function: &openai.FunctionDefinition{
		Name:        "web_fetch",
		Description: "Fetch and extract readable text content from a web page URL",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"url": map[string]any{"type": "string", "description": "Full URL including https://"},
			},
			"required": []string{"url"},
		},
	}},
	{Type: openai.ToolTypeFunction, Function: &openai.FunctionDefinition{
		Name:        "web_search",
		Description: "Search the web and return result titles, URLs, and descriptions",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{"type": "string", "description": "Search query string"},
			},
			"required": []string{"query"},
		},
	}},
}

// ChatLogic handles the streaming LLM call and conversation management.
type ChatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewChatLogic creates a new ChatLogic instance.
func NewChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatLogic {
	return &ChatLogic{ctx: ctx, svcCtx: svcCtx}
}

// StreamChat performs the agentic streaming LLM call, writing SSE tokens to w.
// It prepends the system prompt for new sessions, appends conversation history,
// and stores the assembled response in the Store after completion.
// When req.SessionId is empty, a UUID is assigned and a new conversation is created.
// A final SSE done event carrying the session ID is emitted after streaming completes.
// The agentic loop calls tools (read_file, write_file, shell_run) when the model
// requests them, appending results to history and re-calling the model up to
// MaxToolCalls iterations.
func (l *ChatLogic) StreamChat(req *types.ChatRequest, w http.ResponseWriter) error {
	ctx, cancel := context.WithTimeout(l.ctx,
		time.Duration(l.svcCtx.Config.TurnTimeoutSeconds)*time.Second)
	defer cancel()

	// Determine if this is a new session or an existing one.
	isNewSession := false
	if req.SessionId == "" {
		req.SessionId = uuid.New().String()
		isNewSession = true
	}

	// Build message history with system prompt for new sessions
	history := l.svcCtx.Store.Get(req.SessionId)
	if len(history) == 0 {
		isNewSession = true
		history = append(history, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: l.svcCtx.Config.Model.SystemPrompt,
		})
	}
	history = append(history, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.Message,
	})

	// Assert flusher before making the LLM call
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported: ResponseWriter does not implement http.Flusher")
	}

	// Agentic loop: each iteration may result in tool calls or a final text response.
	// MaxToolCalls bounds the total number of loop iterations (including the final stop).
	maxIter := l.svcCtx.Config.MaxToolCalls
	if maxIter <= 0 {
		maxIter = 10
	}

	var lastAssistantContent strings.Builder

	for iteration := 0; iteration < maxIter; iteration++ {
		stream, err := l.svcCtx.AIClient.CreateChatCompletionStream(ctx,
			openai.ChatCompletionRequest{
				Model:    l.svcCtx.Config.Model.Name,
				Messages: history,
				Tools:    chatTools,
			})
		if err != nil {
			fmt.Fprintf(w, "data: [ERROR] %s\n\n", err.Error())
			flusher.Flush()
			return fmt.Errorf("create stream: %w", err)
		}

		// Accumulate tool call fragments by index and full text content
		toolCallAccum := map[int]*openai.ToolCall{}
		var textContent strings.Builder
		var finishReason openai.FinishReason

		recvErr := func() error {
			defer stream.Close()
			for {
				resp, err := stream.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					fmt.Fprintf(w, "data: [ERROR] %s\n\n", err.Error())
					flusher.Flush()
					return fmt.Errorf("stream recv: %w", err)
				}
				if len(resp.Choices) == 0 {
					continue
				}
				delta := resp.Choices[0].Delta

				// Accumulate text tokens
				if delta.Content != "" {
					textContent.WriteString(delta.Content)
					fmt.Fprintf(w, "data: %s\n\n", delta.Content)
					flusher.Flush()
				}

				// Accumulate tool call fragments
				for _, tc := range delta.ToolCalls {
					if tc.Index == nil {
						continue
					}
					idx := *tc.Index
					if _, exists := toolCallAccum[idx]; !exists {
						toolCallAccum[idx] = &openai.ToolCall{
							ID:   tc.ID,
							Type: tc.Type,
						}
						toolCallAccum[idx].Function.Name = tc.Function.Name
					}
					// Append ID if we get it in a later chunk
					if tc.ID != "" && toolCallAccum[idx].ID == "" {
						toolCallAccum[idx].ID = tc.ID
					}
					if tc.Function.Name != "" && toolCallAccum[idx].Function.Name == "" {
						toolCallAccum[idx].Function.Name = tc.Function.Name
					}
					toolCallAccum[idx].Function.Arguments += tc.Function.Arguments
				}

				// Track finish reason (may come in same chunk as last delta or separately)
				if resp.Choices[0].FinishReason != "" {
					finishReason = resp.Choices[0].FinishReason
				}
			}
			return nil
		}()
		if recvErr != nil {
			return recvErr
		}

		// If finish reason is not tool_calls, the model produced a final response — break
		if finishReason != openai.FinishReasonToolCalls {
			lastAssistantContent = textContent
			break
		}

		// Convert accumulated tool calls map to ordered slice
		var toolCalls []openai.ToolCall
		for i := 0; i < len(toolCallAccum); i++ {
			if tc, ok := toolCallAccum[i]; ok {
				toolCalls = append(toolCalls, *tc)
			}
		}

		// CRITICAL: append assistant message with ToolCalls to history BEFORE tool results
		history = append(history, openai.ChatCompletionMessage{
			Role:      openai.ChatMessageRoleAssistant,
			ToolCalls: toolCalls,
		})

		// Dispatch each tool call
		for _, tc := range toolCalls {
			// Emit tool_call SSE event
			toolCallEvent, _ := json.Marshal(map[string]any{
				"type": "tool_call",
				"id":   tc.ID,
				"name": tc.Function.Name,
				"args": tc.Function.Arguments,
			})
			fmt.Fprintf(w, "data: %s\n\n", toolCallEvent)
			flusher.Flush()

			var resultContent string

			// Check if shell_run requires approval gate
			if tc.Function.Name == "shell_run" && l.svcCtx.ShellTool != nil {
				// Extract the command from args to check approval
				var shellArgs struct{ Command string }
				if jsonErr := json.Unmarshal([]byte(tc.Function.Arguments), &shellArgs); jsonErr == nil {
					if l.svcCtx.ShellTool.RequiresApproval(shellArgs.Command) {
						approvalID := uuid.New().String()
						gateErr := l.waitForApproval(ctx, w, flusher, approvalID, shellArgs.Command)
						if gateErr != nil {
							resultContent = fmt.Sprintf("error: %s", gateErr.Error())
							// Audit the denial
							if l.svcCtx.AuditStore != nil {
								_ = l.svcCtx.AuditStore.Log(req.SessionId, tc.Function.Name, tc.Function.Arguments, "", resultContent)
							}
							// Emit tool_result with denial
							toolResultEvent, _ := json.Marshal(map[string]any{
								"type":    "tool_result",
								"id":      tc.ID,
								"content": resultContent,
							})
							fmt.Fprintf(w, "data: %s\n\n", toolResultEvent)
							flusher.Flush()
							// Append tool result to history
							history = append(history, openai.ChatCompletionMessage{
								Role:       openai.ChatMessageRoleTool,
								Content:    resultContent,
								ToolCallID: tc.ID,
							})
							continue
						}
					}
				}
			}

			// Execute the tool
			result := l.svcCtx.Executor.Execute(ctx, tc.Function.Name, tc.Function.Arguments)
			if result.Error != "" {
				resultContent = result.Error
			} else {
				resultContent = result.Content
			}
			// Audit the tool execution
			if l.svcCtx.AuditStore != nil {
				auditResult := result.Content
				if len(auditResult) > 2000 {
					auditResult = auditResult[:2000] + "[truncated]"
				}
				_ = l.svcCtx.AuditStore.Log(req.SessionId, tc.Function.Name, tc.Function.Arguments, auditResult, result.Error)
			}

			// Emit tool_result SSE event
			toolResultEvent, _ := json.Marshal(map[string]any{
				"type":    "tool_result",
				"id":      tc.ID,
				"content": resultContent,
			})
			fmt.Fprintf(w, "data: %s\n\n", toolResultEvent)
			flusher.Flush()

			// Append tool result to history
			history = append(history, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    resultContent,
				ToolCallID: tc.ID,
			})
		}
	}

	// Persist the full conversation turn to the store
	finalContent := lastAssistantContent.String()
	history = append(history, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: finalContent,
	})
	l.svcCtx.Store.Set(req.SessionId, history)

	// Create the conversation record for new sessions
	if isNewSession {
		runes := []rune(req.Message)
		if len(runes) > 50 {
			runes = runes[:50]
		}
		title := string(runes)
		_ = l.svcCtx.Store.CreateConversation(req.SessionId, title)
	}

	// Emit the done event with the session ID
	fmt.Fprintf(w, "data: {\"done\":true,\"sessionId\":\"%s\"}\n\n", req.SessionId)
	flusher.Flush()

	return nil
}

// waitForApproval registers an approval channel, emits an approval_request SSE event,
// and blocks until the user approves or denies, or the context is cancelled.
// Returns nil if approved, or an error if denied or context is done.
func (l *ChatLogic) waitForApproval(ctx context.Context, w http.ResponseWriter, flusher http.Flusher, approvalID, command string) error {
	ch := make(chan bool, 1)
	l.svcCtx.ApprovalStore.Register(approvalID, ch)
	defer l.svcCtx.ApprovalStore.Delete(approvalID)

	event, _ := json.Marshal(map[string]any{
		"type":       "approval_request",
		"approvalId": approvalID,
		"tool":       "shell_run",
		"command":    command,
	})
	fmt.Fprintf(w, "data: %s\n\n", event)
	flusher.Flush()

	select {
	case approved := <-ch:
		if !approved {
			return errors.New("user denied command execution")
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
