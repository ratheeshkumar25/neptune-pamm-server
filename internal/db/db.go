// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/db/db.go
// Role: Frameworks & drivers — PostgreSQL (GORM) connection
// Description: Opens a pooled GORM connection from config, verifies it with a ping, and runs
// AutoMigrate for the domain models. Returns errors to the caller (cmd) — never calls
// os.Exit — so startup failures are handled in one place.

package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"neptune-pamm/github.com/ratheeshkumar25/internal/model"
	"neptune-pamm/github.com/ratheeshkumar25/pkg/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Pool defaults — tune via load testing.
const (
	maxOpenConns    = 25
	maxIdleConns    = 5
	connMaxLifetime = 30 * time.Minute
	connMaxIdleTime = 5 * time.Minute
	pingTimeout     = 5 * time.Second
)

// migrateModels is the set AutoMigrate manages. NOTE: AutoMigrate does NOT reproduce the
// multi-schema / CHECK-constrained DDL in 02-DATA-MODEL.md (tables land pluralized in public,
// money is default-precision numeric, enums/CHECKs are absent). It is a prototype convenience;
// the ledger-grade schema still needs versioned SQL migrations.
func migrateModels() []any {
	return []any{
		// auth
		&model.Principal{}, &model.SessionAudit{}, &model.TokenAudit{},
		// accounts
		&model.Account{}, &model.Master{}, &model.Investor{}, &model.Admin{},
		// platform
		&model.PlatformServer{}, &model.AccountPlatformLink{},
		// allocation
		&model.Connection{}, &model.AllocationState{}, &model.Rollover{}, &model.RolloverAllocation{},
		// trading
		&model.Deal{}, &model.DealInvestorShare{}, &model.Position{},
		// ledger
		&model.LedgerAccount{}, &model.LedgerEntry{}, &model.LedgerPosting{}, &model.LedgerBalance{},
		// operations
		&model.BalanceOperation{}, &model.CreditOperation{}, &model.Transfer{},
		// requests
		&model.Request{},
		// settings
		&model.CommonSettings{}, &model.CallbackSettings{}, &model.SMTPSettings{},
		&model.Currency{}, &model.Language{},
		&model.WebSettings{}, &model.WebColors{}, &model.WebIframe{}, &model.PeriodSetting{},
		// audit / reconciliation / outbox
		&model.ReconciliationRun{}, &model.AuditLog{}, &model.Outbox{},
		// fees
		&model.FeeAccount{}, &model.FeeSchedule{}, &model.PaymentPeriod{}, &model.FeeCharge{},
	}
}

// ConnectDB opens a pooled, verified GORM connection to PostgreSQL.
func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	slog.Info("connecting to postgres",
		"host", cfg.Postgres.DBHost, "port", cfg.Postgres.DBPort, "db", cfg.Postgres.DBName)

	gdb, err := gorm.Open(postgres.Open(cfg.Postgres.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	// Configure the underlying connection pool.
	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, fmt.Errorf("access sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	// Verify connectivity rather than discovering it on the first query.
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	slog.Info("postgres connected")

	if err := gdb.AutoMigrate(migrateModels()...); err != nil {
		return nil, fmt.Errorf("auto-migrate: %w", err)
	}
	slog.Info("auto-migrate complete", "models", len(migrateModels()))

	return gdb, nil
}
