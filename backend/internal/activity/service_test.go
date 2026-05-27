package activity

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepo struct {
	items  map[uint]Activity
	nextID uint
}

func newFakeRepo() *fakeRepo { return &fakeRepo{items: map[uint]Activity{}} }

func (f *fakeRepo) List(ctx context.Context) ([]Activity, error) {
	out := make([]Activity, 0, len(f.items))
	for _, a := range f.items {
		out = append(out, a)
	}
	return out, nil
}
func (f *fakeRepo) ListForParent(ctx context.Context, parentType string, parentID uint) ([]Activity, error) {
	var out []Activity
	for _, a := range f.items {
		if a.ParentType == parentType && a.ParentID == parentID {
			out = append(out, a)
		}
	}
	return out, nil
}
func (f *fakeRepo) Get(ctx context.Context, id uint) (Activity, error) {
	if a, ok := f.items[id]; ok {
		return a, nil
	}
	return Activity{}, ErrNotFound
}
func (f *fakeRepo) Create(ctx context.Context, a *Activity) error {
	f.nextID++
	a.ID = f.nextID
	f.items[a.ID] = *a
	return nil
}
func (f *fakeRepo) Update(ctx context.Context, a *Activity) error {
	if _, ok := f.items[a.ID]; !ok {
		return ErrNotFound
	}
	f.items[a.ID] = *a
	return nil
}

func TestLogDefaultsTaskStatusToOpen(t *testing.T) {
	svc := NewService(newFakeRepo())

	got, err := svc.Log(context.Background(), Activity{Type: TypeTask, Subject: "Call back", ParentType: "account", ParentID: 1})
	if err != nil {
		t.Fatalf("Log returned error: %v", err)
	}
	if got.ID == 0 || got.Status != StatusOpen {
		t.Errorf("activity = %+v, want assigned id and open status", got)
	}
}

func TestLogRequiresSubjectAndValidType(t *testing.T) {
	svc := NewService(newFakeRepo())

	if _, err := svc.Log(context.Background(), Activity{Type: TypeCall, ParentType: "account", ParentID: 1}); !errors.Is(err, ErrSubjectRequired) {
		t.Errorf("missing subject error = %v, want ErrSubjectRequired", err)
	}
	if _, err := svc.Log(context.Background(), Activity{Type: "banana", Subject: "x", ParentType: "account", ParentID: 1}); !errors.Is(err, ErrInvalidType) {
		t.Errorf("bad type error = %v, want ErrInvalidType", err)
	}
}

func TestListForParentFiltersByParent(t *testing.T) {
	svc := NewService(newFakeRepo())
	ctx := context.Background()
	svc.Log(ctx, Activity{Type: TypeCall, Subject: "A", ParentType: "account", ParentID: 1})
	svc.Log(ctx, Activity{Type: TypeMeeting, Subject: "B", ParentType: "account", ParentID: 1})
	svc.Log(ctx, Activity{Type: TypeTask, Subject: "C", ParentType: "contact", ParentID: 1})

	got, err := svc.ListForParent(ctx, "account", 1)
	if err != nil {
		t.Fatalf("ListForParent error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d activities for account 1, want 2", len(got))
	}
}

func TestCompleteMarksTaskDone(t *testing.T) {
	svc := NewService(newFakeRepo())
	ctx := context.Background()
	created, _ := svc.Log(ctx, Activity{Type: TypeTask, Subject: "Follow up", ParentType: "account", ParentID: 1})

	done, err := svc.Complete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Complete error: %v", err)
	}
	if done.Status != StatusDone {
		t.Errorf("status = %q, want done", done.Status)
	}
}

func TestOpenTasksForUserReturnsOnlyMyOpenTasksSoonestFirst(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo)
	ctx := context.Background()
	me, other := uint(5), uint(6)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mk := func(typ Type, st Status, user uint, occ time.Time, subj string) {
		u := user
		repo.Create(ctx, &Activity{Type: typ, Status: st, AssignedUserID: &u, OccurredAt: occ, Subject: subj})
	}
	mk(TypeTask, StatusOpen, me, base.AddDate(0, 0, 2), "later task")
	mk(TypeTask, StatusOpen, me, base.AddDate(0, 0, 1), "sooner task")
	mk(TypeTask, StatusDone, me, base, "done task")     // excluded: completed
	mk(TypeCall, StatusOpen, me, base, "a call")        // excluded: not a task
	mk(TypeTask, StatusOpen, other, base, "their task") // excluded: not mine

	got, err := svc.OpenTasksForUser(ctx, me)
	if err != nil {
		t.Fatalf("OpenTasksForUser error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d tasks, want 2; %+v", len(got), got)
	}
	if got[0].Subject != "sooner task" || got[1].Subject != "later task" {
		t.Errorf("order = [%q, %q], want sooner then later (by OccurredAt)", got[0].Subject, got[1].Subject)
	}
}
