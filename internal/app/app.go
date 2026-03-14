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
