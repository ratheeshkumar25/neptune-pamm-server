// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/services/auth_service.go
// Role: Application — auth use cases (registration, verification, login, logout)
// Description: The auth service orchestrates domain rules across the repositories, the password
// hasher, the JWT manager, the session cache and the event bus. It depends ONLY on ports
// (interfaces) so the composition root can wire concrete drivers and tests can substitute fakes.
// Registration creates the account + role satellite + principal inside a single transaction via
// Store.Atomic — either all three rows land or none do.

package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"neptune-pamm/github.com/ratheeshkumar25/internal/model"
	"neptune-pamm/github.com/ratheeshkumar25/internal/repo"
	"neptune-pamm/github.com/ratheeshkumar25/pkg/utilis"

	"github.com/shopspring/decimal"
)

// verificationTTL is how long an email-verification token stays valid.
const verificationTTL = 24 * time.Hour

// minPasswordLength is the floor enforced at registration (OWASP minimum).
const minPasswordLength = 8

// SessionCache stores short-lived auth artifacts: email-verification tokens (token → principal)
// and the access-token revocation list (jti → revoked). Backed by Redis in production.
type SessionCache interface {
	PutVerificationToken(ctx context.Context, token string, principalID int64, ttl time.Duration) error
	// ConsumeVerificationToken atomically reads and deletes the token, returning utilis.ErrCacheMiss
	// when it is absent or expired.
	ConsumeVerificationToken(ctx context.Context, token string) (int64, error)
	RevokeAccessToken(ctx context.Context, jti string, ttl time.Duration) error
	IsAccessTokenRevoked(ctx context.Context, jti string) (bool, error)
}

// EventPublisher publishes notification/domain events. Defined as a port so the service depends
// on the contract, not on the NATS driver.
type EventPublisher interface {
	Publish(subject utilis.Subject, payload any) error
}

// RegisterResult is the outcome of a successful registration.
type RegisterResult struct {
	AccountID                 int64
	Username                  string
	Role                      model.Role
	EmailVerificationRequired bool
}

// LoginInput is the transport-agnostic login request. TenantID is resolved by the handler from
// the authenticated context/host — never from client-controlled body fields.
type LoginInput struct {
	TenantID int64
	Username string
	Password string
}

// TokenResult is a freshly issued session.
type TokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // access-token TTL, seconds
	AccountID    int64
	Role         model.Role
}

// AuthService is the auth application API consumed by the gRPC handler.
type AuthService interface {
	Register(ctx context.Context, in model.RegisterInput) (*RegisterResult, error)
	VerifyEmail(ctx context.Context, token string) error
	Login(ctx context.Context, in LoginInput) (*TokenResult, error)
	Logout(ctx context.Context, accessToken string) error
	ListCurrencies(ctx context.Context, tenantID int64) ([]model.Currency, error)
}

// authService is the concrete implementation. All collaborators are injected ports.
type authService struct {
	store repo.Store
	cache SessionCache
	bus   EventPublisher
	jwt   *utilis.JWTManager
	log   *slog.Logger
}

// NewAuthService wires the auth service from its ports.
func NewAuthService(store repo.Store, cache SessionCache, bus EventPublisher, jwt *utilis.JWTManager, log *slog.Logger) AuthService {
	return &authService{store: store, cache: cache, bus: bus, jwt: jwt, log: log}
}

