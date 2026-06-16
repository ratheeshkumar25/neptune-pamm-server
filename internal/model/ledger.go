// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/model/ledger.go
// Role: Domain models — core schema double-entry ledger (source of truth for balances)
// Description: Chart of accounts, balanced journal entries, postings, and materialized balances.

package model

import "time"

// LedgerAccountType enumerates the real-world buckets a ledger account represents.
type LedgerAccountType string

const (
	LedgerInvestorBalance LedgerAccountType = "InvestorBalance"
	LedgerInvestorCredit  LedgerAccountType = "InvestorCredit"
	LedgerMasterBalance   LedgerAccountType = "MasterBalance"
	LedgerFeeAccount      LedgerAccountType = "FeeAccount"
	LedgerBrokerSuspense  LedgerAccountType = "BrokerSuspense"
	LedgerPnLClearing     LedgerAccountType = "PnLClearing"
	LedgerSwapClearing    LedgerAccountType = "SwapClearing"
)

// LedgerEntryKind enumerates the business reason for a journal entry.
type LedgerEntryKind string

const (
	EntryDeposit    LedgerEntryKind = "Deposit"
	EntryWithdraw   LedgerEntryKind = "Withdraw"
	EntryCredit     LedgerEntryKind = "Credit"
	EntryTransfer   LedgerEntryKind = "Transfer"
	EntryPnL        LedgerEntryKind = "PnL"
	EntrySwap       LedgerEntryKind = "Swap"
	EntryFee        LedgerEntryKind = "Fee"
	EntryCorrection LedgerEntryKind = "Correction"
	EntryReversal   LedgerEntryKind = "Reversal"
)

// LedgerEntrySource enumerates where an entry originated.
type LedgerEntrySource string

const (
	SourceBalanceOp   LedgerEntrySource = "balance_op"
	SourceRollover    LedgerEntrySource = "rollover"
	SourceFeeCharge   LedgerEntrySource = "fee_charge"
	SourceRequest     LedgerEntrySource = "request"
	SourceMaintenance LedgerEntrySource = "maintenance"
)

// LedgerAccount is one chart-of-accounts bucket.
type LedgerAccount struct {
	ID             int64             `db:"id"               json:"id"`
	TenantID       int64             `db:"tenant_id"        json:"tenant_id"`
	OwnerAccountID *int64            `db:"owner_account_id" json:"owner_account_id,omitempty"` // NULL for system accounts
	Type           LedgerAccountType `db:"type"             json:"type"`
	Currency       string            `db:"currency"         json:"currency"`
	CreatedAt      time.Time         `db:"created_at"       json:"created_at"`
}

// LedgerEntry is a balanced journal entry (Σ debit == Σ credit). Reversible, never edited.
type LedgerEntry struct {
	ID             int64              `db:"id"              json:"id"`
	TenantID       int64              `db:"tenant_id"       json:"tenant_id"`
	Kind           LedgerEntryKind    `db:"kind"            json:"kind"`
	Description    *string            `db:"description"     json:"description,omitempty"`
	IdempotencyKey string             `db:"idempotency_key" json:"idempotency_key"` // dedupe; one entry per business op
	ReversalOf     *int64             `db:"reversal_of"     json:"reversal_of,omitempty"`
	Source         *LedgerEntrySource `db:"source"          json:"source,omitempty"`
	SourceID       *int64             `db:"source_id"       json:"source_id,omitempty"`
	PostedAt       time.Time          `db:"posted_at"       json:"posted_at"`
	CreatedBy      *int64             `db:"created_by"      json:"created_by,omitempty"`
}

// LedgerPosting is an individual debit/credit leg. App enforces per-entry balance.
type LedgerPosting struct {
	ID              int64  `db:"id"                json:"id"`
	EntryID         int64  `db:"entry_id"          json:"entry_id"`
	LedgerAccountID int64  `db:"ledger_account_id" json:"ledger_account_id"`
	Debit           Money  `db:"debit"             json:"debit"`
	Credit          Money  `db:"credit"            json:"credit"`
	Currency        string `db:"currency"          json:"currency"`
}

// LedgerBalance is the materialized running balance per ledger account (rebuildable).
type LedgerBalance struct {
	LedgerAccountID int64     `gorm:"primaryKey" db:"ledger_account_id" json:"ledger_account_id"`
	Balance         Money     `db:"balance"           json:"balance"`
	UpdatedAt       time.Time `db:"updated_at"        json:"updated_at"`
}
