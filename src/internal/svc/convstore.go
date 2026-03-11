package svc

import (
	"sync"

	openai "github.com/sashabaranov/go-openai"
)

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
