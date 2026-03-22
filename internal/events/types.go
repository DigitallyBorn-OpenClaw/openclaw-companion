package events

import (
	"context"
	"time"
)

type GmailNotification struct {
	Account      string            `json:"account"`
	HistoryID    string            `json:"history_id"`
	ReceivedAt   time.Time         `json:"received_at"`
	PubSubID     string            `json:"pubsub_id,omitempty"`
	Subscription string            `json:"subscription,omitempty"`
	Attributes   map[string]string `json:"attributes,omitempty"`
}

type Handler func(context.Context, GmailNotification) error

type Worker interface {
	Run(context.Context) error
}
