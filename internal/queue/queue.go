package queue

import (
	"bufio"
	"fmt"
	"os"
	"time"
)


func CreateQueue(id string, filename string) (*Queue, error) {
	filepath := fmt.Sprintf("./data/%s", filename)
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return nil, err
	}

	q := &Queue{
		items:    make([]QueueItem, 0),
		dataFile: file,
		writer:  bufio.NewWriter(file),
	}

	q.loadFromFile()

	go func() {
		ticker := time.NewTicker((100 * time.Millisecond))

		for range ticker.C {
			q.mu.Lock()
			q.writer.Flush()
			q.mu.Unlock()
		}
	}()

	queues[id] = q

	return q, nil
}

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
	new_q, err := CreateQueue(id, fmt.Sprintf("%s.log", id))

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
