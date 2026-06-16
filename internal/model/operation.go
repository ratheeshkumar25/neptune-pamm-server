// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/model/operation.go
// Role: Domain models — core schema balance/credit operations & transfers (TFB parity)
// Description: Money-movement records that post through the double-entry ledger.

package model

import "time"

// BalanceOpType enumerates balance-operation kinds.
type BalanceOpType string

const (
	OpBalance           BalanceOpType = "Balance"
	OpBalanceCorrection BalanceOpType = "BalanceCorrection"
	OpProfitCorrection  BalanceOpType = "ProfitCorrection"
)

// WithdrawalMethod enumerates how a withdrawal is sourced on the platform.
type WithdrawalMethod string

const (
	MethodPartialClose WithdrawalMethod = "PartialClose"
	MethodFreeMargin   WithdrawalMethod = "FreeMargin"
)

// BalanceOperation is a deposit/withdraw/correction against an account. Posts to the ledger.
type BalanceOperation struct {
	ID              int64             `db:"id"                  json:"id"`
	TenantID        int64             `db:"tenant_id"           json:"tenant_id"`
	AccountID       int64             `db:"account_id"          json:"account_id"`
	Amount          Money             `db:"amount"              json:"amount"`
	Currency        *string           `db:"currency"            json:"currency,omitempty"`
	OperationType   BalanceOpType     `db:"operation_type"      json:"operation_type"`
	Method          *WithdrawalMethod `db:"method"              json:"method,omitempty"`
	Comment         *string           `db:"comment"             json:"comment,omitempty"`
	CheckMargin     bool              `db:"check_margin"        json:"check_margin"`
	Synced          bool              `db:"synced"              json:"synced"` // mirrored to platform?
	MTSyncedOrderID *int64            `db:"mt_synced_order_id"  json:"mt_synced_order_id,omitempty"`
	LedgerEntryID   *int64            `db:"ledger_entry_id"     json:"ledger_entry_id,omitempty"`
	CreatedAt       time.Time         `db:"created_at"          json:"created_at"`
}

// CreditOperation is a bonus/credit with optional expiry.
type CreditOperation struct {
	ID              int64      `db:"id"                 json:"id"`
	TenantID        int64      `db:"tenant_id"          json:"tenant_id"`
	AccountID       int64      `db:"account_id"         json:"account_id"`
	Amount          Money      `db:"amount"             json:"amount"`
	Currency        *string    `db:"currency"           json:"currency,omitempty"`
	Expiration      *time.Time `db:"expiration"         json:"expiration,omitempty"`
	OperationType   *string    `db:"operation_type"     json:"operation_type,omitempty"`
	Synced          bool       `db:"synced"             json:"synced"`
	MTSyncedOrderID *int64     `db:"mt_synced_order_id" json:"mt_synced_order_id,omitempty"`
	LedgerEntryID   *int64     `db:"ledger_entry_id"    json:"ledger_entry_id,omitempty"`
	CreatedAt       time.Time  `db:"created_at"         json:"created_at"`
}

// Transfer is an investor→investor transfer.
type Transfer struct {
	ID             int64     `db:"id"               json:"id"`
	TenantID       int64     `db:"tenant_id"        json:"tenant_id"`
	FromInvestorID int64     `db:"from_investor_id" json:"from_investor_id"`
	ToInvestorID   int64     `db:"to_investor_id"   json:"to_investor_id"`
	Amount         Money     `db:"amount"           json:"amount"`
	Currency       *string   `db:"currency"         json:"currency,omitempty"`
	Method         *string   `db:"method"           json:"method,omitempty"`
	Comment        *string   `db:"comment"          json:"comment,omitempty"`
	LedgerEntryID  *int64    `db:"ledger_entry_id"  json:"ledger_entry_id,omitempty"`
	CreatedAt      time.Time `db:"created_at"       json:"created_at"`
}
