package queue

import (
	"fmt"
)

func GetOrCreateQueue(id string) (*Queue, error) {
	q, exists := GetQueue(id)

	if exists {
		return q, nil
	}

	// If the queue doesn't exist, acquire a write lock to create it
	queuesMu.Lock()
	defer queuesMu.Unlock()

	// Double-check if the queue was created by another goroutine while waiting for the lock
	q, exists = queues[id]

	if exists {
		return q, nil
	}

	// If the queue doesn't exist, create a new one
	new_q, err := NewQueue(id)

	if err != nil {
		return nil, fmt.Errorf("Failed to create queue: %v", err)
	}

	q = new_q
	queues[id] = q

	return q, nil
}

func GetQueue(id string) (*Queue, bool) {
	// Fetch the queue (using a read lock)
	queuesMu.RLock()
	q, exists := queues[id]
	queuesMu.RUnlock()

	if !exists {
		return nil, false
	}

	return q, true
}
