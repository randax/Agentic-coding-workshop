// Package seed populates the database with demo data. It is idempotent: it only
// inserts data when the relevant tables are empty, so it is safe to run on every
// startup. The walking skeleton seeds a couple of customers; richer demo data
// arrives in a later slice.
package seed

import (
	"time"

	"ispcrm/internal/customer"

	"gorm.io/gorm"
)

// Demo seeds demo customers if none exist yet.
func Demo(db *gorm.DB) error {
	var count int64
	if err := db.Model(&customer.Customer{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	customers := []customer.Customer{
		{
			Name:           "Ada Lovelace",
			Email:          "ada@example.com",
			Phone:          "+47 900 00 001",
			ServiceAddress: "Storgata 1, 0155 Oslo",
			AccountNumber:  "ISP-1001",
			CustomerSince:  time.Date(2021, time.March, 14, 0, 0, 0, 0, time.UTC),
			Status:         customer.StatusActive,
		},
		{
			Name:           "Alan Turing",
			Email:          "alan@example.com",
			Phone:          "+47 900 00 002",
			ServiceAddress: "Kongens gate 7, 0153 Oslo",
			AccountNumber:  "ISP-1002",
			CustomerSince:  time.Date(2022, time.June, 23, 0, 0, 0, 0, time.UTC),
			Status:         customer.StatusActive,
		},
		{
			Name:           "Grace Hopper",
			Email:          "grace@example.com",
			Phone:          "+47 900 00 003",
			ServiceAddress: "Havnegata 12, 7010 Trondheim",
			AccountNumber:  "ISP-1003",
			CustomerSince:  time.Date(2020, time.December, 9, 0, 0, 0, 0, time.UTC),
			Status:         customer.StatusSuspended,
		},
	}
	return db.Create(&customers).Error
}
