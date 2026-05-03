package queue

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type QueueItem struct {
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

type Queue struct {
	mu         sync.Mutex // Ensure thread-safe access to the queue with mutual exclusion
	items      []QueueItem
	dataFile   *os.File
	writer     *bufio.Writer
	popCounter int
}

type LogEntry struct {
	Operation string          `json:"operation"`
	Timestamp time.Time       `json:"timestamp,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

var (
	queues   = make(map[string]*Queue)
	queuesMu sync.RWMutex
)

func NewQueue(id string) (*Queue, error) {
	filename := fmt.Sprintf("%s.log", id)
	dataDir := os.Getenv("DATA_DIR")

	if dataDir == "" {
		dataDir = "./data"
	}

	_, err := os.Stat(dataDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(dataDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("Failed to create data directory: %v", err)
		}
	}

	filepath := fmt.Sprintf("%s/%s", dataDir, filename)
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return nil, err
	}

	q := &Queue{
		items:    make([]QueueItem, 0),
		dataFile: file,
		writer:   bufio.NewWriter(file),
	}

	err = q.loadFromFile()
	if err != nil {
		file.Close()
		return nil, err
	}

	go func() {
		ticker := time.NewTicker((100 * time.Millisecond))

		for range ticker.C {
			q.mu.Lock()
			q.writer.Flush()
			q.mu.Unlock()
		}
	}()

	queues[id] = q

	fmt.Printf("[Queue~LoadQueue] Loaded queue with ID \"%s\" and data file \"%s\" (%d items)\n", id, filepath, len(q.items))

	return q, nil
}

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

		var logEntry LogEntry
		err := json.Unmarshal([]byte(line), &logEntry)

		if err != nil {
			fmt.Printf("[Queue~loadFromFile] Failed to unmarshal log entry: %v\n", err)
			continue
		}

		if logEntry.Operation == "PUSH" {
			items = append(items, QueueItem{
				Payload:   logEntry.Payload,
				Timestamp: logEntry.Timestamp,
			})
		} else if logEntry.Operation == "POP" {
			if len(items) == 0 {
				fmt.Println("[Queue~loadFromFile] Warning: POP operation found but queue is already empty")
				continue
			}

			items = items[1:]
		} else {
			fmt.Println("[Queue~loadFromFile] Unknown log entry:", line)
		}
	}

	q.items = items

	fmt.Printf("[Queue~loadFromFile] Loaded %d items from file \"%s\"\n", len(items), q.dataFile.Name())

	return nil
}

func (q *Queue) Push(payload json.RawMessage) (*QueueItem, error) {
	// Create a new QueueItem with the decoded payload and current timestamp
	var item = QueueItem{
		Payload:   payload,
		Timestamp: time.Now(),
	}

	// Lock the queue for thread-safe access
	q.mu.Lock()
	defer q.mu.Unlock()

	// Log the PUSH operation to the file
	logEntry := LogEntry{
		Operation: "PUSH",
		Timestamp: item.Timestamp,
		Payload:   item.Payload,
	}
	logEntryJson, err := json.Marshal(logEntry)

	if err != nil {
		fmt.Printf("[Queue~Push] Failed to marshal log entry: %v\n", err)
		return nil, err
	}

	_, err = fmt.Fprintf(q.writer, "%s\n", logEntryJson)
	if err != nil {
		return nil, err
	}

	// Push the item to the queue after it has been queued for writing to the log
	q.items = append(q.items, item)

	return &item, nil
}

func (q *Queue) Pop() (*QueueItem, error) {
	// Lock the queue for thread-safe access
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check if the queue is empty
	if len(q.items) == 0 {
		return nil, nil
	}

	// Pop the first item from the queue
	item := q.items[0]
	var err error

	if len(q.items) == 1 {
		// On empty queue, truncate the file to remove all entries
		err = q.dataFile.Truncate(0)
		if err != nil {
			return nil, err
		}

		_, err = q.dataFile.Seek(0, 0)
		if err != nil {
			return nil, err
		}

		q.writer.Reset(q.dataFile)

		fmt.Println("[Queue~Pop] Queue is empty, file truncated")
	} else {
		// Log the POP operation to the file
		logEntry := LogEntry{
			Operation: "POP",
			Timestamp: time.Now(),
		}
		logEntryJson, err := json.Marshal(logEntry)

		if err != nil {
			fmt.Printf("[Queue~Pop] Failed to marshal log entry: %v\n", err)
			return nil, err
		}

		_, err = fmt.Fprintf(q.writer, "%s\n", logEntryJson)
		if err != nil {
			return nil, err
		}

	}

	q.items = q.items[1:]

	q.popCounter++

	if q.popCounter >= 1000 {
		q.popCounter = 0
		go q.Compact()
	}

	return &item, nil
}

func (q *Queue) Compact() error {
	fmt.Printf("[Queue~Compact] Starting compaction for file \"%s\" with %d items\n", q.dataFile.Name(), len(q.items))

	q.mu.Lock()
	defer q.mu.Unlock()

	// Write all pending changes
	err := q.writer.Flush()
	if err != nil {
		return err
	}

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
		logEntry := LogEntry{
			Operation: "PUSH",
			Timestamp: item.Timestamp,
			Payload:   item.Payload,
		}
		logEntryJson, err := json.Marshal(logEntry)

		if err != nil {
			fmt.Printf("[Queue~Compact] Failed to marshal log entry: %v\n", err)
			return err
		}

		_, err = fmt.Fprintf(tempWriter, "%s\n", logEntryJson)
		if err != nil {
			return err
		}
	}

	// Write buffer to disk and close the temporary file
	err = tempWriter.Flush()
	if err != nil {
		return err
	}

	// Ensure all data is written to disk before renaming
	err = tempFile.Sync()
	if err != nil {
		return err
	}

	err = tempFile.Close()
	if err != nil {
		return err
	}

	// Replace the original file with the compacted file
	filename := q.dataFile.Name()
	err = q.dataFile.Close()
	if err != nil {
		return err
	}

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
