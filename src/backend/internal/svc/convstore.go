package svc

import (
	"sync"

	openai "github.com/sashabaranov/go-openai"
)

// Conversation represents a conversation metadata record.
type Conversation struct {
	ID        string
	Title     string
	CreatedAt int64
	UpdatedAt int64
}

// ConversationStore is the interface for storing and retrieving conversations.
type ConversationStore interface {
	Get(sessionID string) []openai.ChatCompletionMessage
	Set(sessionID string, msgs []openai.ChatCompletionMessage)
	ListConversations() ([]Conversation, error)
	GetConversation(id string) (*Conversation, error)
	DeleteConversation(id string) error
	CreateConversation(id, title string) error
}

// ConvStore is a thread-safe in-memory conversation store keyed by session ID.
type ConvStore struct {
	mu   sync.RWMutex
	data map[string][]openai.ChatCompletionMessage
}

// NewConvStore creates a new empty ConvStore.
func NewConvStore() *ConvStore {
	return &ConvStore{data: make(map[string][]openai.ChatCompletionMessage)}
}

// Get returns a copy of the messages for the given sessionID, or nil if not found.
func (s *ConvStore) Get(sessionID string) []openai.ChatCompletionMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	msgs := s.data[sessionID]
	if msgs == nil {
		return nil
	}
	out := make([]openai.ChatCompletionMessage, len(msgs))
	copy(out, msgs)
	return out
}

// Set replaces the stored messages for the given sessionID.
func (s *ConvStore) Set(sessionID string, msgs []openai.ChatCompletionMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[sessionID] = msgs
}

// ListConversations returns an empty list (stub for in-memory store).
func (s *ConvStore) ListConversations() ([]Conversation, error) {
	return []Conversation{}, nil
}

// GetConversation returns nil (stub for in-memory store).
func (s *ConvStore) GetConversation(id string) (*Conversation, error) {
	return nil, nil
}

// DeleteConversation is a no-op for in-memory store.
func (s *ConvStore) DeleteConversation(id string) error {
	return nil
}

// CreateConversation is a no-op for in-memory store.
func (s *ConvStore) CreateConversation(id, title string) error {
	return nil
}
