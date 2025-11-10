package configs

import (
	"context"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Wrapper struct {
	db *gorm.DB
}

func NewDBWrapper(dsn string) (*Wrapper, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Wrapper{db: db}, nil
}

func (d *Wrapper) DB() *gorm.DB {
	return d.db
}

func (d *Wrapper) Close(_ context.Context) error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
