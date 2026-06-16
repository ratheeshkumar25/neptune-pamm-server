// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/pkg/utilis/token.go
// Role: Security — cryptographically-secure random tokens
// Description: URL-safe opaque tokens for email-verification, password-reset and SSO one-time
// flows. Uses crypto/rand (never math/rand).

package utilis

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// RandomToken returns a URL-safe random token with nBytes of entropy (32 = 256 bits).
func RandomToken(nBytes int) (string, error) {
	if nBytes <= 0 {
		nBytes = 32
	}
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("random token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
