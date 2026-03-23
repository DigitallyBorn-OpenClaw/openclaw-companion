package tools

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"mime"
	"net/mail"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const gmailReadonlyScope = gmail.GmailReadonlyScope

type gmailMessagesAPI interface {
	Get(userID string, messageID string) gmailMessageGetCall
}

type gmailMessageGetCall interface {
	Format(format string) gmailMessageGetCall
	Context(ctx context.Context) gmailMessageGetCall
	Do(...googleCallOption) (*gmail.Message, error)
}

type googleCallOption = googleapi.CallOption

type gmailAPIService struct {
	messages gmailMessagesAPI
}

func NewGmailService(ctx context.Context, credentialsFile string) (GmailService, error) {
	options := []option.ClientOption{option.WithScopes(gmailReadonlyScope)}
	if credentialsFile = strings.TrimSpace(credentialsFile); credentialsFile != "" {
		options = append(options, option.WithCredentialsFile(credentialsFile))
	}

	service, err := gmail.NewService(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("create gmail service: %w", err)
	}

	return &gmailAPIService{messages: &gmailMessagesClient{service: service.Users.Messages}}, nil
}

func (s *gmailAPIService) GetMessage(ctx context.Context, messageID string) (GmailMessage, error) {
	message, err := s.messages.Get("me", messageID).Format("metadata").Context(ctx).Do()
	if err != nil {
		return GmailMessage{}, fmt.Errorf("gmail messages.get %q: %w", messageID, err)
	}

	return normalizeGmailMessage(message)
}

type gmailMessagesClient struct {
	service *gmail.UsersMessagesService
}

func (c *gmailMessagesClient) Get(userID string, messageID string) gmailMessageGetCall {
	return &gmailMessageCall{call: c.service.Get(userID, messageID)}
}

type gmailMessageCall struct {
	call *gmail.UsersMessagesGetCall
}

func (c *gmailMessageCall) Format(format string) gmailMessageGetCall {
	c.call = c.call.Format(format)
	return c
}

func (c *gmailMessageCall) Context(ctx context.Context) gmailMessageGetCall {
	c.call = c.call.Context(ctx)
	return c
}

func (c *gmailMessageCall) Do(opts ...googleCallOption) (*gmail.Message, error) {
	return c.call.Do(opts...)
}

func normalizeGmailMessage(message *gmail.Message) (GmailMessage, error) {
	if message == nil {
		return GmailMessage{}, errors.New("gmail message is nil")
	}

	receivedAt, err := parseReceivedAt(message.InternalDate, headerValue(message.Payload, "Date"))
	if err != nil {
		return GmailMessage{}, err
	}

	return GmailMessage{
		ID:         strings.TrimSpace(message.Id),
		ThreadID:   strings.TrimSpace(message.ThreadId),
		From:       decodeHeader(headerValue(message.Payload, "From")),
		To:         decodeHeader(headerValue(message.Payload, "To")),
		Subject:    decodeHeader(headerValue(message.Payload, "Subject")),
		Snippet:    normalizeSnippet(message.Snippet, message.Payload),
		ReceivedAt: receivedAt,
	}, nil
}

func headerValue(payload *gmail.MessagePart, name string) string {
	if payload == nil {
		return ""
	}

	for _, header := range payload.Headers {
		if strings.EqualFold(strings.TrimSpace(header.Name), name) {
			return strings.TrimSpace(header.Value)
		}
	}

	return ""
}

func decodeHeader(value string) string {
	decoded := strings.TrimSpace(value)
	if decoded == "" {
		return ""
	}

	decoder := new(mime.WordDecoder)
	if result, err := decoder.DecodeHeader(decoded); err == nil {
		return result
	}

	return decoded
}

func parseReceivedAt(internalDate int64, dateHeader string) (time.Time, error) {
	if internalDate > 0 {
		return time.UnixMilli(internalDate).UTC(), nil
	}

	if value := strings.TrimSpace(dateHeader); value != "" {
		parsed, err := mailParseDate(value)
		if err != nil {
			return time.Time{}, fmt.Errorf("parse date header: %w", err)
		}

		return parsed.UTC(), nil
	}

	return time.Time{}, errors.New("gmail message missing received timestamp")
}

func normalizeSnippet(snippet string, payload *gmail.MessagePart) string {
	snippet = strings.TrimSpace(snippet)
	if snippet != "" {
		return snippet
	}

	if body := strings.TrimSpace(extractTextBody(payload)); body != "" {
		return body
	}

	return ""
}

func extractTextBody(part *gmail.MessagePart) string {
	if part == nil {
		return ""
	}

	if isTextBodyPart(part) {
		if decoded := decodeBody(part.Body); decoded != "" {
			return decoded
		}
	}

	for _, child := range part.Parts {
		if body := extractTextBody(child); body != "" {
			return body
		}
	}

	return ""
}

func isTextBodyPart(part *gmail.MessagePart) bool {
	mimeType := strings.ToLower(strings.TrimSpace(part.MimeType))
	return mimeType == "text/plain" || mimeType == "text/html"
}

func decodeBody(body *gmail.MessagePartBody) string {
	if body == nil || strings.TrimSpace(body.Data) == "" {
		return ""
	}

	decoded, err := base64.URLEncoding.DecodeString(body.Data)
	if err != nil {
		return ""
	}

	text := strings.TrimSpace(string(decoded))
	if text == "" {
		return ""
	}

	text = strings.NewReplacer("\r\n", "\n", "\r", "\n", "\t", " ").Replace(text)
	if strings.Contains(strings.ToLower(text), "<html") || strings.Contains(strings.ToLower(text), "<body") {
		text = stripHTML(text)
	}

	return strings.Join(strings.Fields(text), " ")
}

func stripHTML(value string) string {
	var builder strings.Builder
	inTag := false
	for _, r := range value {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
			builder.WriteRune(' ')
		default:
			if !inTag {
				builder.WriteRune(r)
			}
		}
	}

	return html.UnescapeString(builder.String())
}

var mailParseDate = func(value string) (time.Time, error) {
	return mail.ParseDate(value)
}
