package supportcase

import (
	"context"
	"errors"
	"testing"
)

// fakeRepo is an in-memory Repository for unit-testing the service in isolation.
type fakeRepo struct {
	cases         []Case
	comments      []CaseComment
	nextID        uint
	nextCommentID uint
	err           error
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

func (f *fakeRepo) Get(ctx context.Context, id uint) (Case, error) {
	if f.err != nil {
		return Case{}, f.err
	}
	for _, c := range f.cases {
		if c.ID == id {
			// Append comments added via CreateComment in insertion order,
			// mirroring the real repository's chronological preload.
			for _, cm := range f.comments {
				if cm.CaseID == id {
					c.Comments = append(c.Comments, cm)
				}
			}
			return c, nil
		}
	}
	return Case{}, ErrNotFound
}

func (f *fakeRepo) CreateComment(ctx context.Context, cm *CaseComment) error {
	if f.err != nil {
		return f.err
	}
	f.nextCommentID++
	cm.ID = f.nextCommentID
	f.comments = append(f.comments, *cm)
	return nil
}

func (f *fakeRepo) Update(ctx context.Context, c *Case) error {
	if f.err != nil {
		return f.err
	}
	for i := range f.cases {
		if f.cases[i].ID == c.ID {
			f.cases[i] = *c
			return nil
		}
	}
	return ErrNotFound
}

func (f *fakeRepo) Create(ctx context.Context, c *Case) error {
	if f.err != nil {
		return f.err
	}
	f.nextID++
	c.ID = f.nextID
	f.cases = append(f.cases, *c)
	return nil
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

func TestGetReturnsCaseWithComments(t *testing.T) {
	repo := &fakeRepo{cases: []Case{
		{ID: 5, CustomerID: 7, Subject: "No internet", Status: StatusOpen,
			Comments: []CaseComment{{ID: 1, CaseID: 5, Body: "Looking into it"}}},
	}}
	svc := NewService(repo)

	got, err := svc.Get(context.Background(), 5)
	if err != nil {
		t.Fatalf("Get returned unexpected error: %v", err)
	}
	if got.Subject != "No internet" || len(got.Comments) != 1 {
		t.Errorf("Get(5) = %+v, want subject 'No internet' with 1 comment", got)
	}
}

func TestGetUnknownReturnsNotFound(t *testing.T) {
	svc := NewService(&fakeRepo{})

	_, err := svc.Get(context.Background(), 999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get error = %v, want ErrNotFound", err)
	}
}

func TestCreateOpensCaseWithStatusOpenAndAssignsID(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	got, err := svc.Create(context.Background(), Case{
		CustomerID: 7, Subject: "No internet", Description: "Down since 8am",
		Category: CategoryConnectivity, Priority: PriorityHigh,
	})
	if err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}
	if got.ID == 0 {
		t.Errorf("Create did not assign an ID: %+v", got)
	}
	if got.Status != StatusOpen {
		t.Errorf("Create status = %q, want %q", got.Status, StatusOpen)
	}
	if got.CustomerID != 7 || got.Subject != "No internet" {
		t.Errorf("Create = %+v, want customer 7 subject 'No internet'", got)
	}
	if len(repo.cases) != 1 {
		t.Errorf("repo holds %d cases, want 1", len(repo.cases))
	}
}

func TestCreateRejectsMissingSubject(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), Case{
		CustomerID: 7, Subject: "  ", Category: CategoryConnectivity, Priority: PriorityHigh,
	})
	if !errors.Is(err, ErrSubjectRequired) {
		t.Fatalf("Create error = %v, want ErrSubjectRequired", err)
	}
	if len(repo.cases) != 0 {
		t.Errorf("invalid case was persisted: %+v", repo.cases)
	}
}

func TestCreateRejectsInvalidCategory(t *testing.T) {
	svc := NewService(&fakeRepo{})

	_, err := svc.Create(context.Background(), Case{
		CustomerID: 7, Subject: "Help", Category: Category("nonsense"), Priority: PriorityHigh,
	})
	if !errors.Is(err, ErrInvalidCategory) {
		t.Fatalf("Create error = %v, want ErrInvalidCategory", err)
	}
}

func TestCreateRejectsInvalidPriority(t *testing.T) {
	svc := NewService(&fakeRepo{})

	_, err := svc.Create(context.Background(), Case{
		CustomerID: 7, Subject: "Help", Category: CategoryBilling, Priority: Priority("whenever"),
	})
	if !errors.Is(err, ErrInvalidPriority) {
		t.Fatalf("Create error = %v, want ErrInvalidPriority", err)
	}
}

