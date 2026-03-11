package types

// ChatRequest represents a single chat message from the client.
type ChatRequest struct {
	SessionId string `json:"sessionId"`
	Message   string `json:"message"`
}

// ConversationResponse represents a conversation record in API responses.
type ConversationResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

// MessageResponse represents a single chat message in API responses.
type MessageResponse struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ApproveRequest is the request body for POST /api/chat/approve.
type ApproveRequest struct {
	ApprovalID string `json:"approvalId"`
	Approved   bool   `json:"approved"`
}

// ConfigResponse is the response for GET /api/config.
type ConfigResponse struct {
	BaseURL      string `json:"baseURL"`
	Name         string `json:"name"`
	APIKey       string `json:"apiKey"`
	SystemPrompt string `json:"systemPrompt"`
}

// UpdateConfigRequest is the request body for PUT /api/config.
type UpdateConfigRequest struct {
	BaseURL      string `json:"baseURL"`
	Name         string `json:"name"`
	APIKey       string `json:"apiKey"`
	SystemPrompt string `json:"systemPrompt"`
}

// SearchConvsRequest is parsed from GET /api/conversations/search?q=<term>.
type SearchConvsRequest struct {
	Query string `form:"q"`
}

// SearchResult is one matching conversation in a search response.
type SearchResult struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	UpdatedAt int64  `json:"updatedAt"`
	Snippet   string `json:"snippet"`
}
