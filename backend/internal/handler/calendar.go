package handler

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whatacotton/tarumi/internal/middleware"
	"github.com/whatacotton/tarumi/internal/repository"
	"github.com/whatacotton/tarumi/internal/service"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func HandleCalendar(r *gin.Engine) {
	calendarGroup := r.Group("/calendar")
	calendarGroup.Use(middleware.AuthMiddleware)

	h := NewCalendarHandler()
	calendarGroup.POST("/calendars", h.GetCalendars)
	calendarGroup.POST("/events", h.GetEvents)
	calendarGroup.POST("/events/create", h.CreateEvent)
	calendarGroup.POST("/debug/first-event", h.DebugFirstEvent)
	calendarGroup.POST("/events/monthly", h.GetMonthlyEvents)
	calendarGroup.POST("/gen-todo", h.GenerateTodosToCalendar)
}

type CalendarHandler struct{}

type OAuthPayload struct {
	AccessToken string `json:"access_token" binding:"required"`
	TokenType   string `json:"token_type"`
}

type GetCalendarsRequest struct {
	OAuthPayload
}

type GetEventsRequest struct {
	OAuthPayload
	CalendarID string `json:"calendar_id"`
	TimeMin    string `json:"time_min"`
	TimeMax    string `json:"time_max"`
}

func NewCalendarHandler() *CalendarHandler {
	return &CalendarHandler{}
}

// createCalendarService creates a Google Calendar service from OAuth token
func (h *CalendarHandler) createCalendarService(c *gin.Context, accessToken string) (*calendar.Service, error) {
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

// GetCalendars retrieves the list of calendars for the authenticated user
func (h *CalendarHandler) GetCalendars(c *gin.Context) {
	var req GetCalendarsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	calendarService, err := h.createCalendarService(c, req.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create calendar service", "details": err.Error()})
		return
	}

	cs := service.NewCalendarService(calendarService)
	calendars, err := cs.GetCalendars()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve calendars", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"calendars": calendars})
}

// GetEvents retrieves events from a specific calendar
func (h *CalendarHandler) GetEvents(c *gin.Context) {
	var req GetEventsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	calendarService, err := h.createCalendarService(c, req.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create calendar service", "details": err.Error()})
		return
	}

	calendarID := req.CalendarID
	if calendarID == "" {
		calendarID = "primary" // Default to primary calendar
	}

	// Parse time range parameters
	var timeMin, timeMax time.Time

	if req.TimeMin != "" {
		timeMin, err = time.Parse(time.RFC3339, req.TimeMin)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time_min format. Use RFC3339 format."})
			return
		}
	} else {
		timeMin = time.Now().AddDate(0, 0, -7) // Default to 7 days ago
	}

	if req.TimeMax != "" {
		timeMax, err = time.Parse(time.RFC3339, req.TimeMax)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time_max format. Use RFC3339 format."})
			return
		}
	} else {
		timeMax = time.Now().AddDate(0, 0, 7) // Default to 7 days from now
	}

	cs := service.NewCalendarService(calendarService)
	events, err := cs.GetEvents(calendarID, timeMin, timeMax)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve events", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}

type CreateEventRequest struct {
	OAuthPayload
	CalendarID  string    `json:"calendar_id"`
	Summary     string    `json:"summary" binding:"required"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required"`
}

// CreateEvent creates a new event in Google Calendar
func (h *CalendarHandler) CreateEvent(c *gin.Context) {
	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	calendarService, err := h.createCalendarService(c, req.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create calendar service", "details": err.Error()})
		return
	}

	if req.CalendarID == "" {
		req.CalendarID = "primary"
	}

	if req.EndTime.Before(req.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "End time must be after start time"})
		return
	}

	cs := service.NewCalendarService(calendarService)
	event := &service.CalendarEvent{
		Summary:     req.Summary,
		Description: req.Description,
		Location:    req.Location,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	}

	createdEvent, err := cs.CreateEvent(req.CalendarID, event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"event": createdEvent})
}

