package lead

import (
	"context"
	"errors"
	"testing"
)

type fakeRepo struct {
	items  map[uint]Lead
	nextID uint
}

func newFakeRepo() *fakeRepo { return &fakeRepo{items: map[uint]Lead{}} }

func (f *fakeRepo) List(ctx context.Context) ([]Lead, error) {
	out := make([]Lead, 0, len(f.items))
	for _, l := range f.items {
		out = append(out, l)
	}
	return out, nil
}
func (f *fakeRepo) Get(ctx context.Context, id uint) (Lead, error) {
	if l, ok := f.items[id]; ok {
		return l, nil
	}
	return Lead{}, ErrNotFound
}
func (f *fakeRepo) Create(ctx context.Context, l *Lead) error {
	f.nextID++
	l.ID = f.nextID
	f.items[l.ID] = *l
	return nil
}
func (f *fakeRepo) Update(ctx context.Context, l *Lead) error {
	if _, ok := f.items[l.ID]; !ok {
		return ErrNotFound
	}
	f.items[l.ID] = *l
	return nil
}

func TestCreateDefaultsStatusToNew(t *testing.T) {
	svc := NewService(newFakeRepo())

	got, err := svc.Create(context.Background(), Lead{Name: "Ada Lovelace"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if got.ID == 0 || got.Status != StatusNew {
		t.Errorf("created lead = %+v, want assigned id and status new", got)
	}
}

func TestCreateRequiresName(t *testing.T) {
	svc := NewService(newFakeRepo())

	_, err := svc.Create(context.Background(), Lead{Company: "Globex"})
	if !errors.Is(err, ErrNameRequired) {
		t.Fatalf("Create error = %v, want ErrNameRequired", err)
	}
}

func TestCreateRejectsInvalidStatus(t *testing.T) {
	svc := NewService(newFakeRepo())

	_, err := svc.Create(context.Background(), Lead{Name: "Ada", Status: "banana"})
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("Create error = %v, want ErrInvalidStatus", err)
	}
}

func TestUpdateChangesStatus(t *testing.T) {
	svc := NewService(newFakeRepo())
	ctx := context.Background()
	created, _ := svc.Create(ctx, Lead{Name: "Ada"})

	created.Status = StatusQualified
	updated, err := svc.Update(ctx, created)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Status != StatusQualified {
		t.Errorf("updated status = %q, want qualified", updated.Status)
	}
}

func TestUpdateUnknownReturnsNotFound(t *testing.T) {
	svc := NewService(newFakeRepo())

	_, err := svc.Update(context.Background(), Lead{ID: 999, Name: "Ghost"})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Update error = %v, want ErrNotFound", err)
	}
}

func TestUpdateCannotManuallySetConvertedStatus(t *testing.T) {
	// `converted` is terminal and only the conversion workflow may set it (it
	// also wires up ConvertedAccountID). A plain edit must not be able to flip a
	// lead to converted, which would leave a converted lead with no account.
	svc := NewService(newFakeRepo())
	ctx := context.Background()
	created, _ := svc.Create(ctx, Lead{Name: "Ada", Company: "Globex"})

	created.Status = StatusConverted
	_, err := svc.Update(ctx, created)
	if !errors.Is(err, ErrConvertedProtected) {
		t.Fatalf("Update error = %v, want ErrConvertedProtected", err)
	}
}

func TestUpdateCannotEditAConvertedLead(t *testing.T) {
	// A converted lead is terminal: editing it (e.g. reverting status) must be
	// rejected so its link to the account it became stays consistent.
	repo := newFakeRepo()
	acc := uint(42)
	repo.items[1] = Lead{ID: 1, Name: "Sofia", Company: "Polar Foods", Status: StatusConverted, ConvertedAccountID: &acc}
	repo.nextID = 1
	svc := NewService(repo)

	_, err := svc.Update(context.Background(), Lead{ID: 1, Name: "Sofia (edited)", Company: "Polar Foods", Status: StatusQualified})
	if !errors.Is(err, ErrConvertedProtected) {
		t.Fatalf("Update error = %v, want ErrConvertedProtected", err)
	}
}
