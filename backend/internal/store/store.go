// Package store provides SQLite-backed persistence. It opens the database,
// runs migrations, and exposes GORM-backed implementations of the domain
// repository interfaces.
package store

import (
	"ispcrm/internal/customer"
	"ispcrm/internal/product"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open opens (and creates if needed) a SQLite database at the given DSN.
// Use a file path for persistence or ":memory:" for ephemeral data.
func Open(dsn string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

// Migrate creates/updates the schema for all domain models.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&customer.Customer{}, &product.Product{})
}
