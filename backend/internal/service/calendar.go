package service

import (
	"log"
	"time"

	"google.golang.org/api/calendar/v3"
)

type CalendarService struct {
	service *calendar.Service
}

type CalendarEvent struct {
	ID          string    `json:"id"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Location    string    `json:"location"`
}

func NewCalendarService(calendarService *calendar.Service) *CalendarService {
	return &CalendarService{
		service: calendarService,
	}
}

// GetEvents retrieves events from Google Calendar within the specified time range
func (cs *CalendarService) GetEvents(calendarID string, timeMin, timeMax time.Time) ([]*CalendarEvent, error) {
	log.Printf("[CALENDAR_SERVICE] GetEvents started for calendar: %s", calendarID)
	log.Printf("[CALENDAR_SERVICE] Time range: %s to %s", timeMin.Format(time.RFC3339), timeMax.Format(time.RFC3339))

	events, err := cs.service.Events.List(calendarID).
		TimeMin(timeMin.Format(time.RFC3339)).
		TimeMax(timeMax.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Do()
	if err != nil {
		log.Printf("[CALENDAR_SERVICE] Error listing events from Google Calendar API: %v", err)
		return nil, err
	}

	log.Printf("[CALENDAR_SERVICE] Retrieved %d events from Google Calendar API", len(events.Items))

	var calendarEvents []*CalendarEvent
	for i, event := range events.Items {
		log.Printf("[CALENDAR_SERVICE] Processing event %d: %s (ID: %s)", i+1, event.Summary, event.Id)

		calendarEvent := &CalendarEvent{
			ID:          event.Id,
			Summary:     event.Summary,
			Description: event.Description,
			Location:    event.Location,
		}

		// Parse start time
		if event.Start.DateTime != "" {
			if startTime, err := time.Parse(time.RFC3339, event.Start.DateTime); err == nil {
				calendarEvent.StartTime = startTime
				log.Printf("[CALENDAR_SERVICE] Event start time (datetime): %s", startTime.Format(time.RFC3339))
			} else {
				log.Printf("[CALENDAR_SERVICE] Warning: Failed to parse event start datetime: %v", err)
			}
		} else if event.Start.Date != "" {
			if startTime, err := time.Parse("2006-01-02", event.Start.Date); err == nil {
				calendarEvent.StartTime = startTime
				log.Printf("[CALENDAR_SERVICE] Event start time (date): %s", startTime.Format("2006-01-02"))
			} else {
				log.Printf("[CALENDAR_SERVICE] Warning: Failed to parse event start date: %v", err)
			}
		}

		// Parse end time
		if event.End.DateTime != "" {
			if endTime, err := time.Parse(time.RFC3339, event.End.DateTime); err == nil {
				calendarEvent.EndTime = endTime
				log.Printf("[CALENDAR_SERVICE] Event end time (datetime): %s", endTime.Format(time.RFC3339))
			} else {
				log.Printf("[CALENDAR_SERVICE] Warning: Failed to parse event end datetime: %v", err)
			}
		} else if event.End.Date != "" {
			if endTime, err := time.Parse("2006-01-02", event.End.Date); err == nil {
				calendarEvent.EndTime = endTime
				log.Printf("[CALENDAR_SERVICE] Event end time (date): %s", endTime.Format("2006-01-02"))
			} else {
				log.Printf("[CALENDAR_SERVICE] Warning: Failed to parse event end date: %v", err)
			}
		}

		calendarEvents = append(calendarEvents, calendarEvent)
	}

	log.Printf("[CALENDAR_SERVICE] GetEvents completed successfully. Processed %d events", len(calendarEvents))
	return calendarEvents, nil
}

// CreateEvent creates a new event in Google Calendar
func (cs *CalendarService) CreateEvent(calendarID string, event *CalendarEvent) (*CalendarEvent, error) {
	log.Printf("[CALENDAR_SERVICE] CreateEvent started")
	log.Printf("[CALENDAR_SERVICE] Calendar ID: %s", calendarID)
	log.Printf("[CALENDAR_SERVICE] Event details:")
	log.Printf("[CALENDAR_SERVICE]   - Summary: %s", event.Summary)
	log.Printf("[CALENDAR_SERVICE]   - Description: %s", event.Description)
	log.Printf("[CALENDAR_SERVICE]   - Location: %s", event.Location)
	log.Printf("[CALENDAR_SERVICE]   - Start: %s", event.StartTime.Format(time.RFC3339))
	log.Printf("[CALENDAR_SERVICE]   - End: %s", event.EndTime.Format(time.RFC3339))

	googleEvent := &calendar.Event{
		Summary:     event.Summary,
		Description: event.Description,
		Location:    event.Location,
		Start: &calendar.EventDateTime{
			DateTime: event.StartTime.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: event.EndTime.Format(time.RFC3339),
		},
	}

	log.Printf("[CALENDAR_SERVICE] Calling Google Calendar API to create event...")
	createdEvent, err := cs.service.Events.Insert(calendarID, googleEvent).Do()
	if err != nil {
		log.Printf("[CALENDAR_SERVICE] Error creating event via Google Calendar API: %v", err)
		return nil, err
	}

	log.Printf("[CALENDAR_SERVICE] Event created successfully with ID: %s", createdEvent.Id)

	resultEvent := &CalendarEvent{
		ID:          createdEvent.Id,
		Summary:     createdEvent.Summary,
		Description: createdEvent.Description,
		Location:    createdEvent.Location,
		StartTime:   event.StartTime,
		EndTime:     event.EndTime,
	}

	log.Printf("[CALENDAR_SERVICE] CreateEvent completed successfully")
	return resultEvent, nil
}

// UpdateEvent updates an existing event in Google Calendar
func (cs *CalendarService) UpdateEvent(calendarID, eventID string, event *CalendarEvent) (*CalendarEvent, error) {
	log.Printf("[CALENDAR_SERVICE] UpdateEvent started")
	log.Printf("[CALENDAR_SERVICE] Calendar ID: %s, Event ID: %s", calendarID, eventID)
	log.Printf("[CALENDAR_SERVICE] Updated event details:")
	log.Printf("[CALENDAR_SERVICE]   - Summary: %s", event.Summary)
	log.Printf("[CALENDAR_SERVICE]   - Description: %s", event.Description)
	log.Printf("[CALENDAR_SERVICE]   - Location: %s", event.Location)
	log.Printf("[CALENDAR_SERVICE]   - Start: %s", event.StartTime.Format(time.RFC3339))
	log.Printf("[CALENDAR_SERVICE]   - End: %s", event.EndTime.Format(time.RFC3339))

	googleEvent := &calendar.Event{
		Summary:     event.Summary,
		Description: event.Description,
		Location:    event.Location,
		Start: &calendar.EventDateTime{
			DateTime: event.StartTime.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: event.EndTime.Format(time.RFC3339),
		},
	}

	log.Printf("[CALENDAR_SERVICE] Calling Google Calendar API to update event...")
	updatedEvent, err := cs.service.Events.Update(calendarID, eventID, googleEvent).Do()
	if err != nil {
		log.Printf("[CALENDAR_SERVICE] Error updating event via Google Calendar API: %v", err)
		return nil, err
	}

	log.Printf("[CALENDAR_SERVICE] Event updated successfully with ID: %s", updatedEvent.Id)

	resultEvent := &CalendarEvent{
		ID:          updatedEvent.Id,
		Summary:     updatedEvent.Summary,
		Description: updatedEvent.Description,
		Location:    updatedEvent.Location,
		StartTime:   event.StartTime,
		EndTime:     event.EndTime,
	}

	log.Printf("[CALENDAR_SERVICE] UpdateEvent completed successfully")
	return resultEvent, nil
}

// DeleteEvent deletes an event from Google Calendar
func (cs *CalendarService) DeleteEvent(calendarID, eventID string) error {
	log.Printf("[CALENDAR_SERVICE] DeleteEvent started")
	log.Printf("[CALENDAR_SERVICE] Calendar ID: %s, Event ID: %s", calendarID, eventID)

	log.Printf("[CALENDAR_SERVICE] Calling Google Calendar API to delete event...")
	err := cs.service.Events.Delete(calendarID, eventID).Do()
	if err != nil {
		log.Printf("[CALENDAR_SERVICE] Error deleting event via Google Calendar API: %v", err)
		return err
	}

	log.Printf("[CALENDAR_SERVICE] Event deleted successfully")
	log.Printf("[CALENDAR_SERVICE] DeleteEvent completed successfully")
	return nil
}

// GetCalendars retrieves the list of calendars for the authenticated user
func (cs *CalendarService) GetCalendars() ([]*calendar.CalendarListEntry, error) {
	log.Printf("[CALENDAR_SERVICE] GetCalendars started")

	log.Printf("[CALENDAR_SERVICE] Calling Google Calendar API to list calendars...")
	calendarList, err := cs.service.CalendarList.List().Do()
	if err != nil {
		log.Printf("[CALENDAR_SERVICE] Error listing calendars via Google Calendar API: %v", err)
		return nil, err
	}

	log.Printf("[CALENDAR_SERVICE] Retrieved %d calendars from Google Calendar API", len(calendarList.Items))

	for i, cal := range calendarList.Items {
		log.Printf("[CALENDAR_SERVICE] Calendar %d: %s (ID: %s, Primary: %v)", i+1, cal.Summary, cal.Id, cal.Primary)
	}

	log.Printf("[CALENDAR_SERVICE] GetCalendars completed successfully")
	return calendarList.Items, nil
}

// CreateCalendar creates a new calendar for the user
func (cs *CalendarService) CreateCalendar(summary, description string) (*calendar.Calendar, error) {
	log.Printf("[CALENDAR_SERVICE] CreateCalendar started")
	log.Printf("[CALENDAR_SERVICE] Calendar details:")
	log.Printf("[CALENDAR_SERVICE]   - Summary: %s", summary)
	log.Printf("[CALENDAR_SERVICE]   - Description: %s", description)
	log.Printf("[CALENDAR_SERVICE]   - TimeZone: Asia/Tokyo")

	googleCalendar := &calendar.Calendar{
		Summary:     summary,
		Description: description,
		TimeZone:    "Asia/Tokyo", // Default timezone
	}

	log.Printf("[CALENDAR_SERVICE] Calling Google Calendar API to create calendar...")
	createdCalendar, err := cs.service.Calendars.Insert(googleCalendar).Do()
	if err != nil {
		log.Printf("[CALENDAR_SERVICE] Error creating calendar via Google Calendar API: %v", err)
		return nil, err
	}

	log.Printf("[CALENDAR_SERVICE] Calendar created successfully with ID: %s", createdCalendar.Id)
	log.Printf("[CALENDAR_SERVICE] Created calendar details:")
	log.Printf("[CALENDAR_SERVICE]   - ID: %s", createdCalendar.Id)
	log.Printf("[CALENDAR_SERVICE]   - Summary: %s", createdCalendar.Summary)
	log.Printf("[CALENDAR_SERVICE]   - Description: %s", createdCalendar.Description)
	log.Printf("[CALENDAR_SERVICE]   - TimeZone: %s", createdCalendar.TimeZone)

	log.Printf("[CALENDAR_SERVICE] CreateCalendar completed successfully")
	return createdCalendar, nil
}
