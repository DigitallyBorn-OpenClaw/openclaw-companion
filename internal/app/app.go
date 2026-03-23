package app

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"path/filepath"

	"github.com/ricky/oc-companion/internal/api"
	"github.com/ricky/oc-companion/internal/config"
	"github.com/ricky/oc-companion/internal/events"
	"github.com/ricky/oc-companion/internal/tools"
)

type App struct {
	cfg config.Config
}

func New(cfg config.Config) *App {
	return &App{cfg: cfg}
}

func (a *App) Run(ctx context.Context) error {
	logger := slog.Default()
	registry := api.NewRegistry()
	services, err := newServices(ctx, a.cfg)
	if err != nil {
		return err
	}

	if err := tools.Register(registry, services); err != nil {
		return err
	}

	eventWorker, err := newEventWorker(ctx, a.cfg, logger)
	if err != nil {
		return err
	}

	if err := ensureSocketDir(a.cfg.SocketPath); err != nil {
		return err
	}

	if err := removeStaleSocket(a.cfg.SocketPath); err != nil {
		return err
	}

	listener, err := net.Listen("unix", a.cfg.SocketPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = listener.Close()
		_ = os.Remove(a.cfg.SocketPath)
	}()

	if err := os.Chmod(a.cfg.SocketPath, 0o660); err != nil {
		return err
	}

	slog.Info("oc-companion started", "config", a.cfg.Summary())
	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	errCh := make(chan error, 2)
	go func() {
		errCh <- eventWorker.Run(ctx)
	}()
	go func() {
		errCh <- serveSocket(ctx, listener, logger, registry)
	}()

	for {
		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func newServices(ctx context.Context, cfg config.Config) (tools.Services, error) {
	gmailService, err := tools.NewGmailService(ctx, tools.GmailServiceConfig{
		CredentialsFile:  cfg.GCPCredentialsFile,
		UserID:           cfg.GmailUserID,
		DelegatedSubject: cfg.GmailDelegatedSubject,
	})
	if err != nil {
		return tools.Services{}, err
	}

	services := tools.NewUnavailableServices()
	services.Gmail = gmailService
	return services, nil
}

func newEventWorker(ctx context.Context, cfg config.Config, logger *slog.Logger) (events.Worker, error) {
	receiver, err := events.NewGCPPubSubReceiver(ctx, cfg, logger)
	if err != nil {
		return nil, err
	}

	return events.NewService(logger, receiver, events.NewWebhookNotifier(cfg.GmailWebhookURL, cfg.GmailWebhookToken)), nil
}

func serveSocket(ctx context.Context, listener net.Listener, logger *slog.Logger, registry *api.Registry) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			if isTemporaryError(err) {
				logger.Warn("temporary accept error", "error", err)
				continue
			}

			return err
		}

		go api.ServeConnection(ctx, logger, conn, registry)
	}
}

func isTemporaryError(err error) bool {
	var netError net.Error
	if errors.As(err, &netError) {
		return netError.Temporary()
	}

	return false
}

func ensureSocketDir(socketPath string) error {
	directory := filepath.Dir(socketPath)
	return os.MkdirAll(directory, 0o750)
}

func removeStaleSocket(socketPath string) error {
	info, err := os.Lstat(socketPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	if info.Mode()&os.ModeSocket == 0 {
		return errors.New("socket path exists and is not a socket")
	}

	return os.Remove(socketPath)
}
