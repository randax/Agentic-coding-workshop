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

func TestListForCustomerPropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	svc := NewService(&fakeRepo{err: repoErr})

	_, err := svc.ListForCustomer(context.Background(), 7)
	if !errors.Is(err, repoErr) {
		t.Fatalf("ListForCustomer error = %v, want %v", err, repoErr)
	}
}
