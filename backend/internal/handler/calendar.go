package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whatacotton/tarumi/internal/middleware"
	"github.com/whatacotton/tarumi/internal/service"
	"google.golang.org/api/calendar/v3"
)

func HandleCalendar(r *gin.Engine) {
	r.Use(middleware.AuthMiddleware)
	r.Use(middleware.CalendarMiddleware)
	h := NewCalendarHandler()
	r.GET("/calendars", h.GetCalendars)
	r.GET("/events", h.GetEvents)
	r.POST("/events", h.CreateEvent)
}

type CalendarHandler struct{}

func NewCalendarHandler() *CalendarHandler {
	return &CalendarHandler{}
}

// GetCalendars retrieves the list of calendars for the authenticated user
func (h *CalendarHandler) GetCalendars(c *gin.Context) {
	calendarService, exists := c.Get("calendarService")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Calendar service not available"})
		return
	}

	googleCalendarService, ok := calendarService.(*calendar.Service)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid calendar service type"})
		return
	}

	cs := service.NewCalendarService(googleCalendarService)
	calendars, err := cs.GetCalendars()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve calendars", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"calendars": calendars})
}

// GetEvents retrieves events from a specific calendar
func (h *CalendarHandler) GetEvents(c *gin.Context) {
	calendarService, exists := c.Get("calendarService")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Calendar service not available"})
		return
	}

	googleCalendarService, ok := calendarService.(*calendar.Service)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid calendar service type"})
		return
	}

	calendarID := c.Query("calendar_id")
	if calendarID == "" {
		calendarID = "primary" // Default to primary calendar
	}

	// Parse time range parameters
	timeMinStr := c.Query("time_min")
	timeMaxStr := c.Query("time_max")

	var timeMin, timeMax time.Time
	var err error

	if timeMinStr != "" {
		timeMin, err = time.Parse(time.RFC3339, timeMinStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time_min format. Use RFC3339 format."})
			return
		}
	} else {
		timeMin = time.Now().AddDate(0, 0, -7) // Default to 7 days ago
	}

	if timeMaxStr != "" {
		timeMax, err = time.Parse(time.RFC3339, timeMaxStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time_max format. Use RFC3339 format."})
			return
		}
	} else {
		timeMax = time.Now().AddDate(0, 0, 7) // Default to 7 days from now
	}

	cs := service.NewCalendarService(googleCalendarService)
	events, err := cs.GetEvents(calendarID, timeMin, timeMax)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve events", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}

type CreateEventRequest struct {
	CalendarID  string    `json:"calendar_id"`
	Summary     string    `json:"summary" binding:"required"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required"`
}

// CreateEvent creates a new event in Google Calendar
func (h *CalendarHandler) CreateEvent(c *gin.Context) {
	calendarService, exists := c.Get("calendarService")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Calendar service not available"})
		return
	}

	googleCalendarService, ok := calendarService.(*calendar.Service)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid calendar service type"})
		return
	}

	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	if req.CalendarID == "" {
		req.CalendarID = "primary"
	}

	if req.EndTime.Before(req.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "End time must be after start time"})
		return
	}

	cs := service.NewCalendarService(googleCalendarService)
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

type UpdateEventRequest struct {
	Summary     *string    `json:"summary"`
	Description *string    `json:"description"`
	Location    *string    `json:"location"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
}

// UpdateEvent updates an existing event in Google Calendar
func (h *CalendarHandler) UpdateEvent(c *gin.Context) {
	calendarService, exists := c.Get("calendarService")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Calendar service not available"})
		return
	}

	googleCalendarService, ok := calendarService.(*calendar.Service)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid calendar service type"})
		return
	}

	calendarID := c.Query("calendar_id")
	if calendarID == "" {
		calendarID = "primary"
	}

	eventID := c.Param("eventId")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event ID is required"})
		return
	}

	var req UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// Get existing event to preserve unchanged fields
	cs := service.NewCalendarService(googleCalendarService)

	// For simplicity, we'll require all fields in this example
	// In a real implementation, you'd fetch the existing event and merge changes
	if req.Summary == nil || req.StartTime == nil || req.EndTime == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Summary, start_time, and end_time are required"})
		return
	}

	if req.EndTime.Before(*req.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "End time must be after start time"})
		return
	}

	event := &service.CalendarEvent{
		Summary:   *req.Summary,
		StartTime: *req.StartTime,
		EndTime:   *req.EndTime,
	}

	if req.Description != nil {
		event.Description = *req.Description
	}
	if req.Location != nil {
		event.Location = *req.Location
	}

	updatedEvent, err := cs.UpdateEvent(calendarID, eventID, event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"event": updatedEvent})
}

// DeleteEvent deletes an event from Google Calendar
func (h *CalendarHandler) DeleteEvent(c *gin.Context) {
	calendarService, exists := c.Get("calendarService")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Calendar service not available"})
		return
	}

	googleCalendarService, ok := calendarService.(*calendar.Service)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid calendar service type"})
		return
	}

	calendarID := c.Query("calendar_id")
	if calendarID == "" {
		calendarID = "primary"
	}

	eventID := c.Param("eventId")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event ID is required"})
		return
	}

	cs := service.NewCalendarService(googleCalendarService)
	err := cs.DeleteEvent(calendarID, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}
