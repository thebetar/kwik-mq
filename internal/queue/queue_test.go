package queue

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestGetOrCreateQueue(t *testing.T) {
	test_queue_id := "test_queue_" + time.Now().Format("20060102150405")
	tempDir := os.TempDir()
	os.Setenv("DATA_DIR", tempDir)

	// Create new queue
	q, err := GetOrCreateQueue(test_queue_id)

	if err != nil {
		t.Fatalf(("Expected no error, got %v"), err)
	}

	if q == nil {
		t.Fatal("Expected a queue instance, got nil")
	}

	// Check if file exists
	expectedFilePath := fmt.Sprintf("%s/%s.log", tempDir, test_queue_id)
	if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
		t.Errorf("Expected data file \"%s\" to exist, but it does not", expectedFilePath)
	}

	// Get queue from map
	retrieved_q, exists := GetQueue(test_queue_id)

	if !exists {
		t.Errorf("Expected queue with ID \"%s\" to exist, but it does not", test_queue_id)
	}

	if retrieved_q != q {
		t.Errorf("Expected retrieved queue to be the same instance as the created queue")
	}
}

func TestGetOrCreateQueueWithRows(t *testing.T) {
    test_queue_id := "test_queue_with_rows_" + time.Now().Format("20060102150405")
    tempDir := t.TempDir() 
    os.Setenv("DATA_DIR", tempDir) // Assuming your code now checks for this

    // 1. Manually craft the log file first
    logFilePath := tempDir + "/" + test_queue_id + ".log"
    
    now := time.Now().Unix()

	test_items := []string{
		`{"test": "data1"}`,
		`{"test": "data2"}`,
		`{"test": "data3"}`,
	}

    logContent := fmt.Sprintf("PUSH:%d:%s\n", now-10, test_items[0])
	logContent += fmt.Sprintf("PUSH:%d:%s\n", now-5, test_items[1])
	logContent += "POP\n"  // Removes the first payload
	logContent += fmt.Sprintf("PUSH:%d:%s\n", now, test_items[2])

    // Write this raw log content to disk BEFORE initializing the queue
    err := os.WriteFile(logFilePath, []byte(logContent), 0644)
    if err != nil {
        t.Fatalf("Failed to write mock log file: %v", err)
    }

    // Create new queue from existing file
    q, err := GetOrCreateQueue(test_queue_id)
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }

    if q == nil {
        t.Fatal("Expected a queue instance, got nil")
    }

	if len(q.items) != 2 {
		t.Fatalf("Expected 2 items in the queue after loading from file, got %d", len(q.items))
	}

	// Verify that the loaded items match the expected payloads
	for i, item := range q.items {
		if string(item.Payload) == test_items[i+1] { // Skip the first item which was popped
			continue
		}
		
		t.Errorf("Expected payload %s at index %d, got %s", test_items[i+1], i, string(item.Payload))
	}
}

func TestGetQueue(t *testing.T) {
	test_queue_id := "test_queue_get_" + time.Now().Format("20060102150405")
	tempDir := t.TempDir() 
	os.Setenv("DATA_DIR", tempDir) // Assuming your code now checks for this

	q, err := GetOrCreateQueue(test_queue_id)

	if err != nil {
		t.Fatalf(("Expected no error, got %v"), err)
	}

	if q == nil {
		t.Fatal("Expected a queue instance, got nil")
	}

	retrieved_q, exists := GetQueue(test_queue_id)

	if !exists {
		t.Errorf("Expected queue with ID \"%s\" to exist, but it does not", test_queue_id)
	}

	if retrieved_q != q {
		t.Errorf("Expected retrieved queue to be the same instance as the created queue")
	}
}

func TestGetQueueFail(t *testing.T) {
	test_queue_id := "non_existent_queue_" + time.Now().Format("20060102150405")

	q, exists := GetQueue(test_queue_id)

	if exists {
		t.Errorf("Expected queue with ID \"%s\" to not exist, but it does", test_queue_id)
	}

	if q != nil {
		t.Errorf("Expected returned queue to be nil when it does not exist, got %v", q)
	}
}