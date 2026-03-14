package tools

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ricky/oc-companion/internal/api"
)

type fakeGmailService struct {
	message GmailMessage
	err     error
}

func (f fakeGmailService) GetMessage(context.Context, string) (GmailMessage, error) {
	if f.err != nil {
		return GmailMessage{}, f.err
	}

	return f.message, nil
}

type fakeCalendarService struct {
	events []CalendarEvent
	err    error
}

func (f fakeCalendarService) ListEvents(context.Context, string, time.Time, time.Time, int) ([]CalendarEvent, error) {
	if f.err != nil {
		return nil, f.err
	}

	return f.events, nil
}

func TestRegister_ExposesToolMethods(t *testing.T) {
	registry := api.NewRegistry()
	err := Register(registry, Services{
		Gmail:    fakeGmailService{},
		Calendar: fakeCalendarService{},
	})
	if err != nil {
		t.Fatalf("expected register success: %v", err)
	}

	discovery := registry.Discover()
	if len(discovery) < 4 {
		t.Fatalf("expected tool methods in discovery output, got %d methods", len(discovery))
	}
}

func TestGmailGetMessageHandler_ValidatesParams(t *testing.T) {
	handler := newGmailGetMessageHandler(fakeGmailService{})
	_, apiErr := handler(context.Background(), json.RawMessage(`{}`))
	if apiErr == nil || apiErr.Message != "message_id is required" {
		t.Fatalf("expected missing message_id error, got %+v", apiErr)
	}
}

func TestCalendarListEventsHandler_ValidatesWindow(t *testing.T) {
	handler := newCalendarListEventsHandler(fakeCalendarService{})
	_, apiErr := handler(context.Background(), json.RawMessage(`{"start":"2026-03-15T00:00:00Z","end":"2026-03-14T00:00:00Z"}`))
	if apiErr == nil || apiErr.Message != "end must be after start" {
		t.Fatalf("expected end-after-start error, got %+v", apiErr)
	}
}
