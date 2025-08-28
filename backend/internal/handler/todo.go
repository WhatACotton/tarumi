package handler

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whatacotton/tarumi/internal/middleware"
	"github.com/whatacotton/tarumi/internal/models"
	"github.com/whatacotton/tarumi/internal/repository"
	"github.com/whatacotton/tarumi/internal/service"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func HandleTodo(r *gin.Engine) {
	r.Use(middleware.AuthMiddleware)
	authHandler := r.Group("/todo")
	h := &TodoHandler{
		repo: repository.NewTodoRepository(),
	}
	authHandler.Use(middleware.AuthMiddleware)

	authHandler.POST("/create", h.CreateTodo)
	authHandler.POST("/update", h.UpdateTodo)
	authHandler.GET("/complete", h.GetCompleteTodos)
	authHandler.GET("/incomplete", h.GetIncompleteTodos)
	authHandler.POST("/complete", h.CompleteTodo)
	authHandler.POST("/incomplete", h.InCompleteTodo)
	authHandler.GET("/delete", h.DeleteTodo)
	authHandler.GET("/get", h.GetTodoByID)
	authHandler.GET("/list", h.GetTodosByUserID)
	authHandler.GET("/group", h.GetTodosByGroupID)
	authHandler.POST("/schedule-to-calendar", h.ScheduleTodosToCalendar)
	authHandler.GET("/summary", h.GetTodoSummary)
}

type TodoHandler struct {
	repo repository.TodoRepository
}

func (h *TodoHandler) CreateTodo(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	todoPayload := models.TodoRegisterPayload{}
	if err := c.ShouldBindJSON(&todoPayload); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid request payload"})
		return
	}
	todo, err := h.repo.CreateTodo(
		userID.(string),
		todoPayload,
	)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to create todo"})
		return
	}
	c.JSON(200, gin.H{"todo": todo})
}

func (h *TodoHandler) UpdateTodo(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	todoID, exists := c.GetQuery("id")
	if !exists || todoID == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Todo ID not found"})
		return
	}
	todoPayload := models.TodoUpdatePayload{}
	if err := c.ShouldBindJSON(&todoPayload); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid request payload"})
		return
	}
	todo, err := h.repo.UpdateTodo(
		userID.(string),
		todoID,
		todoPayload,
	)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to update todo"})
		return
	}
	c.JSON(200, gin.H{"todo": todo})
}

func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	todoID, exists := c.GetQuery("id")
	if !exists || todoID == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Todo ID not found"})
		return
	}
	err := h.repo.DeleteTodo(userID.(string), todoID)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to delete todo"})
		return
	}
	c.JSON(200, gin.H{"message": "Todo deleted successfully"})
}

func (h *TodoHandler) GetTodoByID(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	todoID, exists := c.GetQuery("id")
	if !exists || todoID == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Todo ID not found"})
		return
	}
	todo, err := h.repo.GetTodoByID(userID.(string), todoID)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to retrieve todo"})
		return
	}
	c.JSON(200, gin.H{"todo": todo})
}

func (h *TodoHandler) GetTodosByUserID(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	todos, err := h.repo.GetTodosByUserID(userID.(string))
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to retrieve todos"})
		return
	}
	c.JSON(200, gin.H{"todos": todos})
}

func (h *TodoHandler) GetTodosByGroupID(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	groupID, exists := c.GetQuery("group_id")
	if !exists || groupID == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Group ID not found"})
		return
	}
	todos, err := h.repo.GetTodosByGroupID(userID.(string), groupID)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to retrieve todos by group"})
		return
	}
	c.JSON(200, gin.H{"todos": todos})
}

func (r *TodoHandler) CompleteTodo(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	todoID, exists := c.GetQuery("id")
	if !exists || todoID == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Todo ID not found"})
		return
	}

	// CompleteTodoは完了フラグとcompletedAtを設定する専用のハンドラとして扱う
	completedTrue := true
	now := time.Now()

	p := models.TodoUpdatePayload{
		IsCompleted: &completedTrue,
		CompletedAt: &now,
	}

	// リクエストボディに追加のフィールドがあれば取得
	requestPayload := models.TodoUpdatePayload{}
	if err := c.ShouldBindJSON(&requestPayload); err == nil {
		// 完了時間以外のフィールドがリクエストに含まれている場合はマージ
		if requestPayload.ConsumedTime != nil {
			p.ConsumedTime = requestPayload.ConsumedTime
		}
		if requestPayload.Title != nil {
			p.Title = requestPayload.Title
		}
		if requestPayload.Description != nil {
			p.Description = requestPayload.Description
		}
		if requestPayload.DueDate != nil {
			p.DueDate = requestPayload.DueDate
		}
		if requestPayload.ParentID != nil {
			p.ParentID = requestPayload.ParentID
		}
		if requestPayload.GroupID != nil {
			p.GroupID = requestPayload.GroupID
		}
	}

	todo, err := r.repo.UpdateTodo(
		userID.(string),
		todoID,
		p,
	)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to complete todo"})
		return
	}
	c.JSON(200, gin.H{"todo": todo})
}

