package opportunity

import (
	"context"
	"errors"
	"testing"
)

type fakeRepo struct {
	items  map[uint]Opportunity
	nextID uint
}

func newFakeRepo() *fakeRepo { return &fakeRepo{items: map[uint]Opportunity{}} }

func (f *fakeRepo) List(ctx context.Context) ([]Opportunity, error) {
	out := make([]Opportunity, 0, len(f.items))
	for _, o := range f.items {
		out = append(out, o)
	}
	return out, nil
}
func (f *fakeRepo) Get(ctx context.Context, id uint) (Opportunity, error) {
	if o, ok := f.items[id]; ok {
		return o, nil
	}
	return Opportunity{}, ErrNotFound
}
func (f *fakeRepo) Create(ctx context.Context, o *Opportunity) error {
	f.nextID++
	o.ID = f.nextID
	f.items[o.ID] = *o
	return nil
}
func (f *fakeRepo) Update(ctx context.Context, o *Opportunity) error {
	if _, ok := f.items[o.ID]; !ok {
		return ErrNotFound
	}
	f.items[o.ID] = *o
	return nil
}

func TestCreateDefaultsStageAndProbability(t *testing.T) {
	svc := NewService(newFakeRepo())

	got, err := svc.Create(context.Background(), Opportunity{Name: "Globex fiber", AccountID: 1, Amount: 50000})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if got.Stage != StageProspecting {
		t.Errorf("default stage = %q, want prospecting", got.Stage)
	}
	if got.Probability != 10 {
		t.Errorf("default probability = %d, want 10 (prospecting)", got.Probability)
	}
}

func TestCreateRejectsNegativeAmountAndMissingName(t *testing.T) {
	svc := NewService(newFakeRepo())

	if _, err := svc.Create(context.Background(), Opportunity{AccountID: 1, Amount: 10}); !errors.Is(err, ErrNameRequired) {
		t.Errorf("missing name error = %v, want ErrNameRequired", err)
	}
	if _, err := svc.Create(context.Background(), Opportunity{Name: "X", AccountID: 1, Amount: -1}); !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("negative amount error = %v, want ErrInvalidAmount", err)
	}
}

func TestUpdateStageResetsProbabilityToStageDefault(t *testing.T) {
	svc := NewService(newFakeRepo())
	ctx := context.Background()
	created, _ := svc.Create(ctx, Opportunity{Name: "Globex", AccountID: 1, Amount: 1000})

	created.Stage = StageClosedWon
	updated, err := svc.Update(ctx, created)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Probability != 100 {
		t.Errorf("probability after Closed Won = %d, want 100", updated.Probability)
	}
}

func TestUpdateRejectsInvalidStage(t *testing.T) {
	svc := NewService(newFakeRepo())
	ctx := context.Background()
	created, _ := svc.Create(ctx, Opportunity{Name: "Globex", AccountID: 1, Amount: 1000})

	created.Stage = "banana"
	if _, err := svc.Update(ctx, created); !errors.Is(err, ErrInvalidStage) {
		t.Fatalf("Update error = %v, want ErrInvalidStage", err)
	}
}

func TestUpdateUnknownReturnsNotFound(t *testing.T) {
	svc := NewService(newFakeRepo())

	_, err := svc.Update(context.Background(), Opportunity{ID: 999, Name: "Ghost", AccountID: 1, Stage: StageProspecting})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Update error = %v, want ErrNotFound", err)
	}
}
