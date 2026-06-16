// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/model/fees.go
// Role: Domain models — fees schema (fee_account, fee_schedule, payment_period, fee_charge)
// Description: Fee accounts/rates, tenant defaults, payment periods, and computed charges.
// The fee engine COMPUTES; the user-service ledger POSTS.

package model

import "time"

// FeeOwnerKind enumerates who owns a fee account.
type FeeOwnerKind string

const (
	FeeOwnerMaster   FeeOwnerKind = "Master"
	FeeOwnerInvestor FeeOwnerKind = "Investor"
)

// FeeType enumerates chargeable fee types.
type FeeType string

const (
	FeePerformance      FeeType = "Performance"
	FeeManagement       FeeType = "Management"
	FeeAnnualManagement FeeType = "AnnualManagement"
	FeePerLot           FeeType = "PerLot"
	FeePerDeal          FeeType = "PerDeal"
	FeeConnection       FeeType = "Connection"
)

// FeeChargeStatus enumerates the lifecycle of a computed fee.
type FeeChargeStatus string

const (
	FeeComputed FeeChargeStatus = "Computed"
	FeePosting  FeeChargeStatus = "Posting"
	FeePosted   FeeChargeStatus = "Posted"
	FeeFailed   FeeChargeStatus = "Failed"
)

// FeeAccount is the MT account that receives charged fees (≤5 per MM).
type FeeAccount struct {
	ID                  int64        `db:"id"                    json:"id"`
	TenantID            int64        `db:"tenant_id"             json:"tenant_id"`
	OwnerAccountID      int64        `db:"owner_account_id"      json:"owner_account_id"`
	OwnerKind           FeeOwnerKind `db:"owner_kind"            json:"owner_kind"`
	PlatformServerID    *int64       `db:"platform_server_id"    json:"platform_server_id,omitempty"`
	MTLogin             int64        `db:"mt_login"              json:"mt_login"`
	MTGroup             *string      `db:"mt_group"              json:"mt_group,omitempty"`
	IncentiveFee        *Money       `db:"incentive_fee"          json:"incentive_fee,omitempty"` // performance %
	ManagementFee       *Money       `db:"management_fee"         json:"management_fee,omitempty"`
	AnnualManagementFee *Money       `db:"annual_management_fee"  json:"annual_management_fee,omitempty"`
	PerLotFee           *Money       `db:"per_lot_fee"            json:"per_lot_fee,omitempty"`
	PerDealFee          *Money       `db:"per_deal_fee"           json:"per_deal_fee,omitempty"`
	ConnectionFee       *Money       `db:"connection_fee"         json:"connection_fee,omitempty"`
	Mode                *string      `db:"mode"                   json:"mode,omitempty"`
	CreatedAt           time.Time    `db:"created_at"             json:"created_at"`
}

// FeeSchedule holds tenant-level default + max fee rates (TFB SetCommonSettings defaults).
type FeeSchedule struct {
	TenantID                   int64     `gorm:"primaryKey" db:"tenant_id"                     json:"tenant_id"`
	DefaultIncentiveFee        *Money    `db:"default_incentive_fee"         json:"default_incentive_fee,omitempty"`
	MaxIncentiveFee            *Money    `db:"max_incentive_fee"             json:"max_incentive_fee,omitempty"`
	DefaultManagementFee       *Money    `db:"default_management_fee"        json:"default_management_fee,omitempty"`
	MaxManagementFee           *Money    `db:"max_management_fee"            json:"max_management_fee,omitempty"`
	DefaultAnnualManagementFee *Money    `db:"default_annual_management_fee" json:"default_annual_management_fee,omitempty"`
	MaxAnnualManagementFee     *Money    `db:"max_annual_management_fee"     json:"max_annual_management_fee,omitempty"`
	DefaultPerLotFee           *Money    `db:"default_per_lot_fee"           json:"default_per_lot_fee,omitempty"`
	DefaultPerDealFee          *Money    `db:"default_per_deal_fee"          json:"default_per_deal_fee,omitempty"`
	DefaultConnectionFee       *Money    `db:"default_connection_fee"        json:"default_connection_fee,omitempty"`
	PerformanceFeeCalcMode     *string   `db:"performance_fee_calc_mode"     json:"performance_fee_calc_mode,omitempty"` // HWM selector
	UpdatedAt                  time.Time `db:"updated_at"                    json:"updated_at"`
}

// PaymentPeriod defines when a fee is charged (TFB PaymentPeriodSettings).
type PaymentPeriod struct {
	ID                    int64       `db:"id"                       json:"id"`
	TenantID              int64       `db:"tenant_id"                json:"tenant_id"`
	AccountID             *int64      `db:"account_id"               json:"account_id,omitempty"` // NULL = tenant default
	CommissionType        *string     `db:"commission_type"          json:"commission_type,omitempty"`
	ChargeForCurrentMonth *bool       `db:"charge_for_current_month" json:"charge_for_current_month,omitempty"`
	PeriodType            *PeriodType `db:"period_type"              json:"period_type,omitempty"`
	PeriodDay             *int        `db:"period_day"               json:"period_day,omitempty"`
	ReversedDays          *bool       `db:"reversed_days"            json:"reversed_days,omitempty"`
	PeriodTime            *string     `db:"period_time"              json:"period_time,omitempty"`
}

// FeeCharge is a computed/charged fee. fees-service writes this journal; user-service posts
// the ledger entry after pamm.fee.charge.requested.
type FeeCharge struct {
	ID             int64           `db:"id"              json:"id"`
	TenantID       int64           `db:"tenant_id"       json:"tenant_id"`
	AccountID      int64           `db:"account_id"      json:"account_id"` // investor charged
	FeeAccountID   *int64          `db:"fee_account_id"  json:"fee_account_id,omitempty"`
	FeeType        FeeType         `db:"fee_type"        json:"fee_type"`
	Amount         Money           `db:"amount"          json:"amount"`
	Basis          *Money          `db:"basis"           json:"basis,omitempty"` // equity/profit/volume basis
	HWMBefore      *Money          `db:"hwm_before"      json:"hwm_before,omitempty"`
	HWMAfter       *Money          `db:"hwm_after"       json:"hwm_after,omitempty"`
	PeriodStart    *time.Time      `db:"period_start"    json:"period_start,omitempty"`
	PeriodEnd      *time.Time      `db:"period_end"      json:"period_end,omitempty"`
	IdempotencyKey string          `db:"idempotency_key" json:"idempotency_key"`
	Status         FeeChargeStatus `db:"status"          json:"status"`
	LedgerEntryID  *int64          `db:"ledger_entry_id" json:"ledger_entry_id,omitempty"` // from pamm.ledger.posted
	CreatedAt      time.Time       `db:"created_at"      json:"created_at"`
}
