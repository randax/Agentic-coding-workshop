package studio

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type fakeLayoutRepo struct {
	defs map[string]LayoutDef // keyed by module+"/"+view
}

func newFakeLayoutRepo() *fakeLayoutRepo { return &fakeLayoutRepo{defs: map[string]LayoutDef{}} }

func (f *fakeLayoutRepo) Upsert(ctx context.Context, d *LayoutDef) error {
	f.defs[d.Module+"/"+d.View] = *d
	return nil
}

func (f *fakeLayoutRepo) GetByModule(ctx context.Context, module string) ([]LayoutDef, error) {
	var out []LayoutDef
	for _, d := range f.defs {
		if d.Module == module {
			out = append(out, d)
		}
	}
	return out, nil
}

func TestSetLayoutValidates(t *testing.T) {
	svc := NewService(newFakeRepo(), newFakeLayoutRepo())
	ctx := context.Background()

	if err := svc.SetLayout(ctx, "accounts", "banana", []string{"name"}); !errors.Is(err, ErrInvalidView) {
		t.Errorf("bad view error = %v, want ErrInvalidView", err)
	}
	if err := svc.SetLayout(ctx, "", "list", []string{"name"}); !errors.Is(err, ErrModuleRequired) {
		t.Errorf("missing module error = %v, want ErrModuleRequired", err)
	}
	for _, view := range []string{"list", "detail", "edit"} {
		if err := svc.SetLayout(ctx, "accounts", view, []string{"name"}); err != nil {
			t.Errorf("SetLayout(%q) unexpected error: %v", view, err)
		}
	}
}

func TestGetLayoutsReturnsSavedByView(t *testing.T) {
	svc := NewService(newFakeRepo(), newFakeLayoutRepo())
	ctx := context.Background()
	svc.SetLayout(ctx, "accounts", "list", []string{"status", "name"})
	svc.SetLayout(ctx, "accounts", "edit", []string{"name"})
	svc.SetLayout(ctx, "contacts", "list", []string{"email"})

	got, err := svc.GetLayouts(ctx, "accounts")
	if err != nil {
		t.Fatalf("GetLayouts error: %v", err)
	}
	want := map[string][]string{"list": {"status", "name"}, "edit": {"name"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("layouts = %+v, want %+v (contacts excluded)", got, want)
	}
}

func TestSetLayoutUpsertsView(t *testing.T) {
	svc := NewService(newFakeRepo(), newFakeLayoutRepo())
	ctx := context.Background()
	svc.SetLayout(ctx, "accounts", "list", []string{"name"})
	svc.SetLayout(ctx, "accounts", "list", []string{"status", "name"})

	got, _ := svc.GetLayouts(ctx, "accounts")
	if want := []string{"status", "name"}; !reflect.DeepEqual(got["list"], want) {
		t.Errorf("list layout = %v, want %v (replaced, not duplicated)", got["list"], want)
	}
}
