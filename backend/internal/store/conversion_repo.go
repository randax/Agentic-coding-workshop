package store

import (
	"context"
	"strconv"
	"strings"
	"time"

	"saltcrm/internal/conversion"
	"saltcrm/internal/customer"
	"saltcrm/internal/lead"

	"gorm.io/gorm"
)

// ConversionRepository is the GORM-backed implementation of
// conversion.Repository. It is the only place the conversion's all-or-nothing
// persistence touches the database.
type ConversionRepository struct {
	db    *gorm.DB
	leads *LeadRepository
}

// NewConversionRepository wires a repository to a GORM database handle.
func NewConversionRepository(db *gorm.DB) *ConversionRepository {
	return &ConversionRepository{db: db, leads: NewLeadRepository(db)}
}

// GetLead returns a single lead by ID. It reuses the lead repository so lead
// lookup (and its not-found translation) stays defined in exactly one place.
func (r *ConversionRepository) GetLead(ctx context.Context, id uint) (lead.Lead, error) {
	return r.leads.Get(ctx, id)
}

// Persist creates the account and contact and flips the lead inside one
// transaction, so a failure at any step rolls the whole conversion back (no
// orphan account/contact, lead unchanged). The account number is minted here —
// inside the transaction — so concurrent converts can't collide on its unique
// index.
func (r *ConversionRepository) Persist(ctx context.Context, plan conversion.Plan) (conversion.Result, error) {
	var result conversion.Result
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		seq, err := nextAccountSeq(tx)
		if err != nil {
			return err
		}
		plan.Account.AccountNumber = customer.FormatAccountNumber(seq)
		if plan.Account.CustomerSince.IsZero() {
			plan.Account.CustomerSince = time.Now()
		}
		if err := tx.Create(&plan.Account).Error; err != nil {
			return err
		}

		plan.Contact.AccountID = plan.Account.ID
		if err := tx.Create(&plan.Contact).Error; err != nil {
			return err
		}

		// Flip the lead as the authoritative eligibility gate, inside the tx: only
		// a still-qualified, still-unconverted lead is updated. The condition makes
		// double-conversion safe under concurrency — a racing convert that already
		// flipped the lead matches 0 rows here, so this whole transaction (account
		// and contact included) rolls back. A 0-row result (lead converted
		// concurrently, deleted, or no longer qualified) is reported as a conflict.
		flip := tx.Model(&lead.Lead{}).
			Where("id = ? AND converted_account_id IS NULL AND status = ?", plan.LeadID, lead.StatusQualified).
			Updates(map[string]any{
				"status":               lead.StatusConverted,
				"converted_account_id": plan.Account.ID,
			})
		if flip.Error != nil {
			return flip.Error
		}
		if flip.RowsAffected != 1 {
			return conversion.ErrAlreadyConverted
		}

		result = conversion.Result{AccountID: plan.Account.ID, ContactID: plan.Contact.ID}
		return nil
	})
	if err != nil {
		return conversion.Result{}, err
	}
	return result, nil
}

// nextAccountSeq returns the next account-number sequence: one past the highest
// existing ISP-#### suffix, or the seed base (1001) when there are none. It runs
// inside the conversion transaction against tx.
func nextAccountSeq(tx *gorm.DB) (int, error) {
	const seedBase = 1001
	var numbers []string
	if err := tx.Model(&customer.Customer{}).
		Where("account_number LIKE ?", "ISP-%").
		Pluck("account_number", &numbers).Error; err != nil {
		return 0, err
	}
	max := seedBase - 1
	for _, n := range numbers {
		suffix := strings.TrimPrefix(n, "ISP-")
		v, err := strconv.Atoi(suffix)
		if err != nil {
			continue // ignore any non-numeric account numbers
		}
		if v > max {
			max = v
		}
	}
	return max + 1, nil
}