func (r *TodoHandler) InCompleteTodo(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	todoID, exists := c.GetQuery("id")
	if !exists || todoID == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Todo ID not found"})
		return
	}
	t := false
	p := models.TodoUpdatePayload{
		IsCompleted: &t,
	}
	todo, err := r.repo.UpdateTodo(
		userID.(string),
		todoID,
		p,
	)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to incomplete todo"})
		return
	}
	c.JSON(200, gin.H{"todo": todo})
}

func (r *TodoHandler) GetCompleteTodos(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	todos, err := r.repo.GetCompleteTodos(userID.(string))
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to retrieve complete todos"})
		return
	}
	c.JSON(200, gin.H{"todos": todos})
}

func (r *TodoHandler) GetIncompleteTodos(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	todos, err := r.repo.GetIncompleteTodos(userID.(string))
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to retrieve incomplete todos"})
		return
	}
	c.JSON(200, gin.H{"todos": todos})
}

type ScheduleTodosRequest struct {
	AccessToken string `json:"access_token" binding:"required"`
}

// ScheduleTodosToCalendar schedules incomplete todos to calendar avoiding existing events
func (h *TodoHandler) ScheduleTodosToCalendar(c *gin.Context) {
	log.Printf("[TODO_HANDLER] ScheduleTodosToCalendar started")

	// Parse request
	var req ScheduleTodosRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[TODO_HANDLER] Error binding request: %v", err)
		c.JSON(400, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// Get user ID
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		log.Printf("[TODO_HANDLER] User ID not found in context")
		c.JSON(500, gin.H{"error": "User ID not found"})
		return
	}
	userIDStr := userID.(string)
	log.Printf("[TODO_HANDLER] Processing request for user: %s", userIDStr)

	// Create calendar service
	calendarService, err := h.createCalendarService(c, req.AccessToken)
	if err != nil {
		log.Printf("[TODO_HANDLER] Error creating calendar service: %v", err)
		c.JSON(500, gin.H{"error": "Failed to create calendar service", "details": err.Error()})
		return
	}

	cs := service.NewCalendarService(calendarService)

	// Find tarumi_calendar
	tarumiCalendar, err := cs.FindTarumiCalendar()
	if err != nil {
		log.Printf("[TODO_HANDLER] Error finding tarumi_calendar: %v", err)
		c.JSON(500, gin.H{"error": "Failed to find tarumi_calendar", "details": err.Error()})
		return
	}

	var calendarID string
	if tarumiCalendar == nil {
		// Create tarumi_calendar if it doesn't exist
		log.Printf("[TODO_HANDLER] tarumi_calendar not found, creating new one")
		newCalendar, err := cs.CreateCalendar("tarumi_calendar", "Master calendar for tarumi todos")
		if err != nil {
			log.Printf("[TODO_HANDLER] Error creating tarumi_calendar: %v", err)
			c.JSON(500, gin.H{"error": "Failed to create tarumi_calendar", "details": err.Error()})
			return
		}
		calendarID = newCalendar.Id
		log.Printf("[TODO_HANDLER] Created tarumi_calendar with ID: %s", calendarID)
	} else {
		// Delete existing tarumi_calendar and create a new one for fresh scheduling
		log.Printf("[TODO_HANDLER] tarumi_calendar found, deleting and recreating for fresh scheduling")
		oldCalendarID := tarumiCalendar.Id

		err := cs.DeleteCalendar(oldCalendarID)
		if err != nil {
			log.Printf("[TODO_HANDLER] Warning: Failed to delete existing tarumi_calendar: %v", err)
			// Continue anyway, try to create new calendar
		} else {
			log.Printf("[TODO_HANDLER] Successfully deleted old tarumi_calendar: %s", oldCalendarID)
		}

		// Create new tarumi_calendar
		newCalendar, err := cs.CreateCalendar("tarumi_calendar", "Master calendar for tarumi todos")
		if err != nil {
			log.Printf("[TODO_HANDLER] Error creating new tarumi_calendar: %v", err)
			c.JSON(500, gin.H{"error": "Failed to create new tarumi_calendar", "details": err.Error()})
			return
		}
		calendarID = newCalendar.Id
		log.Printf("[TODO_HANDLER] Created new tarumi_calendar with ID: %s", calendarID)
	}

	// Get current time and set time range for next week
	now := time.Now()
	oneWeekLater := now.AddDate(0, 0, 7)
	log.Printf("[TODO_HANDLER] Scheduling time range: %s to %s", now.Format(time.RFC3339), oneWeekLater.Format(time.RFC3339))

	// Since we created a fresh calendar, generate all available business hour slots
	availableSlots := cs.GenerateBusinessHourSlots(now, oneWeekLater)
	log.Printf("[TODO_HANDLER] Generated %d business hour slots for fresh calendar", len(availableSlots))

	// Get incomplete todos
	todos, err := h.repo.GetIncompleteTodos(userIDStr)
	if err != nil {
		log.Printf("[TODO_HANDLER] Error getting incomplete todos: %v", err)
		c.JSON(500, gin.H{"error": "Failed to get incomplete todos", "details": err.Error()})
		return
	}
	log.Printf("[TODO_HANDLER] Found %d incomplete todos", len(todos))

	// Convert todos to tasks to schedule
	var tasksToSchedule []service.TaskToSchedule
	for _, todo := range todos {
		// Ensure minimum duration of 30 minutes
		duration := todo.Content.Duration
		if duration < 30 {
			duration = 30
		}

		// Debug: Log the original DueDate and timezone info
		log.Printf("[TODO_HANDLER] Todo '%s': DueDate=%s (UTC: %s, JST: %s)",
			todo.Content.Title,
			todo.Content.DueDate.Format("2006-01-02 15:04:05 MST"),
			todo.Content.DueDate.UTC().Format("2006-01-02 15:04:05 MST"),
			todo.Content.DueDate.In(time.FixedZone("JST", 9*3600)).Format("2006-01-02 15:04:05 MST"))

		task := service.TaskToSchedule{
			Title:           todo.Content.Title,
			Description:     todo.Content.Description,
			DurationMinutes: duration,
			Deadline:        todo.Content.DueDate,
		}
		tasksToSchedule = append(tasksToSchedule, task)
	}

	// Schedule tasks using intelligent scheduling with fallback to multiple days (max 7 days)
	scheduledEvents, err := cs.ScheduleTasksIntelligentlyWithFallback(calendarID, tasksToSchedule, 7)
	if err != nil {
		log.Printf("[TODO_HANDLER] Error scheduling tasks: %v", err)
		c.JSON(500, gin.H{"error": "Failed to schedule tasks", "details": err.Error()})
		return
	}

	log.Printf("[TODO_HANDLER] Successfully scheduled %d events", len(scheduledEvents))

	c.JSON(200, gin.H{
		"message":          "Todos scheduled successfully",
		"calendar_id":      calendarID,
		"calendar_name":    "tarumi_calendar",
		"scheduled_count":  len(scheduledEvents),
		"total_todos":      len(todos),
		"available_slots":  len(availableSlots),
		"scheduled_events": scheduledEvents,
	})
}

