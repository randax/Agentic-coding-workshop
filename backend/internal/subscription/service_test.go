package subscription

import (
	"context"
	"errors"
	"testing"

	"saltcrm/internal/product"
)

// fakeRepo is an in-memory Repository for unit-testing the service.
type fakeRepo struct {
	items  map[uint]Subscription
	nextID uint
	err    error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{items: map[uint]Subscription{}}
}

func (f *fakeRepo) ListByCustomer(ctx context.Context, customerID uint) ([]Subscription, error) {
	if f.err != nil {
		return nil, f.err
	}
	var out []Subscription
	for _, s := range f.items {
		if s.CustomerID == customerID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (f *fakeRepo) Create(ctx context.Context, s *Subscription) error {
	if f.err != nil {
		return f.err
	}
	f.nextID++
	s.ID = f.nextID
	f.items[s.ID] = *s
	return nil
}

func (f *fakeRepo) Get(ctx context.Context, id uint) (Subscription, error) {
	if f.err != nil {
		return Subscription{}, f.err
	}
	if s, ok := f.items[id]; ok {
		return s, nil
	}
	return Subscription{}, ErrNotFound
}

func (f *fakeRepo) Update(ctx context.Context, s *Subscription) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.items[s.ID]; !ok {
		return ErrNotFound
	}
	f.items[s.ID] = *s
	return nil
}

// fakeProducts is an in-memory ProductReader for unit-testing the service
// without the catalog's database.
type fakeProducts struct {
	items map[uint]product.Product
	err   error
}

func (f *fakeProducts) Get(ctx context.Context, id uint) (product.Product, error) {
	if f.err != nil {
		return product.Product{}, f.err
	}
	if p, ok := f.items[id]; ok {
		return p, nil
	}
	return product.Product{}, product.ErrNotFound
}

func newFakeProducts(products ...product.Product) *fakeProducts {
	items := map[uint]product.Product{}
	for _, p := range products {
		items[p.ID] = p
	}
	return &fakeProducts{items: items}
}

func TestListForCustomerReturnsThatCustomersSubscriptions(t *testing.T) {
	repo := newFakeRepo()
	repo.items = map[uint]Subscription{
		1: {ID: 1, CustomerID: 7, Status: StatusActive, Quantity: 1},
		2: {ID: 2, CustomerID: 7, Status: StatusCancelled, Quantity: 2},
		3: {ID: 3, CustomerID: 9, Status: StatusActive, Quantity: 1},
	}
	svc := NewService(repo, newFakeProducts())

	got, err := svc.ListForCustomer(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListForCustomer returned unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d subscriptions, want 2", len(got))
	}
}

func TestSnapshotPriceIsStableAcrossLaterCatalogChanges(t *testing.T) {
	repo := newFakeRepo()
	products := newFakeProducts(product.Product{
		ID: 5, Name: "Fiber 500", Category: product.CategoryFiber, MonthlyPrice: 499, Available: true,
	})
	svc := NewService(repo, products)

	sub, err := svc.Assign(context.Background(), 7, 5, 1)
	if err != nil {
		t.Fatalf("Assign returned unexpected error: %v", err)
	}

	// The catalog price later rises; the subscription's snapshot must not move.
	products.items[5] = product.Product{
		ID: 5, Name: "Fiber 500", Category: product.CategoryFiber, MonthlyPrice: 999, Available: true,
	}

	stored, err := repo.Get(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("Get returned unexpected error: %v", err)
	}
	if stored.MonthlyPriceSnapshot != 499 {
		t.Errorf("snapshot price = %v after catalog change, want 499 (stable)", stored.MonthlyPriceSnapshot)
	}
}

func TestAssignRejectsRetiredProduct(t *testing.T) {
	repo := newFakeRepo()
	products := newFakeProducts(product.Product{
		ID: 5, Name: "Fiber 500", Category: product.CategoryFiber, MonthlyPrice: 499, Available: false,
	})
	svc := NewService(repo, products)

	_, err := svc.Assign(context.Background(), 7, 5, 1)
	if !errors.Is(err, ErrProductRetired) {
		t.Fatalf("Assign error = %v, want ErrProductRetired", err)
	}
	if len(repo.items) != 0 {
		t.Errorf("a rejected assignment should not persist a subscription, got %d", len(repo.items))
	}
}

func TestCancelSetsStatusCancelledAndEndDate(t *testing.T) {
	repo := newFakeRepo()
	products := newFakeProducts(product.Product{
		ID: 5, Name: "Fiber 500", Category: product.CategoryFiber, MonthlyPrice: 499, Available: true,
	})
	svc := NewService(repo, products)
	sub, _ := svc.Assign(context.Background(), 7, 5, 1)

	got, err := svc.Cancel(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("Cancel returned unexpected error: %v", err)
	}
	if got.Status != StatusCancelled {
		t.Errorf("status = %q, want cancelled", got.Status)
	}
	if got.EndDate == nil || got.EndDate.IsZero() {
		t.Errorf("cancelled subscription should have an end date, got %v", got.EndDate)
	}

	stored, _ := repo.Get(context.Background(), sub.ID)
	if stored.Status != StatusCancelled || stored.EndDate == nil {
		t.Errorf("cancellation not persisted: %+v", stored)
	}
}

func TestCancelUnknownSubscriptionReturnsNotFound(t *testing.T) {
	svc := NewService(newFakeRepo(), newFakeProducts())

	_, err := svc.Cancel(context.Background(), 9999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Cancel error = %v, want ErrNotFound", err)
	}
}

func TestCancelIsIdempotentOnAlreadyCancelled(t *testing.T) {
	repo := newFakeRepo()
	products := newFakeProducts(product.Product{
		ID: 5, Name: "Fiber 500", Category: product.CategoryFiber, MonthlyPrice: 499, Available: true,
	})
	svc := NewService(repo, products)
	sub, _ := svc.Assign(context.Background(), 7, 5, 1)

	first, err := svc.Cancel(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("first Cancel returned unexpected error: %v", err)
	}

	second, err := svc.Cancel(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("re-cancelling should be a no-op, got error: %v", err)
	}
	if second.Status != StatusCancelled {
		t.Errorf("status = %q, want cancelled", second.Status)
	}
	if first.EndDate == nil || second.EndDate == nil || !second.EndDate.Equal(*first.EndDate) {
		t.Errorf("re-cancelling moved the end date: first=%v second=%v", first.EndDate, second.EndDate)
	}
}

func TestAssignRejectsQuantityBelowOne(t *testing.T) {
	repo := newFakeRepo()
	products := newFakeProducts(product.Product{
		ID: 5, Name: "Fiber 500", Category: product.CategoryFiber, MonthlyPrice: 499, Available: true,
	})
	svc := NewService(repo, products)

	for _, q := range []int{0, -1} {
		_, err := svc.Assign(context.Background(), 7, 5, q)
		if !errors.Is(err, ErrInvalidQuantity) {
			t.Errorf("Assign(quantity=%d) error = %v, want ErrInvalidQuantity", q, err)
		}
	}
	if len(repo.items) != 0 {
		t.Errorf("a rejected assignment should not persist a subscription, got %d", len(repo.items))
	}
}

func TestAssignCapturesCurrentCatalogPriceAsSnapshot(t *testing.T) {
	repo := newFakeRepo()
	products := newFakeProducts(product.Product{
		ID: 5, Name: "Fiber 500", Category: product.CategoryFiber, MonthlyPrice: 499, Available: true,
	})
	svc := NewService(repo, products)

	got, err := svc.Assign(context.Background(), 7, 5, 2)
	if err != nil {
		t.Fatalf("Assign returned unexpected error: %v", err)
	}
	if got.ID == 0 {
		t.Errorf("Assign should persist and assign an ID, got 0")
	}
	if got.MonthlyPriceSnapshot != 499 {
		t.Errorf("snapshot price = %v, want 499 (the current catalog price)", got.MonthlyPriceSnapshot)
	}
	if got.Status != StatusActive {
		t.Errorf("new subscription status = %q, want active", got.Status)
	}
	if got.CustomerID != 7 || got.ProductID != 5 {
		t.Errorf("got customer=%d product=%d, want customer=7 product=5", got.CustomerID, got.ProductID)
	}
	if got.Quantity != 2 {
		t.Errorf("quantity = %d, want 2", got.Quantity)
	}
	if got.StartDate.IsZero() {
		t.Errorf("new subscription should have a start date")
	}
}
