package svc_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"open-jarvis/internal/svc"
)

func TestApprovalStore_RegisterResolve(t *testing.T) {
	store := svc.NewApprovalStore()
	ch := make(chan bool, 1)
	store.Register("req-1", ch)

	ok := store.Resolve("req-1", true)
	require.True(t, ok, "Resolve should return true for known ID")

	got := <-ch
	assert.True(t, got, "channel should receive the approved value")
}

func TestApprovalStore_ResolveUnknown(t *testing.T) {
	store := svc.NewApprovalStore()

	// Resolve on unknown ID must return false without panicking
	ok := store.Resolve("unknown-id", true)
	assert.False(t, ok, "Resolve should return false for unknown ID")
}

func TestApprovalStore_Delete(t *testing.T) {
	store := svc.NewApprovalStore()
	ch := make(chan bool, 1)
	store.Register("req-2", ch)

	// Delete then resolve should fail
	store.Delete("req-2")
	ok := store.Resolve("req-2", true)
	assert.False(t, ok, "Resolve after Delete should return false")
}

func TestApprovalStore_Concurrent(t *testing.T) {
	store := svc.NewApprovalStore()
	const n = 10
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := string(rune('a' + i)) // unique IDs: "a", "b", ...
			ch := make(chan bool, 1)
			store.Register(id, ch)
			ok := store.Resolve(id, false)
			assert.True(t, ok, "concurrent Resolve should succeed for registered ID")
			<-ch
		}(i)
	}

	wg.Wait()
}
