package contact

import (
	"context"
	"errors"
	"testing"
)

// fakeRepo is an in-memory Repository for unit-testing the service.
type fakeRepo struct {
	items  map[uint]Contact
	nextID uint
}

func newFakeRepo() *fakeRepo { return &fakeRepo{items: map[uint]Contact{}} }

func (f *fakeRepo) List(ctx context.Context) ([]Contact, error) {
	out := make([]Contact, 0, len(f.items))
	for _, c := range f.items {
		out = append(out, c)
	}
	return out, nil
}
func (f *fakeRepo) Get(ctx context.Context, id uint) (Contact, error) {
	if c, ok := f.items[id]; ok {
		return c, nil
	}
	return Contact{}, ErrNotFound
}
func (f *fakeRepo) Create(ctx context.Context, c *Contact) error {
	f.nextID++
	c.ID = f.nextID
	f.items[c.ID] = *c
	return nil
}
func (f *fakeRepo) Update(ctx context.Context, c *Contact) error {
	if _, ok := f.items[c.ID]; !ok {
		return ErrNotFound
	}
	f.items[c.ID] = *c
	return nil
}
func (f *fakeRepo) ListByAccount(ctx context.Context, accountID uint) ([]Contact, error) {
	var out []Contact
	for _, c := range f.items {
		if c.AccountID == accountID {
			out = append(out, c)
		}
	}
	return out, nil
}

func TestCreateStoresContact(t *testing.T) {
	svc := NewService(newFakeRepo())

	got, err := svc.Create(context.Background(), Contact{Name: "Ada Lovelace", Email: "ada@x.example", AccountID: 7})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if got.ID == 0 || got.AccountID != 7 {
		t.Errorf("created contact = %+v, want assigned id and account 7", got)
	}
}

func TestCreateRequiresName(t *testing.T) {
	svc := NewService(newFakeRepo())

	_, err := svc.Create(context.Background(), Contact{Email: "x@x.example", AccountID: 1})
	if !errors.Is(err, ErrNameRequired) {
		t.Fatalf("Create error = %v, want ErrNameRequired", err)
	}
}

func TestUpdateEditsContact(t *testing.T) {
	svc := NewService(newFakeRepo())
	ctx := context.Background()
	created, _ := svc.Create(ctx, Contact{Name: "Ada", Email: "ada@x.example", AccountID: 1})

	created.Name = "Ada Lovelace"
	updated, err := svc.Update(ctx, created)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Name != "Ada Lovelace" {
		t.Errorf("updated = %+v, want name Ada Lovelace", updated)
	}
}

func TestUpdateUnknownReturnsNotFound(t *testing.T) {
	svc := NewService(newFakeRepo())

	_, err := svc.Update(context.Background(), Contact{ID: 999, Name: "Ghost", AccountID: 1})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Update error = %v, want ErrNotFound", err)
	}
}

func TestListByAccountFiltersByAccount(t *testing.T) {
	svc := NewService(newFakeRepo())
	ctx := context.Background()
	svc.Create(ctx, Contact{Name: "A", AccountID: 1})
	svc.Create(ctx, Contact{Name: "B", AccountID: 1})
	svc.Create(ctx, Contact{Name: "C", AccountID: 2})

	got, err := svc.ListByAccount(ctx, 1)
	if err != nil {
		t.Fatalf("ListByAccount error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("ListByAccount(1) = %d contacts, want 2", len(got))
	}
}
