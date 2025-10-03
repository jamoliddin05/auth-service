package configs

import (
	"context"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Wrapper struct {
	db *gorm.DB
}

// NewDBWrapper opens the database and returns a wrapper that satisfies Closer.
func NewDBWrapper(dsn string) (*Wrapper, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Wrapper{db: db}, nil
}

// DB exposes the underlying gorm.DB so repositories can use it.
func (d *Wrapper) DB() *gorm.DB {
	return d.db
}

// Close shuts down the underlying sql.DB connection.
func (d *Wrapper) Close(_ context.Context) error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}

	// sqlDB.Close() doesn’t take a context, so we just ignore ctx
	return sqlDB.Close()
}
