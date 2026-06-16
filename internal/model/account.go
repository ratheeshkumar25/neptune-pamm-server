// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/model/account.go
// Role: Domain models — core schema accounts (account, master, investor, admin)
// Description: The polymorphic account row and the three role satellites.

package model

import (
	"encoding/json"
	"time"
)

// AccountKind discriminates the polymorphic account row.
type AccountKind string

const (
	AccountAdmin    AccountKind = "Admin"
	AccountMaster   AccountKind = "Master"
	AccountInvestor AccountKind = "Investor"
)

// AccountStatus enumerates account states.
type AccountStatus string

const (
	AccountActive   AccountStatus = "Active"
	AccountDisabled AccountStatus = "Disabled"
	AccountArchived AccountStatus = "Archived"
)

// RiskMode enumerates risk-level modes (*_mode columns). FromMm is Investor so_mode only.
type RiskMode string

const (
	RiskOff     RiskMode = "Off"
	RiskPercent RiskMode = "Percent"
	RiskMoney   RiskMode = "Money"
	RiskFromMm  RiskMode = "FromMm"
)

// AllocationLogic is the per-master correction logic (03-ALLOCATION-AND-FEES §1).
type AllocationLogic string

const (
	AllocationReallocation   AllocationLogic = "Reallocation"
	AllocationAutoCorrection AllocationLogic = "AutoCorrection"
)

// Allocation modes (core.master.allocation_mode).
const (
	AllocationModeByBalance int16 = 0
	AllocationModeByEquity  int16 = 1
)

// Account is the base row shared by all three roles. account.id == auth.principal.account_id.
type Account struct {
	ID             int64           `db:"id"              json:"id"`
	TenantID       int64           `db:"tenant_id"       json:"tenant_id"`
	Kind           AccountKind     `db:"kind"            json:"kind"`
	Name           *string         `db:"name"            json:"name,omitempty"`
	Email          *string         `db:"email"           json:"email,omitempty"`
	Country        *string         `db:"country"         json:"country,omitempty"`
	Phone          *string         `db:"phone"           json:"phone,omitempty"`
	Currency       *string         `db:"currency"        json:"currency,omitempty"` // ISO 4217 / BTC etc.
	Status         AccountStatus   `db:"status"          json:"status"`
	IsTest         bool            `db:"is_test"         json:"is_test"`
	AccountOptions json.RawMessage `db:"account_options" json:"account_options,omitempty"`
	CreatedAt      time.Time       `db:"created_at"      json:"created_at"`
	UpdatedAt      *time.Time      `db:"updated_at"      json:"updated_at,omitempty"`
}

// Master is the money-manager profile.
type Master struct {
	AccountID           int64            `db:"account_id"          json:"account_id"`
	TenantID            int64            `db:"tenant_id"           json:"tenant_id"`
	AllocationMode      int16            `db:"allocation_mode"     json:"allocation_mode"` // 0=ByBalance 1=ByEquity
	MinInvestment       Money            `db:"min_investment"      json:"min_investment"`
	Leverage            *int             `db:"leverage"            json:"leverage,omitempty"`
	Private             bool             `db:"private"             json:"private"`
	Invisible           bool             `db:"invisible"           json:"invisible"`
	FreeMarginCoef      *Money           `db:"free_margin_coef"    json:"free_margin_coef,omitempty"`
	So                  *Money           `db:"so"                  json:"so,omitempty"` // Investments-Loss / Stop-Out
	SoMode              *RiskMode        `db:"so_mode"             json:"so_mode,omitempty"`
	EquityCallLevel     *Money           `db:"equity_call_level"   json:"equity_call_level,omitempty"`
	EquityCallMode      *RiskMode        `db:"equity_call_level_mode" json:"equity_call_level_mode,omitempty"`
	Info                *string          `db:"info"                json:"info,omitempty"`
	Bio                 *string          `db:"bio"                 json:"bio,omitempty"`
	TermsConditions     *string          `db:"terms_conditions"    json:"terms_conditions,omitempty"`
	OwnFundsAccountID   *int64           `db:"own_funds_account_id" json:"own_funds_account_id,omitempty"`
	AllocationLogic     *AllocationLogic `db:"allocation_logic"    json:"allocation_logic,omitempty"`
	UseDefaultPayment   bool             `db:"use_default_payment_settings"   json:"use_default_payment_settings"`
	UseDefaultStatement bool             `db:"use_default_statement_settings" json:"use_default_statement_settings"`
	CreatedAt           time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt           *time.Time       `db:"updated_at" json:"updated_at,omitempty"`
}

// Investor is the investor profile.
type Investor struct {
	AccountID           int64      `db:"account_id" json:"account_id"`
	TenantID            int64      `db:"tenant_id"  json:"tenant_id"`
	MasterID            *int64     `db:"master_id"  json:"master_id,omitempty"` // current connection (NULL = unconnected)
	Sl                  *Money     `db:"sl"         json:"sl,omitempty"`        // Max Loss
	SlMode              *RiskMode  `db:"sl_mode"    json:"sl_mode,omitempty"`
	Tp                  *Money     `db:"tp"         json:"tp,omitempty"` // Max Profit
	TpMode              *RiskMode  `db:"tp_mode"    json:"tp_mode,omitempty"`
	So                  *Money     `db:"so"         json:"so,omitempty"` // Investments-Loss (may be FromMm)
	SoMode              *RiskMode  `db:"so_mode"    json:"so_mode,omitempty"`
	EquityCallLevel     *Money     `db:"equity_call_level"      json:"equity_call_level,omitempty"`
	EquityCallMode      *RiskMode  `db:"equity_call_level_mode" json:"equity_call_level_mode,omitempty"`
	HighWaterMark       Money      `db:"high_water_mark" json:"high_water_mark"` // performance-fee baseline
	ReferralCode        *string    `db:"referral_code"   json:"referral_code,omitempty"`
	LinkToken           *string    `db:"link_token"      json:"link_token,omitempty"`
	UseDefaultStatement bool       `db:"use_default_statement_settings" json:"use_default_statement_settings"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

// Admin is the admin profile.
type Admin struct {
	AccountID int64 `db:"account_id" json:"account_id"`
	TenantID  int64 `db:"tenant_id"  json:"tenant_id"`
	ViewOnly  bool  `db:"view_only"  json:"view_only"`
}
