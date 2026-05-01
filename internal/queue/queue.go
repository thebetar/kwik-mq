package queue

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

type QueueItem struct {
	Payload json.RawMessage `json:"payload"`
}

type Queue struct {
	mu sync.Mutex // Ensure thread-safe access to the queue with mutual exclusion
	items []QueueItem
	dataFile *os.File
}

var queues = make(map[string]*Queue)

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
			items = append(items, QueueItem{Payload: json.RawMessage(payload)})
		} else if line == "POP" {
			if len(items) > 0 {
				items = items[1:]
			}
		}
	}

	q.items = items

	return nil
}

func CreateQueue(id string, filename string) (*Queue, error) {
	filepath := fmt.Sprintf("./data/%s", filename)
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return nil, err
	}

	q := &Queue{
		items:    make([]QueueItem, 0),
		dataFile: file,
	}

	q.loadFromFile()

	queues[id] = q

	return q, nil
}


func QueuePush(w http.ResponseWriter, req *http.Request) {
	// Get the ID from the query parameter
	id := req.URL.Query().Get("id")

	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "Missing queue ID"}`, http.StatusBadRequest)
		return
	}

	// Fetch or create the queue
	q, exists := queues[id]

	if !exists {
		// If the queue doesn't exist, create a new one
		new_q, err := CreateQueue(id, fmt.Sprintf("%s.log", id))

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error": "Failed to create queue"}`, http.StatusInternalServerError)
			return
		}

		q = new_q
	}

	var item QueueItem

	// Decode the JSON payload from the request bo
	var raw json.RawMessage
	err := json.NewDecoder(req.Body).Decode(&raw)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "Invalid JSON payload"}`, http.StatusBadRequest)
		return
	}

	item.Payload = raw

	// Lock the queue for thread-safe access
	q.mu.Lock()
	defer q.mu.Unlock()

	// Push the item to the queue and log it to the file
	q.items = append(q.items, item)
	// Log the PUSH operation to the file
	fmt.Fprintf(q.dataFile, "PUSH:%s\n", string(item.Payload))

	// Return a success response
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Pushed to queue %s"}`, id)
}

func QueuePop(w http.ResponseWriter, req *http.Request) {
	// Get the ID from the query parameter
	id := req.URL.Query().Get("id")

	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "Missing queue ID"}`, http.StatusBadRequest)
		return
	}

	// Fetch the queue
	q, exists := queues[id]

	if !exists {
		// If the queue doesn't exist, return a 404 error
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "Queue not found"}`, http.StatusNotFound)
		return
	}

	// Lock the queue for thread-safe access
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check if the queue is empty
	if len(q.items) == 0 {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "Queue is empty"}`, http.StatusNoContent)
		return
	}

	// Pop the first item from the queue
	item := q.items[0]
	q.items = q.items[1:]

	if len(q.items) == 0 {
		// On empty queue, truncate the file to remove all entries
		q.dataFile.Truncate(0)
	} else {
		// Log the POP operation to the file
		fmt.Fprintf(q.dataFile, "POP\n")
	}

	// Return the popped item as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}