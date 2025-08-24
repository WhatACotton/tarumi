package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type OpenRouterService struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

type OpenRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenRouterResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func NewOpenRouterService() *OpenRouterService {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	log.Printf("[OpenRouter] Initializing service with API key present: %t", apiKey != "")

	if apiKey == "" {
		log.Printf("[OpenRouter] Warning: No OPENROUTER_API_KEY environment variable found")
		apiKey = "your-openrouter-api-key" // デフォルト値
	}

	service := &OpenRouterService{
		apiKey:  apiKey,
		baseURL: "https://openrouter.ai/api/v1",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	log.Printf("[OpenRouter] Service initialized with baseURL: %s", service.baseURL)
	return service
}

// Chat sends a chat completion request to OpenRouter API
func (s *OpenRouterService) Chat(model string, messages []Message) (*OpenRouterResponse, error) {
	log.Printf("[OpenRouter] Starting chat completion request with model: %s", model)

	request := OpenRouterRequest{
		Model:    model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Printf("[OpenRouter] Failed to marshal request: %v", err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("[OpenRouter] Request payload size: %d bytes", len(jsonData))
	log.Printf("[OpenRouter] Making POST request to: %s", s.baseURL+"/chat/completions")

	req, err := http.NewRequest("POST", s.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[OpenRouter] Failed to create HTTP request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/WhatACotton/tarumi")
	req.Header.Set("X-Title", "Tarumi Todo App")

	// ログ用にAPIキーをマスク
	maskedKey := s.apiKey
	if len(s.apiKey) > 10 {
		maskedKey = s.apiKey[:10] + "..."
	}
	log.Printf("[OpenRouter] Using API key: %s", maskedKey)

	log.Printf("[OpenRouter] Request headers set, sending request...")

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("[OpenRouter] HTTP request failed: %v", err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[OpenRouter] Received response with status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[OpenRouter] API error response body: %s", string(body))
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response OpenRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("[OpenRouter] Failed to decode response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("[OpenRouter] Successfully received response with %d choices", len(response.Choices))
	if len(response.Choices) > 0 {
		log.Printf("[OpenRouter] First choice content length: %d characters", len(response.Choices[0].Message.Content))
	}

	return &response, nil
}

// EstimateTaskDuration uses AI to estimate task duration in minutes with retry logic
func (s *OpenRouterService) EstimateTaskDuration(title, description string) (int, error) {
	log.Printf("[OpenRouter] EstimateTaskDuration called with title: '%s', description length: %d", title, len(description))

	// Fallback calculation for when AI fails
	fallbackDuration := s.calculateFallbackDuration(title, description)
	log.Printf("[OpenRouter] Fallback duration calculated: %d minutes", fallbackDuration)

	messages := []Message{
		{
			Role: "system",
			Content: `You are a task duration estimation expert. Given a task title and description, estimate how many minutes it would take for an average person to complete the task. 

Rules:
- Return only a number (the estimated minutes)
- Consider the complexity and typical time requirements
- Be realistic but not overly conservative
- For simple tasks (5-15 min), for medium tasks (15-60 min), for complex tasks (60+ min)`,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Task: %s\nDescription: %s\n\nHow many minutes will this task take?", title, description),
		},
	}

	// Try with retry logic
	maxRetries := 2
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("[OpenRouter] Attempt %d/%d to estimate duration...", attempt, maxRetries)

		response, err := s.Chat("openai/gpt-oss-20b:free", messages)
		if err != nil {
			log.Printf("[OpenRouter] Attempt %d failed: %v", attempt, err)

			// Check if it's a rate limit error and we have more attempts
			if attempt < maxRetries && s.isRateLimitError(err) {
				waitTime := time.Duration(attempt*2) * time.Second
				log.Printf("[OpenRouter] Rate limited, waiting %v before retry...", waitTime)
				time.Sleep(waitTime)
				continue
			}

			// If it's the last attempt or not a rate limit error, use fallback
			log.Printf("[OpenRouter] All attempts failed, using fallback duration: %d", fallbackDuration)
			return fallbackDuration, nil
		}

		if len(response.Choices) == 0 {
			log.Printf("[OpenRouter] No choices received in response")
			continue
		}

		// Parse the response to extract minutes
		content := response.Choices[0].Message.Content
		log.Printf("[OpenRouter] AI response content: '%s'", content)

		var minutes int
		if _, err := fmt.Sscanf(content, "%d", &minutes); err != nil {
			log.Printf("[OpenRouter] Failed to parse duration from response: %v", err)
			log.Printf("[OpenRouter] Using fallback duration: %d", fallbackDuration)
			return fallbackDuration, nil
		}

		log.Printf("[OpenRouter] Parsed duration: %d minutes", minutes)

		// Ensure reasonable bounds (1 minute to 8 hours)
		if minutes < 1 {
			log.Printf("[OpenRouter] Duration too low (%d), setting to 1 minute", minutes)
			minutes = 1
		} else if minutes > 480 {
			log.Printf("[OpenRouter] Duration too high (%d), capping at 480 minutes", minutes)
			minutes = 480
		}

		log.Printf("[OpenRouter] Final estimated duration: %d minutes", minutes)
		return minutes, nil
	}

	// If we get here, all attempts failed
	log.Printf("[OpenRouter] All attempts exhausted, using fallback duration: %d", fallbackDuration)
	return fallbackDuration, nil
}

// calculateFallbackDuration provides smart fallback duration calculation
func (s *OpenRouterService) calculateFallbackDuration(title, description string) int {
	titleLen := len(title)
	descLen := len(description)

	// Base duration calculation
	var baseDuration int

	// Analyze title for complexity indicators
	complexityWords := []string{"設計", "開発", "実装", "分析", "研究", "調査", "プロジェクト", "会議", "打ち合わせ"}
	simpleWords := []string{"買い物", "食事", "休憩", "電話", "メール", "確認"}

	isComplex := false
	isSimple := false

	for _, word := range complexityWords {
		if contains(title+description, word) {
			isComplex = true
			break
		}
	}

	for _, word := range simpleWords {
		if contains(title+description, word) {
			isSimple = true
			break
		}
	}

	if isComplex {
		baseDuration = 90 // Complex tasks: 1.5 hours
	} else if isSimple {
		baseDuration = 15 // Simple tasks: 15 minutes
	} else {
		// Medium complexity based on length
		if titleLen < 10 {
			baseDuration = 20
		} else if titleLen < 20 {
			baseDuration = 35
		} else {
			baseDuration = 50
		}
	}

	// Adjust based on description length
	if descLen > 100 {
		baseDuration += 20
	} else if descLen > 50 {
		baseDuration += 10
	}

	// Ensure bounds
	if baseDuration < 5 {
		baseDuration = 5
	} else if baseDuration > 240 {
		baseDuration = 240
	}

	return baseDuration
}

// isRateLimitError checks if the error is a rate limit error
func (s *OpenRouterService) isRateLimitError(err error) bool {
	return err != nil && (contains(err.Error(), "429") ||
		contains(err.Error(), "rate") ||
		contains(err.Error(), "rate-limited"))
}

// contains checks if a string contains a substring (helper function)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsMiddle(s, substr))))
}

// containsMiddle checks if substr is in the middle of s
func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GenerateTaskSuggestions generates task breakdown suggestions
func (s *OpenRouterService) GenerateTaskSuggestions(title, description string) ([]string, error) {
	messages := []Message{
		{
			Role: "system",
			Content: `You are a productivity assistant. Given a task title and description, break it down into 3-5 smaller, actionable subtasks.

Rules:
- Return each subtask on a new line
- Keep subtasks specific and actionable
- Each subtask should be completable in 15-30 minutes
- Use simple, clear language`,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Task: %s\nDescription: %s\n\nBreak this down into smaller subtasks:", title, description),
		},
	}

	response, err := s.Chat("openai/gpt-oss-20b:free", messages)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response choices received")
	}

	content := response.Choices[0].Message.Content
	// Split by newlines and clean up
	suggestions := []string{}
	for _, line := range bytes.Split([]byte(content), []byte("\n")) {
		suggestion := string(bytes.TrimSpace(line))
		if suggestion != "" && len(suggestion) > 3 {
			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions, nil
}
