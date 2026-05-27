package store

import (
	"context"
	"errors"

	"saltcrm/internal/conversion"
	"saltcrm/internal/lead"

	"gorm.io/gorm"
)

// ConversionRepository persists lead conversions atomically.
type ConversionRepository struct {
	db *gorm.DB
}

// NewConversionRepository wires a repository to a GORM database handle.
func NewConversionRepository(db *gorm.DB) *ConversionRepository {
	return &ConversionRepository{db: db}
}

// GetLead loads a lead by id, translating GORM's not-found error.
func (r *ConversionRepository) GetLead(ctx context.Context, id uint) (lead.Lead, error) {
	var l lead.Lead
	if err := r.db.WithContext(ctx).First(&l, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return lead.Lead{}, lead.ErrNotFound
		}
		return lead.Lead{}, err
	}
	return l, nil
}

// Persist creates the account, then the contact + opportunity referencing it,
// and marks the lead converted (linking the created records) — all in one
// transaction, so any failure rolls the whole conversion back.
func (r *ConversionRepository) Persist(ctx context.Context, plan conversion.Plan) (conversion.Result, error) {
	var res conversion.Result
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		account := plan.Account
		if err := tx.Create(&account).Error; err != nil {
			return err
		}
		contact := plan.Contact
		contact.AccountID = account.ID
		if err := tx.Create(&contact).Error; err != nil {
			return err
		}
		opp := plan.Opportunity
		opp.AccountID = account.ID
		if err := tx.Create(&opp).Error; err != nil {
			return err
		}
		if err := tx.Model(&lead.Lead{}).Where("id = ?", plan.Lead.ID).Updates(map[string]any{
			"status":                   lead.StatusConverted,
			"converted_account_id":     account.ID,
			"converted_contact_id":     contact.ID,
			"converted_opportunity_id": opp.ID,
		}).Error; err != nil {
			return err
		}
		res = conversion.Result{AccountID: account.ID, ContactID: contact.ID, OpportunityID: opp.ID}
		return nil
	})
	if err != nil {
		return conversion.Result{}, err
	}
	return res, nil
}
