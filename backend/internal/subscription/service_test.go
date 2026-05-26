package subscription

import (
	"context"
	"testing"
)

// fakeRepo is an in-memory Repository for unit-testing the service.
type fakeRepo struct {
	byCustomer map[uint][]Subscription
	err        error
}

func (f *fakeRepo) ListByCustomer(ctx context.Context, customerID uint) ([]Subscription, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.byCustomer[customerID], nil
}

func TestListForCustomerReturnsThatCustomersSubscriptions(t *testing.T) {
	repo := &fakeRepo{byCustomer: map[uint][]Subscription{
		7: {
			{ID: 1, CustomerID: 7, Status: StatusActive, Quantity: 1},
			{ID: 2, CustomerID: 7, Status: StatusCancelled, Quantity: 2},
		},
	}}
	svc := NewService(repo)

	got, err := svc.ListForCustomer(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListForCustomer returned unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d subscriptions, want 2", len(got))
	}
}
