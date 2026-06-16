// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/pkg/utilis/nats_events.go
// Role: Frameworks & drivers — NATS subjects & event envelope
// Description: The canonical NATS subject catalogue (01-ARCHITECTURE §4) and a small JSON
// publish helper. Subjects are transport-level constants so publishers/consumers never
// hand-type subject strings.

package utilis

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

// Subject is a NATS subject name.
type Subject string

// Platform ingress (gateway republishes connector webhooks). Scope: MT5 only.
const (
	SubjPlatformDeal       Subject = "pamm.platform.mt5.deal"
	SubjPlatformPosition   Subject = "pamm.platform.mt5.position"
	SubjPlatformAccount    Subject = "pamm.platform.mt5.account"
	SubjPlatformBalanceAck Subject = "pamm.platform.mt5.balanceop.ack"
	SubjPlatformHeartbeat  Subject = "pamm.platform.mt5.heartbeat"
)

// Trade / allocation / ledger / fee lifecycle.
const (
	SubjTradeDealIngested    Subject = "pamm.trade.deal.ingested"
	SubjRolloverStarted      Subject = "pamm.allocation.rollover.started"
	SubjRolloverCompleted    Subject = "pamm.allocation.rollover.completed"
	SubjLedgerPostingRequest Subject = "pamm.ledger.posting.requested"
	SubjLedgerPosted         Subject = "pamm.ledger.posted"
	SubjFeeChargeRequested   Subject = "pamm.fee.charge.requested"
	SubjFeeCharged           Subject = "pamm.fee.charged"
)

// Account / request lifecycle (also fanned out to UIs over websocket).
const (
	SubjAccountConnected    Subject = "pamm.account.connected"
	SubjAccountDisconnected Subject = "pamm.account.disconnected"
	SubjAccountEquityCall   Subject = "pamm.account.equitycall"
	SubjAccountAutoDisconn  Subject = "pamm.account.autodisconnect"
	SubjRequestCreated      Subject = "pamm.request.created"
	SubjRequestApproved     Subject = "pamm.request.approved"
	SubjRequestExecuted     Subject = "pamm.request.executed"
	SubjRequestRejected     Subject = "pamm.request.rejected"
)

// Notifications (email/portal events consumed by the notification worker).
const SubjNotifyPrefix Subject = "pamm.notify." // append the concrete event, e.g. "welcome"

// String returns the subject as a plain string.
func (s Subject) String() string { return string(s) }

// PublishJSON marshals v and publishes it on a core NATS subject (at-most-once).
func PublishJSON(nc *nats.Conn, subj Subject, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", subj, err)
	}
	if err := nc.Publish(subj.String(), data); err != nil {
		return fmt.Errorf("publish %s: %w", subj, err)
	}
	return nil
}

// PublishJSONJS marshals v and publishes it durably via JetStream (at-least-once).
func PublishJSONJS(js nats.JetStreamContext, subj Subject, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", subj, err)
	}
	if _, err := js.Publish(subj.String(), data); err != nil {
		return fmt.Errorf("publish(js) %s: %w", subj, err)
	}
	return nil
}
