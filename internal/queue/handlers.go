package queue

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func QueuePushHandler(w http.ResponseWriter, req *http.Request) {
	// Get the ID from the query parameter
	id := req.URL.Query().Get("id")

	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "Missing queue ID"}`, http.StatusBadRequest)
		return
	}

	q, q_err := GetOrCreateQueue(id)

	if q_err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "`+q_err.Error()+`"}`, http.StatusNotFound)
		return
	}

	// Decode the JSON payload from the request body
	var raw json.RawMessage
	d_err := json.NewDecoder(req.Body).Decode(&raw)

	if d_err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "Invalid JSON payload"}`, http.StatusBadRequest)
		return
	}

	// Create a new QueueItem with the decoded payload and current timestamp
	_, err := q.Push(raw)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// Return a success response
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Pushed to queue %s"}`, id)
}

func QueuePopHandler(w http.ResponseWriter, req *http.Request) {
	// Get the ID from the query parameter
	id := req.URL.Query().Get("id")

	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "Missing queue ID"}`, http.StatusBadRequest)
		return
	}

	// Fetch the queue
	q, q_err := GetOrCreateQueue(id)

	if q_err != nil {
		// If the queue doesn't exist, return a 404 error
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "`+q_err.Error()+`"}`, http.StatusNotFound)
		return
	}

	item, err := q.Pop()

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	if item == nil {
		// If the queue is empty, return a 204 No Content response
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Return the popped item as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item.Payload)
}
