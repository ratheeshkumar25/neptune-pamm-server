// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/repo/currency_repo.go
// Role: Interface adapters — currency persistence (core.currency)
// Description: Backs the public ListCurrencies endpoint. Tenant-scoped.

package repo

import (
	"context"

	"neptune-pamm/github.com/ratheeshkumar25/internal/model"

	"gorm.io/gorm"
)

// CurrencyRepository queries the tenant's configured currencies.
type CurrencyRepository interface {
	ListByTenant(ctx context.Context, tenantID int64) ([]model.Currency, error)
}

type currencyRepo struct{ db *gorm.DB }

// NewCurrencyRepository builds a GORM-backed CurrencyRepository.
func NewCurrencyRepository(db *gorm.DB) CurrencyRepository { return &currencyRepo{db: db} }

func (r *currencyRepo) ListByTenant(ctx context.Context, tenantID int64) ([]model.Currency, error) {
	var currencies []model.Currency
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("name ASC").
		Find(&currencies).Error
	if err != nil {
		return nil, model.NewInternal(err)
	}
	return currencies, nil
}
