// Package seed populates the database with demo data. It is idempotent: it only
// inserts data when the relevant tables are empty, so it is safe to run on every
// startup. The walking skeleton seeds a couple of customers; richer demo data
// arrives in a later slice.
package seed

import (
	"fmt"
	"strings"
	"time"

	"saltcrm/internal/activity"
	"saltcrm/internal/agent"
	"saltcrm/internal/contact"
	"saltcrm/internal/customer"
	"saltcrm/internal/identity"
	"saltcrm/internal/lead"
	"saltcrm/internal/opportunity"
	"saltcrm/internal/product"
	"saltcrm/internal/subscription"
	"saltcrm/internal/supportcase"
	"saltcrm/internal/team"

	"gorm.io/gorm"
)

// Demo seeds demo data. Each entity is seeded independently and only when its
// table is empty, so Demo is safe to run on every startup.
func Demo(db *gorm.DB) error {
	// Agents (and their team) first, so customers can be assigned to the team.
	if err := seedAgents(db); err != nil {
		return err
	}
	if err := seedCustomers(db); err != nil {
		return err
	}
	if err := seedContacts(db); err != nil {
		return err
	}
	if err := seedLeads(db); err != nil {
		return err
	}
	if err := seedProducts(db); err != nil {
		return err
	}
	if err := seedOpportunities(db); err != nil {
		return err
	}
	if err := seedSubscriptions(db); err != nil {
		return err
	}
	if err := seedCases(db); err != nil {
		return err
	}
	if err := seedActivities(db); err != nil {
		return err
	}
	return seedCaseComments(db)
}

