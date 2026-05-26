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

	got, err := svc.List(context.Background(), Filter{})
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

func TestServiceListFiltersByPartialNameCaseInsensitive(t *testing.T) {
	repo := &fakeRepo{customers: []Customer{
		{ID: 1, Name: "Ada Lovelace", AccountNumber: "ACME-001"},
		{ID: 2, Name: "Alan Turing", AccountNumber: "ACME-002"},
	}}
	svc := NewService(repo)

	got, err := svc.List(context.Background(), Filter{Search: "lOV"})
	if err != nil {
		t.Fatalf("List returned unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("List(search=lOV) returned %d customers, want 1: %+v", len(got), got)
	}
	if got[0].Name != "Ada Lovelace" {
		t.Errorf("got %q, want %q", got[0].Name, "Ada Lovelace")
	}
}

func TestServiceListFiltersByPartialEmail(t *testing.T) {
	repo := &fakeRepo{customers: []Customer{
		{ID: 1, Name: "Ada Lovelace", Email: "ada@analytical.example"},
		{ID: 2, Name: "Alan Turing", Email: "alan@bletchley.example"},
	}}
	svc := NewService(repo)

	got, err := svc.List(context.Background(), Filter{Search: "bletchley"})
	if err != nil {
		t.Fatalf("List returned unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Alan Turing" {
		t.Fatalf("List(search=bletchley) = %+v, want only Alan Turing", got)
	}
}

func TestServiceListFiltersByPartialAccountNumber(t *testing.T) {
	repo := &fakeRepo{customers: []Customer{
		{ID: 1, Name: "Ada Lovelace", AccountNumber: "ACME-001"},
		{ID: 2, Name: "Alan Turing", AccountNumber: "GLOBEX-7"},
	}}
	svc := NewService(repo)

	got, err := svc.List(context.Background(), Filter{Search: "globex"})
	if err != nil {
		t.Fatalf("List returned unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Alan Turing" {
		t.Fatalf("List(search=globex) = %+v, want only Alan Turing", got)
	}
}

func TestServiceListFiltersByStatus(t *testing.T) {
	repo := &fakeRepo{customers: []Customer{
		{ID: 1, Name: "Ada Lovelace", Status: StatusActive},
		{ID: 2, Name: "Alan Turing", Status: StatusSuspended},
		{ID: 3, Name: "Grace Hopper", Status: StatusActive},
	}}
	svc := NewService(repo)

	got, err := svc.List(context.Background(), Filter{Status: StatusSuspended})
	if err != nil {
		t.Fatalf("List returned unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Alan Turing" {
		t.Fatalf("List(status=suspended) = %+v, want only Alan Turing", got)
	}
}

func TestServiceListCombinesSearchAndStatus(t *testing.T) {
	repo := &fakeRepo{customers: []Customer{
		{ID: 1, Name: "Ada Lovelace", Status: StatusActive},
		{ID: 2, Name: "Ada Byron", Status: StatusSuspended},
		{ID: 3, Name: "Grace Hopper", Status: StatusActive},
	}}
	svc := NewService(repo)

	// "ada" matches customers 1 and 2; status=active narrows to customer 1.
	got, err := svc.List(context.Background(), Filter{Search: "ada", Status: StatusActive})
	if err != nil {
		t.Fatalf("List returned unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Ada Lovelace" {
		t.Fatalf("List(search=ada,status=active) = %+v, want only Ada Lovelace", got)
	}
}

func TestServiceListPropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	svc := NewService(&fakeRepo{err: repoErr})

	_, err := svc.List(context.Background(), Filter{})
	if !errors.Is(err, repoErr) {
		t.Fatalf("List error = %v, want %v", err, repoErr)
	}
}
