package customer

import (
	"context"
	"errors"
	"testing"
)

// fakeRepo is an in-memory Repository for unit-testing the service in isolation,
// without a database. This is the convention for service-layer unit tests.
type fakeRepo struct {
	customers []Customer
	err       error
}

func (f *fakeRepo) List(ctx context.Context) ([]Customer, error) {
	return f.customers, f.err
}

func (f *fakeRepo) Get(ctx context.Context, id uint) (Customer, error) {
	if f.err != nil {
		return Customer{}, f.err
	}
	for _, c := range f.customers {
		if c.ID == id {
			return c, nil
		}
	}
	return Customer{}, ErrNotFound
}

func TestServiceListReturnsCustomersFromRepository(t *testing.T) {
	want := []Customer{
		{ID: 1, Name: "Ada Lovelace", AccountNumber: "ACME-001"},
		{ID: 2, Name: "Alan Turing", AccountNumber: "ACME-002"},
	}
	svc := NewService(&fakeRepo{customers: want})

	got, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List returned unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("List returned %d customers, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Name != want[i].Name {
			t.Errorf("customer[%d].Name = %q, want %q", i, got[i].Name, want[i].Name)
		}
	}
}

func TestServiceGetReturnsCustomerByID(t *testing.T) {
	repo := &fakeRepo{customers: []Customer{
		{ID: 1, Name: "Ada Lovelace"},
		{ID: 2, Name: "Alan Turing"},
	}}
	svc := NewService(repo)

	got, err := svc.Get(context.Background(), 2)
	if err != nil {
		t.Fatalf("Get returned unexpected error: %v", err)
	}
	if got.Name != "Alan Turing" {
		t.Errorf("Get(2).Name = %q, want %q", got.Name, "Alan Turing")
	}
}

func TestServiceListPropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	svc := NewService(&fakeRepo{err: repoErr})

	_, err := svc.List(context.Background())
	if !errors.Is(err, repoErr) {
		t.Fatalf("List error = %v, want %v", err, repoErr)
	}
}
