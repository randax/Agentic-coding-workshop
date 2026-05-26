package seed_test

import (
	"path/filepath"
	"testing"

	"ispcrm/internal/agent"
	"ispcrm/internal/customer"
	"ispcrm/internal/product"
	"ispcrm/internal/seed"
	"ispcrm/internal/store"
	"ispcrm/internal/subscription"
	"ispcrm/internal/supportcase"

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

func TestDemoSeedsAgents(t *testing.T) {
	db := freshDB(t)

	if err := seed.Demo(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var n int64
	db.Model(&agent.Agent{}).Count(&n)
	if n < 2 {
		t.Fatalf("after seeding got %d agents, want at least 2", n)
	}
}

func TestDemoSeedsProductsCoveringEveryCategory(t *testing.T) {
	db := freshDB(t)

	if err := seed.Demo(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	for _, cat := range []product.Category{
		product.CategoryFiber, product.CategoryRouter, product.CategoryTV,
	} {
		var n int64
		db.Model(&product.Product{}).Where("category = ?", cat).Count(&n)
		if n == 0 {
			t.Errorf("expected at least one seeded product in category %q", cat)
		}
	}
}

func TestDemoSeedsSubscriptionsLinkingCustomersToProducts(t *testing.T) {
	db := freshDB(t)

	if err := seed.Demo(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var subs []subscription.Subscription
	db.Find(&subs)
	if len(subs) == 0 {
		t.Fatal("expected demo subscriptions to be seeded")
	}
	for _, s := range subs {
		if s.CustomerID == 0 || s.ProductID == 0 {
			t.Errorf("seeded subscription %d must reference a customer and product: %+v", s.ID, s)
		}
		if s.MonthlyPriceSnapshot == 0 || s.Quantity == 0 {
			t.Errorf("seeded subscription %d must have a price snapshot and quantity: %+v", s.ID, s)
		}
	}
}

func TestDemoSeedsCasesLinkedToCustomersAndAgents(t *testing.T) {
	db := freshDB(t)

	if err := seed.Demo(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var cases []supportcase.Case
	db.Find(&cases)
	if len(cases) == 0 {
		t.Fatal("expected demo cases to be seeded")
	}
	for _, c := range cases {
		if c.CustomerID == 0 {
			t.Errorf("seeded case %d must reference a customer: %+v", c.ID, c)
		}
		if c.Subject == "" || c.Status == "" || c.Priority == "" || c.Category == "" {
			t.Errorf("seeded case %d must have subject and taxonomy fields set: %+v", c.ID, c)
		}
	}

	var assigned int64
	db.Model(&supportcase.Case{}).Where("assigned_agent_id IS NOT NULL").Count(&assigned)
	if assigned == 0 {
		t.Error("expected at least one seeded case assigned to an agent")
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
