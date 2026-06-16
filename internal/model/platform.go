// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/model/platform.go
// Role: Domain models — core schema platform servers & account links
// Description: MT5 trading servers and the mapping of accounts to platform logins.

package model

import "time"

// PlatformType enumerates supported trading platforms. Scope: MT5 only.
type PlatformType string

const (
	PlatformMT5 PlatformType = "MT5"
)

// PlatformServer is a configured MT5 trading server the broker runs.
type PlatformServer struct {
	ID                   int64        `db:"id"                      json:"id"`
	TenantID             int64        `db:"tenant_id"               json:"tenant_id"`
	ServerGUID           string       `db:"server_guid"             json:"server_guid"`
	PlatformType         PlatformType `db:"platform_type"           json:"platform_type"`
	Name                 string       `db:"name"                    json:"name"`
	DefaultGroup         *string      `db:"default_group"           json:"default_group,omitempty"`
	FreeMarginCoef       *Money       `db:"free_margin_coef"        json:"free_margin_coef,omitempty"`
	Enabled              bool         `db:"enabled"                 json:"enabled"`
	EnableLoginSequence  bool         `db:"enable_login_sequence"   json:"enable_login_sequence"`
	FirstLoginInSequence *int64       `db:"first_login_in_sequence" json:"first_login_in_sequence,omitempty"`
	ConnectorEndpoint    *string      `db:"connector_endpoint"      json:"connector_endpoint,omitempty"`
	CreatedAt            time.Time    `db:"created_at"              json:"created_at"`
}

// PlatformLinkRole enumerates the role an account's platform login plays.
type PlatformLinkRole string

const (
	LinkMaster   PlatformLinkRole = "Master"
	LinkInvestor PlatformLinkRole = "Investor"
	LinkOwnFunds PlatformLinkRole = "OwnFunds"
	LinkFee      PlatformLinkRole = "Fee"
)

// AccountPlatformLink links a Neptune account to its MT5 login on a platform server.
type AccountPlatformLink struct {
	ID               int64            `db:"id"                 json:"id"`
	TenantID         int64            `db:"tenant_id"          json:"tenant_id"`
	AccountID        int64            `db:"account_id"         json:"account_id"`
	PlatformServerID int64            `db:"platform_server_id" json:"platform_server_id"`
	MTLogin          int64            `db:"mt_login"           json:"mt_login"` // platform-side login/account id
	MTGroup          *string          `db:"mt_group"           json:"mt_group,omitempty"`
	Role             PlatformLinkRole `db:"role"               json:"role"`
	CreatedAt        time.Time        `db:"created_at"         json:"created_at"`
}
