package service

import (
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
	events, err := cs.service.Events.List(calendarID).
		TimeMin(timeMin.Format(time.RFC3339)).
		TimeMax(timeMax.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, err
	}

	var calendarEvents []*CalendarEvent
	for _, event := range events.Items {
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
			}
		} else if event.Start.Date != "" {
			if startTime, err := time.Parse("2006-01-02", event.Start.Date); err == nil {
				calendarEvent.StartTime = startTime
			}
		}

		// Parse end time
		if event.End.DateTime != "" {
			if endTime, err := time.Parse(time.RFC3339, event.End.DateTime); err == nil {
				calendarEvent.EndTime = endTime
			}
		} else if event.End.Date != "" {
			if endTime, err := time.Parse("2006-01-02", event.End.Date); err == nil {
				calendarEvent.EndTime = endTime
			}
		}

		calendarEvents = append(calendarEvents, calendarEvent)
	}

	return calendarEvents, nil
}

// CreateEvent creates a new event in Google Calendar
func (cs *CalendarService) CreateEvent(calendarID string, event *CalendarEvent) (*CalendarEvent, error) {
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

	createdEvent, err := cs.service.Events.Insert(calendarID, googleEvent).Do()
	if err != nil {
		return nil, err
	}

	return &CalendarEvent{
		ID:          createdEvent.Id,
		Summary:     createdEvent.Summary,
		Description: createdEvent.Description,
		Location:    createdEvent.Location,
		StartTime:   event.StartTime,
		EndTime:     event.EndTime,
	}, nil
}

// UpdateEvent updates an existing event in Google Calendar
func (cs *CalendarService) UpdateEvent(calendarID, eventID string, event *CalendarEvent) (*CalendarEvent, error) {
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

	updatedEvent, err := cs.service.Events.Update(calendarID, eventID, googleEvent).Do()
	if err != nil {
		return nil, err
	}

	return &CalendarEvent{
		ID:          updatedEvent.Id,
		Summary:     updatedEvent.Summary,
		Description: updatedEvent.Description,
		Location:    updatedEvent.Location,
		StartTime:   event.StartTime,
		EndTime:     event.EndTime,
	}, nil
}

// DeleteEvent deletes an event from Google Calendar
func (cs *CalendarService) DeleteEvent(calendarID, eventID string) error {
	return cs.service.Events.Delete(calendarID, eventID).Do()
}

// GetCalendars retrieves the list of calendars for the authenticated user
func (cs *CalendarService) GetCalendars() ([]*calendar.CalendarListEntry, error) {
	calendarList, err := cs.service.CalendarList.List().Do()
	if err != nil {
		return nil, err
	}

	return calendarList.Items, nil
}
