// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/model/auth.go
// Role: Domain models — auth schema (auth.principal, auth.session_audit, auth.token_audit)
// Description: Login identities, session/token audit trails, and the registration input.

package model

import "time"

// Role enumerates principal roles (auth.principal.role).
type Role string

const (
	RoleAdmin        Role = "Admin"
	RoleMoneyManager Role = "MoneyManager"
	RoleInvestor     Role = "Investor"
)

// PrincipalStatus enumerates login-account states.
type PrincipalStatus string

const (
	PrincipalActive   PrincipalStatus = "Active"
	PrincipalDisabled PrincipalStatus = "Disabled"
	PrincipalLocked   PrincipalStatus = "Locked"
)

// Principal is a login identity (auth.principal). One row per authenticatable
// Admin / MoneyManager / Investor; the business profile lives in core.* by account_id.
type Principal struct {
	ID           int64           `db:"id"            json:"id"`
	TenantID     int64           `db:"tenant_id"     json:"tenant_id"`
	AccountID    int64           `db:"account_id"    json:"account_id"`
	Username     string          `db:"username"      json:"username"`
	Email        *string         `db:"email"         json:"email,omitempty"`
	PasswordHash *string         `db:"password_hash" json:"-"` // argon2id; NULL when MT-only login
	Role         Role            `db:"role"          json:"role"`
	ViewOnly     bool            `db:"view_only"     json:"view_only"`
	Status       PrincipalStatus `db:"status"        json:"status"`
	MFASecret    *string         `db:"mfa_secret"    json:"-"`

	// --- registration fields (added; mirror into 02-DATA-MODEL.md auth.principal) ---
	EmailVerified   bool       `db:"email_verified"    json:"email_verified"` // gate login until verified
	TermsAccepted   bool       `db:"terms_accepted"    json:"terms_accepted"` // AML/KYC compliance
	TermsAcceptedAt *time.Time `db:"terms_accepted_at" json:"terms_accepted_at,omitempty"`
	LastLoginAt     *time.Time `db:"last_login_at"     json:"last_login_at,omitempty"`

	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

// SessionMode mirrors the login mode (Pamm native vs MT credentials).
type SessionMode string

const (
	SessionModePamm SessionMode = "Pamm"
	SessionModeMT   SessionMode = "MT"
)

// SessionAudit is the durable trail of issued sessions (live sessions live in Redis).
type SessionAudit struct {
	ID          int64       `db:"id"           json:"id"`
	TenantID    int64       `db:"tenant_id"    json:"tenant_id"`
	PrincipalID int64       `db:"principal_id" json:"principal_id"`
	Mode        SessionMode `db:"mode"         json:"mode"`
	MTServerID  *int64      `db:"mt_server_id" json:"mt_server_id,omitempty"`
	IssuedAt    time.Time   `db:"issued_at"    json:"issued_at"`
	RevokedAt   *time.Time  `db:"revoked_at"   json:"revoked_at,omitempty"`
	IP          *string     `db:"ip"           json:"ip,omitempty"`
	UserAgent   *string     `db:"user_agent"   json:"user_agent,omitempty"`
}

// TokenKind enumerates short-lived token types recorded for audit.
type TokenKind string

const (
	TokenPasswordReset     TokenKind = "PasswordReset"
	TokenSsoOneTime        TokenKind = "SsoOneTime"
	TokenEmailVerification TokenKind = "EmailVerification" // added for registration
)

// TokenAudit records issuance of password-reset / SSO / email-verification tokens.
type TokenAudit struct {
	ID          int64      `db:"id"           json:"id"`
	TenantID    int64      `db:"tenant_id"    json:"tenant_id"`
	PrincipalID *int64     `db:"principal_id" json:"principal_id,omitempty"`
	Kind        TokenKind  `db:"kind"         json:"kind"`
	IssuedAt    time.Time  `db:"issued_at"    json:"issued_at"`
	ConsumedAt  *time.Time `db:"consumed_at"  json:"consumed_at,omitempty"`
}

// RegisterInput is the payload for the registration use case. Plaintext password is
// hashed before persistence and never stored. Lives here as the use-case input model.
type RegisterInput struct {
	TenantID      int64  `json:"tenant_id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	Phone         string `json:"phone,omitempty"`
	Country       string `json:"country,omitempty"`
	Password      string `json:"password"` // plaintext in; never persisted
	Role          Role   `json:"role"`
	Currency      string `json:"currency,omitempty"`
	ReferralCode  string `json:"referral_code,omitempty"`
	TermsAccepted bool   `json:"terms_accepted"`
}
