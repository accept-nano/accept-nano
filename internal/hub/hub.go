package hub

// This file is copied from https://github.com/cenkalti/hub/blob/master/hub.go then modified for the project.

import "sync"

type Key = string

// Event is an interface for published events.
type Event interface {
	Key() Key
}

// Hub is an event dispatcher, publishes events to the subscribers
// which are subscribed for a specific event type.
// Optimized for publish calls.
// The handlers may be called in order different than they are registered.
type Hub struct {
	subscribers map[Key][]handler
	m           sync.RWMutex
	seq         uint64
}

type handler struct {
	f  func(Event)
	id uint64
}

// Subscribe registers f for the event of a specific key.
func (h *Hub) Subscribe(key Key, f func(Event)) (cancel func()) {
	var cancelled bool
	h.m.Lock()
	h.seq++
	id := h.seq
	if h.subscribers == nil {
		h.subscribers = make(map[Key][]handler)
	}
	h.subscribers[key] = append(h.subscribers[key], handler{id: id, f: f})
	h.m.Unlock()
	return func() {
		h.m.Lock()
		if cancelled {
			h.m.Unlock()
			return
		}
		cancelled = true
		a := h.subscribers[key]
		for i, f := range a {
			if f.id == id {
				a[i], h.subscribers[key] = a[len(a)-1], a[:len(a)-1]
				break
			}
		}
		if len(a) == 0 {
			delete(h.subscribers, key)
		}
		h.m.Unlock()
	}
}

// Publish an event to the subscribers.
func (h *Hub) Publish(e Event) {
	h.m.RLock()
	if handlers, ok := h.subscribers[e.Key()]; ok {
		for _, h := range handlers {
			h.f(e)
		}
	}
	h.m.RUnlock()
}
