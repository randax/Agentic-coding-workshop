// Command server is the SaltCRM API entry point. It wires the persistence,
// service, and HTTP layers together, seeds demo data, and serves the API.
package main

import (
	"cmp"
	"log"
	"net/http"
	"os"

	"saltcrm/internal/agent"
	"saltcrm/internal/api"
	"saltcrm/internal/contact"
	"saltcrm/internal/customer"
	"saltcrm/internal/identity"
	"saltcrm/internal/product"
	"saltcrm/internal/seed"
	"saltcrm/internal/store"
	"saltcrm/internal/subscription"
	"saltcrm/internal/supportcase"
)

func main() {
	dsn := cmp.Or(os.Getenv("SALTCRM_DB"), "saltcrm.db")
	addr := cmp.Or(os.Getenv("SALTCRM_ADDR"), ":8080")

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
	subscriptions := subscription.NewService(store.NewSubscriptionRepository(db), products)
	agentRepo := store.NewAgentRepository(db)
	agents := agent.NewService(agentRepo)
	cases := supportcase.NewService(store.NewCaseRepository(db))
	identitySvc := identity.NewService(agentRepo, store.NewSessionRepository(db))
	contacts := contact.NewService(store.NewContactRepository(db))
	router := api.NewRouter(customers, products, subscriptions, agents, cases, identitySvc, contacts)

	log.Printf("SaltCRM API listening on %s (db: %s)", addr, dsn)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
