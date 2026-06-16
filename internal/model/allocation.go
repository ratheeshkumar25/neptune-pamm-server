// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/model/allocation.go
// Role: Domain models — core schema connections & allocation state
// Description: Investor↔master connections, NAV/unit allocation state, and rollovers.

package model

import "time"

// ConnectionStatus enumerates the connection lifecycle.
type ConnectionStatus string

const (
	StatusConnected     ConnectionStatus = "Connected"
	StatusDisconnecting ConnectionStatus = "Disconnecting"
	StatusDisconnected  ConnectionStatus = "Disconnected"
)

// Connection is an investor↔master connection (connect/disconnect history).
type Connection struct {
	ID                   int64            `db:"id"                     json:"id"`
	TenantID             int64            `db:"tenant_id"              json:"tenant_id"`
	MasterID             int64            `db:"master_id"              json:"master_id"`
	InvestorID           int64            `db:"investor_id"            json:"investor_id"`
	Status               ConnectionStatus `db:"status"                 json:"status"`
	ConnectedAt          time.Time        `db:"connected_at"           json:"connected_at"`
	DisconnectedAt       *time.Time       `db:"disconnected_at"        json:"disconnected_at,omitempty"`
	DisconnectReason     *string          `db:"disconnect_reason"      json:"disconnect_reason,omitempty"`
	DisconnectReasonCode *int             `db:"disconnect_reason_code" json:"disconnect_reason_code,omitempty"`
}

// AllocationState is a per-investor share of a master pool (NAV/unit accounting).
type AllocationState struct {
	ID             int64     `db:"id"               json:"id"`
	TenantID       int64     `db:"tenant_id"        json:"tenant_id"`
	MasterID       int64     `db:"master_id"        json:"master_id"`
	InvestorID     int64     `db:"investor_id"      json:"investor_id"`
	Units          Units     `db:"units"            json:"units"`       // fund units held
	ShareRatio     Ratio     `db:"share_ratio"      json:"share_ratio"` // equity/Σequity at last rollover (Σ=1)
	LastRolloverID *int64    `db:"last_rollover_id" json:"last_rollover_id,omitempty"`
	UpdatedAt      time.Time `db:"updated_at"       json:"updated_at"`
}

// RolloverTrigger enumerates what initiated a rollover.
type RolloverTrigger string

const (
	TriggerScheduled   RolloverTrigger = "Scheduled"
	TriggerManual      RolloverTrigger = "Manual"
	TriggerTransaction RolloverTrigger = "Transaction"
)

// RolloverStatus enumerates rollover processing states.
type RolloverStatus string

const (
	RolloverPending   RolloverStatus = "Pending"
	RolloverRunning   RolloverStatus = "Running"
	RolloverCompleted RolloverStatus = "Completed"
	RolloverError     RolloverStatus = "Error"
)

// Rollover is a non-trading settlement event that distributes P&L/swap and recomputes shares.
type Rollover struct {
	ID                int64           `db:"id"                  json:"id"`
	TenantID          int64           `db:"tenant_id"           json:"tenant_id"`
	MasterID          int64           `db:"master_id"           json:"master_id"`
	RolloverTime      time.Time       `db:"rollover_time"       json:"rollover_time"`
	Trigger           RolloverTrigger `db:"trigger"             json:"trigger"`
	TotalEquityBefore *Money          `db:"total_equity_before" json:"total_equity_before,omitempty"`
	TotalUnits        *Units          `db:"total_units"         json:"total_units,omitempty"`
	NavPerUnit        *Units          `db:"nav_per_unit"        json:"nav_per_unit,omitempty"`
	PnLDistributed    *Money          `db:"pnl_distributed"     json:"pnl_distributed,omitempty"`
	SwapDistributed   *Money          `db:"swap_distributed"    json:"swap_distributed,omitempty"`
	Status            RolloverStatus  `db:"status"              json:"status"`
	CreatedAt         time.Time       `db:"created_at"          json:"created_at"`
}

// RolloverAllocation is the per-investor result of one rollover (audit of the distribution).
type RolloverAllocation struct {
	ID           int64  `db:"id"            json:"id"`
	RolloverID   int64  `db:"rollover_id"   json:"rollover_id"`
	InvestorID   int64  `db:"investor_id"   json:"investor_id"`
	EquityBefore *Money `db:"equity_before" json:"equity_before,omitempty"`
	ShareRatio   *Ratio `db:"share_ratio"   json:"share_ratio,omitempty"`
	PnLShare     *Money `db:"pnl_share"     json:"pnl_share,omitempty"`
	SwapShare    *Money `db:"swap_share"    json:"swap_share,omitempty"`
	FeeCharged   Money  `db:"fee_charged"   json:"fee_charged"`
	EquityAfter  *Money `db:"equity_after"  json:"equity_after,omitempty"`
}