func TestAddCommentAppendsCommentAttributedToAgent(t *testing.T) {
	repo := &fakeRepo{cases: []Case{{ID: 5, CustomerID: 7, Subject: "No internet", Status: StatusOpen}}}
	svc := NewService(repo)

	got, err := svc.AddComment(context.Background(), 5, 2, "Looking into it")
	if err != nil {
		t.Fatalf("AddComment returned unexpected error: %v", err)
	}
	if got.ID == 0 || got.CaseID != 5 || got.Body != "Looking into it" {
		t.Errorf("AddComment = %+v, want case 5 body 'Looking into it' with ID", got)
	}
	if got.AuthorAgentID == nil || *got.AuthorAgentID != 2 {
		t.Errorf("comment not attributed to agent 2: %+v", got.AuthorAgentID)
	}
}

func TestAddCommentTimelineKeepsOrderAndAttribution(t *testing.T) {
	repo := &fakeRepo{cases: []Case{{ID: 5, CustomerID: 7, Subject: "No internet", Status: StatusOpen}}}
	svc := NewService(repo)

	if _, err := svc.AddComment(context.Background(), 5, 1, "First from agent 1"); err != nil {
		t.Fatalf("AddComment 1: %v", err)
	}
	if _, err := svc.AddComment(context.Background(), 5, 2, "Second from agent 2"); err != nil {
		t.Fatalf("AddComment 2: %v", err)
	}

	got, err := svc.Get(context.Background(), 5)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(got.Comments) != 2 {
		t.Fatalf("got %d comments, want 2", len(got.Comments))
	}
	if got.Comments[0].Body != "First from agent 1" || *got.Comments[0].AuthorAgentID != 1 {
		t.Errorf("comment[0] = %+v, want first from agent 1", got.Comments[0])
	}
	if got.Comments[1].Body != "Second from agent 2" || *got.Comments[1].AuthorAgentID != 2 {
		t.Errorf("comment[1] = %+v, want second from agent 2", got.Comments[1])
	}
}

func TestAddCommentRejectsEmptyBody(t *testing.T) {
	repo := &fakeRepo{cases: []Case{{ID: 5, Status: StatusOpen}}}
	svc := NewService(repo)

	_, err := svc.AddComment(context.Background(), 5, 1, "   ")
	if !errors.Is(err, ErrCommentBodyRequired) {
		t.Fatalf("AddComment error = %v, want ErrCommentBodyRequired", err)
	}
	if len(repo.comments) != 0 {
		t.Errorf("empty comment was persisted: %+v", repo.comments)
	}
}

func TestAddCommentToUnknownCaseReturnsNotFound(t *testing.T) {
	svc := NewService(&fakeRepo{})

	_, err := svc.AddComment(context.Background(), 999, 1, "Hello")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("AddComment error = %v, want ErrNotFound", err)
	}
}

func TestChangeStatusAdvancesOpenToInProgress(t *testing.T) {
	repo := &fakeRepo{cases: []Case{{ID: 5, CustomerID: 7, Subject: "No internet", Status: StatusOpen}}}
	svc := NewService(repo)

	got, err := svc.ChangeStatus(context.Background(), 5, StatusInProgress)
	if err != nil {
		t.Fatalf("ChangeStatus returned unexpected error: %v", err)
	}
	if got.Status != StatusInProgress {
		t.Errorf("status = %q, want %q", got.Status, StatusInProgress)
	}
	if repo.cases[0].Status != StatusInProgress {
		t.Errorf("persisted status = %q, want %q", repo.cases[0].Status, StatusInProgress)
	}
}

func TestChangeStatusAllowsEveryLegalTransition(t *testing.T) {
	legal := []struct{ from, to Status }{
		{StatusOpen, StatusInProgress},
		{StatusInProgress, StatusResolved},
		{StatusResolved, StatusClosed},
		{StatusResolved, StatusInProgress}, // reopen
	}
	for _, tc := range legal {
		t.Run(string(tc.from)+"_to_"+string(tc.to), func(t *testing.T) {
			repo := &fakeRepo{cases: []Case{{ID: 1, Status: tc.from}}}
			svc := NewService(repo)

			got, err := svc.ChangeStatus(context.Background(), 1, tc.to)
			if err != nil {
				t.Fatalf("ChangeStatus(%s→%s) returned error: %v", tc.from, tc.to, err)
			}
			if got.Status != tc.to {
				t.Errorf("status = %q, want %q", got.Status, tc.to)
			}
		})
	}
}

