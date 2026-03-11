package toolexec

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// WebFetchTool tests
// ---------------------------------------------------------------------------

func TestWebFetchTool_EmptyURL(t *testing.T) {
	tool := NewWebFetchTool(30)
	result := tool.Fetch(context.Background(), `{"URL":""}`)
	assert.Equal(t, "url is required", result.Error)
	assert.Empty(t, result.Content)
}

func TestWebFetchTool_InvalidArgsJSON(t *testing.T) {
	tool := NewWebFetchTool(30)
	result := tool.Fetch(context.Background(), `not-json`)
	assert.Contains(t, result.Error, "invalid args:")
	assert.Empty(t, result.Content)
}

func TestWebFetchTool_FetchReturnsReadableText(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		const htmlBody = `<!DOCTYPE html>
<html>
<head><title>Test Article Title</title></head>
<body>
  <article>
    <h1>Test Article Title</h1>
    <p>This is the article body with some readable content about Go programming.</p>
    <p>Another paragraph with more information.</p>
  </article>
</body>
</html>`
		fmt.Fprint(w, htmlBody) //nolint:errcheck // test handler; write errors are not relevant
	}))
	defer srv.Close()

	tool := NewWebFetchTool(30)
	argsJSON, _ := json.Marshal(map[string]string{"URL": srv.URL})
	result := tool.Fetch(context.Background(), string(argsJSON))

	require.Empty(t, result.Error, "expected no error, got: %s", result.Error)
	assert.NotEmpty(t, result.Content)
	// Should contain readable text (title or article text)
	assert.True(t,
		strings.Contains(result.Content, "Test Article Title") ||
			strings.Contains(result.Content, "readable content"),
		"expected content to contain article text, got: %s", result.Content)
	// Should NOT contain raw HTML tags
	assert.NotContains(t, result.Content, "<html>")
	assert.NotContains(t, result.Content, "<body>")
	assert.NotContains(t, result.Content, "<div>")
	assert.NotContains(t, result.Content, "<script>")
}

func TestWebFetchTool_InvalidURL(t *testing.T) {
	tool := NewWebFetchTool(5)
	argsJSON, _ := json.Marshal(map[string]string{"URL": "http://localhost:1"})
	result := tool.Fetch(context.Background(), string(argsJSON))

	assert.NotEmpty(t, result.Error)
	assert.Contains(t, result.Error, "fetch failed:")
}

func TestWebFetchTool_DefaultTimeout(t *testing.T) {
	// timeoutSeconds <= 0 should use default of 30s
	tool := NewWebFetchTool(0)
	assert.Equal(t, int64(30), int64(tool.timeout.Seconds()))
}

// ---------------------------------------------------------------------------
// WebSearchTool tests
// ---------------------------------------------------------------------------

func TestWebSearchTool_EmptyAPIKey(t *testing.T) {
	tool := NewWebSearchTool("")
	result := tool.Search(context.Background(), `{"Query":"golang"}`)
	assert.Equal(t, "web_search not configured: set BraveSearchAPIKey in config", result.Error)
	assert.Empty(t, result.Content)
}

func TestWebSearchTool_InvalidArgsJSON(t *testing.T) {
	tool := NewWebSearchTool("some-key")
	result := tool.Search(context.Background(), `not-json`)
	assert.Contains(t, result.Error, "invalid args:")
	assert.Empty(t, result.Content)
}

func TestWebSearchTool_SuccessfulSearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		assert.Equal(t, "test-api-key", r.Header.Get("X-Subscription-Token"))
		assert.Contains(t, r.Header.Get("Accept"), "application/json")

		// Return mock Brave API response
		resp := braveSearchResponse{}
		resp.Web.Results = []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
		}{
			{Title: "Go Programming Language", URL: "https://golang.org", Description: "The Go programming language website."},
			{Title: "Go Tour", URL: "https://tour.golang.org", Description: "An interactive tour of Go."},
			{Title: "Go Blog", URL: "https://go.dev/blog", Description: "The Go blog."},
			{Title: "Go Packages", URL: "https://pkg.go.dev", Description: "Go package documentation."},
			{Title: "Go Playground", URL: "https://play.golang.org", Description: "Try Go in your browser."},
		}
		json.NewEncoder(w).Encode(resp) //nolint:errcheck // test handler; encode errors not relevant
	}))
	defer srv.Close()

	tool := NewWebSearchTool("test-api-key")
	tool.baseURL = srv.URL + "/res/v1/web/search"

	argsJSON, _ := json.Marshal(map[string]string{"Query": "golang"})
	result := tool.Search(context.Background(), string(argsJSON))

	require.Empty(t, result.Error, "expected no error, got: %s", result.Error)
	assert.Contains(t, result.Content, "1. Go Programming Language")
	assert.Contains(t, result.Content, "https://golang.org")
	assert.Contains(t, result.Content, "The Go programming language website.")
	assert.Contains(t, result.Content, "2. Go Tour")
	assert.Contains(t, result.Content, "5. Go Playground")
}

func TestWebSearchTool_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	tool := NewWebSearchTool("bad-key")
	tool.baseURL = srv.URL + "/res/v1/web/search"

	argsJSON, _ := json.Marshal(map[string]string{"Query": "golang"})
	result := tool.Search(context.Background(), string(argsJSON))

	assert.Equal(t, "search API error: 401", result.Error)
	assert.Empty(t, result.Content)
}

func TestWebSearchTool_EmptyResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := braveSearchResponse{}
		json.NewEncoder(w).Encode(resp) //nolint:errcheck // test handler; encode errors not relevant
	}))
	defer srv.Close()

	tool := NewWebSearchTool("test-key")
	tool.baseURL = srv.URL + "/res/v1/web/search"

	argsJSON, _ := json.Marshal(map[string]string{"Query": "no results query"})
	result := tool.Search(context.Background(), string(argsJSON))

	require.Empty(t, result.Error)
	assert.Empty(t, result.Content)
}
