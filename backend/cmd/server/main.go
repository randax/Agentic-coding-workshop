// Command server is the ISP CRM API entry point. It wires the persistence,
// service, and HTTP layers together, seeds demo data, and serves the API.
package main

import (
	"cmp"
	"log"
	"net/http"
	"os"

	"ispcrm/internal/api"
	"ispcrm/internal/customer"
	"ispcrm/internal/product"
	"ispcrm/internal/seed"
	"ispcrm/internal/store"
)

func main() {
	dsn := cmp.Or(os.Getenv("ISPCRM_DB"), "ispcrm.db")
	addr := cmp.Or(os.Getenv("ISPCRM_ADDR"), ":8080")

	db, err := store.Open(dsn)
	if err != nil {
		log.Fatalf("open database %q: %v", dsn, err)
	}
	if err := store.Migrate(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}
	if err := seed.Demo(db); err != nil {
		log.Fatalf("seed database: %v", err)
	}

	customers := customer.NewService(store.NewCustomerRepository(db))
	products := product.NewService(store.NewProductRepository(db))
	router := api.NewRouter(customers, products)

	log.Printf("ISP CRM API listening on %s (db: %s)", addr, dsn)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
