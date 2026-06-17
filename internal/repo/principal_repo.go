// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/repo/principal_repo.go
// Role: Interface adapters — auth principal persistence (auth.principal)
// Description: Login-identity persistence. Lookups are tenant-scoped. The repo stores the
// already-hashed password (argon2id) it is given and NEVER hashes or compares — that is the
// service layer's job; the repo must never receive or store a plaintext password.

package repo

import (
	"context"
	"time"

	"neptune-pamm/github.com/ratheeshkumar25/internal/model"

	"gorm.io/gorm"
)

// PrincipalRepository persists and queries login identities.
type PrincipalRepository interface {
	Create(ctx context.Context, p *model.Principal) error
	GetByUsername(ctx context.Context, tenantID int64, username string) (*model.Principal, error)
	GetByEmail(ctx context.Context, tenantID int64, email string) (*model.Principal, error)
	GetByAccountID(ctx context.Context, tenantID, accountID int64) (*model.Principal, error)
	ExistsByUsername(ctx context.Context, tenantID int64, username string) (bool, error)

	SetEmailVerified(ctx context.Context, principalID int64) error
	UpdateLastLogin(ctx context.Context, principalID int64, at time.Time) error
	UpdatePasswordHash(ctx context.Context, principalID int64, passwordHash string) error
}

type principalRepo struct{ db *gorm.DB }

// NewPrincipalRepository builds a GORM-backed PrincipalRepository.
func NewPrincipalRepository(db *gorm.DB) PrincipalRepository { return &principalRepo{db: db} }

func (r *principalRepo) Create(ctx context.Context, p *model.Principal) error {
	return writeErr(r.db.WithContext(ctx).Create(p).Error)
}

// GetByUsername retrieves a principal by tenant-scoped username. Returns model.ErrNotFound if not found.
func (r *principalRepo) GetByUsername(ctx context.Context, tenantID int64, username string) (*model.Principal, error) {
	var p model.Principal
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND username = ?", tenantID, username).
		First(&p).Error
	if err != nil {
		return nil, readErr(err, "principal", username)
	}
	return &p, nil
}

// GetByEmail retrieves a principal by tenant-scoped email. Returns model.ErrNotFound if not found.
func (r *principalRepo) GetByEmail(ctx context.Context, tenantID int64, email string) (*model.Principal, error) {
	var p model.Principal
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND email = ?", tenantID, email).
		First(&p).Error
	if err != nil {
		return nil, readErr(err, "principal", email)
	}
	return &p, nil
}

// GetByAccountID retrieves a principal by tenant-scoped account ID. Returns model.ErrNotFound if not found.
func (r *principalRepo) GetByAccountID(ctx context.Context, tenantID, accountID int64) (*model.Principal, error) {
	var p model.Principal
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND account_id = ?", tenantID, accountID).
		First(&p).Error
	if err != nil {
		return nil, readErr(err, "principal", accountID)
	}
	return &p, nil
}

// ExistsByUsername checks if a principal with the given username exists in the tenant. Returns true if found, false otherwise.
func (r *principalRepo) ExistsByUsername(ctx context.Context, tenantID int64, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Principal{}).
		Where("tenant_id = ? AND username = ?", tenantID, username).
		Count(&count).Error
	if err != nil {
		return false, model.NewInternal(err)
	}
	return count > 0, nil
}

// SetEmailVerified marks the principal's email as verified. Returns model.ErrNotFound if the principal does not exist.
func (r *principalRepo) SetEmailVerified(ctx context.Context, principalID int64) error {
	res := r.db.WithContext(ctx).Model(&model.Principal{}).
		Where("id = ?", principalID).
		Update("email_verified", true)
	if res.Error != nil {
		return model.NewInternal(res.Error)
	}
	if res.RowsAffected == 0 {
		return model.NewNotFound("principal", principalID)
	}
	return nil
}

func (r *principalRepo) UpdateLastLogin(ctx context.Context, principalID int64, at time.Time) error {
	res := r.db.WithContext(ctx).Model(&model.Principal{}).
		Where("id = ?", principalID).
		Update("last_login_at", at)
	if res.Error != nil {
		return model.NewInternal(res.Error)
	}
	if res.RowsAffected == 0 {
		return model.NewNotFound("principal", principalID)
	}
	return nil
}

// UpdatePasswordHash updates the principal's password hash. Returns model.ErrNotFound if the principal does not exist.
func (r *principalRepo) UpdatePasswordHash(ctx context.Context, principalID int64, passwordHash string) error {
	res := r.db.WithContext(ctx).Model(&model.Principal{}).
		Where("id = ?", principalID).
		Update("password_hash", passwordHash)
	if res.Error != nil {
		return model.NewInternal(res.Error)
	}
	if res.RowsAffected == 0 {
		return model.NewNotFound("principal", principalID)
	}
	return nil
}
