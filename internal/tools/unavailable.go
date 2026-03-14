package tools

import (
	"context"
	"errors"
	"time"
)

var errIntegrationNotConfigured = errors.New("integration not configured")

type unavailableGmailService struct{}

func (unavailableGmailService) GetMessage(context.Context, string) (GmailMessage, error) {
	return GmailMessage{}, errIntegrationNotConfigured
}

type unavailableCalendarService struct{}

func (unavailableCalendarService) ListEvents(context.Context, string, time.Time, time.Time, int) ([]CalendarEvent, error) {
	return nil, errIntegrationNotConfigured
}
