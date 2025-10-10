package tmigo

import (
	"time"
)

// Queue implements a delayed task queue for rate limiting
type Queue struct {
	queue        []queueItem
	index        int
	defaultDelay time.Duration
}

type queueItem struct {
	fn    func()
	delay time.Duration
}

// NewQueue creates a new queue with the specified default delay
func NewQueue(defaultDelay time.Duration) *Queue {
	if defaultDelay == 0 {
		defaultDelay = 3 * time.Second
	}
	return &Queue{
		queue:        make([]queueItem, 0),
		index:        0,
		defaultDelay: defaultDelay,
	}
}

// Add adds a new function to the queue with an optional custom delay
func (q *Queue) Add(fn func(), delay ...time.Duration) {
	item := queueItem{fn: fn}
	if len(delay) > 0 {
		item.delay = delay[0]
	}
	q.queue = append(q.queue, item)
}

// Next processes the next item in the queue
func (q *Queue) Next() {
	if q.index >= len(q.queue) {
		return
	}

	item := q.queue[q.index]
	q.index++

	// Execute the current function
	item.fn()

	// Schedule the next item if it exists
	if q.index < len(q.queue) {
		nextItem := q.queue[q.index]
		delay := nextItem.delay
		if delay == 0 {
			delay = q.defaultDelay
		}

		time.AfterFunc(delay, func() {
			q.Next()
		})
	}
}
