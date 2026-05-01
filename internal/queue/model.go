package queue

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type QueueItem struct {
	Payload json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

type Queue struct {
	mu sync.Mutex // Ensure thread-safe access to the queue with mutual exclusion
	items []QueueItem
	dataFile *os.File
	writer *bufio.Writer
}

var (
	queues = make(map[string]*Queue)
	queuesMu sync.RWMutex
)

func (q *Queue) loadFromFile() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	_, err := q.dataFile.Seek(0, 0)
	if err != nil {
		return err
	}

	items := make([]QueueItem, 0)
	scanner := bufio.NewScanner(q.dataFile)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "PUSH:") {
			payload := line[5:]
			before_cut, after_cut, ok := strings.Cut(payload, ":")
			
			if !ok {
				continue
			}

			timestampStr := before_cut
			payloadStr := after_cut

			timestampInt, err := strconv.ParseInt(timestampStr, 10, 64)
			if err != nil {
				continue
			}

			items = append(items, QueueItem{
				Payload: json.RawMessage(payloadStr), 
				Timestamp: time.Unix(timestampInt, 0),
			})
		} else if line == "POP" {
			if len(items) > 0 {
				items = items[1:]
			}
		}
	}

	q.items = items

	return nil
}

func (q *Queue) Push(payload json.RawMessage) *QueueItem {
	// Create a new QueueItem with the decoded payload and current timestamp
	var item = QueueItem{
		Payload : payload,
		Timestamp : time.Now(),
	}

	// Lock the queue for thread-safe access
	q.mu.Lock()
	defer q.mu.Unlock()

	// Push the item to the queue and log it to the file
	q.items = append(q.items, item)
	
	// Log the PUSH operation to the file
	fmt.Fprintf(q.writer, "PUSH:%d:%s\n", item.Timestamp.Unix(), string(item.Payload))

	return &item
}

func (q *Queue) Pop() *QueueItem {
// Lock the queue for thread-safe access
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check if the queue is empty
	if len(q.items) == 0 {
		return nil
	}

	// Pop the first item from the queue
	item := q.items[0]
	q.items = q.items[1:]

	if len(q.items) == 0 {
		// On empty queue, truncate the file to remove all entries
		q.writer.Flush()
		q.dataFile.Truncate(0)
		q.dataFile.Seek(0, 0)
		q.writer.Reset(q.dataFile)
	} else {
		// Log the POP operation to the file
		fmt.Fprintf(q.writer, "POP\n")
	}

	return &item
}