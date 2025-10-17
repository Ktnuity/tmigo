package tmigo

import (
	"sync"
	"testing"
	"time"
)

func TestNewEventEmitter(t *testing.T) {
	ee := NewEventEmitter()
	if ee == nil {
		t.Fatal("NewEventEmitter() returned nil")
	}
	if ee.events == nil {
		t.Error("events map not initialized")
	}
}

func TestEventEmitter_On(t *testing.T) {
	ee := NewEventEmitter()
	called := false

	handler := func(args ...any) {
		called = true
	}

	ee.On("test", handler)

	if ee.ListenerCount("test") != 1 {
		t.Errorf("ListenerCount = %d, want 1", ee.ListenerCount("test"))
	}

	ee.Emit("test")

	if !called {
		t.Error("Event handler was not called")
	}
}

func TestEventEmitter_Emit(t *testing.T) {
	ee := NewEventEmitter()
	var receivedArgs []any

	handler := func(args ...any) {
		receivedArgs = args
	}

	ee.On("test", handler)

	result := ee.Emit("test", "arg1", 42, true)

	if !result {
		t.Error("Emit should return true when listeners exist")
	}

	if len(receivedArgs) != 3 {
		t.Fatalf("Expected 3 args, got %d", len(receivedArgs))
	}

	if receivedArgs[0] != "arg1" || receivedArgs[1] != 42 || receivedArgs[2] != true {
		t.Errorf("Args = %v, want [arg1 42 true]", receivedArgs)
	}
}

func TestEventEmitter_EmitNoListeners(t *testing.T) {
	ee := NewEventEmitter()

	result := ee.Emit("nonexistent")

	if result {
		t.Error("Emit should return false when no listeners exist")
	}
}

