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
