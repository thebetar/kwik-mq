package queue

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestPush(t *testing.T) {
	test_queue_id := "test_queue_push_" + time.Now().Format("20060102150405")
	tempDir := t.TempDir() 
	os.Setenv("DATA_DIR", tempDir) // Assuming your code now checks for this

	q, err := NewQueue(test_queue_id)

	if err != nil {
		t.Fatalf(("Expected no error, got %v"), err)
	}

	if q == nil {
		t.Fatal("Expected a queue instance, got nil")
	}

	var test_payload json.RawMessage = json.RawMessage(`{"test": "data"}`)
	q.Push(test_payload)

	if len(q.items) != 1 {
		t.Errorf("Expected queue to have 1 item after push, got %d", len(q.items))
	}

	if string(q.items[0].Payload) != `{"test": "data"}` {
		t.Errorf("Expected payload of first item to be {\"test\": \"data\"}, got %v", q.items[0].Payload)
	}
}

func TestPop(t *testing.T) {
	test_queue_id := "test_queue_pop_" + time.Now().Format("20060102150405")
	tempDir := t.TempDir() 
	os.Setenv("DATA_DIR", tempDir) // Assuming your code now checks for this

	q, err := NewQueue(test_queue_id)

	if err != nil {
		t.Fatalf(("Expected no error, got %v"), err)
	}

	if q == nil {
		t.Fatal("Expected a queue instance, got nil")
	}

	var test_payload_1 json.RawMessage = json.RawMessage(`{"test": "data"}`)
	q.Push(test_payload_1)

	var test_payload_2 json.RawMessage = json.RawMessage(`{"test": "data2"}`)
	q.Push(test_payload_2)

	item1 := q.Pop()
	item2 := q.Pop()
	item3 := q.Pop() // This should be nil since the queue is now empty

	if item1 == nil || string(item1.Payload) != `{"test": "data"}` {
		t.Errorf("Expected first popped item to have payload {\"test\": \"data\"}, got %v", item1)
	}

	if item2 == nil || string(item2.Payload) != `{"test": "data2"}` {
		t.Errorf("Expected second popped item to have payload {\"test\": \"data2\"}, got %v", item2)
	}

	if item3 != nil {
		t.Errorf("Expected third popped item to be nil since the queue should be empty, got %v", item3)
	}
}