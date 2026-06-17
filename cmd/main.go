// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/cmd/main.go
// Role: Entry point — server-service process
// Description: Builds the dependency container, starts the gRPC server, and blocks until an
// interrupt/terminate signal triggers a graceful shutdown. All wiring lives in di; all transport
// in server — main only sequences startup and teardown.

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"neptune-pamm/github.com/ratheeshkumar25/internal/di"
	"neptune-pamm/github.com/ratheeshkumar25/internal/server"
)

func main() {
	c, err := di.New()
	if err != nil {
		slog.Error("startup failed", "err", err)
		os.Exit(1)
	}
	defer func() {
		if cerr := c.Close(); cerr != nil {
			c.Logger.Error("shutdown cleanup failed", "err", cerr)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := server.New(c.Config.Server.ServerPort, c.Auth, c.JWT, c.Cache, c.Logger)
	if err := srv.Run(ctx); err != nil {
		c.Logger.Error("server stopped with error", "err", err)
		os.Exit(1)
	}
	c.Logger.Info("shutdown complete")
}
