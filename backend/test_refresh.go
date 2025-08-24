package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Example usage of the refresh token endpoint
func main() {
	fmt.Println("Testing OAuth Token Refresh Endpoint:")
	fmt.Println("")

	// Example: Refresh token request
	// Note: This requires a valid authorization token and existing refresh token in database

	refreshURL := "http://localhost:8080/token/refresh"

	// Create request
	req, err := http.NewRequest("POST", refreshURL, bytes.NewBuffer([]byte("{}")))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}

	// Add headers (example - replace with actual auth token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer YOUR_JWT_TOKEN_HERE")

	fmt.Printf("Endpoint: POST %s\n", refreshURL)
	fmt.Println("Headers:")
	fmt.Println("  Content-Type: application/json")
	fmt.Println("  Authorization: Bearer YOUR_JWT_TOKEN_HERE")
	fmt.Println("")
	fmt.Println("Expected Response (on success):")

	// Example successful response
	successResponse := map[string]interface{}{
		"access_token": "ya29.a0ARrda...",
		"token_type":   "Bearer",
		"expires_at":   1692700800,
		"message":      "Token refreshed successfully",
	}

	responseJSON, _ := json.MarshalIndent(successResponse, "", "  ")
	fmt.Println(string(responseJSON))

	fmt.Println("")
	fmt.Println("Expected Response (on error):")

	// Example error response
	errorResponse := map[string]interface{}{
		"error": "No token found for user",
	}

	errorJSON, _ := json.MarshalIndent(errorResponse, "", "  ")
	fmt.Println(string(errorJSON))

	fmt.Println("")
	fmt.Println("How to use:")
	fmt.Println("1. User must be authenticated (valid JWT token in Authorization header)")
	fmt.Println("2. User must have a saved OAuth token with valid refresh_token in database")
	fmt.Println("3. Send POST request to /token/refresh")
	fmt.Println("4. New access token will be returned and saved to database")
	fmt.Println("5. Original refresh token is preserved unless Google provides a new one")
}
