// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/pkg/utilis/redis.go
// Role: Frameworks & drivers — Redis client
// Description: Builds a production-tuned, verified go-redis client from config. Redis holds
// sessions, SSO/reset tokens, hot equity, share-ratio cache, idempotency guards, rate-limit
// counters and the ws connection registry (02-DATA-MODEL §E). Password is optional (empty =
// no auth) — local/dev Redis usually runs without one.

package utilis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"neptune-pamm/github.com/ratheeshkumar25/pkg/config"

	"github.com/redis/go-redis/v9"
)

// Production pool / timeout / retry defaults — tune via load testing.
const (
	redisDialTimeout     = 5 * time.Second
	redisReadTimeout     = 3 * time.Second
	redisWriteTimeout    = 3 * time.Second
	redisPoolTimeout     = 4 * time.Second
	redisPoolSize        = 20
	redisMinIdleConns    = 5
	redisMaxIdleConns    = 10
	redisConnMaxIdleTime = 5 * time.Minute
	redisConnMaxLifetime = 30 * time.Minute
	redisMaxRetries      = 3
	redisMinRetryBackoff = 8 * time.Millisecond
	redisMaxRetryBackoff = 512 * time.Millisecond
	redisPingTimeout     = 5 * time.Second
)

// NewRedisClient opens, tunes and pings a Redis client. Caller owns Close().
// Password is taken from config only if set; an empty value means no AUTH.
func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.RedisPassword, // optional; "" = no auth
		DB:       cfg.Redis.RedisDB,

		// timeouts
		DialTimeout:  redisDialTimeout,
		ReadTimeout:  redisReadTimeout,
		WriteTimeout: redisWriteTimeout,
		PoolTimeout:  redisPoolTimeout,

		// pool
		PoolSize:        redisPoolSize,
		MinIdleConns:    redisMinIdleConns,
		MaxIdleConns:    redisMaxIdleConns,
		ConnMaxIdleTime: redisConnMaxIdleTime,
		ConnMaxLifetime: redisConnMaxLifetime,

		// retries (transient errors)
		MaxRetries:      redisMaxRetries,
		MinRetryBackoff: redisMinRetryBackoff,
		MaxRetryBackoff: redisMaxRetryBackoff,
	})

	if err := PingRedis(context.Background(), client); err != nil {
		_ = client.Close()
		return nil, err
	}

	slog.Info("redis connected",
		"addr", cfg.Redis.Addr(), "db", cfg.Redis.RedisDB, "pool_size", redisPoolSize)
	return client, nil
}

// PingRedis verifies connectivity within a bounded timeout. Use it at startup and from the
// readiness probe so the service reports unhealthy when Redis is unreachable.
func PingRedis(ctx context.Context, client *redis.Client) error {
	ctx, cancel := context.WithTimeout(ctx, redisPingTimeout)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}
	return nil
}