func TestEventEmitter_MultipleListeners(t *testing.T) {
	ee := NewEventEmitter()
	callCount := 0
	var mu sync.Mutex

	handler1 := func(args ...any) {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	handler2 := func(args ...any) {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	ee.On("test", handler1)
	ee.On("test", handler2)

	if ee.ListenerCount("test") != 2 {
		t.Errorf("ListenerCount = %d, want 2", ee.ListenerCount("test"))
	}

	ee.Emit("test")

	mu.Lock()
	count := callCount
	mu.Unlock()

	if count != 2 {
		t.Errorf("callCount = %d, want 2", count)
	}
}

func TestEventEmitter_Once(t *testing.T) {
	ee := NewEventEmitter()
	callCount := 0

	handler := func(args ...any) {
		callCount++
	}

	ee.Once("test", handler)

	ee.Emit("test")
	ee.Emit("test")
	ee.Emit("test")

	if callCount != 1 {
		t.Errorf("callCount = %d, want 1 (handler should only be called once)", callCount)
	}

	// Listener should be removed after first call
	if ee.ListenerCount("test") != 0 {
		t.Errorf("ListenerCount = %d, want 0 after once handler executed", ee.ListenerCount("test"))
	}
}

func TestEventEmitter_RemoveAllListeners(t *testing.T) {
	ee := NewEventEmitter()

	ee.On("test1", func(args ...any) {})
	ee.On("test1", func(args ...any) {})
	ee.On("test2", func(args ...any) {})

	if ee.ListenerCount("test1") != 2 {
		t.Errorf("ListenerCount(test1) = %d, want 2", ee.ListenerCount("test1"))
	}

	ee.RemoveAllListeners("test1")

	if ee.ListenerCount("test1") != 0 {
		t.Errorf("ListenerCount(test1) = %d, want 0 after removal", ee.ListenerCount("test1"))
	}

	if ee.ListenerCount("test2") != 1 {
		t.Errorf("ListenerCount(test2) = %d, want 1 (should not be affected)", ee.ListenerCount("test2"))
	}
}

func TestEventEmitter_RemoveAllListenersNoArg(t *testing.T) {
	ee := NewEventEmitter()

	ee.On("test1", func(args ...any) {})
	ee.On("test2", func(args ...any) {})
	ee.On("test3", func(args ...any) {})

	ee.RemoveAllListeners()

	if ee.ListenerCount("test1") != 0 {
		t.Error("All listeners should be removed")
	}
	if ee.ListenerCount("test2") != 0 {
		t.Error("All listeners should be removed")
	}
	if ee.ListenerCount("test3") != 0 {
		t.Error("All listeners should be removed")
	}
}

func TestEventEmitter_SetMaxListeners(t *testing.T) {
	ee := NewEventEmitter()
	ee.SetMaxListeners(2)

	ee.On("test", func(args ...any) {})
	ee.On("test", func(args ...any) {})
	ee.On("test", func(args ...any) {})

	// With max listeners set, the third one should not be added
	if ee.ListenerCount("test") != 2 {
		t.Errorf("ListenerCount = %d, want 2 (max listeners enforced)", ee.ListenerCount("test"))
	}
}

func TestEventEmitter_SetMaxListenersZero(t *testing.T) {
	ee := NewEventEmitter()
	ee.SetMaxListeners(0) // Unlimited

	for i := 0; i < 100; i++ {
		ee.On("test", func(args ...any) {})
	}

	if ee.ListenerCount("test") != 100 {
		t.Errorf("ListenerCount = %d, want 100 (unlimited)", ee.ListenerCount("test"))
	}
}

func TestEventEmitter_Listeners(t *testing.T) {
	ee := NewEventEmitter()

	handler1 := func(args ...any) {}
	handler2 := func(args ...any) {}

	ee.On("test", handler1)
	ee.On("test", handler2)

	listeners := ee.Listeners("test")

	if len(listeners) != 2 {
		t.Errorf("len(Listeners) = %d, want 2", len(listeners))
	}
}

func TestEventEmitter_ListenersNonexistent(t *testing.T) {
	ee := NewEventEmitter()

	listeners := ee.Listeners("nonexistent")

	if len(listeners) != 0 {
		t.Errorf("len(Listeners) = %d, want 0 for nonexistent event", len(listeners))
	}
}

func TestEventEmitter_AddListener(t *testing.T) {
	ee := NewEventEmitter()
	called := false

	handler := func(args ...any) {
		called = true
	}

	ee.AddListener("test", handler)
	ee.Emit("test")

	if !called {
		t.Error("AddListener should work as alias for On")
	}
}

func TestEventEmitter_Off(t *testing.T) {
	ee := NewEventEmitter()

	handler := func(args ...any) {}

	ee.On("test", handler)

	if ee.ListenerCount("test") != 1 {
		t.Errorf("ListenerCount = %d, want 1", ee.ListenerCount("test"))
	}

	ee.Off("test", handler)

	if ee.ListenerCount("test") != 0 {
		t.Errorf("ListenerCount = %d, want 0 after Off", ee.ListenerCount("test"))
	}
}

func TestEventEmitter_Emits(t *testing.T) {
	ee := NewEventEmitter()

	var received1, received2 []any

	ee.On("event1", func(args ...any) {
		received1 = args
	})

	ee.On("event2", func(args ...any) {
		received2 = args
	})

	ee.Emits(
		[]string{"event1", "event2"},
		[][]any{{"arg1"}, {"arg2"}},
	)

	if len(received1) != 1 || received1[0] != "arg1" {
		t.Errorf("event1 received %v, want [arg1]", received1)
	}

	if len(received2) != 1 || received2[0] != "arg2" {
		t.Errorf("event2 received %v, want [arg2]", received2)
	}
}

func TestEventEmitter_Concurrency(t *testing.T) {
	ee := NewEventEmitter()
	callCount := 0
	var mu sync.Mutex

	handler := func(args ...any) {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	ee.On("test", handler)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ee.Emit("test")
		}()
	}

	wg.Wait()

	mu.Lock()
	count := callCount
	mu.Unlock()

	if count != 100 {
		t.Errorf("callCount = %d, want 100 (concurrent emits)", count)
	}
}

func TestEventEmitter_RemoveListenerDuringEmit(t *testing.T) {
	ee := NewEventEmitter()
	called := false

	var handler EventHandler
	handler = func(args ...any) {
		called = true
		ee.RemoveListener("test", handler)
	}

	ee.On("test", handler)

	// First emit should work
	ee.Emit("test")

	if !called {
		t.Error("Handler should be called on first emit")
	}

	// Listener should be removed
	if ee.ListenerCount("test") != 0 {
		t.Errorf("ListenerCount = %d, want 0 after self-removal", ee.ListenerCount("test"))
	}

	called = false
	ee.Emit("test")

	if called {
		t.Error("Handler should not be called after self-removal")
	}
}

func TestEventEmitter_EmitsWithFewerValues(t *testing.T) {
	ee := NewEventEmitter()

	var received3 []any

	ee.On("event1", func(args ...any) {
		// Just register the handler
	})

	ee.On("event2", func(args ...any) {
		// Just register the handler
	})

	ee.On("event3", func(args ...any) {
		received3 = args
	})

	// Only provide 2 value arrays for 3 events
	ee.Emits(
		[]string{"event1", "event2", "event3"},
		[][]any{{"arg1"}, {"arg2"}},
	)

	// event3 should receive the last value array (arg2)
	if len(received3) != 1 || received3[0] != "arg2" {
		t.Errorf("event3 received %v, want [arg2] (should reuse last value)", received3)
	}
}

func TestEventEmitter_ConcurrentAddRemove(t *testing.T) {
	ee := NewEventEmitter()

	var wg sync.WaitGroup

	// Concurrently add listeners
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ee.On("test", func(args ...any) {})
		}()
	}

	// Concurrently emit events
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ee.Emit("test")
		}()
	}

	// Give some time for goroutines to run
	time.Sleep(100 * time.Millisecond)

	wg.Wait()

	// Test passed if no race conditions detected
}
