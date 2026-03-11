package svc

import "sync"

// ApprovalStore holds pending approval channels keyed by approval ID.
// Thread-safe for concurrent use from multiple goroutines.
type ApprovalStore struct {
	mu      sync.Mutex
	pending map[string]chan bool
}

// NewApprovalStore returns an initialised ApprovalStore.
func NewApprovalStore() *ApprovalStore {
	return &ApprovalStore{pending: make(map[string]chan bool)}
}

// Register stores ch under id. The ChatLogic goroutine calls this before
// emitting an approval_request SSE event.
func (a *ApprovalStore) Register(id string, ch chan bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.pending[id] = ch
}

// Resolve sends approved to the channel registered under id.
// Returns false if id is not found (already deleted or never registered).
func (a *ApprovalStore) Resolve(id string, approved bool) bool {
	a.mu.Lock()
	ch, ok := a.pending[id]
	a.mu.Unlock()
	if !ok {
		return false
	}
	ch <- approved
	return true
}

// Delete removes the channel registered under id.
// Called via defer in ChatLogic after the approval gate completes or times out.
func (a *ApprovalStore) Delete(id string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.pending, id)
}