func TestChangeStatusRejectsIllegalTransitions(t *testing.T) {
	illegal := []struct{ from, to Status }{
		{StatusOpen, StatusResolved},      // skip In-progress
		{StatusOpen, StatusClosed},        // skip to terminal
		{StatusInProgress, StatusOpen},    // backward
		{StatusInProgress, StatusClosed},  // skip Resolved
		{StatusResolved, StatusOpen},      // backward past In-progress
		{StatusClosed, StatusInProgress},  // Closed is terminal
		{StatusClosed, StatusResolved},    // Closed is terminal
		{StatusOpen, StatusOpen},          // same-status is not a transition
		{StatusResolved, Status("bogus")}, // unknown target
	}
	for _, tc := range illegal {
		t.Run(string(tc.from)+"_to_"+string(tc.to), func(t *testing.T) {
			repo := &fakeRepo{cases: []Case{{ID: 1, Status: tc.from}}}
			svc := NewService(repo)

			_, err := svc.ChangeStatus(context.Background(), 1, tc.to)
			if !errors.Is(err, ErrIllegalTransition) {
				t.Fatalf("ChangeStatus(%s→%s) error = %v, want ErrIllegalTransition", tc.from, tc.to, err)
			}
			if repo.cases[0].Status != tc.from {
				t.Errorf("status changed to %q on an illegal transition; want unchanged %q", repo.cases[0].Status, tc.from)
			}
		})
	}
}

func TestChangeStatusUnknownCaseReturnsNotFound(t *testing.T) {
	svc := NewService(&fakeRepo{})

	_, err := svc.ChangeStatus(context.Background(), 999, StatusInProgress)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("ChangeStatus error = %v, want ErrNotFound", err)
	}
}

func ptr[T any](v T) *T { return &v }

func TestUpdateMetadataChangesPriorityOnly(t *testing.T) {
	repo := &fakeRepo{cases: []Case{
		{ID: 5, CustomerID: 7, Subject: "No internet", Status: StatusOpen,
			Category: CategoryConnectivity, Priority: PriorityLow},
	}}
	svc := NewService(repo)

	got, err := svc.UpdateMetadata(context.Background(), 5, MetadataPatch{Priority: ptr(PriorityUrgent)})
	if err != nil {
		t.Fatalf("UpdateMetadata returned unexpected error: %v", err)
	}
	if got.Priority != PriorityUrgent {
		t.Errorf("priority = %q, want %q", got.Priority, PriorityUrgent)
	}
	// Untouched fields stay put.
	if got.Category != CategoryConnectivity || got.Status != StatusOpen {
		t.Errorf("UpdateMetadata changed untouched fields: %+v", got)
	}
	if repo.cases[0].Priority != PriorityUrgent {
		t.Errorf("persisted priority = %q, want %q", repo.cases[0].Priority, PriorityUrgent)
	}
}

func TestUpdateMetadataAssignsAgentAndCategory(t *testing.T) {
	repo := &fakeRepo{cases: []Case{
		{ID: 5, Status: StatusOpen, Category: CategoryGeneral, Priority: PriorityLow},
	}}
	svc := NewService(repo)

	got, err := svc.UpdateMetadata(context.Background(), 5, MetadataPatch{
		Category:        ptr(CategoryBilling),
		AssignedAgentID: ptr(uint(3)),
	})
	if err != nil {
		t.Fatalf("UpdateMetadata returned unexpected error: %v", err)
	}
	if got.Category != CategoryBilling {
		t.Errorf("category = %q, want %q", got.Category, CategoryBilling)
	}
	if got.AssignedAgentID == nil || *got.AssignedAgentID != 3 {
		t.Errorf("case not assigned to agent 3: %+v", got.AssignedAgentID)
	}
}

func TestUpdateMetadataRejectsInvalidPriority(t *testing.T) {
	repo := &fakeRepo{cases: []Case{{ID: 5, Status: StatusOpen, Priority: PriorityLow}}}
	svc := NewService(repo)

	_, err := svc.UpdateMetadata(context.Background(), 5, MetadataPatch{Priority: ptr(Priority("whenever"))})
	if !errors.Is(err, ErrInvalidPriority) {
		t.Fatalf("error = %v, want ErrInvalidPriority", err)
	}
	if repo.cases[0].Priority != PriorityLow {
		t.Errorf("priority changed on invalid patch: %q", repo.cases[0].Priority)
	}
}

func TestUpdateMetadataRejectsInvalidCategory(t *testing.T) {
	repo := &fakeRepo{cases: []Case{{ID: 5, Status: StatusOpen, Category: CategoryGeneral}}}
	svc := NewService(repo)

	_, err := svc.UpdateMetadata(context.Background(), 5, MetadataPatch{Category: ptr(Category("nonsense"))})
	if !errors.Is(err, ErrInvalidCategory) {
		t.Fatalf("error = %v, want ErrInvalidCategory", err)
	}
}

func TestUpdateMetadataUnknownCaseReturnsNotFound(t *testing.T) {
	svc := NewService(&fakeRepo{})

	_, err := svc.UpdateMetadata(context.Background(), 999, MetadataPatch{Priority: ptr(PriorityHigh)})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
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
