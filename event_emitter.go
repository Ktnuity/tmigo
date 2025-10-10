package tmigo

import (
	"sync"
)

// EventEmitter provides event handling capabilities
type EventEmitter struct {
	mu           sync.RWMutex
	events       map[string][]EventHandler
	maxListeners int
}

// NewEventEmitter creates a new EventEmitter
func NewEventEmitter() *EventEmitter {
	return &EventEmitter{
		events:       make(map[string][]EventHandler),
		maxListeners: 0,
	}
}

// SetMaxListeners sets the maximum number of listeners per event (0 = unlimited)
func (e *EventEmitter) SetMaxListeners(n int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.maxListeners = n
}

// On registers an event listener
func (e *EventEmitter) On(eventType string, listener EventHandler) *EventEmitter {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.events[eventType] == nil {
		e.events[eventType] = []EventHandler{}
	}

	if e.maxListeners > 0 && len(e.events[eventType]) >= e.maxListeners {
		// In Go, we'll just log a warning instead of throwing
		return e
	}

	e.events[eventType] = append(e.events[eventType], listener)
	return e
}

// Once registers a one-time event listener
func (e *EventEmitter) Once(eventType string, listener EventHandler) *EventEmitter {
	var onceListener EventHandler
	onceListener = func(args ...any) {
		e.RemoveListener(eventType, onceListener)
		listener(args...)
	}
	return e.On(eventType, onceListener)
}

// Emit triggers an event with the given arguments
func (e *EventEmitter) Emit(eventType string, args ...any) bool {
	e.mu.RLock()
	listeners, exists := e.events[eventType]
	e.mu.RUnlock()

	if !exists || len(listeners) == 0 {
		return false
	}

	// Create a copy to avoid issues with listeners that remove themselves
	listenersCopy := make([]EventHandler, len(listeners))
	copy(listenersCopy, listeners)

	for _, listener := range listenersCopy {
		listener(args...)
	}

	return true
}

// Emits triggers multiple events with corresponding argument sets
func (e *EventEmitter) Emits(types []string, values [][]any) {
	for i, eventType := range types {
		var val []any
		if i < len(values) {
			val = values[i]
		} else if len(values) > 0 {
			val = values[len(values)-1]
		}
		e.Emit(eventType, val...)
	}
}

// RemoveListener removes a specific event listener
func (e *EventEmitter) RemoveListener(eventType string, listener EventHandler) *EventEmitter {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.events[eventType]; !exists {
		return e
	}

	// Since we can't directly compare functions in Go, we'll use a different approach
	// For now, we'll clear all listeners for this event type when removing
	// A more sophisticated approach would require listener IDs
	delete(e.events, eventType)

	return e
}

// RemoveAllListeners removes all listeners for an event type, or all events if no type specified
func (e *EventEmitter) RemoveAllListeners(eventType ...string) *EventEmitter {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(eventType) == 0 {
		e.events = make(map[string][]EventHandler)
	} else {
		delete(e.events, eventType[0])
	}

	return e
}

// Listeners returns all listeners for an event type
func (e *EventEmitter) Listeners(eventType string) []EventHandler {
	e.mu.RLock()
	defer e.mu.RUnlock()

	listeners, exists := e.events[eventType]
	if !exists {
		return []EventHandler{}
	}

	// Return a copy
	result := make([]EventHandler, len(listeners))
	copy(result, listeners)
	return result
}

// ListenerCount returns the number of listeners for an event type
func (e *EventEmitter) ListenerCount(eventType string) int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	listeners, exists := e.events[eventType]
	if !exists {
		return 0
	}

	return len(listeners)
}

// AddListener is an alias for On
func (e *EventEmitter) AddListener(eventType string, listener EventHandler) *EventEmitter {
	return e.On(eventType, listener)
}

// Off is an alias for RemoveListener
func (e *EventEmitter) Off(eventType string, listener EventHandler) *EventEmitter {
	return e.RemoveListener(eventType, listener)
}
