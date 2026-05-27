package store_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"saltcrm/internal/contact"
	"saltcrm/internal/conversion"
	"saltcrm/internal/customer"
	"saltcrm/internal/lead"
	"saltcrm/internal/store"

	"gorm.io/gorm"
)

func freshDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := store.Migrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return db
}

func planFor(leadID uint) conversion.Plan {
	return conversion.Plan{
		LeadID:  leadID,
		Account: customer.Customer{Name: "Polar Foods", Email: "sofia@polarfoods.example", Status: customer.StatusActive},
		Contact: contact.Contact{Name: "Sofia Berg", Email: "sofia@polarfoods.example"},
	}
}

// TestPersistRejectsSecondConvertOfSameLead makes the double-convert race
// deterministic: two Persist calls for the same qualified lead (as two racing
// requests that both passed the service-level guard would do). The second must
// be rejected by the in-transaction guard, and must leave no second account or
// contact behind.
func TestPersistRejectsSecondConvertOfSameLead(t *testing.T) {
	db := freshDB(t)
	l := lead.Lead{Name: "Sofia Berg", Company: "Polar Foods", Status: lead.StatusQualified}
	db.Create(&l)
	repo := store.NewConversionRepository(db)
	ctx := context.Background()

	first, err := repo.Persist(ctx, planFor(l.ID))
	if err != nil {
		t.Fatalf("first Persist returned error: %v", err)
	}
	if first.AccountID == 0 {
		t.Fatalf("first Persist did not create an account: %+v", first)
	}

	_, err = repo.Persist(ctx, planFor(l.ID))
	if !errors.Is(err, conversion.ErrAlreadyConverted) {
		t.Fatalf("second Persist error = %v, want conversion.ErrAlreadyConverted", err)
	}

	// The rejected second conversion left nothing behind: exactly one account and
	// one contact exist, and the lead still points at the first account.
	var accounts, contacts int64
	db.Model(&customer.Customer{}).Count(&accounts)
	db.Model(&contact.Contact{}).Count(&contacts)
	if accounts != 1 || contacts != 1 {
		t.Errorf("after a rejected re-convert: %d accounts, %d contacts; want 1 and 1", accounts, contacts)
	}
	var reloaded lead.Lead
	db.First(&reloaded, l.ID)
	if reloaded.ConvertedAccountID == nil || *reloaded.ConvertedAccountID != first.AccountID {
		t.Errorf("lead.ConvertedAccountID = %v, want it still pointing at the first account %d", reloaded.ConvertedAccountID, first.AccountID)
	}
}
