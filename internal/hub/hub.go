package hub

// This file is copied from https://github.com/cenkalti/hub/blob/master/hub.go then modified for the project.

import "sync"

type Account string

// Event is an interface for published events.
type Event interface {
	Account() Account
}

// Hub is an event dispatcher, publishes events to the subscribers
// which are subscribed for a specific event type.
// Optimized for publish calls.
// The handlers may be called in order different than they are registered.
type Hub struct {
	subscribers map[Account][]handler
	m           sync.RWMutex
	seq         uint64
}

type handler struct {
	f  func(Event)
	id uint64
}

// Subscribe registers f for the event of a specific account.
func (h *Hub) Subscribe(account Account, f func(Event)) (cancel func()) {
	var cancelled bool
	h.m.Lock()
	h.seq++
	id := h.seq
	if h.subscribers == nil {
		h.subscribers = make(map[Account][]handler)
	}
	h.subscribers[account] = append(h.subscribers[account], handler{id: id, f: f})
	h.m.Unlock()
	return func() {
		h.m.Lock()
		if cancelled {
			h.m.Unlock()
			return
		}
		cancelled = true
		a := h.subscribers[account]
		for i, f := range a {
			if f.id == id {
				a[i], h.subscribers[account] = a[len(a)-1], a[:len(a)-1]
				break
			}
		}
		if len(a) == 0 {
			delete(h.subscribers, account)
		}
		h.m.Unlock()
	}
}

// Publish an event to the subscribers.
func (h *Hub) Publish(e Event) {
	h.m.RLock()
	if handlers, ok := h.subscribers[e.Account()]; ok {
		for _, h := range handlers {
			h.f(e)
		}
	}
	h.m.RUnlock()
}
