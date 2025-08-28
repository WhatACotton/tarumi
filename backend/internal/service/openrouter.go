// InferDurationAndGenreWithOpenRouter infers both duration and genre in a single API call
package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ...existing code...

// InferDurationAndGenreWithOpenRouter infers both duration and genre in a single API call
func InferDurationAndGenreWithOpenRouter(title, description string) (int, string, error) {
	svc := NewOpenRouterService()
	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant for todo tasks. Given a title and description, respond ONLY with a valid JSON object like: {\"duration\": <minutes>, \"genre\": <genre>} where duration is the estimated minutes to complete and genre is one of: work, study, life, meal, health, shopping, other.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Title: %s\nDescription: %s\n\nEstimate duration and genre.", title, description),
		},
	}
	resp, err := svc.Chat("openai/gpt-oss-20b:free", messages)
	if err != nil {
		return 45, "other", err
	}
	if len(resp.Choices) == 0 {
		return 45, "other", fmt.Errorf("no response from OpenRouter")
	}
	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	var result struct {
		Duration int    `json:"duration"`
		Genre    string `json:"genre"`
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// fallback: try to parse manually
		var duration int
		var genre string
		_, scanErr := fmt.Sscanf(content, "{duration:%d, genre:%s}", &duration, &genre)
		if scanErr != nil {
			return 45, "other", fmt.Errorf("failed to parse response: %v, content: %s", err, content)
		}
		return duration, genre, nil
	}
	genre := strings.ToLower(strings.TrimSpace(result.Genre))
	if genre == "" {
		genre = "other"
	}
	if result.Duration < 1 {
		result.Duration = 45
	}
	return result.Duration, genre, nil
}

type OpenRouterService struct {
	apiKey         string
	baseURL        string
	client         *http.Client
	fallbackModels []string
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
		fallbackModels: []string{
			"mistralai/mistral-7b-instruct:free",
			"microsoft/phi-3-mini-128k-instruct:free",
			"meta-llama/llama-3.2-1b-instruct:free",
			"gryphe/mythomist-7b:free",
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

// ScheduleTasksWithAI uses AI to create a complete schedule with specific time slots
func (s *OpenRouterService) ScheduleTasksWithAI(tasks []*TaskToSchedule, currentTime time.Time) ([]*AIScheduledTask, error) {
	return s.ScheduleTasksWithAIAndStartTime(tasks, currentTime, 6, 30) // デフォルト 6:30 AM
}

// ScheduleTasksWithAIAndStartTime uses AI to create a complete schedule with custom start time
func (s *OpenRouterService) ScheduleTasksWithAIAndStartTime(tasks []*TaskToSchedule, currentTime time.Time, startHour, startMinute int) ([]*AIScheduledTask, error) {
	if len(tasks) == 0 {
		return nil, fmt.Errorf("no tasks provided")
	}

	// Format current time in JST
	jst := time.FixedZone("JST", 9*60*60)
	currentJST := currentTime.In(jst)

	// Build task list for AI
	taskList := ""
	for i, task := range tasks {
		deadlineStr := "No deadline"
		if !task.Deadline.IsZero() {
			deadlineJST := task.Deadline.In(jst)
			deadlineStr = deadlineJST.Format("2006-01-02 15:04")
		}

		taskList += fmt.Sprintf("%d. Title: \"%s\"\n   Description: \"%s\"\n   Duration: %d minutes\n   Deadline: %s\n\n",
			i+1, task.Title, task.Description, task.DurationMinutes, deadlineStr)
	}

	messages := []Message{
		{
			Role: "system",
			Content: fmt.Sprintf(`You are an intelligent scheduling assistant. Create a complete daily schedule with specific time slots for the given tasks.

CURRENT TIME: %s (JST)

SCHEDULING RULES (STRICT!):
1. Schedule tasks from current time onward (don't schedule in the past)
2. Available time: %02d:%02d-23:00 (11pm), reasonable working hours only
3. Prioritize tasks with earlier deadlines
4. Respect task duration requirements exactly
5. Consider appropriate time slots:
			 - Morning (%02d:%02d-10:00): breakfast, morning routines, exercise
			 - Mid-morning (10:00-12:00): productive tasks, work
			 - Lunch (12:00-14:00): lunch, break, rest
			 - Afternoon (14:00-17:00): work, meetings, productive tasks
			 - Evening (17:00-21:00): dinner, evening routines
			 - Night (21:00-23:00): relaxation, personal time
6. Leave reasonable gaps between tasks (15-30 minutes)
7. If a task cannot fit before its deadline, set start_time and end_time to null
8. CRITICAL: 食事系タスク（"朝ごはん", "昼飯", "晩ごはん" など）は必ず該当時間帯（朝ごはん: Morning, 昼飯: Lunch, 晩ごはん: Evening）に割り当てること。他の時間帯に割り当ててはならない。絶対に守ること。
	 MEAL TASKS ("breakfast", "lunch", "dinner") MUST be scheduled ONLY in their correct time slots (breakfast: Morning, lunch: Lunch, dinner: Evening). DO NOT schedule meal tasks in any other time slot. THIS IS STRICTLY ENFORCED.
9. DO NOT add, invent, or create any new tasks. Only schedule the tasks provided in the list. 新しいタスクを絶対に追加しないこと。与えられたタスクのみをスケジュールすること。For example, do NOT add tasks like "whatacotton invoker", "hahaha", or any other task not in the original list. 例: 「whatacotton invoker」「hahaha」など元リストにないタスクは絶対に追加しない。
10. All date/time values (start_time, end_time) MUST be in RFC3339 format (e.g. "2025-08-25T14:00:00+09:00"). Do NOT use any other format. 日時は必ずRFC3339形式で返すこと。

CRITICAL: Always use the EXACT original title and description provided. Do not modify, translate, or change the task titles. タイトルや説明文は絶対に変更・翻訳・加工しないこと。

Return ONLY a valid JSON array with this exact format:
[
	{
		"title": "EXACT original task title (DO NOT CHANGE)",
		"description": "EXACT original task description (DO NOT CHANGE)",
		"start_time": "2025-08-25T14:00:00+09:00",
		"end_time": "2025-08-25T14:45:00+09:00",
		"reasoning": "Why this time slot was chosen",
		"priority": 1-4,
		"fits_deadline": true
	}
]

If a task cannot be scheduled before its deadline, set start_time and end_time to null and fits_deadline to false.

IMPORTANT: Return only the JSON array, no other text.`, currentJST.Format("2006-01-02 15:04")),
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Create a schedule for these tasks:\n\n%sSchedule them appropriately starting from the current time.", taskList),
		},
	}

	// Use fallback models
	response, err := s.chatWithFallback(messages)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI scheduling response: %w", err)
	}

	// Parse the JSON response
	scheduledTasks, err := s.parseAIScheduledTasksResponse(response.Choices[0].Message.Content, jst)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI scheduling response: %w", err)
	}

	log.Printf("[OpenRouter] Successfully generated AI schedule for %d tasks", len(scheduledTasks))
	return scheduledTasks, nil
}