// DebugFirstEvent retrieves and logs information about the first calendar's first event
func (h *CalendarHandler) GetMonthlyEvents(c *gin.Context) {
	log.Printf("[CALENDAR] Getting monthly events")

	var req OAuthPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[CALENDAR] Error binding request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	calendarService, err := h.createCalendarService(c, req.AccessToken)
	if err != nil {
		log.Printf("[CALENDAR] Error creating calendar service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create calendar service", "details": err.Error()})
		return
	}

	// 現在時刻から1ヶ月後の時刻を計算
	now := time.Now()
	oneMonthLater := now.AddDate(0, 1, 0)

	log.Printf("[CALENDAR] Time range: %s to %s", now.Format(time.RFC3339), oneMonthLater.Format(time.RFC3339))

	// プライマリカレンダーから全てのイベントを取得
	events, err := calendarService.Events.List("primary").
		TimeMin(now.Format(time.RFC3339)).
		TimeMax(oneMonthLater.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		MaxResults(1000).
		Do()

	if err != nil {
		log.Printf("[CALENDAR] Error listing events: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list events"})
		return
	}

	log.Printf("[CALENDAR] Found %d events in the next month", len(events.Items))

	// service.CalendarEvent形式に変換
	var monthlyEvents []service.CalendarEvent
	for i, item := range events.Items {
		event := service.CalendarEvent{
			ID:          item.Id,
			Summary:     item.Summary,
			Description: item.Description,
			Location:    item.Location,
		}

		// 開始時間の処理
		if item.Start.DateTime != "" {
			if t, err := time.Parse(time.RFC3339, item.Start.DateTime); err == nil {
				event.StartTime = t
			}
		} else if item.Start.Date != "" {
			if t, err := time.Parse("2006-01-02", item.Start.Date); err == nil {
				event.StartTime = t
			}
		}

		// 終了時間の処理
		if item.End.DateTime != "" {
			if t, err := time.Parse(time.RFC3339, item.End.DateTime); err == nil {
				event.EndTime = t
			}
		} else if item.End.Date != "" {
			if t, err := time.Parse("2006-01-02", item.End.Date); err == nil {
				event.EndTime = t
			}
		}

		monthlyEvents = append(monthlyEvents, event)

		// 最初の5個のイベントを詳細ログに出力
		if i < 5 {
			log.Printf("[CALENDAR] Event %d: %s (Start: %v, End: %v)", i+1, item.Summary, event.StartTime, event.EndTime)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"events": monthlyEvents,
		"count":  len(monthlyEvents),
		"period": gin.H{
			"start": now.Format(time.RFC3339),
			"end":   oneMonthLater.Format(time.RFC3339),
		},
	})
}

func (h *CalendarHandler) DebugFirstEvent(c *gin.Context) {
	var req GetCalendarsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	calendarService, err := h.createCalendarService(c, req.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create calendar service", "details": err.Error()})
		return
	}

	// Step 1: Get list of calendars
	log.Println("=== STEP 1: Getting calendars ===")
	cs := service.NewCalendarService(calendarService)
	calendars, err := cs.GetCalendars()
	if err != nil {
		log.Printf("Error getting calendars: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve calendars", "details": err.Error()})
		return
	}

	if len(calendars) == 0 {
		log.Println("No calendars found")
		c.JSON(http.StatusNotFound, gin.H{"error": "No calendars found"})
		return
	}

	// Log first calendar info
	firstCalendar := calendars[0]
	log.Printf("First calendar found:")
	log.Printf("  - ID: %s", firstCalendar.Id)
	log.Printf("  - Summary: %s", firstCalendar.Summary)
	log.Printf("  - Description: %s", firstCalendar.Description)
	log.Printf("  - Primary: %v", firstCalendar.Primary)
	log.Printf("  - TimeZone: %s", firstCalendar.TimeZone)

	// Step 2: Get events from first calendar
	log.Println("=== STEP 2: Getting events from first calendar ===")
	timeMin := time.Now().AddDate(0, -1, 0) // 1 month ago
	timeMax := time.Now().AddDate(0, 1, 0)  // 1 month from now

	log.Printf("Time range: %s to %s", timeMin.Format(time.RFC3339), timeMax.Format(time.RFC3339))

	events, err := cs.GetEvents(firstCalendar.Id, timeMin, timeMax)
	if err != nil {
		log.Printf("Error getting events: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve events", "details": err.Error()})
		return
	}

	if len(events) == 0 {
		log.Println("No events found in the first calendar")
		c.JSON(http.StatusNotFound, gin.H{"error": "No events found in the first calendar"})
		return
	}

	// Log first event info
	firstEvent := events[0]
	log.Printf("First event found:")
	log.Printf("  - ID: %s", firstEvent.ID)
	log.Printf("  - Summary: %s", firstEvent.Summary)
	log.Printf("  - Description: %s", firstEvent.Description)
	log.Printf("  - Location: %s", firstEvent.Location)
	log.Printf("  - Start: %s", firstEvent.StartTime.Format(time.RFC3339))
	log.Printf("  - End: %s", firstEvent.EndTime.Format(time.RFC3339))

	// Return summary info
	c.JSON(http.StatusOK, gin.H{
		"message": "Calendar and event information logged successfully",
		"calendar": gin.H{
			"id":      firstCalendar.Id,
			"summary": firstCalendar.Summary,
		},
		"event": gin.H{
			"id":      firstEvent.ID,
			"summary": firstEvent.Summary,
			"start":   firstEvent.StartTime,
			"end":     firstEvent.EndTime,
		},
	})
}

type GenerateTodosRequest struct {
	OAuthPayload
	CalendarName        string `json:"calendar_name"`
	CalendarDescription string `json:"calendar_description"`
}

