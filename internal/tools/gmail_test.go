package tools

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"google.golang.org/api/gmail/v1"
)

type fakeGmailMessagesAPI struct {
	userID    string
	messageID string
	call      *fakeGmailMessageGetCall
}

func (f *fakeGmailMessagesAPI) Get(userID string, messageID string) gmailMessageGetCall {
	f.userID = userID
	f.messageID = messageID
	return f.call
}

type fakeGmailMessageGetCall struct {
	format  string
	ctx     context.Context
	message *gmail.Message
	err     error
}

func (f *fakeGmailMessageGetCall) Format(format string) gmailMessageGetCall {
	f.format = format
	return f
}

func (f *fakeGmailMessageGetCall) Context(ctx context.Context) gmailMessageGetCall {
	f.ctx = ctx
	return f
}

func (f *fakeGmailMessageGetCall) Do(...googleCallOption) (*gmail.Message, error) {
	if f.err != nil {
		return nil, f.err
	}

	return f.message, nil
}

func TestGmailAPIServiceGetMessage_UsesMetadataFormatAndNormalizesMessage(t *testing.T) {
	call := &fakeGmailMessageGetCall{
		message: &gmail.Message{
			Id:           "msg-1",
			ThreadId:     "thread-1",
			InternalDate: 1710117296000,
			Snippet:      "",
			Payload: &gmail.MessagePart{
				MimeType: "multipart/alternative",
				Headers: []*gmail.MessagePartHeader{
					{Name: "From", Value: "Sender <sender@example.com>"},
					{Name: "To", Value: "Receiver <receiver@example.com>"},
					{Name: "Subject", Value: "=?UTF-8?Q?Hello_=E2=9C=93?="},
				},
				Parts: []*gmail.MessagePart{
					{
						MimeType: "text/plain",
						Body: &gmail.MessagePartBody{
							Data: base64.URLEncoding.EncodeToString([]byte("Hello from Gmail body")),
						},
					},
				},
			},
		},
	}
	api := &fakeGmailMessagesAPI{call: call}
	service := &gmailAPIService{messages: api, userID: "mailbox@example.com"}

	message, err := service.GetMessage(context.Background(), "msg-1")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if api.userID != "mailbox@example.com" {
		t.Fatalf("expected configured userID, got %q", api.userID)
	}
	if api.messageID != "msg-1" {
		t.Fatalf("expected message id to be forwarded, got %q", api.messageID)
	}
	if call.format != "full" {
		t.Fatalf("expected full format, got %q", call.format)
	}
	if message.Subject != "Hello ✓" {
		t.Fatalf("expected decoded subject, got %q", message.Subject)
	}
	if message.Snippet != "Hello from Gmail body" {
		t.Fatalf("expected normalized snippet from body, got %q", message.Snippet)
	}
	expectedReceivedAt := time.UnixMilli(1710117296000).UTC()
	if !message.ReceivedAt.Equal(expectedReceivedAt) {
		t.Fatalf("expected received_at %s, got %s", expectedReceivedAt, message.ReceivedAt)
	}
}

func TestResolveGmailUserID(t *testing.T) {
	tests := []struct {
		name             string
		userID           string
		delegatedSubject string
		want             string
	}{
		{
			name: "defaults to me",
			want: "me",
		},
		{
			name:   "uses explicit user id",
			userID: "mailbox@example.com",
			want:   "mailbox@example.com",
		},
		{
			name:             "falls back to delegated subject",
			delegatedSubject: "delegate@example.com",
			want:             "delegate@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveGmailUserID(tt.userID, tt.delegatedSubject); got != tt.want {
				t.Fatalf("expected user id %q, got %q", tt.want, got)
			}
		})
	}
}

func TestGmailAPIServiceGetMessage_ReturnsWrappedError(t *testing.T) {
	service := &gmailAPIService{
		messages: &fakeGmailMessagesAPI{
			call: &fakeGmailMessageGetCall{err: errors.New("boom")},
		},
	}

	_, err := service.GetMessage(context.Background(), "msg-2")
	if err == nil || err.Error() != `gmail messages.get "msg-2": boom` {
		t.Fatalf("expected wrapped gmail get error, got %v", err)
	}
}

func TestNormalizeGmailMessage_FallsBackToDateHeader(t *testing.T) {
	message, err := normalizeGmailMessage(&gmail.Message{
		Id:       "msg-3",
		ThreadId: "thread-3",
		Snippet:  "Header fallback",
		Payload: &gmail.MessagePart{
			Headers: []*gmail.MessagePartHeader{
				{Name: "Date", Value: "Mon, 11 Mar 2024 15:14:56 +0000"},
			},
		},
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	expected := time.Date(2024, time.March, 11, 15, 14, 56, 0, time.UTC)
	if !message.ReceivedAt.Equal(expected) {
		t.Fatalf("expected received_at %s, got %s", expected, message.ReceivedAt)
	}
}

func TestNormalizeGmailMessage_RequiresTimestamp(t *testing.T) {
	_, err := normalizeGmailMessage(&gmail.Message{Id: "msg-4"})
	if err == nil || err.Error() != "gmail message missing received timestamp" {
		t.Fatalf("expected missing timestamp error, got %v", err)
	}
}
