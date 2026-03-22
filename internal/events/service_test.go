package events

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
)

type fakeReceiver struct {
	notification GmailNotification
	err          error
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func (f fakeReceiver) Run(ctx context.Context, handler Handler) error {
	if f.err != nil {
		return f.err
	}

	return handler(ctx, f.notification)
}

func TestWebhookNotifier_Notify(t *testing.T) {
	var authHeader string
	var eventHeader string
	var payload webhookPayload

	notifier := NewWebhookNotifier("http://example.invalid/hooks/gmail", "secret-token")
	notifier.client = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			authHeader = r.Header.Get("Authorization")
			eventHeader = r.Header.Get("X-OpenClaw-Event")
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read request body: %v", err)
			}
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}

			return &http.Response{
				StatusCode: http.StatusAccepted,
				Status:     "202 Accepted",
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}, nil
		}),
	}
	err := notifier.Notify(context.Background(), GmailNotification{
		Account:      "me@example.com",
		HistoryID:    "12345",
		ReceivedAt:   time.Date(2026, 3, 22, 14, 0, 0, 0, time.UTC),
		PubSubID:     "pubsub-1",
		Subscription: "oc-companion-gmail-test",
	})
	if err != nil {
		t.Fatalf("expected notify success: %v", err)
	}

	if authHeader != "Bearer secret-token" {
		t.Fatalf("expected bearer auth header, got %q", authHeader)
	}
	if eventHeader != "gmail.new_message" {
		t.Fatalf("expected event header, got %q", eventHeader)
	}
	if payload.Event != "gmail.new_message" || payload.HistoryID != "12345" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestService_RunDeliversReceivedNotification(t *testing.T) {
	service := NewService(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		fakeReceiver{notification: GmailNotification{
			Account:    "me@example.com",
			HistoryID:  "12345",
			ReceivedAt: time.Now().UTC(),
		}},
		&WebhookNotifier{
			url:   "http://example.invalid/hooks/gmail",
			token: "secret-token",
			client: &http.Client{
				Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     "204 No Content",
						Body:       io.NopCloser(strings.NewReader("")),
						Header:     make(http.Header),
					}, nil
				}),
			},
		},
	)

	if err := service.Run(context.Background()); err != nil {
		t.Fatalf("expected service run success: %v", err)
	}
}

func TestParseGmailNotification(t *testing.T) {
	notification, err := parseGmailNotification("sub-1", &pubsub.Message{
		ID:          "pubsub-1",
		Data:        []byte(`{"emailAddress":"me@example.com","historyId":"12345"}`),
		PublishTime: time.Date(2026, 3, 22, 14, 30, 0, 0, time.UTC),
		Attributes:  map[string]string{"trace": "abc"},
	})
	if err != nil {
		t.Fatalf("expected parse success: %v", err)
	}

	if notification.Account != "me@example.com" || notification.HistoryID != "12345" {
		t.Fatalf("unexpected notification: %+v", notification)
	}
	if notification.Subscription != "sub-1" || notification.Attributes["trace"] != "abc" {
		t.Fatalf("expected subscription metadata to be carried through: %+v", notification)
	}
}
