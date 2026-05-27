package studio

import (
	"context"
	"errors"
	"testing"
)

type fakeRepo struct {
	defs   map[uint]FieldDef
	nextID uint
}

func newFakeRepo() *fakeRepo { return &fakeRepo{defs: map[uint]FieldDef{}} }

func (f *fakeRepo) Create(ctx context.Context, d *FieldDef) error {
	f.nextID++
	d.ID = f.nextID
	f.defs[d.ID] = *d
	return nil
}
func (f *fakeRepo) ListByModule(ctx context.Context, module string) ([]FieldDef, error) {
	var out []FieldDef
	for _, d := range f.defs {
		if d.Module == module {
			out = append(out, d)
		}
	}
	return out, nil
}

func TestAddFieldStoresDefinition(t *testing.T) {
	svc := NewService(newFakeRepo())

	got, err := svc.AddField(context.Background(), FieldDef{Module: "accounts", Name: "churnRisk", Type: "enum", Label: "Churn risk", Options: []string{"low", "high"}})
	if err != nil {
		t.Fatalf("AddField error: %v", err)
	}
	if got.ID == 0 {
		t.Errorf("AddField should assign an id, got %+v", got)
	}
}

func TestAddFieldValidates(t *testing.T) {
	svc := NewService(newFakeRepo())
	ctx := context.Background()

	if _, err := svc.AddField(ctx, FieldDef{Module: "accounts", Type: "string", Label: "X"}); !errors.Is(err, ErrNameRequired) {
		t.Errorf("missing name error = %v, want ErrNameRequired", err)
	}
	if _, err := svc.AddField(ctx, FieldDef{Module: "accounts", Name: "x", Type: "banana", Label: "X"}); !errors.Is(err, ErrInvalidType) {
		t.Errorf("bad type error = %v, want ErrInvalidType", err)
	}
	if _, err := svc.AddField(ctx, FieldDef{Module: "", Name: "x", Type: "string", Label: "X"}); !errors.Is(err, ErrModuleRequired) {
		t.Errorf("missing module error = %v, want ErrModuleRequired", err)
	}
}

func TestListByModuleFilters(t *testing.T) {
	svc := NewService(newFakeRepo())
	ctx := context.Background()
	svc.AddField(ctx, FieldDef{Module: "accounts", Name: "a", Type: "string", Label: "A"})
	svc.AddField(ctx, FieldDef{Module: "accounts", Name: "b", Type: "string", Label: "B"})
	svc.AddField(ctx, FieldDef{Module: "contacts", Name: "c", Type: "string", Label: "C"})

	got, err := svc.ListByModule(ctx, "accounts")
	if err != nil {
		t.Fatalf("ListByModule error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d defs for accounts, want 2", len(got))
	}
}
