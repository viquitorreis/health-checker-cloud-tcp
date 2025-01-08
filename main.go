package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	apiUrl := os.Getenv("API_URL")
	if apiUrl == "" {
		log.Fatalf("API_URL not set in environment variables")
	}

	opts := NewTCPOpts("cloud", nil, 0, 0, apiUrl, true)
	checker := NewTCPChecker(opts)
	checker.Timeout = 5 * time.Second

	logOutput := log.Writer()
	result := checker.CheckWithRetries(-1, 10*time.Second, logOutput)

	println("Result: ", result.Message)
}