// GenerateTodosToCalendar creates a new calendar and registers user's todos as events
func (h *CalendarHandler) GenerateTodosToCalendar(c *gin.Context) {
	log.Printf("[CALENDAR] GenerateTodosToCalendar started")

	// Log the raw request body for debugging
	if body, err := c.GetRawData(); err == nil {
		log.Printf("[CALENDAR] Raw request body: %s", string(body))
		// Reset the request body so it can be read again
		c.Request.Body = io.NopCloser(strings.NewReader(string(body)))
	}

	var req GenerateTodosRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[CALENDAR] Error binding request: %v", err)
		log.Printf("[CALENDAR] Request validation failed")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"required_fields": map[string]string{
				"access_token":         "string (required)",
				"calendar_name":        "string (optional, default: 'tarumi calendar')",
				"calendar_description": "string (optional)",
			},
		})
		return
	}

	// Set default values if not provided
	if req.CalendarName == "" {
		req.CalendarName = "tarumi calendar"
		log.Printf("[CALENDAR] No calendar name provided, using default: '%s'", req.CalendarName)
	}
	if req.CalendarDescription == "" {
		req.CalendarDescription = "Calendar automatically generated from todo tasks"
		log.Printf("[CALENDAR] No calendar description provided, using default: '%s'", req.CalendarDescription)
	}

	log.Printf("[CALENDAR] Successfully parsed request:")
	log.Printf("[CALENDAR]   - Calendar Name: '%s'", req.CalendarName)
	log.Printf("[CALENDAR]   - Calendar Description: '%s'", req.CalendarDescription)
	log.Printf("[CALENDAR]   - Access Token present: %t", req.AccessToken != "")

	// Get user ID from middleware
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists {
		log.Printf("[CALENDAR] User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		log.Printf("[CALENDAR] Invalid user ID type")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	log.Printf("[CALENDAR] Processing request for user: %s", userIDStr)

	// Create calendar service
	calendarService, err := h.createCalendarService(c, req.AccessToken)
	if err != nil {
		log.Printf("[CALENDAR] Error creating calendar service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create calendar service", "details": err.Error()})
		return
	}

	cs := service.NewCalendarService(calendarService)

	// Create new calendar
	log.Printf("[CALENDAR] Creating new calendar: %s", req.CalendarName)
	newCalendar, err := cs.CreateCalendar(req.CalendarName, req.CalendarDescription)
	if err != nil {
		log.Printf("[CALENDAR] Error creating calendar: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create calendar", "details": err.Error()})
		return
	}

	log.Printf("[CALENDAR] Created calendar with ID: %s", newCalendar.Id)

	// Get user's todos
	todoRepo := repository.NewTodoRepository()
	todos, err := todoRepo.GetIncompleteTodos(userIDStr)
	if err != nil {
		log.Printf("[CALENDAR] Error getting todos: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get todos", "details": err.Error()})
		return
	}

	log.Printf("[CALENDAR] Found %d incomplete todos", len(todos))

	// Create events for each todo
	var createdEvents []service.CalendarEvent
	eventCreationTime := time.Now()

	for i, todo := range todos {
		log.Printf("[CALENDAR] Processing todo %d: %s", i+1, todo.Content.Title)

		// Calculate event duration from todo duration (convert minutes to hours)
		eventDuration := time.Duration(todo.Content.Duration) * time.Minute
		if eventDuration < 30*time.Minute {
			eventDuration = 30 * time.Minute // Minimum 30 minutes
		}

		// Create event starting from creation time + offset for each todo
		startTime := eventCreationTime.Add(time.Duration(i) * time.Hour)
		endTime := startTime.Add(eventDuration)

		event := &service.CalendarEvent{
			Summary:     todo.Content.Title,
			Description: todo.Content.Description,
			StartTime:   startTime,
			EndTime:     endTime,
		}

		createdEvent, err := cs.CreateEvent(newCalendar.Id, event)
		if err != nil {
			log.Printf("[CALENDAR] Error creating event for todo %s: %v", todo.ID, err)
			continue // Continue with next todo instead of failing completely
		}

		createdEvents = append(createdEvents, *createdEvent)
		log.Printf("[CALENDAR] Created event: %s (Start: %v, End: %v)", createdEvent.Summary, createdEvent.StartTime, createdEvent.EndTime)
	}

	log.Printf("[CALENDAR] Successfully created %d events in calendar %s", len(createdEvents), newCalendar.Id)

	c.JSON(http.StatusOK, gin.H{
		"message": "Calendar created and todos registered successfully",
		"calendar": gin.H{
			"id":          newCalendar.Id,
			"summary":     newCalendar.Summary,
			"description": newCalendar.Description,
		},
		"events_created":  len(createdEvents),
		"events":          createdEvents,
		"todos_processed": len(todos),
	})
}
