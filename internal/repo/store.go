// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/repo/store.go
// Role: Interface adapters — persistence (repository ports + unit-of-work)
// Description: Store is the unit-of-work aggregate. It exposes every repository and can run
// several of them inside a single transaction via Atomic (registration creates account +
// principal + role-satellite atomically). All errors are translated to model.AppError so
// internal driver details never leak past this layer.

package repo

import (
	"context"
	"errors"

	"neptune-pamm/github.com/ratheeshkumar25/internal/model"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// Store is the unit-of-work: repositories plus an Atomic transaction boundary.
type Store interface {
	Account() AccountRepository
	Principal() PrincipalRepository
	Currency() CurrencyRepository

	// Atomic runs fn inside one transaction; every repo obtained from the passed Store
	// shares that transaction and is rolled back together on error.
	Atomic(ctx context.Context, fn func(s Store) error) error
}

type gormStore struct{ db *gorm.DB }

// NewStore builds a Store backed by GORM.
func NewStore(db *gorm.DB) Store { return &gormStore{db: db} }

func (s *gormStore) Account() AccountRepository     { return &accountRepo{db: s.db} }
func (s *gormStore) Principal() PrincipalRepository { return &principalRepo{db: s.db} }
func (s *gormStore) Currency() CurrencyRepository   { return &currencyRepo{db: s.db} }

func (s *gormStore) Atomic(ctx context.Context, fn func(s Store) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&gormStore{db: tx})
	})
}

// --- error translation (keep driver details out of higher layers) ----------------

// pgUniqueViolation is the PostgreSQL SQLSTATE for a unique-constraint violation.
const pgUniqueViolation = "23505"

// writeErr maps a write error to a domain error: unique violations become Conflict,
// everything else an Internal error that wraps (but does not expose) the cause.
func writeErr(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return model.ErrConflict.Wrap(err)
	}
	return model.NewInternal(err)
}

// readErr maps a read error: gorm "record not found" becomes a domain NotFound for the
// given entity/id; anything else an Internal error.
func readErr(err error, entity string, id any) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.NewNotFound(entity, id)
	}
	return model.NewInternal(err)
}
