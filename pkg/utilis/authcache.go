// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/pkg/utilis/authcache.go
// Role: Frameworks & drivers — Redis-backed auth cache
// Description: Concrete Redis implementation of the auth session cache: single-use email
// verification tokens and the access-token (jti) revocation list used by logout. It stays
// domain-agnostic — a missing key surfaces as the sentinel ErrCacheMiss and the service layer
// maps that to a domain error.

package utilis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrCacheMiss is returned when a looked-up key is absent or expired.
var ErrCacheMiss = errors.New("cache miss")

// Redis key namespaces. Keep prefixes stable — other services read the same conventions.
const (
	verifyTokenPrefix = "auth:verify:"  // + token  → principal id
	revokedJTIPrefix  = "auth:revoked:" // + jti    → "1" until token expiry
)

// AuthCache implements the auth session cache over Redis.
type AuthCache struct {
	rdb *redis.Client
}

// NewAuthCache wraps a Redis client as an AuthCache.
func NewAuthCache(rdb *redis.Client) *AuthCache { return &AuthCache{rdb: rdb} }

// PutVerificationToken stores token → principalID with a TTL.
func (c *AuthCache) PutVerificationToken(ctx context.Context, token string, principalID int64, ttl time.Duration) error {
	if err := c.rdb.Set(ctx, verifyTokenPrefix+token, principalID, ttl).Err(); err != nil {
		return fmt.Errorf("put verification token: %w", err)
	}
	return nil
}

// ConsumeVerificationToken atomically reads and deletes the token (GETDEL), so it can be redeemed
// only once. Returns ErrCacheMiss when the token is absent or already expired/used.
func (c *AuthCache) ConsumeVerificationToken(ctx context.Context, token string) (int64, error) {
	principalID, err := c.rdb.GetDel(ctx, verifyTokenPrefix+token).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, ErrCacheMiss
		}
		return 0, fmt.Errorf("consume verification token: %w", err)
	}
	return principalID, nil
}

// RevokeAccessToken denies a token's jti until ttl elapses (matching the token's own expiry, so
// the entry self-cleans). Auth middleware must reject any access token whose jti is denied.
func (c *AuthCache) RevokeAccessToken(ctx context.Context, jti string, ttl time.Duration) error {
	if err := c.rdb.Set(ctx, revokedJTIPrefix+jti, 1, ttl).Err(); err != nil {
		return fmt.Errorf("revoke access token: %w", err)
	}
	return nil
}

// IsAccessTokenRevoked reports whether the jti has been denied.
func (c *AuthCache) IsAccessTokenRevoked(ctx context.Context, jti string) (bool, error) {
	n, err := c.rdb.Exists(ctx, revokedJTIPrefix+jti).Result()
	if err != nil {
		return false, fmt.Errorf("check revoked token: %w", err)
	}
	return n > 0, nil
}
