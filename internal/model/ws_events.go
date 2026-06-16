// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/model/ws_events.go
// Role: Domain models — realtime WebSocket events
// Description: The messages the gateway fans out to UIs over WebSocket (01-ARCHITECTURE §5):
// equity ticks, rollover results, request-status changes and risk alerts. Clients subscribe to
// authorized channels (investor:{id}, master:{id}, admin:requests, admin:health).

package model

import (
	"fmt"
	"time"
)

// WSEventType discriminates a realtime message.
type WSEventType string

const (
	WSEquityTick          WSEventType = "equity_tick"
	WSRolloverCompleted   WSEventType = "rollover_completed"
	WSRequestStatus       WSEventType = "request_status"
	WSRiskAlert           WSEventType = "risk_alert"
	WSAccountConnected    WSEventType = "account_connected"
	WSAccountDisconnected WSEventType = "account_disconnected"
)

// WSMessage is the envelope pushed to a subscribed client.
type WSMessage struct {
	Type    WSEventType `json:"type"`
	Channel string      `json:"channel"`
	Time    time.Time   `json:"time"`
	Data    any         `json:"data"`
}

// NewWSMessage builds an envelope for a channel. The time is supplied by the caller (the
// domain layer avoids hidden clocks; pass time.Now() at the edge).
func NewWSMessage(t WSEventType, channel string, at time.Time, data any) WSMessage {
	return WSMessage{Type: t, Channel: channel, Time: at, Data: data}
}

// --- channel helpers -------------------------------------------------------------

const (
	WSChanAdminRequests = "admin:requests"
	WSChanAdminHealth   = "admin:health"
)

// InvestorChannel returns the per-investor channel name.
func InvestorChannel(investorID int64) string { return fmt.Sprintf("investor:%d", investorID) }

// MasterChannel returns the per-master channel name.
func MasterChannel(masterID int64) string { return fmt.Sprintf("master:%d", masterID) }

// --- event payloads --------------------------------------------------------------

// EquityTick is a mark-to-market equity update for an account.
type EquityTick struct {
	AccountID   int64 `json:"account_id"`
	Equity      Money `json:"equity"`
	Balance     Money `json:"balance"`
	FloatingPnL Money `json:"floating_pnl"`
}

// RolloverResult summarises a completed rollover for a master's subscribers.
type RolloverResult struct {
	MasterID       int64  `json:"master_id"`
	RolloverID     int64  `json:"rollover_id"`
	NavPerUnit     Units  `json:"nav_per_unit"`
	PnLDistributed Money  `json:"pnl_distributed"`
	Trigger        string `json:"trigger"`
}

// RequestStatusChange notifies a request lifecycle transition.
type RequestStatusChange struct {
	RequestID int64         `json:"request_id"`
	Type      RequestType   `json:"type"`
	Status    RequestStatus `json:"status"`
}

// RiskTrigger enumerates what fired a risk alert.
type RiskTrigger string

const (
	RiskTriggerEquityCall     RiskTrigger = "EquityCall"
	RiskTriggerAutoDisconnect RiskTrigger = "AutoDisconnect"
	RiskTriggerMaxLoss        RiskTrigger = "MaxLoss"
	RiskTriggerMaxProfit      RiskTrigger = "MaxProfit"
)

// RiskAlert notifies a risk-control event against an account.
type RiskAlert struct {
	AccountID int64       `json:"account_id"`
	Trigger   RiskTrigger `json:"trigger"`
	Equity    Money       `json:"equity"`
	Level     *Money      `json:"level,omitempty"`
}
