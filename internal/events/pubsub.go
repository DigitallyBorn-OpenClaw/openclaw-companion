package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/ricky/oc-companion/internal/config"
	"google.golang.org/api/option"
)

const subscriptionDeleteTimeout = 10 * time.Second

type GCPPubSubReceiver struct {
	logger          *slog.Logger
	client          *pubsub.Client
	topicID         string
	subscriptionID  string
	shutdownTimeout time.Duration
}

type gmailPushMessage struct {
	EmailAddress string `json:"emailAddress"`
	HistoryID    string `json:"historyId"`
}

func NewGCPPubSubReceiver(ctx context.Context, cfg config.Config, logger *slog.Logger) (*GCPPubSubReceiver, error) {
	options := make([]option.ClientOption, 0, 1)
	if cfg.GCPCredentialsFile != "" {
		options = append(options, option.WithCredentialsFile(cfg.GCPCredentialsFile))
	}

	client, err := pubsub.NewClient(ctx, cfg.GCPProjectID, options...)
	if err != nil {
		return nil, fmt.Errorf("create pubsub client: %w", err)
	}

	return &GCPPubSubReceiver{
		logger:          logger,
		client:          client,
		topicID:         cfg.GmailPubSubTopicID,
		subscriptionID:  buildSubscriptionID(cfg.PubSubSubscriptionPrefix),
		shutdownTimeout: cfg.ShutdownTimeout,
	}, nil
}

func (r *GCPPubSubReceiver) Run(ctx context.Context, handler Handler) error {
	defer func() {
		_ = r.client.Close()
	}()

	topic := r.client.Topic(r.topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return fmt.Errorf("check pubsub topic %q: %w", r.topicID, err)
	}
	if !exists {
		return fmt.Errorf("pubsub topic %q does not exist", r.topicID)
	}

	subscription, err := r.client.CreateSubscription(ctx, r.subscriptionID, pubsub.SubscriptionConfig{
		Topic:       topic,
		AckDeadline: 20 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("create pubsub subscription %q: %w", r.subscriptionID, err)
	}

	r.logger.Info("created gmail pubsub subscription", "topic", r.topicID, "subscription", r.subscriptionID)
	defer func() {
		if cleanupErr := r.deleteSubscription(); cleanupErr != nil {
			r.logger.Error("failed deleting gmail pubsub subscription", "subscription", r.subscriptionID, "error", cleanupErr)
		}
	}()

	subscription.ReceiveSettings.NumGoroutines = 1
	subscription.ReceiveSettings.MaxOutstandingMessages = 1

	err = subscription.Receive(ctx, func(msgCtx context.Context, msg *pubsub.Message) {
		notification, parseErr := parseGmailNotification(r.subscriptionID, msg)
		if parseErr != nil {
			r.logger.Error("invalid gmail pubsub message", "error", parseErr, "pubsub_id", msg.ID)
			msg.Nack()
			return
		}

		if err := handler(msgCtx, notification); err != nil {
			msg.Nack()
			return
		}

		msg.Ack()
	})
	if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
		return ctx.Err()
	}
	if err != nil {
		return fmt.Errorf("receive pubsub messages: %w", err)
	}

	return nil
}

func (r *GCPPubSubReceiver) deleteSubscription() error {
	ctx, cancel := context.WithTimeout(context.Background(), minDuration(r.shutdownTimeout, subscriptionDeleteTimeout))
	defer cancel()

	err := r.client.Subscription(r.subscriptionID).Delete(ctx)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
		return err
	}

	r.logger.Info("deleted gmail pubsub subscription", "subscription", r.subscriptionID)
	return nil
}

func parseGmailNotification(subscriptionID string, msg *pubsub.Message) (GmailNotification, error) {
	var payload gmailPushMessage
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		return GmailNotification{}, fmt.Errorf("decode gmail push payload: %w", err)
	}
	if strings.TrimSpace(payload.EmailAddress) == "" {
		return GmailNotification{}, errors.New("emailAddress is required")
	}
	if strings.TrimSpace(payload.HistoryID) == "" {
		return GmailNotification{}, errors.New("historyId is required")
	}

	receivedAt := msg.PublishTime.UTC()
	if receivedAt.IsZero() {
		receivedAt = time.Now().UTC()
	}

	attributes := make(map[string]string, len(msg.Attributes))
	for key, value := range msg.Attributes {
		attributes[key] = value
	}

	return GmailNotification{
		Account:      payload.EmailAddress,
		HistoryID:    payload.HistoryID,
		ReceivedAt:   receivedAt,
		PubSubID:     msg.ID,
		Subscription: subscriptionID,
		Attributes:   attributes,
	}, nil
}

func buildSubscriptionID(prefix string) string {
	hostname, err := os.Hostname()
	if err != nil || strings.TrimSpace(hostname) == "" {
		hostname = "host"
	}

	hostname = sanitizeSubscriptionComponent(hostname)
	return fmt.Sprintf("%s-%s-%d", prefix, hostname, time.Now().UTC().Unix())
}

func sanitizeSubscriptionComponent(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-':
			return r
		default:
			return '-'
		}
	}, value)
	value = strings.Trim(value, "-")
	if value == "" {
		return "host"
	}

	return value
}

func minDuration(a time.Duration, b time.Duration) time.Duration {
	if a <= 0 || a > b {
		return b
	}

	return a
}