// createCalendarService creates a Google Calendar service from OAuth token
func (h *TodoHandler) createCalendarService(c *gin.Context, accessToken string) (*calendar.Service, error) {
	// Create OAuth2 token
	token := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(1 * time.Hour),
	}

	// Create OAuth2 config
	var clientID, clientSecret string
	clientID = os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	clientSecret = os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{calendar.CalendarScope, calendar.CalendarEventsScope},
		Endpoint:     google.Endpoint,
	}

	// Create HTTP client with OAuth2 token
	client := config.Client(context.Background(), token)

	// Create calendar service
	return calendar.NewService(c.Request.Context(), option.WithHTTPClient(client))
}

func (h *TodoHandler) GetTodoSummary(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}

	todos, err := h.repo.GetTodosByUserID(userID.(string))
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to retrieve todos"})
		return
	}

	// DBから取得したジャンルで集計
	genreMap := make(map[string]int)
	for _, todo := range todos {
		genre := todo.Content.Genre
		if genre == "" {
			genre = "other"
		}
		genreMap[genre]++
	}

	// 集計したジャンルごとの件数をTodoSummaryスライスに変換
	var summaries []TodoSummary
	for genre, count := range genreMap {
		summaries = append(summaries, TodoSummary{
			Genre: genre,
			Count: count,
		})
	}
	fmt.Printf("Inferred genres: %v\n", genreMap)

	// JSONレスポンスとして返却
	c.JSON(200, gin.H{"summaries": summaries})

}

type TodoSummary struct {
	Genre string `json:"genre"`
	Count int    `json:"count"`
}
