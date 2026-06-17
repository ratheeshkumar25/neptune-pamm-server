// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/handler/interceptor.go
// Role: Interface adapters — auth gRPC interceptor
// Description: A unary server interceptor that enforces bearer-token auth on protected RPCs.
// Public RPCs (register/login/verify/reset/sso-login/currency, plus self-authenticating logout)
// are allow-listed and skipped. For everything else it parses the access token, rejects it if its
// jti is on the Redis revocation list (so logout takes effect immediately), and injects the
// validated claims into the context for downstream handlers.

package handler

import (
	"context"
	"strings"

	authv1 "neptune-pamm/github.com/ratheeshkumar25/internal/proto/auth/v1"
	"neptune-pamm/github.com/ratheeshkumar25/internal/services"
	"neptune-pamm/github.com/ratheeshkumar25/pkg/utilis"

	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	mdAuthorization = "authorization"
	bearerPrefix    = "Bearer "
)

// ctxKey is an unexported context-key type so values can't collide with other packages'.
type ctxKey string

const claimsCtxKey ctxKey = "auth.claims"

// AuthInterceptor authenticates protected unary RPCs.
type AuthInterceptor struct {
	jwt    *utilis.JWTManager
	cache  services.SessionCache
	log    *slog.Logger
	public map[string]bool
}

// NewAuthInterceptor builds the interceptor with the default public-method allow-list.
func NewAuthInterceptor(jwt *utilis.JWTManager, cache services.SessionCache, log *slog.Logger) *AuthInterceptor {
	return &AuthInterceptor{jwt: jwt, cache: cache, log: log, public: defaultPublicMethods()}
}

// defaultPublicMethods lists the RPCs that must be reachable without a session. Logout is public
// because it authenticates itself via the access token carried in its request body.
func defaultPublicMethods() map[string]bool {
	return map[string]bool{
		authv1.AuthService_Register_FullMethodName:       true,
		authv1.AuthService_Login_FullMethodName:          true,
		authv1.AuthService_Logout_FullMethodName:         true,
		authv1.AuthService_VerifyEmail_FullMethodName:    true,
		authv1.AuthService_ResetPassword_FullMethodName:  true,
		authv1.AuthService_LoginByToken_FullMethodName:   true,
		authv1.AuthService_ListCurrencies_FullMethodName: true,
	}
}

// Unary returns the grpc.UnaryServerInterceptor.
func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if i.public[info.FullMethod] {
			return handler(ctx, req)
		}
		claims, err := i.authenticate(ctx)
		if err != nil {
			return nil, err
		}
		return handler(WithClaims(ctx, claims), req)
	}
}

// authenticate validates the bearer token from the authorization metadata and checks revocation.
func (i *AuthInterceptor) authenticate(ctx context.Context) (*utilis.Claims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing request metadata")
	}
	vals := md.Get(mdAuthorization)
	if len(vals) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization header")
	}
	token, ok := strings.CutPrefix(vals[0], bearerPrefix)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authorization header must be a Bearer token")
	}

	claims, err := i.jwt.Parse(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
	}
	if claims.Kind != utilis.TokenAccess {
		return nil, status.Error(codes.Unauthenticated, "an access token is required")
	}

	revoked, err := i.cache.IsAccessTokenRevoked(ctx, claims.ID)
	if err != nil {
		i.log.Error("check token revocation", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}
	if revoked {
		return nil, status.Error(codes.Unauthenticated, "token has been revoked")
	}
	return claims, nil
}

// WithClaims returns a child context carrying the authenticated claims.
func WithClaims(ctx context.Context, claims *utilis.Claims) context.Context {
	return context.WithValue(ctx, claimsCtxKey, claims)
}

// ClaimsFromContext extracts the authenticated claims injected by the interceptor.
func ClaimsFromContext(ctx context.Context) (*utilis.Claims, bool) {
	claims, ok := ctx.Value(claimsCtxKey).(*utilis.Claims)
	return claims, ok
}
