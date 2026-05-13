package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path"
)

var BuildTimeAccessToken string

func generateRandomToken(length int) string {
	bytes := make([]byte, (length+1)/2)
	if _, err := rand.Read(bytes); err != nil {
		panic(fmt.Sprintf("failed to generate random token: %v", err))
	}
	return hex.EncodeToString(bytes)[:length]
}

func addTokenToEnv(token string) {
	projectPath := os.Getenv("PROJECT_PATH")
	envPath := path.Join(projectPath, ".env")

	f, err := os.OpenFile(envPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error creating .env file:", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString("ACCESS_TOKEN=" + token + "\n"); err != nil {
		fmt.Println("Error writing to .env file:", err)
	}
}


func checkAccessTokenExists() string {
	accessToken := os.Getenv("ACCESS_TOKEN")

	if accessToken == "" && BuildTimeAccessToken != "" {
		accessToken = BuildTimeAccessToken
		os.Setenv("ACCESS_TOKEN", accessToken)
	}

	if accessToken == "" {
		accessToken = generateRandomToken(32)
		os.Setenv("ACCESS_TOKEN", accessToken)
		addTokenToEnv(accessToken)
		fmt.Printf("Warning: ACCESS_TOKEN not set. Generated token: %s (saved to .env)\n", accessToken)
	}

	return accessToken
}