func seedActivities(db *gorm.DB) error {
	var count int64
	if err := db.Model(&activity.Activity{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var customers []customer.Customer
	if err := db.Order("id").Limit(6).Find(&customers).Error; err != nil {
		return err
	}
	types := []activity.Type{activity.TypeCall, activity.TypeMeeting, activity.TypeTask}
	now := time.Now()
	var items []activity.Activity
	for i, c := range customers {
		typ := types[i%len(types)]
		status := activity.StatusDone
		if typ == activity.TypeTask {
			status = activity.StatusOpen
		}
		items = append(items, activity.Activity{
			Type:           typ,
			Subject:        "Initial " + string(typ) + " with " + c.Name,
			Status:         status,
			ParentType:     "account",
			ParentID:       c.ID,
			OccurredAt:     now.AddDate(0, 0, -i),
			TeamID:         c.TeamID,
			AssignedUserID: c.AssignedUserID,
		})
	}
	if len(items) == 0 {
		return nil
	}
	return db.Create(&items).Error
}

func seedAgents(db *gorm.DB) error {
	var count int64
	if err := db.Model(&agent.Agent{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	// One demo team for everyone, and a shared demo password so the seeded
	// agents can log in. Roles span admin/manager/agent.
	t := team.Team{Name: "Support"}
	if err := db.Create(&t).Error; err != nil {
		return err
	}
	hash, err := identity.HashPassword("password")
	if err != nil {
		return err
	}
	agents := []agent.Agent{
		{Name: "Sam Carter", Email: "sam@isp.example", PasswordHash: hash, Role: agent.RoleAdmin, TeamID: &t.ID},
		{Name: "Robin Diaz", Email: "robin@isp.example", PasswordHash: hash, Role: agent.RoleManager, TeamID: &t.ID},
		{Name: "Lee Nakamura", Email: "lee@isp.example", PasswordHash: hash, Role: agent.RoleAgent, TeamID: &t.ID},
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

	// Assign all demo accounts to the seeded team and round-robin an owner among
	// the agents, so logged-in demo users see them (own-or-team visibility).
	var supportTeam team.Team
	db.First(&supportTeam)
	var agents []agent.Agent
	db.Order("id").Find(&agents)

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
		if supportTeam.ID != 0 {
			customers[i].TeamID = &supportTeam.ID
		}
		if len(agents) > 0 {
			customers[i].AssignedUserID = &agents[i%len(agents)].ID
		}
	}
	return db.Create(&customers).Error
}

func seedContacts(db *gorm.DB) error {
	var count int64
	if err := db.Model(&contact.Contact{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var customers []customer.Customer
	if err := db.Order("id").Find(&customers).Error; err != nil {
		return err
	}
	titles := []string{"Billing contact", "Technical contact"}
	var contacts []contact.Contact
	for _, c := range customers {
		// One primary contact per account, inheriting its team/owner so it's
		// visible to the same users.
		contacts = append(contacts, contact.Contact{
			Name:           c.Name,
			Email:          c.Email,
			Phone:          c.Phone,
			Title:          titles[int(c.ID)%len(titles)],
			AccountID:      c.ID,
			TeamID:         c.TeamID,
			AssignedUserID: c.AssignedUserID,
		})
	}
	if len(contacts) == 0 {
		return nil
	}
	return db.Create(&contacts).Error
}

func seedLeads(db *gorm.DB) error {
	var count int64
	if err := db.Model(&lead.Lead{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var supportTeam team.Team
	db.First(&supportTeam)
	var teamID *uint
	if supportTeam.ID != 0 {
		teamID = &supportTeam.ID
	}

	leads := []lead.Lead{
		{Name: "Priya Patel", Company: "Fjord Logistics", Email: "priya@fjordlog.example", Phone: "+47 901 11 222", Status: lead.StatusNew, TeamID: teamID},
		{Name: "Marco Rossi", Company: "Nordlys Media", Email: "marco@nordlys.example", Phone: "+47 902 33 444", Status: lead.StatusWorking, TeamID: teamID},
		{Name: "Sofia Berg", Company: "Polar Foods", Email: "sofia@polarfoods.example", Phone: "+47 903 55 666", Status: lead.StatusQualified, TeamID: teamID},
		{Name: "Jonas Vik", Company: "Byfjord Eiendom", Email: "jonas@byfjord.example", Phone: "+47 904 77 888", Status: lead.StatusUnqualified, TeamID: teamID},
	}
	return db.Create(&leads).Error
}

func seedOpportunities(db *gorm.DB) error {
	var count int64
	if err := db.Model(&opportunity.Opportunity{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var customers []customer.Customer
	if err := db.Order("id").Limit(6).Find(&customers).Error; err != nil {
		return err
	}
	if len(customers) == 0 {
		return nil
	}

	stages := []opportunity.Stage{
		opportunity.StageProspecting, opportunity.StageQualification, opportunity.StageProposal,
		opportunity.StageNegotiation, opportunity.StageClosedWon, opportunity.StageClosedLost,
	}
	base := time.Now().AddDate(0, 1, 0)
	var opps []opportunity.Opportunity
	for i, c := range customers {
		stage := stages[i%len(stages)]
		opps = append(opps, opportunity.Opportunity{
			Name:              c.Name + " — fiber upgrade",
			AccountID:         c.ID,
			Amount:            float64((i + 1) * 12000),
			Stage:             stage,
			Probability:       stage.Probability(),
			ExpectedCloseDate: base.AddDate(0, 0, i*15),
			TeamID:            c.TeamID,
			AssignedUserID:    c.AssignedUserID,
		})
	}
	if err := db.Create(&opps).Error; err != nil {
		return err
	}

	// Add a product line item to each opportunity (snapshotting name + price).
	var products []product.Product
	db.Order("id").Find(&products)
	if len(products) == 0 {
		return nil
	}
	var items []opportunity.LineItem
	for i, o := range opps {
		p := products[i%len(products)]
		qty := (i % 3) + 1
		items = append(items, opportunity.LineItem{
			OpportunityID: o.ID,
			ProductID:     p.ID,
			ProductName:   p.Name,
			UnitPrice:     p.MonthlyPrice,
			Quantity:      qty,
			LineTotal:     p.MonthlyPrice * float64(qty),
		})
	}
	return db.Create(&items).Error
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
