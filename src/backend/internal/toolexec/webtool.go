package toolexec

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	readability "github.com/go-shiori/go-readability"
)

// WebFetchTool fetches a URL and returns readable article text using go-readability.
// NOTE: go-shiori/go-readability is deprecated in favour of codeberg.org/readeck/go-readability/v2
// but remains functional and is listed on pkg.go.dev. Migrate in a future phase if needed.
type WebFetchTool struct {
	timeout time.Duration
}

// NewWebFetchTool returns a WebFetchTool with the given timeout.
// If timeoutSeconds <= 0, defaults to 30 seconds.
func NewWebFetchTool(timeoutSeconds int) *WebFetchTool {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}
	return &WebFetchTool{timeout: time.Duration(timeoutSeconds) * time.Second}
}

// Fetch retrieves the URL in args and returns article title + plain text.
// JSON args: {"URL":"https://..."}
func (w *WebFetchTool) Fetch(_ context.Context, argsJSON string) ToolResult {
	var args struct{ URL string }
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return ToolResult{Error: "invalid args: " + err.Error()}
	}
	if args.URL == "" {
		return ToolResult{Error: "url is required"}
	}
	article, err := readability.FromURL(args.URL, w.timeout)
	if err != nil {
		return ToolResult{Error: "fetch failed: " + err.Error()}
	}
	content := article.TextContent
	if len(content) > 8000 {
		content = content[:8000] + "\n[truncated]"
	}
	result := article.Title
	if result != "" {
		result += "\n\n"
	}
	result += content
	return ToolResult{Content: result}
}

type braveSearchResponse struct {
	Web struct {
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
		} `json:"results"`
	} `json:"web"`
}

// WebSearchTool searches the web using the Brave Search API.
// Requires BraveSearchAPIKey set in config. Returns ToolResult{Error:...} if key is empty.
type WebSearchTool struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewWebSearchTool returns a WebSearchTool configured with the given API key.
// Pass an empty string to create an unconfigured tool (returns error on Search).
func NewWebSearchTool(apiKey string) *WebSearchTool {
	return &WebSearchTool{
		apiKey:  apiKey,
		baseURL: "https://api.search.brave.com/res/v1/web/search",
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

// Search queries Brave Search API and returns numbered results.
// JSON args: {"Query":"search terms"}
func (w *WebSearchTool) Search(ctx context.Context, argsJSON string) ToolResult {
	var args struct{ Query string }
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return ToolResult{Error: "invalid args: " + err.Error()}
	}
	if w.apiKey == "" {
		return ToolResult{Error: "web_search not configured: set BraveSearchAPIKey in config"}
	}
	endpoint := w.baseURL + "?q=" + url.QueryEscape(args.Query) + "&count=5"
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return ToolResult{Error: "request error: " + err.Error()}
	}
	req.Header.Set("X-Subscription-Token", w.apiKey)
	req.Header.Set("Accept", "application/json")
	resp, err := w.client.Do(req)
	if err != nil {
		return ToolResult{Error: "search failed: " + err.Error()}
	}
	defer resp.Body.Close() //nolint:errcheck // cleanup in defer; error logged by http client
	if resp.StatusCode != http.StatusOK {
		return ToolResult{Error: fmt.Sprintf("search API error: %d", resp.StatusCode)}
	}
	var result braveSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ToolResult{Error: "decode error: " + err.Error()}
	}
	var sb strings.Builder
	for i, r := range result.Web.Results {
		fmt.Fprintf(&sb, "%d. %s\n   %s\n   %s\n\n", i+1, r.Title, r.URL, r.Description)
	}
	return ToolResult{Content: sb.String()}
}
