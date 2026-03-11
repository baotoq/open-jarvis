package toolexec

import "context"

// ToolResult is the return value of every tool invocation.
// Content holds the tool output on success; Error is non-empty on failure.
type ToolResult struct {
	Content string
	Error   string
}

// Executor is the interface satisfied by any tool dispatcher.
type Executor interface {
	Execute(ctx context.Context, name, argsJSON string) ToolResult
}

// ToolRegistry implements Executor by dispatching to registered tool functions
// by name.
type ToolRegistry struct {
	tools map[string]func(context.Context, string) ToolResult
}

// NewRegistry returns an empty ToolRegistry ready for tool registration.
func NewRegistry() *ToolRegistry {
	return &ToolRegistry{tools: make(map[string]func(context.Context, string) ToolResult)}
}

// Register adds a named tool function to the registry.
func (r *ToolRegistry) Register(name string, fn func(context.Context, string) ToolResult) {
	r.tools[name] = fn
}

// Execute dispatches the named tool call with argsJSON.
// Returns ToolResult{Error: "unknown tool: <name>"} when no tool is registered.
func (r *ToolRegistry) Execute(ctx context.Context, name, argsJSON string) ToolResult {
	fn, ok := r.tools[name]
	if !ok {
		return ToolResult{Error: "unknown tool: " + name}
	}
	return fn(ctx, argsJSON)
}
