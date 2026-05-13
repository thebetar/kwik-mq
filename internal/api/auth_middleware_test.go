package api

import (
	"net/http"
	"os"
	"testing"
)

// dummyResponseWriter is a simple implementation of http.ResponseWriter for testing purposes
type dummyResponseWriter struct {
	header http.Header
}

func (d *dummyResponseWriter) Header() http.Header {
	if d.header == nil {
		d.header = make(http.Header)
	}
	return d.header
}

func (d *dummyResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (d *dummyResponseWriter) WriteHeader(statusCode int) {
	// No-op for testing
}

func TestGenerateRandomToken(t *testing.T) {
	token1 := generateRandomToken(32)
	token2 := generateRandomToken(32)

	if len(token1) != 32 {
		t.Errorf("Expected token length of 32, got %d", len(token1))
	}

	if len(token2) != 32 {
		t.Errorf("Expected token length of 32, got %d", len(token2))
	}

	if token1 == token2 {
		t.Error("Expected two generated tokens to be different, but they are the same")
	}
}

func TestGetAccessToken(t *testing.T) {
	// Clear the ACCESS_TOKEN environment variable for testing
	os.Unsetenv("ACCESS_TOKEN")

	token := getAccessToken()

	if token == "" {
		t.Error("Expected a non-empty access token, got an empty string")
	}

	// Check if the token is set in the environment variable
	envToken := os.Getenv("ACCESS_TOKEN")
	if envToken != token {
		t.Errorf("Expected ACCESS_TOKEN environment variable to be set to the generated token, got %s", envToken)
	}
}

func TestCheckAccessToken(t *testing.T) {
	// Set a known access token for testing
	testToken := "test_access_token"
	os.Setenv("ACCESS_TOKEN", testToken)

	// Create a dummy request with the correct token
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", testToken)

	// Create a dummy response writer
	w := &dummyResponseWriter{}

	// Define a dummy handler that will be called if the token is valid
	dummyHandlerCalled := false
	dummyHandler := func(w http.ResponseWriter, r *http.Request) {
		dummyHandlerCalled = true
	}

	// Call CheckAccessToken with the dummy request and handler
	result := CheckAccessToken(w, req, dummyHandler)

	if !result {
		t.Error("Expected CheckAccessToken to return true for valid token, got false")
	}

	if !dummyHandlerCalled {
		t.Error("Expected dummy handler to be called for valid token, but it was not called")
	}
}

func TestCheckAccessTokenInvalidToken(t *testing.T) {
	// Set a known access token for testing
	testToken := "test_access_token"
	os.Setenv("ACCESS_TOKEN", testToken)

	// Create a dummy request with an incorrect token
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "invalid_token")

	// Create a dummy response writer
	w := &dummyResponseWriter{}

	// Define a dummy handler that will be called if the token is valid
	dummyHandlerCalled := false
	dummyHandler := func(w http.ResponseWriter, r *http.Request) {
		dummyHandlerCalled = true
	}

	// Call CheckAccessToken with the dummy request and handler
	result := CheckAccessToken(w, req, dummyHandler)

	if result {
		t.Error("Expected CheckAccessToken to return false for invalid token, got true")
	}

	if dummyHandlerCalled {
		t.Error("Expected dummy handler not to be called for invalid token, but it was called")
	}
}