package app

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"path/filepath"

	"github.com/ricky/oc-companion/internal/config"
)

type App struct {
	cfg config.Config
}

func New(cfg config.Config) *App {
	return &App{cfg: cfg}
}

func (a *App) Run(ctx context.Context) error {
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

	<-ctx.Done()

	return ctx.Err()
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
