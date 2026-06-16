// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/pkg/utilis/nats.go
// Role: Frameworks & drivers — NATS / JetStream connection
// Description: Connects to NATS with sane reconnect behaviour and exposes a JetStream context.
// JetStream gives durable streams + replay for the outbox / idempotent-consumer pattern
// (see 01-ARCHITECTURE §4).

package utilis

import (
	"fmt"
	"log/slog"
	"time"

	"neptune-pamm/github.com/ratheeshkumar25/pkg/config"

	"github.com/nats-io/nats.go"
)

// NewNATSConn dials NATS with bounded, infinite-retry reconnection. Caller owns Close().
func NewNATSConn(cfg *config.Config) (*nats.Conn, error) {
	opts := []nats.Option{
		nats.Name("neptune-pamm-server"),
		nats.MaxReconnects(-1), // keep retrying; the broker is a hard dependency
		nats.ReconnectWait(2 * time.Second),
		nats.Timeout(5 * time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			slog.Warn("nats disconnected", "err", err)
		}),
		nats.ReconnectHandler(func(c *nats.Conn) {
			slog.Info("nats reconnected", "url", c.ConnectedUrl())
		}),
	}

	nc, err := nats.Connect(cfg.Nats.NatsURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("connect nats: %w", err)
	}

	slog.Info("nats connected", "url", nc.ConnectedUrl())
	return nc, nil
}

// NewJetStream returns a JetStream context for durable publish/subscribe.
func NewJetStream(nc *nats.Conn) (nats.JetStreamContext, error) {
	js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		return nil, fmt.Errorf("jetstream context: %w", err)
	}
	return js, nil
}
