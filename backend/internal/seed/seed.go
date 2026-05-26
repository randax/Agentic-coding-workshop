// Package seed populates the database with demo data. It is idempotent: it only
// inserts data when the relevant tables are empty, so it is safe to run on every
// startup. The walking skeleton seeds a couple of customers; richer demo data
// arrives in a later slice.
package seed

import (
	"time"

	"ispcrm/internal/customer"
	"ispcrm/internal/product"

	"gorm.io/gorm"
)

// Demo seeds demo data. Each entity is seeded independently and only when its
// table is empty, so Demo is safe to run on every startup.
func Demo(db *gorm.DB) error {
	if err := seedCustomers(db); err != nil {
		return err
	}
	return seedProducts(db)
}

func seedCustomers(db *gorm.DB) error {
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

func seedProducts(db *gorm.DB) error {
	var count int64
	if err := db.Model(&product.Product{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	speed500, speed1000 := 500, 1000
	meshPro, meshMini := "MeshPro X6", "MeshMini M3"
	tvBasic, tvPremium := "Basic", "Premium"

	products := []product.Product{
		{Name: "Fiber 500", Category: product.CategoryFiber, MonthlyPrice: 499, Available: true, SpeedMbps: &speed500},
		{Name: "Fiber 1000", Category: product.CategoryFiber, MonthlyPrice: 699, Available: true, SpeedMbps: &speed1000},
		{Name: "Mesh Router Pro", Category: product.CategoryRouter, MonthlyPrice: 99, Available: true, RouterModel: &meshPro},
		{Name: "Mesh Router Mini", Category: product.CategoryRouter, MonthlyPrice: 59, Available: true, RouterModel: &meshMini},
		{Name: "TV Basic", Category: product.CategoryTV, MonthlyPrice: 199, Available: true, TvPackageTier: &tvBasic},
		{Name: "TV Premium", Category: product.CategoryTV, MonthlyPrice: 399, Available: true, TvPackageTier: &tvPremium},
	}
	return db.Create(&products).Error
}
