// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/model/audit.go
// Role: Domain models — core schema reconciliation, audit log, and outbox
// Description: Reconciliation runs vs the platform, the immutable audit trail, and the
// transactional outbox relayed to NATS.

package model

import (
	"encoding/json"
	"time"
)

// ReconStatus enumerates reconciliation-run outcomes.
type ReconStatus string

const (
	ReconOk            ReconStatus = "Ok"
	ReconDriftDetected ReconStatus = "DriftDetected"
	ReconHalted        ReconStatus = "Halted"
)

// ReconciliationRun records one ledger-vs-platform reconciliation pass and any drift.
type ReconciliationRun struct {
	ID               int64           `db:"id"                 json:"id"`
	TenantID         int64           `db:"tenant_id"          json:"tenant_id"`
	PlatformServerID *int64          `db:"platform_server_id" json:"platform_server_id,omitempty"`
	StartedAt        time.Time       `db:"started_at"         json:"started_at"`
	FinishedAt       *time.Time      `db:"finished_at"        json:"finished_at,omitempty"`
	DriftTotal       *Money          `db:"drift_total"        json:"drift_total,omitempty"`
	Status           ReconStatus     `db:"status"             json:"status"`
	Detail           json.RawMessage `db:"detail"             json:"detail,omitempty"`
}

// AuditLog is the append-only, immutable audit trail of state changes.
type AuditLog struct {
	ID             int64           `db:"id"               json:"id"`
	TenantID       int64           `db:"tenant_id"        json:"tenant_id"`
	ActorAccountID *int64          `db:"actor_account_id" json:"actor_account_id,omitempty"`
	Action         string          `db:"action"           json:"action"`
	EntityType     *string         `db:"entity_type"      json:"entity_type,omitempty"`
	EntityID       *int64          `db:"entity_id"        json:"entity_id,omitempty"`
	Before         json.RawMessage `db:"before"           json:"before,omitempty"`
	After          json.RawMessage `db:"after"            json:"after,omitempty"`
	At             time.Time       `db:"at"               json:"at"`
}

// Outbox is the transactional outbox row relayed to NATS (outbox pattern).
type Outbox struct {
	ID          int64           `db:"id"           json:"id"`
	Subject     string          `db:"subject"      json:"subject"`
	Payload     json.RawMessage `db:"payload"      json:"payload"`
	CreatedAt   time.Time       `db:"created_at"   json:"created_at"`
	PublishedAt *time.Time      `db:"published_at" json:"published_at,omitempty"`
}
