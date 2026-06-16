// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/di/container.go
// Role: Composition root — dependency injection
// Description: Builds every dependency once, in order (config → logger → stores → repos), and
// hands back a Container the rest of the app reads from. This is the ONLY place that knows the
// concrete types; everything else depends on interfaces. Close() releases resources in reverse.

package di

import (
	"errors"
	"fmt"
	"log/slog"

	"neptune-pamm/github.com/ratheeshkumar25/internal/db"
	"neptune-pamm/github.com/ratheeshkumar25/internal/repo"
	"neptune-pamm/github.com/ratheeshkumar25/pkg/config"
	"neptune-pamm/github.com/ratheeshkumar25/pkg/logger"
	"neptune-pamm/github.com/ratheeshkumar25/pkg/utilis"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Container holds the wired dependencies. Build it once at startup.
type Container struct {
	Config *config.Config
	Logger *slog.Logger

	DB    *gorm.DB
	Redis *redis.Client
	NATS  *nats.Conn
	JS    nats.JetStreamContext

	Store repo.Store
}

// New constructs and verifies every dependency. On any failure it closes what was already
// opened and returns the error, so the caller never gets a half-built Container.
func New() (*Container, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	log := logger.Init(logger.Options{
		Level:  envOr(cfg, "info"),
		Format: logger.FormatJSON,
	})

	c := &Container{Config: cfg, Logger: log}

	gdb, err := db.ConnectDB(cfg)
	if err != nil {
		return nil, c.closeOnErr(fmt.Errorf("connect postgres: %w", err))
	}
	c.DB = gdb

	rdb, err := utilis.NewRedisClient(cfg)
	if err != nil {
		return nil, c.closeOnErr(fmt.Errorf("connect redis: %w", err))
	}
	c.Redis = rdb

	nc, err := utilis.NewNATSConn(cfg)
	if err != nil {
		return nil, c.closeOnErr(fmt.Errorf("connect nats: %w", err))
	}
	c.NATS = nc

	js, err := utilis.NewJetStream(nc)
	if err != nil {
		return nil, c.closeOnErr(fmt.Errorf("jetstream: %w", err))
	}
	c.JS = js

	c.Store = repo.NewStore(gdb)

	log.Info("container initialised")
	return c, nil
}

// Close releases resources in reverse order of acquisition. Safe to call on a partial build.
func (c *Container) Close() error {
	var errs []error
	if c.NATS != nil {
		c.NATS.Close() // no error returned
	}
	if c.Redis != nil {
		if err := c.Redis.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close redis: %w", err))
		}
	}
	if c.DB != nil {
		if sqlDB, err := c.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, fmt.Errorf("close postgres: %w", err))
			}
		}
	}
	return errors.Join(errs...)
}

// closeOnErr releases any already-opened resources and returns the original error.
func (c *Container) closeOnErr(orig error) error {
	if cerr := c.Close(); cerr != nil {
		return errors.Join(orig, cerr)
	}
	return orig
}

// envOr returns a log level; placeholder until a Log config group is added (currently "info").
func envOr(_ *config.Config, def string) string { return def }
