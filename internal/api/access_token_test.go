package api

import (
	"os"
	"testing"
)

func TestCheckAccessTokenExists(t *testing.T) {
	// Unset the ACCESS_TOKEN environment variable for testing
	projectPath := os.TempDir()
	os.Setenv("PROJECT_PATH", projectPath)
	os.Unsetenv("ACCESS_TOKEN")

	// Call the function to check if it generates a token
	token := checkAccessTokenExists()

	if token == "" {
		t.Error("Expected a generated token, got an empty string")
	}

	// Check if the token is set in the environment variable
	envToken := os.Getenv("ACCESS_TOKEN")
	if envToken == "" {
		t.Error("Expected ACCESS_TOKEN environment variable to be set, but it is not set")
	}

	if envToken != token {
		t.Errorf("Expected ACCESS_TOKEN environment variable to be set to the generated token, got %s", envToken)
	}
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