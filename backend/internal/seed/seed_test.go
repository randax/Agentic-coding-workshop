package seed_test

import (
	"path/filepath"
	"testing"

	"ispcrm/internal/customer"
	"ispcrm/internal/seed"
	"ispcrm/internal/store"

	"gorm.io/gorm"
)

func freshDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "seed.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := store.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func countCustomers(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var n int64
	db.Model(&customer.Customer{}).Count(&n)
	return n
}

func TestDemoSeedsCustomersIntoEmptyDB(t *testing.T) {
	db := freshDB(t)

	if err := seed.Demo(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if n := countCustomers(t, db); n < 2 {
		t.Fatalf("after seeding got %d customers, want at least 2", n)
	}
}

func TestDemoIsIdempotent(t *testing.T) {
	db := freshDB(t)

	if err := seed.Demo(db); err != nil {
		t.Fatalf("first seed: %v", err)
	}
	first := countCustomers(t, db)

	if err := seed.Demo(db); err != nil {
		t.Fatalf("second seed: %v", err)
	}
	second := countCustomers(t, db)

	if first != second {
		t.Errorf("seeding twice changed count from %d to %d; want idempotent", first, second)
	}
}
