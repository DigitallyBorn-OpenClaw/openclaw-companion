package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ricky/oc-companion/internal/api"
	"github.com/ricky/oc-companion/internal/protocol"
)

const (
	defaultCalendarID  = "primary"
	defaultMaxCalendar = 20
	maxCalendarResults = 100
)

func Register(registry *api.Registry, services Services) error {
	if err := registry.Register(api.Method{
		Name:        "gmail.getMessage",
		Description: "Returns a Gmail message by message ID.",
		Usage:       `{"id":"2","method":"gmail.getMessage","params":{"message_id":"18c2b"}}`,
		Params: map[string]interface{}{
			"message_id": map[string]interface{}{"type": "string", "required": true},
		},
		Handler: newGmailGetMessageHandler(services.Gmail),
	}); err != nil {
		return err
	}

	if err := registry.Register(api.Method{
		Name:        "calendar.listEvents",
		Description: "Returns calendar events for a time window.",
		Usage:       `{"id":"3","method":"calendar.listEvents","params":{"start":"2026-03-14T00:00:00Z","end":"2026-03-15T00:00:00Z","calendar_id":"primary","max_results":20}}`,
		Params: map[string]interface{}{
			"calendar_id": map[string]interface{}{"type": "string", "required": false, "default": "primary"},
			"start":       map[string]interface{}{"type": "RFC3339 timestamp", "required": true},
			"end":         map[string]interface{}{"type": "RFC3339 timestamp", "required": true},
			"max_results": map[string]interface{}{"type": "integer", "required": false, "default": 20, "max": 100},
		},
		Handler: newCalendarListEventsHandler(services.Calendar),
	}); err != nil {
		return err
	}

	return nil
}

type gmailGetMessageParams struct {
	MessageID string `json:"message_id"`
}

func newGmailGetMessageHandler(gmail GmailService) api.Handler {
	return func(ctx context.Context, params json.RawMessage) (interface{}, *protocol.Error) {
		var request gmailGetMessageParams
		if err := json.Unmarshal(params, &request); err != nil {
			return nil, &protocol.Error{Code: protocol.CodeInvalidParams, Message: "invalid params", Details: err.Error()}
		}

		if request.MessageID == "" {
			return nil, &protocol.Error{Code: protocol.CodeInvalidParams, Message: "message_id is required"}
		}

		message, err := gmail.GetMessage(ctx, request.MessageID)
		if err != nil {
			if errors.Is(err, errIntegrationNotConfigured) {
				return nil, &protocol.Error{Code: protocol.CodeInternalError, Message: "gmail integration unavailable"}
			}

			return nil, &protocol.Error{Code: protocol.CodeInternalError, Message: "failed to fetch message", Details: err.Error()}
		}

		return message, nil
	}
}

type calendarListEventsParams struct {
	CalendarID string `json:"calendar_id"`
	Start      string `json:"start"`
	End        string `json:"end"`
	MaxResults int    `json:"max_results"`
}

func newCalendarListEventsHandler(calendar CalendarService) api.Handler {
	return func(ctx context.Context, params json.RawMessage) (interface{}, *protocol.Error) {
		var request calendarListEventsParams
		if err := json.Unmarshal(params, &request); err != nil {
			return nil, &protocol.Error{Code: protocol.CodeInvalidParams, Message: "invalid params", Details: err.Error()}
		}

		if request.CalendarID == "" {
			request.CalendarID = defaultCalendarID
		}

		if request.MaxResults <= 0 {
			request.MaxResults = defaultMaxCalendar
		}

		if request.MaxResults > maxCalendarResults {
			return nil, &protocol.Error{Code: protocol.CodeInvalidParams, Message: fmt.Sprintf("max_results cannot exceed %d", maxCalendarResults)}
		}

		if request.Start == "" || request.End == "" {
			return nil, &protocol.Error{Code: protocol.CodeInvalidParams, Message: "start and end are required"}
		}

		start, err := time.Parse(time.RFC3339, request.Start)
		if err != nil {
			return nil, &protocol.Error{Code: protocol.CodeInvalidParams, Message: "start must be RFC3339", Details: err.Error()}
		}

		end, err := time.Parse(time.RFC3339, request.End)
		if err != nil {
			return nil, &protocol.Error{Code: protocol.CodeInvalidParams, Message: "end must be RFC3339", Details: err.Error()}
		}

		if !end.After(start) {
			return nil, &protocol.Error{Code: protocol.CodeInvalidParams, Message: "end must be after start"}
		}

		events, err := calendar.ListEvents(ctx, request.CalendarID, start, end, request.MaxResults)
		if err != nil {
			if errors.Is(err, errIntegrationNotConfigured) {
				return nil, &protocol.Error{Code: protocol.CodeInternalError, Message: "calendar integration unavailable"}
			}

			return nil, &protocol.Error{Code: protocol.CodeInternalError, Message: "failed to list events", Details: err.Error()}
		}

		return map[string]interface{}{
			"calendar_id": request.CalendarID,
			"start":       start,
			"end":         end,
			"events":      events,
		}, nil
	}
}
