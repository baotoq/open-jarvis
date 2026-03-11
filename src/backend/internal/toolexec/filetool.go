package toolexec

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileTool provides read_file and write_file operations confined to a workspace root.
type FileTool struct {
	workspaceRoot string
}

// NewFileTool returns a FileTool scoped to workspaceRoot.
// workspaceRoot is resolved to an absolute path at construction time.
func NewFileTool(workspaceRoot string) *FileTool {
	abs, err := filepath.Abs(workspaceRoot)
	if err != nil {
		abs = workspaceRoot
	}
	return &FileTool{workspaceRoot: abs}
}

// safePath resolves userPath relative to the workspace root and confirms the
// result stays within the root (path traversal guard).
func (f *FileTool) safePath(userPath string) (string, error) {
	abs, err := filepath.Abs(filepath.Join(f.workspaceRoot, userPath))
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	// Add trailing separator to root so /tmp does not prefix-match /tmpother.
	root := filepath.Clean(f.workspaceRoot) + string(filepath.Separator)
	if !strings.HasPrefix(abs+string(filepath.Separator), root) {
		return "", fmt.Errorf("path escapes workspace root")
	}
	return abs, nil
}

// ReadFile reads a file at args.Path and returns its content.
// JSON args: {"path":"relative/path"}
func (f *FileTool) ReadFile(_ context.Context, argsJSON string) ToolResult {
	var args struct{ Path string }
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return ToolResult{Error: "invalid args: " + err.Error()}
	}
	safe, err := f.safePath(args.Path)
	if err != nil {
		return ToolResult{Error: err.Error()}
	}
	data, err := os.ReadFile(safe)
	if err != nil {
		return ToolResult{Error: err.Error()}
	}
	return ToolResult{Content: string(data)}
}

// WriteFile writes content to args.Path inside the workspace root.
// JSON args: {"path":"relative/path","content":"..."}
func (f *FileTool) WriteFile(_ context.Context, argsJSON string) ToolResult {
	var args struct {
		Path    string
		Content string
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return ToolResult{Error: "invalid args: " + err.Error()}
	}
	safe, err := f.safePath(args.Path)
	if err != nil {
		return ToolResult{Error: err.Error()}
	}
	if err := os.WriteFile(safe, []byte(args.Content), 0644); err != nil {
		return ToolResult{Error: err.Error()}
	}
	return ToolResult{Content: "written"}
}
