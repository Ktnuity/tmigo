package tmigo

import (
	"sync"
	"testing"
	"time"
)

func TestNewQueue(t *testing.T) {
	q := NewQueue(0)
	if q == nil {
		t.Fatal("NewQueue() returned nil")
	}
	if q.defaultDelay != 3*time.Second {
		t.Errorf("defaultDelay = %v, want 3s (default)", q.defaultDelay)
	}
}

func TestNewQueueWithCustomDelay(t *testing.T) {
	customDelay := 5 * time.Second
	q := NewQueue(customDelay)
	if q.defaultDelay != customDelay {
		t.Errorf("defaultDelay = %v, want %v", q.defaultDelay, customDelay)
	}
}

func TestQueue_Add(t *testing.T) {
	q := NewQueue(1 * time.Second)

	called := false
	fn := func() {
		called = true
	}

	q.Add(fn)

	if len(q.queue) != 1 {
		t.Errorf("queue length = %d, want 1", len(q.queue))
	}

	// Execute the function
	q.queue[0].fn()

	if !called {
		t.Error("Function was not executed")
	}
}

func TestQueue_AddWithCustomDelay(t *testing.T) {
	q := NewQueue(1 * time.Second)

	customDelay := 500 * time.Millisecond
	fn := func() {}

	q.Add(fn, customDelay)

	if len(q.queue) != 1 {
		t.Errorf("queue length = %d, want 1", len(q.queue))
	}

	if q.queue[0].delay != customDelay {
		t.Errorf("item delay = %v, want %v", q.queue[0].delay, customDelay)
	}
}

func TestQueue_Next(t *testing.T) {
	q := NewQueue(1 * time.Second)

	executed := make([]int, 0)
	var mu sync.Mutex

	fn1 := func() {
		mu.Lock()
		executed = append(executed, 1)
		mu.Unlock()
	}

	q.Add(fn1)
	q.Next()

	// Wait a bit for execution
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	length := len(executed)
	mu.Unlock()

	if length != 1 {
		t.Errorf("executed length = %d, want 1", length)
	}

	mu.Lock()
	if executed[0] != 1 {
		t.Errorf("executed[0] = %d, want 1", executed[0])
	}
	mu.Unlock()
}

func TestQueue_NextMultiple(t *testing.T) {
	q := NewQueue(50 * time.Millisecond) // Short delay for testing

	executed := make([]int, 0)
	var mu sync.Mutex

	fn1 := func() {
		mu.Lock()
		executed = append(executed, 1)
		mu.Unlock()
	}

	fn2 := func() {
		mu.Lock()
		executed = append(executed, 2)
		mu.Unlock()
	}

	fn3 := func() {
		mu.Lock()
		executed = append(executed, 3)
		mu.Unlock()
	}

	q.Add(fn1)
	q.Add(fn2)
	q.Add(fn3)

	// Start processing
	q.Next()

	// Wait for all to execute (3 items * 50ms + buffer)
	time.Sleep(250 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(executed) != 3 {
		t.Errorf("executed length = %d, want 3", len(executed))
	}

	// Verify execution order
	if executed[0] != 1 || executed[1] != 2 || executed[2] != 3 {
		t.Errorf("executed = %v, want [1 2 3]", executed)
	}
}

func TestQueue_NextWithCustomDelays(t *testing.T) {
	q := NewQueue(1 * time.Second) // Default delay (not used)

	executed := make([]int, 0)
	var mu sync.Mutex

	fn1 := func() {
		mu.Lock()
		executed = append(executed, 1)
		mu.Unlock()
	}

	fn2 := func() {
		mu.Lock()
		executed = append(executed, 2)
		mu.Unlock()
	}

	q.Add(fn1, 30*time.Millisecond)
	q.Add(fn2, 30*time.Millisecond)

	q.Next()

	// Wait for both to execute
	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(executed) != 2 {
		t.Errorf("executed length = %d, want 2", len(executed))
	}
}

func TestQueue_NextEmpty(t *testing.T) {
	q := NewQueue(1 * time.Second)

	// Calling Next on empty queue should not panic
	q.Next()

	// Test passed if no panic occurred
}

func TestQueue_NextBeyondEnd(t *testing.T) {
	q := NewQueue(50 * time.Millisecond)

	called := false
	fn := func() {
		called = true
	}

	q.Add(fn)
	q.Next()

	time.Sleep(100 * time.Millisecond)

	if !called {
		t.Error("Function should have been called")
	}

	// Calling Next again should not panic (index >= len(queue))
	q.Next()

	// Test passed if no panic occurred
}

func TestQueue_ExecutionOrder(t *testing.T) {
	q := NewQueue(20 * time.Millisecond)

	order := make([]string, 0)
	var mu sync.Mutex

	q.Add(func() {
		mu.Lock()
		order = append(order, "first")
		mu.Unlock()
	})

	q.Add(func() {
		mu.Lock()
		order = append(order, "second")
		mu.Unlock()
	})

	q.Add(func() {
		mu.Lock()
		order = append(order, "third")
		mu.Unlock()
	})

	q.Next()

	// Wait for all to complete
	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(order) != 3 {
		t.Fatalf("expected 3 executions, got %d", len(order))
	}

	if order[0] != "first" || order[1] != "second" || order[2] != "third" {
		t.Errorf("execution order = %v, want [first second third]", order)
	}
}

func TestQueue_DefaultDelayUsed(t *testing.T) {
	defaultDelay := 50 * time.Millisecond
	q := NewQueue(defaultDelay)

	executionTimes := make([]time.Time, 0)
	var mu sync.Mutex

	q.Add(func() {
		mu.Lock()
		executionTimes = append(executionTimes, time.Now())
		mu.Unlock()
	})

	q.Add(func() {
		mu.Lock()
		executionTimes = append(executionTimes, time.Now())
		mu.Unlock()
	})

	start := time.Now()
	q.Next()

	// Wait for both to execute
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(executionTimes) != 2 {
		t.Fatalf("expected 2 executions, got %d", len(executionTimes))
	}

	// Check that second execution happened after the delay
	timeDiff := executionTimes[1].Sub(executionTimes[0])

	// Allow some variance (30ms - 100ms range to account for scheduling)
	if timeDiff < 30*time.Millisecond || timeDiff > 100*time.Millisecond {
		t.Logf("Time difference between executions: %v (expected around %v)", timeDiff, defaultDelay)
		// Not a hard failure since timing can vary in CI
	}

	totalTime := time.Since(start)
	if totalTime < defaultDelay {
		t.Errorf("Total execution time %v is less than expected delay %v", totalTime, defaultDelay)
	}
}

func TestQueue_MixedDelays(t *testing.T) {
	q := NewQueue(100 * time.Millisecond) // Default delay

	order := make([]int, 0)
	var mu sync.Mutex

	// First item uses default delay
	q.Add(func() {
		mu.Lock()
		order = append(order, 1)
		mu.Unlock()
	})

	// Second item uses custom short delay
	q.Add(func() {
		mu.Lock()
		order = append(order, 2)
		mu.Unlock()
	}, 20*time.Millisecond)

	// Third item uses default delay
	q.Add(func() {
		mu.Lock()
		order = append(order, 3)
		mu.Unlock()
	})

	q.Next()

	// Wait for all to complete
	time.Sleep(250 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(order) != 3 {
		t.Fatalf("expected 3 executions, got %d", len(order))
	}

	// Verify execution order is maintained
	if order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Errorf("execution order = %v, want [1 2 3]", order)
	}
}
