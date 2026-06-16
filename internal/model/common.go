// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/model/common.go
// Role: Domain models — shared types
// Description: Shared scalar types used across the domain models. Money MUST be an
// exact decimal (never float) to preserve double-entry ledger integrity.

package model

import "github.com/shopspring/decimal"

// Money mirrors SQL NUMERIC(20,8) — exact decimal for balances, P&L and fees.
type Money = decimal.Decimal

// Units mirrors SQL NUMERIC(28,12) — high-precision fund-unit accounting (NAV model).
type Units = decimal.Decimal

// Ratio mirrors SQL NUMERIC(20,12) — share ratios (Σ per master == 1).
type Ratio = decimal.Decimal