// emailVerificationEvent is the notification payload consumed by the email worker.
type emailVerificationEvent struct {
	PrincipalID int64  `json:"principal_id"`
	AccountID   int64  `json:"account_id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	Token       string `json:"token"`
}

// Register creates a new self-service account. The account row, its role satellite and the login
// principal are written in one transaction; on any failure the whole registration rolls back. The
// password is hashed before the transaction (CPU-bound, no DB needed) and never persisted in clear.
// Email delivery happens after commit, best-effort — a failed notification must not fail the signup.
func (s *authService) Register(ctx context.Context, in model.RegisterInput) (*RegisterResult, error) {
	if err := validateRegister(in); err != nil {
		return nil, err
	}

	passwordHash, err := utilis.HashPassword(in.Password)
	if err != nil {
		return nil, model.NewInternal(err)
	}

	verifyToken, err := utilis.RandomToken(32)
	if err != nil {
		return nil, model.NewInternal(err)
	}

	email := strings.TrimSpace(in.Email)
	now := time.Now().UTC()

	var accountID, principalID int64
	err = s.store.Atomic(ctx, func(st repo.Store) error {
		exists, err := st.Principal().ExistsByUsername(ctx, in.TenantID, in.Username)
		if err != nil {
			return err
		}
		if exists {
			return model.NewConflict("username %q is already taken", in.Username)
		}

		account := &model.Account{
			TenantID: in.TenantID,
			Kind:     accountKindForRole(in.Role),
			Name:     &in.Username,
			Email:    &email,
			Country:  ptrIfSet(in.Country),
			Phone:    ptrIfSet(in.Phone),
			Currency: ptrIfSet(in.Currency),
			Status:   model.AccountActive,
		}
		if err := st.Account().Create(ctx, account); err != nil {
			return err
		}

		if err := createRoleSatellite(ctx, st, in, account.ID); err != nil {
			return err
		}

		principal := &model.Principal{
			TenantID:        in.TenantID,
			AccountID:       account.ID,
			Username:        in.Username,
			Email:           &email,
			PasswordHash:    &passwordHash,
			Role:            in.Role,
			Status:          model.PrincipalActive,
			EmailVerified:   false, // gate login until the token is consumed
			TermsAccepted:   in.TermsAccepted,
			TermsAcceptedAt: &now,
		}
		if err := st.Principal().Create(ctx, principal); err != nil {
			return err
		}

		accountID, principalID = account.ID, principal.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	// --- post-commit side effects (best-effort; never roll the signup back) ----------
	if err := s.cache.PutVerificationToken(ctx, verifyToken, principalID, verificationTTL); err != nil {
		s.log.Error("store verification token", "err", err, "principal_id", principalID)
	} else {
		s.publishNotification("email_verification", emailVerificationEvent{
			PrincipalID: principalID,
			AccountID:   accountID,
			Email:       email,
			Username:    in.Username,
			Token:       verifyToken,
		})
	}

	return &RegisterResult{
		AccountID:                 accountID,
		Username:                  in.Username,
		Role:                      in.Role,
		EmailVerificationRequired: true,
	}, nil
}

// VerifyEmail consumes a verification token and flips the principal's email_verified flag,
// unlocking login. The token is single-use: consuming it removes it from the cache.
func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return model.NewValidation("verification token is required")
	}

	principalID, err := s.cache.ConsumeVerificationToken(ctx, token)
	if err != nil {
		if errors.Is(err, utilis.ErrCacheMiss) {
			return model.NewValidation("verification link is invalid or has expired")
		}
		return model.NewInternal(err)
	}

	return s.store.Principal().SetEmailVerified(ctx, principalID)
}

// Login authenticates a native PAMM principal and issues an access/refresh pair. Failure reasons
// are deliberately collapsed into one "invalid credentials" message so the response cannot be used
// to enumerate usernames.
func (s *authService) Login(ctx context.Context, in LoginInput) (*TokenResult, error) {
	if in.Username == "" || in.Password == "" {
		return nil, model.NewValidation("username and password are required")
	}

	principal, err := s.store.Principal().GetByUsername(ctx, in.TenantID, in.Username)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, model.NewUnauthorized("invalid credentials")
		}
		return nil, err
	}

	// A NULL hash means this is an MT-only login that has no native password.
	if principal.PasswordHash == nil {
		return nil, model.NewUnauthorized("invalid credentials")
	}
	ok, err := utilis.VerifyPassword(in.Password, *principal.PasswordHash)
	if err != nil {
		return nil, model.NewInternal(err)
	}
	if !ok {
		return nil, model.NewUnauthorized("invalid credentials")
	}

	if principal.Status != model.PrincipalActive {
		return nil, model.NewForbidden("account is %s", principal.Status)
	}
	if !principal.EmailVerified {
		return nil, model.NewForbidden("email address is not verified")
	}

	now := time.Now().UTC()
	pair, err := s.jwt.Generate(principal.AccountID, principal.TenantID, string(principal.Role), now)
	if err != nil {
		return nil, model.NewInternal(err)
	}

	// Last-login is an audit convenience; a write failure must not fail an otherwise valid login.
	if err := s.store.Principal().UpdateLastLogin(ctx, principal.ID, now); err != nil {
		s.log.Error("update last login", "err", err, "principal_id", principal.ID)
	}

	return &TokenResult{
		AccessToken:  pair.Access,
		RefreshToken: pair.Refresh,
		ExpiresIn:    pair.AccessTTLms,
		AccountID:    principal.AccountID,
		Role:         principal.Role,
	}, nil
}

// Logout revokes the access token by denying its jti until it would have expired anyway. An
// already-expired token is a no-op success.
func (s *authService) Logout(ctx context.Context, accessToken string) error {
	if accessToken == "" {
		return model.NewValidation("access token is required")
	}

	claims, err := s.jwt.Parse(accessToken)
	if err != nil {
		return model.NewUnauthorized("invalid token")
	}

	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil
	}
	if err := s.cache.RevokeAccessToken(ctx, claims.ID, ttl); err != nil {
		return model.NewInternal(err)
	}
	return nil
}

// ListCurrencies returns the tenant's configured currencies (public endpoint).
func (s *authService) ListCurrencies(ctx context.Context, tenantID int64) ([]model.Currency, error) {
	return s.store.Currency().ListByTenant(ctx, tenantID)
}

// publishNotification fans a notification event out on pamm.notify.<event>. Best-effort: a broker
// hiccup is logged, not surfaced to the caller.
func (s *authService) publishNotification(event string, payload any) {
	subject := utilis.SubjNotifyPrefix + utilis.Subject(event)
	if err := s.bus.Publish(subject, payload); err != nil {
		s.log.Error("publish notification", "err", err, "subject", subject.String())
	}
}

// --- helpers ----------------------------------------------------------------------

// validateRegister enforces the self-service signup rules before any DB work.
func validateRegister(in model.RegisterInput) error {
	switch in.Role {
	case model.RoleInvestor, model.RoleMoneyManager:
		// self-service roles
	case model.RoleAdmin:
		return model.NewForbidden("admin accounts cannot be self-registered")
	default:
		return model.NewValidation("role must be Investor or MoneyManager")
	}
	if len(strings.TrimSpace(in.Username)) < 3 {
		return model.NewValidation("username must be at least 3 characters")
	}
	if !looksLikeEmail(in.Email) {
		return model.NewValidation("a valid email address is required")
	}
	if len(in.Password) < minPasswordLength {
		return model.NewValidation("password must be at least %d characters", minPasswordLength)
	}
	if !in.TermsAccepted {
		return model.NewValidation("terms and conditions must be accepted")
	}
	return nil
}

// accountKindForRole maps a login role to its polymorphic account kind. Callers validate the role
// first, so an unexpected value falls back to Investor rather than panicking.
func accountKindForRole(role model.Role) model.AccountKind {
	if role == model.RoleMoneyManager {
		return model.AccountMaster
	}
	return model.AccountInvestor
}

// createRoleSatellite inserts the role-specific profile row keyed by the freshly created account.
func createRoleSatellite(ctx context.Context, st repo.Store, in model.RegisterInput, accountID int64) error {
	switch in.Role {
	case model.RoleMoneyManager:
		return st.Account().CreateMaster(ctx, &model.Master{
			AccountID:      accountID,
			TenantID:       in.TenantID,
			AllocationMode: model.AllocationModeByBalance,
			MinInvestment:  decimal.Zero,
		})
	default: // RoleInvestor
		return st.Account().CreateInvestor(ctx, &model.Investor{
			AccountID:     accountID,
			TenantID:      in.TenantID,
			HighWaterMark: decimal.Zero,
			ReferralCode:  ptrIfSet(in.ReferralCode),
		})
	}
}

// looksLikeEmail is a cheap structural check (full RFC validation is delegated to the verification
// flow — a deliverable address proves more than any regex).
func looksLikeEmail(s string) bool {
	s = strings.TrimSpace(s)
	at := strings.IndexByte(s, '@')
	if at <= 0 || at == len(s)-1 {
		return false
	}
	return strings.IndexByte(s[at+1:], '.') >= 0
}

// ptrIfSet returns a pointer to s, or nil when s is empty after trimming — so optional columns
// stay NULL instead of storing blank strings.
func ptrIfSet(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}
