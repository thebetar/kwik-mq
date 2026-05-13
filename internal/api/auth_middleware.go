package api

import (
	"fmt"
	"net/http"
	"os"
)


func getAccessToken() string {
	accessToken := os.Getenv("ACCESS_TOKEN")

	if accessToken == "" {
		fmt.Printf("Warning: ACCESS_TOKEN not set. Exiting...\n")
		os.Exit(1)
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