type AIScheduledTask struct {
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	StartTime    *time.Time `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	Reasoning    string     `json:"reasoning"`
	Priority     int        `json:"priority"`
	FitsDeadline bool       `json:"fits_deadline"`
}

func (s *OpenRouterService) parseAIScheduledTasksResponse(response string, timezone *time.Location) ([]*AIScheduledTask, error) {
	// Clean the response
	response = strings.TrimSpace(response)
	response = strings.Trim(response, "`")
	if strings.HasPrefix(response, "json") {
		response = strings.TrimPrefix(response, "json")
		response = strings.TrimSpace(response)
	}

	// JSON配列の開始位置を探す
	idx := strings.Index(response, "[")
	if idx != -1 {
		response = response[idx:]
	}

	log.Printf("[OpenRouter] AI scheduling response preview: %s", response[:min(200, len(response))])

	var rawTasks []map[string]interface{}
	if err := json.Unmarshal([]byte(response), &rawTasks); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w\nAI response: %s", err, response)
	}

	var scheduledTasks []*AIScheduledTask
	for _, raw := range rawTasks {
		task := &AIScheduledTask{
			Title:        getString(raw, "title"),
			Description:  getString(raw, "description"),
			Reasoning:    getString(raw, "reasoning"),
			Priority:     getInt(raw, "priority"),
			FitsDeadline: getBool(raw, "fits_deadline"),
		}

		// Parse start_time (can be null)
		if startTimeStr := getString(raw, "start_time"); startTimeStr != "" && startTimeStr != "null" {
			if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
				startTime = startTime.In(timezone)
				task.StartTime = &startTime
			}
		}

		// Parse end_time (can be null)
		if endTimeStr := getString(raw, "end_time"); endTimeStr != "" && endTimeStr != "null" {
			if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
				endTime = endTime.In(timezone)
				task.EndTime = &endTime
			}
		}

		scheduledTasks = append(scheduledTasks, task)
	}

	log.Printf("[OpenRouter] Successfully parsed %d AI scheduled tasks", len(scheduledTasks))
	return scheduledTasks, nil
}

// chatWithFallback tries multiple models in order until one succeeds
func (s *OpenRouterService) chatWithFallback(messages []Message) (*OpenRouterResponse, error) {
	for i, model := range s.fallbackModels {
		log.Printf("[OpenRouter] Trying model %d/%d: %s", i+1, len(s.fallbackModels), model)

		response, err := s.Chat(model, messages)
		if err != nil {
			if s.isRateLimitError(err) && i < len(s.fallbackModels)-1 {
				log.Printf("[OpenRouter] Rate limit hit for %s, trying next model", model)
				continue
			}
			return nil, err
		}

		log.Printf("[OpenRouter] Successfully used model: %s", model)
		return response, nil
	}

	return nil, fmt.Errorf("all fallback models failed")
}

// Helper functions for parsing JSON
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return int(f)
		}
		if i, ok := v.(int); ok {
			return i
		}
	}
	return 0
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// InferGenreWithOpenRouter infers the genre of a task using OpenRouter
func InferGenreWithOpenRouter(title, description string) (string, error) {
	svc := NewOpenRouterService()
	messages := []Message{
		{
			Role:    "system",
			Content: `You are a helpful assistant that classifies todo tasks into genres. Given a title and description, return only the most appropriate genre as a single word (e.g. "work", "study", "life", "health", "shopping", "other").`,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Title: %s\nDescription: %s\n\nWhat is the genre?", title, description),
		},
	}
	resp, err := svc.Chat("openai/gpt-oss-20b:free", messages)
	if err != nil {
		return "other", err
	}
	if len(resp.Choices) == 0 {
		return "other", fmt.Errorf("no response from OpenRouter")
	}
	genre := strings.TrimSpace(resp.Choices[0].Message.Content)
	// 余計な改行や記号を除去
	genre = strings.ToLower(genre)
	genre = strings.Split(genre, "\n")[0]
	return genre, nil
}
