// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/model/trade.go
// Role: Domain models — core schema trading data (deals / positions)
// Description: Ingested master deals, their per-investor P&L splits, and open positions.

package model

import "time"

// DealSide enumerates trade direction.
type DealSide string

const (
	SideBuy  DealSide = "Buy"
	SideSell DealSide = "Sell"
)

// Deal is a master's closed deal, ingested from the MT5 connector (canonical, deduped).
type Deal struct {
	ID               int64      `db:"id"                 json:"id"`
	TenantID         int64      `db:"tenant_id"          json:"tenant_id"`
	MasterID         int64      `db:"master_id"          json:"master_id"`
	PlatformServerID int64      `db:"platform_server_id" json:"platform_server_id"`
	Ticket           int64      `db:"ticket"             json:"ticket"` // platform deal/order ticket
	Symbol           *string    `db:"symbol"             json:"symbol,omitempty"`
	Side             *DealSide  `db:"side"               json:"side,omitempty"`
	VolumeLots       *Money     `db:"volume_lots"        json:"volume_lots,omitempty"`
	OpenTime         *time.Time `db:"open_time"          json:"open_time,omitempty"`
	CloseTime        *time.Time `db:"close_time"         json:"close_time,omitempty"`
	OpenPrice        *Money     `db:"open_price"         json:"open_price,omitempty"`
	ClosePrice       *Money     `db:"close_price"        json:"close_price,omitempty"`
	Profit           *Money     `db:"profit"             json:"profit,omitempty"`
	Swap             *Money     `db:"swap"               json:"swap,omitempty"`
	Commission       *Money     `db:"commission"         json:"commission,omitempty"`
	ExternalID       string     `db:"external_id"        json:"external_id"` // dedupe key = platform:server:ticket
	IngestedAt       time.Time  `db:"ingested_at"        json:"ingested_at"`
}

// DealInvestorShare records how a master deal's P&L was split across investors.
type DealInvestorShare struct {
	ID          int64  `db:"id"           json:"id"`
	DealID      int64  `db:"deal_id"      json:"deal_id"`
	InvestorID  int64  `db:"investor_id"  json:"investor_id"`
	ShareRatio  *Ratio `db:"share_ratio"  json:"share_ratio,omitempty"`
	ProfitShare *Money `db:"profit_share" json:"profit_share,omitempty"`
	SwapShare   *Money `db:"swap_share"   json:"swap_share,omitempty"`
}

// Position is an open-position snapshot (latest per master, mark-to-market for equity).
type Position struct {
	ID               int64     `db:"id"                 json:"id"`
	TenantID         int64     `db:"tenant_id"          json:"tenant_id"`
	MasterID         int64     `db:"master_id"          json:"master_id"`
	PlatformServerID int64     `db:"platform_server_id" json:"platform_server_id"`
	Ticket           int64     `db:"ticket"             json:"ticket"`
	Symbol           *string   `db:"symbol"             json:"symbol,omitempty"`
	Side             *DealSide `db:"side"               json:"side,omitempty"`
	VolumeLots       *Money    `db:"volume_lots"        json:"volume_lots,omitempty"`
	OpenPrice        *Money    `db:"open_price"         json:"open_price,omitempty"`
	CurrentPrice     *Money    `db:"current_price"      json:"current_price,omitempty"`
	FloatingPnL      *Money    `db:"floating_pnl"       json:"floating_pnl,omitempty"`
	Swap             *Money    `db:"swap"               json:"swap,omitempty"`
	UpdatedAt        time.Time `db:"updated_at"         json:"updated_at"`
}
