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
	popCounter int
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
		} else {
			fmt.Println("[Queue~loadFromFile] Unknown log entry:", line)
		}
	}

	q.items = items

	fmt.Printf("[Queue~loadFromFile] Loaded %d items from file \"%s\"\n", len(items), q.dataFile.Name())

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

		fmt.Println("[Queue~Pop] Queue is empty, file truncated")
	} else {
		// Log the POP operation to the file
		fmt.Fprintf(q.writer, "POP\n")
	}

	q.popCounter++

	if q.popCounter >= 1000 {
		q.popCounter = 0
		go q.Compact()
	}

	return &item
}

func (q *Queue) Compact() error {
	fmt.Printf("[Queue~Compact] Starting compaction for file \"%s\" with %d items\n", q.dataFile.Name(), len(q.items))

	q.mu.Lock()
	defer q.mu.Unlock()

	// Write all pending changes
	q.writer.Flush()

	// Create a temporary file to write the compacted data
	tempFilename := q.dataFile.Name() + ".tmp"
    tempFile, err := os.OpenFile(tempFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)

	if err != nil {
		return err
	}

	// Remove file after method returns
	defer os.Remove(tempFilename)

	// Write the current state of the queue to the buffer
	tempWriter := bufio.NewWriter(tempFile)
	for _, item := range q.items {
		fmt.Fprintf(tempWriter, "PUSH:%d:%s\n", item.Timestamp.Unix(), string(item.Payload))
	}

	// Write buffer to disk and close the temporary file
	tempWriter.Flush()

	// Ensure all data is written to disk before renaming
	tempFile.Sync()
	tempFile.Close()

	// Replace the original file with the compacted file
	filename := q.dataFile.Name()
	q.dataFile.Close()

	err = os.Rename(tempFilename, filename)

	if err != nil {
		return err
	}

	// Reopen the data file and reset the writer
	newFile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return err
	}

	q.dataFile = newFile
	q.writer = bufio.NewWriter(q.dataFile)

	fmt.Printf("[Queue~Compact] Compacted queue to file \"%s\" with %d items\n", filename, len(q.items))

	return nil
}
