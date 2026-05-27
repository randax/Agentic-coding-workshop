package report

import (
	"context"
	"errors"
	"testing"
)

// fakeRepo is an in-memory Repository for unit-testing the service in isolation,
// without a database. This is the convention for service-layer unit tests.
type fakeRepo struct {
	saved  []Saved
	nextID uint
	err    error
}

func (f *fakeRepo) Create(ctx context.Context, s *Saved) error {
	if f.err != nil {
		return f.err
	}
	f.nextID++
	s.ID = f.nextID
	f.saved = append(f.saved, *s)
	return nil
}

func (f *fakeRepo) List(ctx context.Context) ([]Saved, error) {
	return f.saved, f.err
}

func (f *fakeRepo) Get(ctx context.Context, id uint) (Saved, error) {
	if f.err != nil {
		return Saved{}, f.err
	}
	for _, s := range f.saved {
		if s.ID == id {
			return s, nil
		}
	}
	return Saved{}, ErrNotFound
}

func TestSavePersistsAndAssignsID(t *testing.T) {
	svc := NewService(&fakeRepo{})

	got, err := svc.Save(context.Background(), Saved{
		Name: "Leads by status",
		Definition: Definition{
			Module: "leads", GroupBy: "status", Aggregation: Count,
		},
	})
	if err != nil {
		t.Fatalf("Save returned unexpected error: %v", err)
	}
	if got.ID == 0 {
		t.Fatalf("saved report = %+v, want a non-zero assigned ID", got)
	}
	if got.Name != "Leads by status" || got.Definition.GroupBy != "status" {
		t.Fatalf("saved report = %+v, want the submitted name and definition preserved", got)
	}
}

func TestSaveRequiresNameAndModule(t *testing.T) {
	svc := NewService(&fakeRepo{})

	_, err := svc.Save(context.Background(), Saved{Definition: Definition{Module: "leads"}})
	if !errors.Is(err, ErrNameRequired) {
		t.Fatalf("err = %v, want ErrNameRequired when the name is blank", err)
	}

	_, err = svc.Save(context.Background(), Saved{Name: "No module"})
	if !errors.Is(err, ErrModuleRequired) {
		t.Fatalf("err = %v, want ErrModuleRequired when the definition has no module", err)
	}
}

func TestListAndGetReturnSavedReports(t *testing.T) {
	svc := NewService(&fakeRepo{})
	a, _ := svc.Save(context.Background(), Saved{Name: "A", Definition: Definition{Module: "leads"}})
	svc.Save(context.Background(), Saved{Name: "B", Definition: Definition{Module: "opportunities"}})

	all, err := svc.List(context.Background())
	if err != nil || len(all) != 2 {
		t.Fatalf("List = %+v (err=%v), want 2 saved reports", all, err)
	}

	got, err := svc.Get(context.Background(), a.ID)
	if err != nil || got.Name != "A" {
		t.Fatalf("Get(%d) = %+v (err=%v), want report A", a.ID, got, err)
	}

	if _, err := svc.Get(context.Background(), 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get(999) err = %v, want ErrNotFound", err)
	}
}
