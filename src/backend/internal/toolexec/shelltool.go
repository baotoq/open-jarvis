package toolexec

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// ShellTool executes shell commands with allowlist/denylist approval logic.
type ShellTool struct {
	allowlist []string
	denylist  []string
}

// NewShellTool returns a ShellTool with the provided allowlist and denylist glob patterns.
// Pass nil (or empty) slices to use defaults (all commands require approval).
func NewShellTool(allowlist, denylist []string) *ShellTool {
	return &ShellTool{allowlist: allowlist, denylist: denylist}
}

// RequiresApproval reports whether command requires human approval before execution.
// Rules (evaluated in order):
//  1. Denylist pattern match → true (requires approval)
//  2. Allowlist pattern match → false (auto-approved)
//  3. No match → true (default: requires approval)
//
// Patterns use filepath.Match glob syntax.
func (s *ShellTool) RequiresApproval(command string) bool {
	for _, pattern := range s.denylist {
		if matched, _ := filepath.Match(pattern, command); matched {
			return true
		}
	}
	for _, pattern := range s.allowlist {
		if matched, _ := filepath.Match(pattern, command); matched {
			return false
		}
	}
	return true
}

// Run executes command and returns combined stdout+stderr.
// JSON args: {"command":"cmd arg1 arg2"}
// Returns an error in ToolResult.Error if the command fails; partial output is
// still included in ToolResult.Content so callers can surface it.
func (s *ShellTool) Run(ctx context.Context, argsJSON string) ToolResult {
	var args struct{ Command string }
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return ToolResult{Error: "invalid args: " + err.Error()}
	}
	parts := strings.Fields(args.Command)
	if len(parts) == 0 {
		return ToolResult{Error: "empty command"}
	}
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return ToolResult{Content: out.String(), Error: fmt.Sprintf("command failed: %v", err)}
	}
	return ToolResult{Content: out.String()}
}
