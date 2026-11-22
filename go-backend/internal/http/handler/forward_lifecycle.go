package handler

import "sync"

type forwardLifecycleLock struct {
	mu    sync.Mutex
	users int
}

// lockForwardLifecycle keeps one forward's persistence, runtime apply and
// rollback ordered so an older failed request cannot undo a newer success.
func (h *Handler) lockForwardLifecycle(id int64) func() {
	if h == nil || id <= 0 {
		return func() {}
	}
	h.forwardLifecycleMu.Lock()
	if h.forwardLifecycleLocks == nil {
		h.forwardLifecycleLocks = make(map[int64]*forwardLifecycleLock)
	}
	lock := h.forwardLifecycleLocks[id]
	if lock == nil {
		lock = &forwardLifecycleLock{}
		h.forwardLifecycleLocks[id] = lock
	}
	lock.users++
	h.forwardLifecycleMu.Unlock()

	lock.mu.Lock()
	return func() {
		h.forwardLifecycleMu.Lock()
		lock.mu.Unlock()
		lock.users--
		if lock.users == 0 {
			delete(h.forwardLifecycleLocks, id)
		}
		h.forwardLifecycleMu.Unlock()
	}
}
