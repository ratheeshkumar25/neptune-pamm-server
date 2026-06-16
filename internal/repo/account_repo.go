// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/repo/account_repo.go
// Role: Interface adapters — account persistence (core.account + role satellites)
// Description: Reads are tenant-scoped so one tenant can never read another's rows (defends
// against IDOR / cross-tenant access). The caller passes tenant_id from the authenticated
// context — never from client-controlled input.

package repo

import (
	"context"

	"neptune-pamm/github.com/ratheeshkumar25/internal/model"

	"gorm.io/gorm"
)

// AccountRepository persists accounts and their role satellites.
type AccountRepository interface {
	Create(ctx context.Context, a *model.Account) error
	GetByID(ctx context.Context, tenantID, id int64) (*model.Account, error)

	CreateInvestor(ctx context.Context, inv *model.Investor) error
	CreateMaster(ctx context.Context, m *model.Master) error
	CreateAdmin(ctx context.Context, a *model.Admin) error

	GetInvestor(ctx context.Context, tenantID, accountID int64) (*model.Investor, error)
	GetMaster(ctx context.Context, tenantID, accountID int64) (*model.Master, error)
}

// accountRepo is the GORM-backed AccountRepository implementation.
type accountRepo struct {
	db *gorm.DB
}

// NewAccountRepository builds a GORM-backed AccountRepository.
func NewAccountRepository(db *gorm.DB) AccountRepository {
	return &accountRepo{db: db}
}

func (r *accountRepo) Create(ctx context.Context, a *model.Account) error {
	return writeErr(r.db.WithContext(ctx).Create(a).Error)
}

func (r *accountRepo) GetByID(ctx context.Context, tenantID, id int64) (*model.Account, error) {
	var a model.Account
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&a).Error
	if err != nil {
		return nil, readErr(err, "account", id)
	}
	return &a, nil
}

func (r *accountRepo) CreateInvestor(ctx context.Context, inv *model.Investor) error {
	return writeErr(r.db.WithContext(ctx).Create(inv).Error)
}

func (r *accountRepo) CreateMaster(ctx context.Context, m *model.Master) error {
	return writeErr(r.db.WithContext(ctx).Create(m).Error)
}

func (r *accountRepo) CreateAdmin(ctx context.Context, a *model.Admin) error {
	return writeErr(r.db.WithContext(ctx).Create(a).Error)
}

func (r *accountRepo) GetInvestor(ctx context.Context, tenantID, accountID int64) (*model.Investor, error) {
	var inv model.Investor
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND account_id = ?", tenantID, accountID).
		First(&inv).Error
	if err != nil {
		return nil, readErr(err, "investor", accountID)
	}
	return &inv, nil
}

func (r *accountRepo) GetMaster(ctx context.Context, tenantID, accountID int64) (*model.Master, error) {
	var m model.Master
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND account_id = ?", tenantID, accountID).
		First(&m).Error
	if err != nil {
		return nil, readErr(err, "master", accountID)
	}
	return &m, nil
}
