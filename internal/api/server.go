package api

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"

	"github.com/ricky/oc-companion/internal/protocol"
)

func ServeConnection(ctx context.Context, logger *slog.Logger, conn net.Conn, registry *Registry) {
	defer func() {
		_ = conn.Close()
	}()

	decoder := json.NewDecoder(bufio.NewReader(conn))
	encoder := json.NewEncoder(conn)

	for {
		if ctx.Err() != nil {
			return
		}

		var request protocol.Request
		if err := decoder.Decode(&request); err != nil {
			if errors.Is(err, io.EOF) {
				return
			}

			logger.Warn("request decode error", "error", err)
			if writeErr := encoder.Encode(protocol.Failure(nil, protocol.CodeParseError, "invalid json request", nil)); writeErr != nil {
				logger.Warn("failed writing parse error response", "error", writeErr)
			}

			return
		}

		response := registry.Dispatch(ctx, request)
		if err := encoder.Encode(response); err != nil {
			logger.Warn("response write error", "error", err)
			return
		}
	}
}
