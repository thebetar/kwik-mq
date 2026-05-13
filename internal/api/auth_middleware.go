package api

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"os"
	"time"
)

func generateRandomToken(length int) string {
	// md5 hash of time
	timestamp := time.Now().String()
	hash := md5.Sum([]byte(timestamp))
	return fmt.Sprintf("%x", hash)[:length]
}

func addTokenToEnv(token string) {
	f, err := os.OpenFile(".env", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error creating .env file:", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString("ACCESS_TOKEN=" + token + "\n"); err != nil {
		fmt.Println("Error writing to .env file:", err)
	}
}

func getAccessToken() string {
	accessToken := os.Getenv("ACCESS_TOKEN")

	if accessToken == "" {
		fmt.Println("Warning: ACCESS_TOKEN environment variable is not set. Generating a random token and putting it in .env file.")
		accessToken = generateRandomToken(32)
		addTokenToEnv(accessToken)
	}

	return accessToken
}

func CheckAccessToken(w http.ResponseWriter, req *http.Request, handler http.HandlerFunc) bool {
	// Get the access token from the query parameter
	token := req.Header.Get("Authorization")

	if token == "" {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "Missing access token"}`, http.StatusUnauthorized)
		return false
	}

	// Get token
	if token != getAccessToken() {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "Invalid access token"}`, http.StatusUnauthorized)
		return false
	}

	handler(w, req)

	return true
}