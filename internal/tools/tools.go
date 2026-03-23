package tools

import (
	"context"
	"time"
)

type GmailMessage struct {
	ID         string    `json:"id"`
	ThreadID   string    `json:"thread_id"`
	From       string    `json:"from"`
	To         string    `json:"to"`
	Subject    string    `json:"subject"`
	Snippet    string    `json:"snippet"`
	ReceivedAt time.Time `json:"received_at"`
}

type CalendarEvent struct {
	ID          string    `json:"id"`
	CalendarID  string    `json:"calendar_id"`
	Summary     string    `json:"summary"`
	Description string    `json:"description,omitempty"`
	StartsAt    time.Time `json:"starts_at"`
	EndsAt      time.Time `json:"ends_at"`
	Location    string    `json:"location,omitempty"`
}

type GmailServiceConfig struct {
	CredentialsFile  string
	UserID           string
	DelegatedSubject string
}

type GmailService interface {
	GetMessage(context.Context, string) (GmailMessage, error)
}

type CalendarService interface {
	ListEvents(context.Context, string, time.Time, time.Time, int) ([]CalendarEvent, error)
}

type Services struct {
	Gmail    GmailService
	Calendar CalendarService
}

func NewUnavailableServices() Services {
	return Services{
		Gmail:    unavailableGmailService{},
		Calendar: unavailableCalendarService{},
	}
}
