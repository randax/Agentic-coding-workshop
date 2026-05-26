// Package seed populates the database with demo data. It is idempotent: it only
// inserts data when the relevant tables are empty, so it is safe to run on every
// startup. The walking skeleton seeds a couple of customers; richer demo data
// arrives in a later slice.
package seed

import (
	"fmt"
	"strings"
	"time"

	"ispcrm/internal/agent"
	"ispcrm/internal/customer"
	"ispcrm/internal/product"
	"ispcrm/internal/subscription"
	"ispcrm/internal/supportcase"

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
	if err := seedSubscriptions(db); err != nil {
		return err
	}
	if err := seedAgents(db); err != nil {
		return err
	}
	if err := seedCases(db); err != nil {
		return err
	}
	return seedCaseComments(db)
}

func seedAgents(db *gorm.DB) error {
	var count int64
	if err := db.Model(&agent.Agent{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	agents := []agent.Agent{
		{Name: "Sam Carter", Email: "sam@isp.example"},
		{Name: "Robin Diaz", Email: "robin@isp.example"},
		{Name: "Lee Nakamura", Email: "lee@isp.example"},
	}
	return db.Create(&agents).Error
}

func seedCases(db *gorm.DB) error {
	var count int64
	if err := db.Model(&supportcase.Case{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var customers []customer.Customer
	if err := db.Order("id").Find(&customers).Error; err != nil {
		return err
	}
	var agents []agent.Agent
	if err := db.Order("id").Find(&agents).Error; err != nil {
		return err
	}
	if len(customers) == 0 || len(agents) == 0 {
		return nil
	}

	agentID := func(i int) *uint {
		id := agents[i%len(agents)].ID
		return &id
	}

	type spec struct {
		custIdx              int
		subject, description string
		category             supportcase.Category
		priority             supportcase.Priority
		status               supportcase.Status
		agentIdx             int
	}
	specs := []spec{
		{0, "No internet since this morning", "Connection dropped around 08:00 and has not come back.", supportcase.CategoryConnectivity, supportcase.PriorityHigh, supportcase.StatusOpen, 0},
		{0, "Charged twice for August", "My August invoice shows two fiber charges.", supportcase.CategoryBilling, supportcase.PriorityMedium, supportcase.StatusInProgress, 1},
		{1, "Mesh router keeps rebooting", "The MeshPro reboots every few hours.", supportcase.CategoryHardware, supportcase.PriorityUrgent, supportcase.StatusOpen, 2},
		{1, "How do I add an extra TV box?", "Want a second TV package decoder.", supportcase.CategoryTV, supportcase.PriorityLow, supportcase.StatusResolved, 0},
		{2, "Requesting a speed upgrade", "Interested in moving from Fiber 500 to Fiber 1000.", supportcase.CategoryGeneral, supportcase.PriorityLow, supportcase.StatusClosed, 1},
		{3, "Intermittent packet loss in the evenings", "Latency spikes between 20:00 and 23:00.", supportcase.CategoryConnectivity, supportcase.PriorityHigh, supportcase.StatusInProgress, 2},
		{4, "Refund for service downtime", "Three days without service last week.", supportcase.CategoryBilling, supportcase.PriorityMedium, supportcase.StatusResolved, 0},
		{5, "Router LED blinking red", "The router shows a steady red light.", supportcase.CategoryHardware, supportcase.PriorityHigh, supportcase.StatusOpen, 1},
		{6, "TV channels missing after update", "Several channels disappeared from the Premium package.", supportcase.CategoryTV, supportcase.PriorityMedium, supportcase.StatusInProgress, 2},
		{7, "Question about contract terms", "When does my current term end?", supportcase.CategoryGeneral, supportcase.PriorityLow, supportcase.StatusClosed, 0},
	}

	cases := make([]supportcase.Case, 0, len(specs))
	for _, s := range specs {
		if s.custIdx >= len(customers) {
			continue
		}
		cases = append(cases, supportcase.Case{
			CustomerID:      customers[s.custIdx].ID,
			Subject:         s.subject,
			Description:     s.description,
			Category:        s.category,
			Priority:        s.priority,
			Status:          s.status,
			AssignedAgentID: agentID(s.agentIdx),
		})
	}

	// Omit the AssignedAgent association so GORM persists only the FK, never a phantom agent.
	return db.Omit("AssignedAgent").Create(&cases).Error
}

func seedCaseComments(db *gorm.DB) error {
	var count int64
	if err := db.Model(&supportcase.CaseComment{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var cases []supportcase.Case
	if err := db.Order("id").Find(&cases).Error; err != nil {
		return err
	}
	var agents []agent.Agent
	if err := db.Order("id").Find(&agents).Error; err != nil {
		return err
	}
	if len(cases) == 0 || len(agents) == 0 {
		return nil
	}

	now := time.Now()
	var comments []supportcase.CaseComment
	// Give the first couple of cases a short back-and-forth timeline.
	scripts := [][]string{
		{
			"Thanks for reporting — I can see repeated dropouts on your line.",
			"Dispatched a line check; please power-cycle the ONT in the meantime.",
			"Line looks stable again after the check. Let us know if it recurs.",
		},
		{
			"Looked at the August invoice; confirming the duplicate fiber charge.",
			"Refund for the duplicate line has been requested.",
		},
		{
			"Fiber 1000 is available at your address — happy to switch you over.",
			"Upgrade scheduled and confirmed. Closing this out.",
		},
		{
			"We're seeing elevated latency on the regional segment in the evenings.",
			"Network team is rebalancing traffic; this should improve within 48h.",
			"Monitoring shows evening latency back to normal.",
		},
		{
			"Confirmed three days of downtime on your account last week.",
			"A credit for the downtime has been applied to your next invoice.",
		},
	}
	for ci, script := range scripts {
		if ci >= len(cases) {
			break
		}
		caseID := cases[ci].ID
		for i, body := range script {
			author := agents[i%len(agents)].ID
			comments = append(comments, supportcase.CaseComment{
				CaseID:        caseID,
				Body:          body,
				AuthorAgentID: &author,
				// Stagger timestamps so the timeline has a clear chronological order.
				CreatedAt: now.Add(time.Duration(ci*10+i) * time.Hour),
			})
		}
	}

	if len(comments) == 0 {
		return nil
	}
	// Omit the AuthorAgent association so GORM persists only the FK, never a phantom agent.
	return db.Omit("AuthorAgent").Create(&comments).Error
}

func seedCustomers(db *gorm.DB) error {
	var count int64
	if err := db.Model(&customer.Customer{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	type spec struct {
		name, address string
		status        customer.Status
	}
	specs := []spec{
		{"Ada Lovelace", "Storgata 1, 0155 Oslo", customer.StatusActive},
		{"Alan Turing", "Kongens gate 7, 0153 Oslo", customer.StatusActive},
		{"Grace Hopper", "Havnegata 12, 7010 Trondheim", customer.StatusSuspended},
		{"Katherine Johnson", "Bryggen 5, 5003 Bergen", customer.StatusActive},
		{"Edsger Dijkstra", "Nygårdsgaten 9, 5015 Bergen", customer.StatusActive},
		{"Margaret Hamilton", "Torggata 20, 0183 Oslo", customer.StatusActive},
		{"Linus Torvalds", "Dronningens gate 3, 7011 Trondheim", customer.StatusActive},
		{"Barbara Liskov", "Kirkegata 14, 4006 Stavanger", customer.StatusActive},
		{"Dennis Ritchie", "Olav Tryggvasons gate 2, 7011 Trondheim", customer.StatusSuspended},
		{"Donald Knuth", "Karl Johans gate 25, 0159 Oslo", customer.StatusActive},
		{"Ken Thompson", "Strandgata 8, 9008 Tromsø", customer.StatusActive},
		{"Radia Perlman", "Storgata 44, 9008 Tromsø", customer.StatusActive},
		{"Tim Berners-Lee", "Skippergata 11, 0152 Oslo", customer.StatusActive},
		{"Vint Cerf", "Prinsens gate 6, 4005 Stavanger", customer.StatusActive},
	}

	base := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	customers := make([]customer.Customer, len(specs))
	for i, s := range specs {
		firstName := strings.ToLower(strings.Fields(s.name)[0])
		customers[i] = customer.Customer{
			Name:           s.name,
			Email:          firstName + "@example.com",
			Phone:          fmt.Sprintf("+47 900 %02d %03d", i+1, (i*37)%1000),
			ServiceAddress: s.address,
			AccountNumber:  fmt.Sprintf("ISP-%04d", 1001+i),
			CustomerSince:  base.AddDate(0, i*2, i),
			Status:         s.status,
		}
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

	fiber, hasFiber := byCat[product.CategoryFiber]
	router, hasRouter := byCat[product.CategoryRouter]
	tv, hasTV := byCat[product.CategoryTV]

	start := time.Now().AddDate(-1, 0, 0)
	ended := time.Now().AddDate(0, -2, 0)

	// Rotate the fiber subscription's status so active/pending/cancelled all
	// appear across the customer base.
	fiberStatuses := []subscription.Status{
		subscription.StatusActive, subscription.StatusActive, subscription.StatusPending,
		subscription.StatusActive, subscription.StatusCancelled,
	}

	var subs []subscription.Subscription
	for i, c := range customers {
		if hasFiber {
			st := fiberStatuses[i%len(fiberStatuses)]
			sub := subscription.Subscription{
				CustomerID: c.ID, ProductID: fiber.ID, Status: st,
				StartDate: start, MonthlyPriceSnapshot: fiber.MonthlyPrice, Quantity: 1,
			}
			if st == subscription.StatusCancelled {
				sub.EndDate = &ended
			}
			subs = append(subs, sub)
		}
		// Every third customer also rents mesh routers, in varied quantities.
		if hasRouter && i%3 == 0 {
			subs = append(subs, subscription.Subscription{
				CustomerID: c.ID, ProductID: router.ID, Status: subscription.StatusActive,
				StartDate: start, MonthlyPriceSnapshot: router.MonthlyPrice, Quantity: 2 + i%2,
			})
		}
		// Every other customer has a TV package.
		if hasTV && i%2 == 0 {
			subs = append(subs, subscription.Subscription{
				CustomerID: c.ID, ProductID: tv.ID, Status: subscription.StatusActive,
				StartDate: start, MonthlyPriceSnapshot: tv.MonthlyPrice, Quantity: 1,
			})
		}
	}

	if len(subs) == 0 {
		return nil
	}
	// Omit the Product association so GORM persists only the FK, never a phantom product.
	return db.Omit("Product").Create(&subs).Error
}
