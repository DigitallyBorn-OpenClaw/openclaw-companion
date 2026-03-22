package events

import (
	"context"
	"log/slog"
)

type Receiver interface {
	Run(context.Context, Handler) error
}

type WebhookDeliverer interface {
	Notify(context.Context, GmailNotification) error
}

type Service struct {
	logger    *slog.Logger
	receiver  Receiver
	deliverer WebhookDeliverer
}

func NewService(logger *slog.Logger, receiver Receiver, deliverer WebhookDeliverer) *Service {
	return &Service{
		logger:    logger,
		receiver:  receiver,
		deliverer: deliverer,
	}
}

func (s *Service) Run(ctx context.Context) error {
	return s.receiver.Run(ctx, func(msgCtx context.Context, notification GmailNotification) error {
		s.logger.Info(
			"received gmail pubsub notification",
			"account", notification.Account,
			"history_id", notification.HistoryID,
			"subscription", notification.Subscription,
			"pubsub_id", notification.PubSubID,
		)

		if err := s.deliverer.Notify(msgCtx, notification); err != nil {
			s.logger.Error("failed delivering gmail notification", "error", err, "pubsub_id", notification.PubSubID)
			return err
		}

		s.logger.Info("delivered gmail notification", "history_id", notification.HistoryID, "pubsub_id", notification.PubSubID)
		return nil
	})
}
