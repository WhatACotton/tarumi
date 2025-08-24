package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	if apiKey == "" {
		apiKey = "your-openrouter-api-key" // デフォルト値
	}

	return &OpenRouterService{
		apiKey:  apiKey,
		baseURL: "https://openrouter.ai/api/v1",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Chat sends a chat completion request to OpenRouter API
func (s *OpenRouterService) Chat(model string, messages []Message) (*OpenRouterResponse, error) {
	request := OpenRouterRequest{
		Model:    model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/WhatACotton/tarumi")
	req.Header.Set("X-Title", "Tarumi Todo App")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response OpenRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// EstimateTaskDuration uses AI to estimate task duration in minutes
func (s *OpenRouterService) EstimateTaskDuration(title, description string) (int, error) {
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

	response, err := s.Chat("openai/gpt-3.5-turbo", messages)
	if err != nil {
		return 0, fmt.Errorf("failed to get AI response: %w", err)
	}

	if len(response.Choices) == 0 {
		return 0, fmt.Errorf("no response choices received")
	}

	// Parse the response to extract minutes
	content := response.Choices[0].Message.Content
	var minutes int
	if _, err := fmt.Sscanf(content, "%d", &minutes); err != nil {
		// If parsing fails, return a default duration based on title length
		if len(title) < 10 {
			return 15, nil // Simple task default
		} else if len(title) < 30 {
			return 30, nil // Medium task default
		} else {
			return 60, nil // Complex task default
		}
	}

	// Ensure reasonable bounds (1 minute to 8 hours)
	if minutes < 1 {
		minutes = 1
	} else if minutes > 480 {
		minutes = 480
	}

	return minutes, nil
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

	response, err := s.Chat("openai/gpt-3.5-turbo", messages)
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
