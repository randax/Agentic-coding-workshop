package supportcase

import (
	"context"
	"errors"
	"testing"
)

// fakeRepo is an in-memory Repository for unit-testing the service in isolation.
type fakeRepo struct {
	cases []Case
	err   error
}

func (f *fakeRepo) ListByCustomer(ctx context.Context, customerID uint) ([]Case, error) {
	if f.err != nil {
		return nil, f.err
	}
	var out []Case
	for _, c := range f.cases {
		if c.CustomerID == customerID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (f *fakeRepo) Get(ctx context.Context, id uint) (Case, error) {
	if f.err != nil {
		return Case{}, f.err
	}
	for _, c := range f.cases {
		if c.ID == id {
			return c, nil
		}
	}
	return Case{}, ErrNotFound
}

func TestListForCustomerReturnsThatCustomersCases(t *testing.T) {
	repo := &fakeRepo{cases: []Case{
		{ID: 1, CustomerID: 7, Subject: "No internet", Status: StatusOpen, Priority: PriorityHigh, Category: CategoryConnectivity},
		{ID: 2, CustomerID: 7, Subject: "Wrong bill", Status: StatusResolved, Priority: PriorityLow, Category: CategoryBilling},
		{ID: 3, CustomerID: 9, Subject: "Router dead", Status: StatusOpen, Priority: PriorityUrgent, Category: CategoryHardware},
	}}
	svc := NewService(repo)

	got, err := svc.ListForCustomer(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListForCustomer returned unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d cases, want 2", len(got))
	}
}

func TestGetReturnsCaseWithComments(t *testing.T) {
	repo := &fakeRepo{cases: []Case{
		{ID: 5, CustomerID: 7, Subject: "No internet", Status: StatusOpen,
			Comments: []CaseComment{{ID: 1, CaseID: 5, Body: "Looking into it"}}},
	}}
	svc := NewService(repo)

	got, err := svc.Get(context.Background(), 5)
	if err != nil {
		t.Fatalf("Get returned unexpected error: %v", err)
	}
	if got.Subject != "No internet" || len(got.Comments) != 1 {
		t.Errorf("Get(5) = %+v, want subject 'No internet' with 1 comment", got)
	}
}

func TestGetUnknownReturnsNotFound(t *testing.T) {
	svc := NewService(&fakeRepo{})

	_, err := svc.Get(context.Background(), 999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get error = %v, want ErrNotFound", err)
	}
}

func TestListForCustomerPropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	svc := NewService(&fakeRepo{err: repoErr})

	_, err := svc.ListForCustomer(context.Background(), 7)
	if !errors.Is(err, repoErr) {
		t.Fatalf("ListForCustomer error = %v, want %v", err, repoErr)
	}
}
