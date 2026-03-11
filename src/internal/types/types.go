package types

// ChatRequest represents a single chat message from the client.
type ChatRequest struct {
	SessionId string `json:"sessionId"`
	Message   string `json:"message"`
}
