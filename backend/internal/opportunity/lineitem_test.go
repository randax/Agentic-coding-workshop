package opportunity

import (
	"context"
	"errors"
	"testing"
)

type fakeLineItems struct {
	items  map[uint]LineItem
	nextID uint
}

func newFakeLineItems() *fakeLineItems { return &fakeLineItems{items: map[uint]LineItem{}} }

func (f *fakeLineItems) Create(ctx context.Context, li *LineItem) error {
	f.nextID++
	li.ID = f.nextID
	f.items[li.ID] = *li
	return nil
}
func (f *fakeLineItems) ListByOpportunity(ctx context.Context, oppID uint) ([]LineItem, error) {
	var out []LineItem
	for _, li := range f.items {
		if li.OpportunityID == oppID {
			out = append(out, li)
		}
	}
	return out, nil
}

func TestAddComputesLineTotal(t *testing.T) {
	svc := NewLineItemService(newFakeLineItems())

	got, err := svc.Add(context.Background(), LineItem{
		OpportunityID: 1, ProductID: 2, ProductName: "Fiber 500", UnitPrice: 499, Quantity: 3,
	})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if got.ID == 0 || got.LineTotal != 1497 {
		t.Errorf("line item = %+v, want assigned id and lineTotal 1497 (499*3)", got)
	}
}

func TestAddRejectsNonPositiveQuantity(t *testing.T) {
	svc := NewLineItemService(newFakeLineItems())

	_, err := svc.Add(context.Background(), LineItem{OpportunityID: 1, ProductID: 2, Quantity: 0})
	if !errors.Is(err, ErrInvalidQuantity) {
		t.Fatalf("Add error = %v, want ErrInvalidQuantity", err)
	}
}

func TestListByOpportunityFiltersAndCanTotal(t *testing.T) {
	svc := NewLineItemService(newFakeLineItems())
	ctx := context.Background()
	svc.Add(ctx, LineItem{OpportunityID: 1, ProductName: "A", UnitPrice: 100, Quantity: 2}) // 200
	svc.Add(ctx, LineItem{OpportunityID: 1, ProductName: "B", UnitPrice: 50, Quantity: 1})  // 50
	svc.Add(ctx, LineItem{OpportunityID: 2, ProductName: "C", UnitPrice: 999, Quantity: 1})

	items, err := svc.ListByOpportunity(ctx, 1)
	if err != nil {
		t.Fatalf("ListByOpportunity error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d line items for opp 1, want 2", len(items))
	}
	var total float64
	for _, li := range items {
		total += li.LineTotal
	}
	if total != 250 {
		t.Errorf("total = %v, want 250", total)
	}
}
