package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const webhookTimeout = 10 * time.Second

type WebhookNotifier struct {
	client *http.Client
	url    string
	token  string
}

type webhookPayload struct {
	Event        string            `json:"event"`
	Account      string            `json:"account"`
	HistoryID    string            `json:"history_id"`
	ReceivedAt   time.Time         `json:"received_at"`
	PubSubID     string            `json:"pubsub_id,omitempty"`
	Subscription string            `json:"subscription,omitempty"`
	Attributes   map[string]string `json:"attributes,omitempty"`
}

func NewWebhookNotifier(url string, token string) *WebhookNotifier {
	return &WebhookNotifier{
		client: &http.Client{Timeout: webhookTimeout},
		url:    strings.TrimSpace(url),
		token:  strings.TrimSpace(token),
	}
}

func (n *WebhookNotifier) Notify(ctx context.Context, notification GmailNotification) error {
	body, err := json.Marshal(webhookPayload{
		Event:        "gmail.new_message",
		Account:      notification.Account,
		HistoryID:    notification.HistoryID,
		ReceivedAt:   notification.ReceivedAt.UTC(),
		PubSubID:     notification.PubSubID,
		Subscription: notification.Subscription,
		Attributes:   notification.Attributes,
	})
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build webhook request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+n.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-OpenClaw-Event", "gmail.new_message")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("send webhook request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return fmt.Errorf("webhook returned %s: %s", resp.Status, strings.TrimSpace(string(responseBody)))
}
