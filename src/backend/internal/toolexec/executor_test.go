package toolexec_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"open-jarvis/internal/toolexec"
)

// ---------------------------------------------------------------------------
// Executor / ToolRegistry tests
// ---------------------------------------------------------------------------

func TestToolRegistry_UnknownTool(t *testing.T) {
	reg := toolexec.NewRegistry()
	result := reg.Execute(context.Background(), "nonexistent", "{}")
	assert.NotEmpty(t, result.Error)
	assert.Contains(t, result.Error, "unknown tool: nonexistent")
}

func TestToolRegistry_RegisterAndDispatch(t *testing.T) {
	reg := toolexec.NewRegistry()
	reg.Register("echo", func(ctx context.Context, argsJSON string) toolexec.ToolResult {
		return toolexec.ToolResult{Content: "echoed: " + argsJSON}
	})
	result := reg.Execute(context.Background(), "echo", `{"msg":"hi"}`)
	assert.Empty(t, result.Error)
	assert.Equal(t, `echoed: {"msg":"hi"}`, result.Content)
}

// ---------------------------------------------------------------------------
// FileTool tests
// ---------------------------------------------------------------------------

func TestFileTool_ReadFile(t *testing.T) {
	dir := t.TempDir()
	ft := toolexec.NewFileTool(dir)

	// Write a file to read back
	require.NoError(t, os.WriteFile(dir+"/hello.txt", []byte("hello world"), 0644))

	t.Run("happy path", func(t *testing.T) {
		result := ft.ReadFile(context.Background(), `{"path":"hello.txt"}`)
		assert.Empty(t, result.Error)
		assert.Equal(t, "hello world", result.Content)
	})

	t.Run("file not found", func(t *testing.T) {
		result := ft.ReadFile(context.Background(), `{"path":"missing.txt"}`)
		assert.NotEmpty(t, result.Error)
	})
}

func TestFileTool_WriteFile(t *testing.T) {
	dir := t.TempDir()
	ft := toolexec.NewFileTool(dir)

	t.Run("happy path", func(t *testing.T) {
		result := ft.WriteFile(context.Background(), `{"path":"out.txt","content":"written content"}`)
		assert.Empty(t, result.Error)
		assert.Equal(t, "written", result.Content)

		// verify written to disk
		readResult := ft.ReadFile(context.Background(), `{"path":"out.txt"}`)
		assert.Empty(t, readResult.Error)
		assert.Equal(t, "written content", readResult.Content)
	})
}

func TestFileTool_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	ft := toolexec.NewFileTool(dir)

	tests := []struct {
		name string
		path string
	}{
		{"parent traversal", "../../etc/passwd"},
		{"mixed traversal", "subdir/../../etc/passwd"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			argsJSON := `{"path":"` + tc.path + `"}`
			result := ft.ReadFile(context.Background(), argsJSON)
			assert.NotEmpty(t, result.Error)
			assert.Contains(t, result.Error, "escapes")
		})
	}
}

// ---------------------------------------------------------------------------
// ShellTool tests
// ---------------------------------------------------------------------------

func TestShellTool_Run(t *testing.T) {
	st := toolexec.NewShellTool(nil, nil)

	t.Run("echo hello", func(t *testing.T) {
		result := st.Run(context.Background(), `{"command":"echo hello"}`)
		assert.Empty(t, result.Error)
		assert.Equal(t, "hello\n", result.Content)
	})

	t.Run("empty command", func(t *testing.T) {
		result := st.Run(context.Background(), `{"command":""}`)
		assert.NotEmpty(t, result.Error)
		assert.Contains(t, result.Error, "empty command")
	})
}

func TestShellTool_Allowlist(t *testing.T) {
	tests := []struct {
		name      string
		allowlist []string
		denylist  []string
		command   string
		want      bool // true = requires approval
	}{
		{
			name:    "no lists defaults to requires approval",
			command: "echo hello",
			want:    true,
		},
		{
			name:      "allowlist match auto-approves",
			allowlist: []string{"echo *"},
			command:   "echo hello",
			want:      false,
		},
		{
			name:     "denylist match requires approval",
			denylist: []string{"rm *"},
			command:  "rm -rf /",
			want:     true,
		},
		{
			name:      "denylist overrides allowlist",
			allowlist: []string{"rm *"},
			denylist:  []string{"rm *"},
			command:   "rm -rf /",
			want:      true,
		},
		{
			name:      "allowlist no match requires approval",
			allowlist: []string{"echo *"},
			command:   "cat file.txt",
			want:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			st := toolexec.NewShellTool(tc.allowlist, tc.denylist)
			got := st.RequiresApproval(tc.command)
			assert.Equal(t, tc.want, got)
		})
	}
}
