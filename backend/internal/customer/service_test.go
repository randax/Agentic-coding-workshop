package customer

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeRepo is an in-memory Repository for unit-testing the service in isolation,
// without a database. This is the convention for service-layer unit tests.
type fakeRepo struct {
	customers []Customer
	nextID    uint
	err       error
}

func (f *fakeRepo) List(ctx context.Context) ([]Customer, error) {
	return f.customers, f.err
}

func (f *fakeRepo) Create(ctx context.Context, c *Customer) error {
	if f.err != nil {
		return f.err
	}
	f.nextID++
	c.ID = f.nextID
	f.customers = append(f.customers, *c)
	return nil
}

func (f *fakeRepo) Update(ctx context.Context, c *Customer) error {
	if f.err != nil {
		return f.err
	}
	for i := range f.customers {
		if f.customers[i].ID == c.ID {
			f.customers[i] = *c
			return nil
		}
	}
	return ErrNotFound
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

func TestServiceCreatePersistsCustomerAndAssignsID(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	got, err := svc.Create(context.Background(), Customer{
		Name: "Grace Hopper", Email: "grace@navy.example", AccountNumber: "ACME-003",
		Status: StatusActive,
	})
	if err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}
	if got.ID == 0 {
		t.Errorf("Create did not assign an ID: %+v", got)
	}
	if got.Name != "Grace Hopper" {
		t.Errorf("Create returned name %q, want %q", got.Name, "Grace Hopper")
	}
	if len(repo.customers) != 1 {
		t.Errorf("repo holds %d customers, want 1", len(repo.customers))
	}
}

func TestServiceCreateRejectsMissingName(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), Customer{
		Email: "x@example.com", AccountNumber: "ACME-9", Status: StatusActive,
	})
	if !errors.Is(err, ErrNameRequired) {
		t.Fatalf("Create error = %v, want ErrNameRequired", err)
	}
	if len(repo.customers) != 0 {
		t.Errorf("invalid customer was persisted: %+v", repo.customers)
	}
}

func TestServiceCreateRejectsMissingEmail(t *testing.T) {
	svc := NewService(&fakeRepo{})

	_, err := svc.Create(context.Background(), Customer{
		Name: "Grace Hopper", AccountNumber: "ACME-9", Status: StatusActive,
	})
	if !errors.Is(err, ErrEmailRequired) {
		t.Fatalf("Create error = %v, want ErrEmailRequired", err)
	}
}

func TestServiceCreateRejectsMissingAccountNumber(t *testing.T) {
	svc := NewService(&fakeRepo{})

	_, err := svc.Create(context.Background(), Customer{
		Name: "Grace Hopper", Email: "grace@navy.example", Status: StatusActive,
	})
	if !errors.Is(err, ErrAccountNumberRequired) {
		t.Fatalf("Create error = %v, want ErrAccountNumberRequired", err)
	}
}

func TestServiceCreateDefaultsStatusToActive(t *testing.T) {
	svc := NewService(&fakeRepo{})

	got, err := svc.Create(context.Background(), Customer{
		Name: "Grace Hopper", Email: "grace@navy.example", AccountNumber: "ACME-9",
		// Status intentionally left empty.
	})
	if err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}
	if got.Status != StatusActive {
		t.Errorf("Create status = %q, want %q", got.Status, StatusActive)
	}
}

func TestServiceCreateRejectsInvalidStatus(t *testing.T) {
	svc := NewService(&fakeRepo{})

	_, err := svc.Create(context.Background(), Customer{
		Name: "Grace Hopper", Email: "grace@navy.example", AccountNumber: "ACME-9",
		Status: Status("frozen"),
	})
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("Create error = %v, want ErrInvalidStatus", err)
	}
}

func TestServiceCreateStampsCustomerSinceWhenUnset(t *testing.T) {
	svc := NewService(&fakeRepo{})

	got, err := svc.Create(context.Background(), Customer{
		Name: "Grace Hopper", Email: "grace@navy.example", AccountNumber: "ACME-9",
	})
	if err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}
	if got.CustomerSince.IsZero() {
		t.Errorf("Create left CustomerSince unset")
	}
}

func TestServiceUpdateChangesFieldsAndPreservesCustomerSince(t *testing.T) {
	since := time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC)
	repo := &fakeRepo{customers: []Customer{
		{ID: 7, Name: "Grace Hopper", Email: "grace@navy.example",
			AccountNumber: "ACME-9", Status: StatusActive, CustomerSince: since},
	}, nextID: 7}
	svc := NewService(repo)

	got, err := svc.Update(context.Background(), Customer{
		ID: 7, Name: "Grace B. Hopper", Email: "ghopper@navy.example",
		Phone: "555-0100", AccountNumber: "ACME-9", Status: StatusSuspended,
	})
	if err != nil {
		t.Fatalf("Update returned unexpected error: %v", err)
	}
	if got.Name != "Grace B. Hopper" || got.Email != "ghopper@navy.example" ||
		got.Phone != "555-0100" || got.Status != StatusSuspended {
		t.Errorf("Update did not apply edits: %+v", got)
	}
	if !got.CustomerSince.Equal(since) {
		t.Errorf("Update changed CustomerSince: got %v, want %v", got.CustomerSince, since)
	}
}

func TestServiceUpdateUnknownIDReturnsNotFound(t *testing.T) {
	svc := NewService(&fakeRepo{})

	_, err := svc.Update(context.Background(), Customer{
		ID: 999, Name: "Nobody", Email: "n@example.com",
		AccountNumber: "X", Status: StatusActive,
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Update error = %v, want ErrNotFound", err)
	}
}

func TestServiceUpdateValidatesFields(t *testing.T) {
	repo := &fakeRepo{customers: []Customer{
		{ID: 7, Name: "Grace Hopper", Email: "grace@navy.example",
			AccountNumber: "ACME-9", Status: StatusActive},
	}, nextID: 7}
	svc := NewService(repo)

	_, err := svc.Update(context.Background(), Customer{
		ID: 7, Name: "  ", Email: "grace@navy.example",
		AccountNumber: "ACME-9", Status: StatusActive,
	})
	if !errors.Is(err, ErrNameRequired) {
		t.Fatalf("Update error = %v, want ErrNameRequired", err)
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
