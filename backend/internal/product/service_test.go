package product

import (
	"context"
	"errors"
	"testing"
)

// fakeRepo is an in-memory Repository for unit-testing the catalog service
// without a database.
type fakeRepo struct {
	items  map[uint]Product
	nextID uint
	err    error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{items: map[uint]Product{}}
}

func (f *fakeRepo) List(ctx context.Context) ([]Product, error) {
	if f.err != nil {
		return nil, f.err
	}
	out := make([]Product, 0, len(f.items))
	for _, p := range f.items {
		out = append(out, p)
	}
	return out, nil
}

func (f *fakeRepo) Get(ctx context.Context, id uint) (Product, error) {
	if f.err != nil {
		return Product{}, f.err
	}
	if p, ok := f.items[id]; ok {
		return p, nil
	}
	return Product{}, ErrNotFound
}

func (f *fakeRepo) Create(ctx context.Context, p *Product) error {
	if f.err != nil {
		return f.err
	}
	f.nextID++
	p.ID = f.nextID
	f.items[p.ID] = *p
	return nil
}

func (f *fakeRepo) Update(ctx context.Context, p *Product) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.items[p.ID]; !ok {
		return ErrNotFound
	}
	f.items[p.ID] = *p
	return nil
}

func TestCreateStoresProductAndDefaultsToAvailable(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo)

	got, err := svc.Create(context.Background(), Product{
		Name:         "Fiber 500",
		Category:     CategoryFiber,
		MonthlyPrice: 499,
	})
	if err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}
	if got.ID == 0 {
		t.Errorf("Create should assign an ID, got 0")
	}
	if !got.Available {
		t.Errorf("a newly created product should be available")
	}
	if _, err := repo.Get(context.Background(), got.ID); err != nil {
		t.Errorf("created product not persisted: %v", err)
	}
}

func TestCreateRejectsInvalidCategory(t *testing.T) {
	svc := NewService(newFakeRepo())

	_, err := svc.Create(context.Background(), Product{
		Name:         "Mystery box",
		Category:     "satellite",
		MonthlyPrice: 100,
	})
	if !errors.Is(err, ErrInvalidCategory) {
		t.Fatalf("Create error = %v, want ErrInvalidCategory", err)
	}
}

func TestCreateRejectsNegativePrice(t *testing.T) {
	svc := NewService(newFakeRepo())

	_, err := svc.Create(context.Background(), Product{
		Name:         "Fiber 500",
		Category:     CategoryFiber,
		MonthlyPrice: -1,
	})
	if !errors.Is(err, ErrInvalidPrice) {
		t.Fatalf("Create error = %v, want ErrInvalidPrice", err)
	}
}

func TestListReturnsAllProducts(t *testing.T) {
	svc := NewService(newFakeRepo())
	ctx := context.Background()
	svc.Create(ctx, Product{Name: "Fiber 500", Category: CategoryFiber, MonthlyPrice: 499})
	svc.Create(ctx, Product{Name: "Mesh Pro", Category: CategoryRouter, MonthlyPrice: 99})

	got, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List returned unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("List returned %d products, want 2", len(got))
	}
}

func TestRetireMarksProductUnavailable(t *testing.T) {
	svc := NewService(newFakeRepo())
	ctx := context.Background()
	created, _ := svc.Create(ctx, Product{Name: "Fiber 500", Category: CategoryFiber, MonthlyPrice: 499})

	if err := svc.Retire(ctx, created.ID); err != nil {
		t.Fatalf("Retire returned unexpected error: %v", err)
	}

	got, _ := svc.Get(ctx, created.ID)
	if got.Available {
		t.Errorf("retired product should not be available")
	}
}

func TestRetireUnknownProductReturnsNotFound(t *testing.T) {
	svc := NewService(newFakeRepo())

	err := svc.Retire(context.Background(), 9999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Retire error = %v, want ErrNotFound", err)
	}
}

func TestUpdateEditsProductFields(t *testing.T) {
	svc := NewService(newFakeRepo())
	ctx := context.Background()
	created, _ := svc.Create(ctx, Product{Name: "Fiber 500", Category: CategoryFiber, MonthlyPrice: 499})

	created.Name = "Fiber 1000"
	created.MonthlyPrice = 699
	updated, err := svc.Update(ctx, created)
	if err != nil {
		t.Fatalf("Update returned unexpected error: %v", err)
	}
	if updated.Name != "Fiber 1000" || updated.MonthlyPrice != 699 {
		t.Errorf("Update returned %+v, want name=Fiber 1000 price=699", updated)
	}

	got, _ := svc.Get(ctx, created.ID)
	if got.Name != "Fiber 1000" || got.MonthlyPrice != 699 {
		t.Errorf("persisted product = %+v, want name=Fiber 1000 price=699", got)
	}
}
