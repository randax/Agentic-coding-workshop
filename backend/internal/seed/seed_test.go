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

	if n := countCustomers(t, db); n < 10 || n > 20 {
		t.Fatalf("after seeding got %d customers, want 10-20", n)
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

func TestDemoSeedsVariedSubscriptions(t *testing.T) {
	db := freshDB(t)

	if err := seed.Demo(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var subs []subscription.Subscription
	db.Find(&subs)

	seen := map[subscription.Status]bool{}
	maxQty := 0
	for _, s := range subs {
		seen[s.Status] = true
		if s.Quantity > maxQty {
			maxQty = s.Quantity
		}
	}
	for _, st := range []subscription.Status{
		subscription.StatusActive, subscription.StatusPending, subscription.StatusCancelled,
	} {
		if !seen[st] {
			t.Errorf("expected at least one %q subscription", st)
		}
	}
	if maxQty < 2 {
		t.Errorf("expected at least one subscription with quantity > 1, max was %d", maxQty)
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

func TestDemoSeedsCasesAcrossStatusesCategoriesPriorities(t *testing.T) {
	db := freshDB(t)

	if err := seed.Demo(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var cases []supportcase.Case
	db.Find(&cases)

	statuses := map[supportcase.Status]bool{}
	cats := map[supportcase.Category]bool{}
	pris := map[supportcase.Priority]bool{}
	for _, c := range cases {
		statuses[c.Status] = true
		cats[c.Category] = true
		pris[c.Priority] = true
	}
	for _, st := range []supportcase.Status{
		supportcase.StatusOpen, supportcase.StatusInProgress,
		supportcase.StatusResolved, supportcase.StatusClosed,
	} {
		if !statuses[st] {
			t.Errorf("expected at least one case with status %q", st)
		}
	}
	if len(cats) < 3 {
		t.Errorf("expected cases across at least 3 categories, got %d", len(cats))
	}
	if len(pris) < 3 {
		t.Errorf("expected cases across at least 3 priorities, got %d", len(pris))
	}
}

func TestDemoSeedsCaseCommentTimelines(t *testing.T) {
	db := freshDB(t)

	if err := seed.Demo(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var comments []supportcase.CaseComment
	db.Find(&comments)
	if len(comments) == 0 {
		t.Fatal("expected demo case comments to be seeded")
	}
	for _, cm := range comments {
		if cm.CaseID == 0 || cm.Body == "" {
			t.Errorf("seeded comment %d must reference a case and have a body: %+v", cm.ID, cm)
		}
	}

	var authored int64
	db.Model(&supportcase.CaseComment{}).Where("author_agent_id IS NOT NULL").Count(&authored)
	if authored == 0 {
		t.Error("expected at least one seeded comment authored by an agent")
	}
}

func TestDemoSeedsMultiCommentTimelines(t *testing.T) {
	db := freshDB(t)

	if err := seed.Demo(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var comments []supportcase.CaseComment
	db.Find(&comments)

	perCase := map[uint]int{}
	for _, cm := range comments {
		perCase[cm.CaseID]++
	}
	multi := 0
	for _, n := range perCase {
		if n >= 2 {
			multi++
		}
	}
	if multi < 3 {
		t.Errorf("expected at least 3 cases with multi-comment timelines, got %d", multi)
	}
}

func TestDemoIsIdempotent(t *testing.T) {
	db := freshDB(t)

	if err := seed.Demo(db); err != nil {
		t.Fatalf("first seed: %v", err)
	}

	type counts struct{ customers, subs, cases, comments, agents int64 }
	snapshot := func() counts {
		var c counts
		db.Model(&customer.Customer{}).Count(&c.customers)
		db.Model(&subscription.Subscription{}).Count(&c.subs)
		db.Model(&supportcase.Case{}).Count(&c.cases)
		db.Model(&supportcase.CaseComment{}).Count(&c.comments)
		db.Model(&agent.Agent{}).Count(&c.agents)
		return c
	}
	first := snapshot()

	if err := seed.Demo(db); err != nil {
		t.Fatalf("second seed: %v", err)
	}
	second := snapshot()

	if first != second {
		t.Errorf("seeding twice changed counts from %+v to %+v; want idempotent", first, second)
	}
}
