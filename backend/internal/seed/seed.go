// Package seed populates the database with demo data. It is idempotent: it only
// inserts data when the relevant tables are empty, so it is safe to run on every
// startup. The walking skeleton seeds a couple of customers; richer demo data
// arrives in a later slice.
package seed

import (
	"time"

	"ispcrm/internal/customer"
	"ispcrm/internal/product"
	"ispcrm/internal/subscription"

	"gorm.io/gorm"
)

// Demo seeds demo data. Each entity is seeded independently and only when its
// table is empty, so Demo is safe to run on every startup.
func Demo(db *gorm.DB) error {
	if err := seedCustomers(db); err != nil {
		return err
	}
	if err := seedProducts(db); err != nil {
		return err
	}
	return seedSubscriptions(db)
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

func seedSubscriptions(db *gorm.DB) error {
	var count int64
	if err := db.Model(&subscription.Subscription{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var customers []customer.Customer
	if err := db.Order("id").Find(&customers).Error; err != nil {
		return err
	}
	var products []product.Product
	if err := db.Order("id").Find(&products).Error; err != nil {
		return err
	}
	if len(customers) == 0 || len(products) == 0 {
		return nil
	}

	// First product seen per category, for convenient linking.
	byCat := map[product.Category]product.Product{}
	for _, p := range products {
		if _, ok := byCat[p.Category]; !ok {
			byCat[p.Category] = p
		}
	}

	start := time.Now().AddDate(-1, 0, 0)
	ended := time.Now().AddDate(0, -2, 0)

	var subs []subscription.Subscription
	if fiber, ok := byCat[product.CategoryFiber]; ok {
		subs = append(subs, subscription.Subscription{
			CustomerID: customers[0].ID, ProductID: fiber.ID, Status: subscription.StatusActive,
			StartDate: start, MonthlyPriceSnapshot: fiber.MonthlyPrice, Quantity: 1,
		})
	}
	if router, ok := byCat[product.CategoryRouter]; ok {
		subs = append(subs, subscription.Subscription{
			CustomerID: customers[0].ID, ProductID: router.ID, Status: subscription.StatusActive,
			StartDate: start, MonthlyPriceSnapshot: router.MonthlyPrice, Quantity: 2,
		})
	}
	if len(customers) > 1 {
		if tv, ok := byCat[product.CategoryTV]; ok {
			subs = append(subs, subscription.Subscription{
				CustomerID: customers[1].ID, ProductID: tv.ID, Status: subscription.StatusActive,
				StartDate: start, MonthlyPriceSnapshot: tv.MonthlyPrice, Quantity: 1,
			})
		}
		if fiber, ok := byCat[product.CategoryFiber]; ok {
			subs = append(subs, subscription.Subscription{
				CustomerID: customers[1].ID, ProductID: fiber.ID, Status: subscription.StatusCancelled,
				StartDate: start, EndDate: &ended, MonthlyPriceSnapshot: fiber.MonthlyPrice, Quantity: 1,
			})
		}
	}

	if len(subs) == 0 {
		return nil
	}
	// Omit the Product association so GORM persists only the FK, never a phantom product.
	return db.Omit("Product").Create(&subs).Error
}
