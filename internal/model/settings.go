// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/model/settings.go
// Role: Domain models — core schema settings & period configuration
// Description: Common/SMTP/web settings, currencies, languages, callbacks, and periods.

package model

import (
	"encoding/json"
	"time"
)

// CommonSettings is the global/common settings blob (TFB SetCommonSettings). One per tenant.
type CommonSettings struct {
	TenantID  int64           `db:"tenant_id"  json:"tenant_id"`
	Settings  json.RawMessage `db:"settings"   json:"settings"` // alloc/withdrawal/fees/risk/plan times…
	UpdatedAt time.Time       `db:"updated_at" json:"updated_at"`
}

// CallbackType enumerates outbound callback event types (TFB Callback).
type CallbackType string

const (
	CallbackPlannedRequest CallbackType = "PlannedRequest"
	CallbackDisconnect     CallbackType = "Disconnect"
)

// CallbackSettings is the outbound webhook callback config.
type CallbackSettings struct {
	ID            int64          `db:"id"             json:"id"`
	TenantID      int64          `db:"tenant_id"      json:"tenant_id"`
	Address       string         `db:"address"        json:"address"`
	CallbackTypes []CallbackType `db:"callback_types" json:"callback_types"`
	Enabled       bool           `db:"enabled"        json:"enabled"`
}

// SMTPSettings holds per-tenant outbound email config.
type SMTPSettings struct {
	TenantID              int64    `db:"tenant_id"              json:"tenant_id"`
	IsEnabled             bool     `db:"is_enabled"            json:"is_enabled"`
	Host                  string   `db:"host"                  json:"host"`
	Port                  int      `db:"port"                  json:"port"`
	Login                 string   `db:"login"                 json:"login"`
	Password              string   `db:"password"              json:"-"`
	SenderAddress         string   `db:"sender_address"        json:"sender_address"`
	EnableSSL             bool     `db:"enable_ssl"            json:"enable_ssl"`
	NotificationAddresses []string `db:"notification_addresses" json:"notification_addresses"`
}

// Currency is a configured currency (per tenant).
type Currency struct {
	TenantID int64   `db:"tenant_id" json:"tenant_id"`
	Name     string  `db:"name"      json:"name"`
	Digits   int     `db:"digits"    json:"digits"`
	Symbol   *string `db:"symbol"    json:"symbol,omitempty"`
}

// Language is a configured UI language (per tenant).
type Language struct {
	TenantID int64  `db:"tenant_id" json:"tenant_id"`
	Code     string `db:"code"      json:"code"`
	Enabled  bool   `db:"enabled"   json:"enabled"`
}

// WebSettings / WebColors / WebIframe hold low-churn JSONB web/theming config.
type WebSettings struct {
	TenantID int64           `db:"tenant_id" json:"tenant_id"`
	Settings json.RawMessage `db:"settings"  json:"settings"`
}

type WebColors struct {
	TenantID int64           `db:"tenant_id" json:"tenant_id"`
	Settings json.RawMessage `db:"settings"  json:"settings"`
}

type WebIframe struct {
	TenantID int64           `db:"tenant_id" json:"tenant_id"`
	Settings json.RawMessage `db:"settings"  json:"settings"`
}

// PeriodKind enumerates scheduled-period kinds.
type PeriodKind string

const (
	PeriodStatement PeriodKind = "Statement"
	PeriodReport    PeriodKind = "Report"
	PeriodEndOfDay  PeriodKind = "EndOfDay"
)

// PeriodType enumerates the cadence of a period.
type PeriodType string

const (
	PeriodDaily     PeriodType = "Daily"
	PeriodWeekly    PeriodType = "Weekly"
	PeriodMonthly   PeriodType = "Monthly"
	PeriodQuarterly PeriodType = "Quarterly"
)

// PeriodSetting is a per-account or global scheduled period.
type PeriodSetting struct {
	ID                   int64       `db:"id"                      json:"id"`
	TenantID             int64       `db:"tenant_id"               json:"tenant_id"`
	AccountID            *int64      `db:"account_id"              json:"account_id,omitempty"` // NULL = tenant default
	Kind                 PeriodKind  `db:"kind"                    json:"kind"`
	PeriodType           *PeriodType `db:"period_type"             json:"period_type,omitempty"`
	PeriodDay            *int        `db:"period_day"              json:"period_day,omitempty"`
	ReversedDays         *bool       `db:"reversed_days"           json:"reversed_days,omitempty"`
	PeriodTime           *string     `db:"period_time"             json:"period_time,omitempty"` // HH:MM:SS UTC
	IncludeAllOpenOrders *bool       `db:"include_all_open_orders" json:"include_all_open_orders,omitempty"`
}
