package svc_test

import (
	"sync"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"open-jarvis/internal/svc"
)

func TestConvStoreGetEmpty(t *testing.T) {
	store := svc.NewConvStore()
	result := store.Get("nonexistent-session")
	assert.Nil(t, result)
}

func TestConvStoreSetGet(t *testing.T) {
	store := svc.NewConvStore()
	msgs := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "hello"},
		{Role: openai.ChatMessageRoleAssistant, Content: "world"},
	}
	store.Set("session-1", msgs)

	result := store.Get("session-1")
	assert.Len(t, result, 2)
	assert.Equal(t, "hello", result[0].Content)
	assert.Equal(t, "world", result[1].Content)
}

func TestConvStoreConcurrent(t *testing.T) {
	store := svc.NewConvStore()
	keys := []string{"session-a", "session-b", "session-c"}
	msgs := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "concurrent"},
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := keys[i%len(keys)]
			store.Set(key, msgs)
			store.Get(key)
		}(i)
	}
	wg.Wait()
}

func TestConvStoreCopyIsolation(t *testing.T) {
	store := svc.NewConvStore()
	original := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "original"},
	}
	store.Set("session-copy", original)

	// Mutate the returned slice
	got := store.Get("session-copy")
	got = append(got, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: "extra"})

	// The stored slice should still have 1 message
	stored := store.Get("session-copy")
	assert.Len(t, stored, 1, "stored slice should not be affected by mutating returned copy")
}
