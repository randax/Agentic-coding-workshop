package conversion

import (
	"context"
	"errors"
	"testing"

	"saltcrm/internal/agent"
	"saltcrm/internal/customer"
	"saltcrm/internal/lead"
)

// fakeRepo is an in-memory Repository for unit-testing the service in isolation,
// without a database. It records the Plan it was asked to persist so tests can
// assert the service's field mapping, and mints ids like the real store would.
type fakeRepo struct {
	lead       lead.Lead
	leadErr    error
	persistErr error

	persisted  *Plan // the plan handed to Persist, captured for assertions
	persistHit bool
}

func (f *fakeRepo) GetLead(ctx context.Context, id uint) (lead.Lead, error) {
	if f.leadErr != nil {
		return lead.Lead{}, f.leadErr
	}
	if f.lead.ID != id {
		return lead.Lead{}, lead.ErrNotFound
	}
	return f.lead, nil
}

func (f *fakeRepo) Persist(ctx context.Context, p Plan) (Result, error) {
	f.persistHit = true
	if f.persistErr != nil {
		return Result{}, f.persistErr
	}
	cp := p
	f.persisted = &cp
	return Result{AccountID: 100, ContactID: 200}, nil
}

func uintPtr(v uint) *uint { return &v }

// qualifiedLead is a visible, convertible lead owned by the viewer below.
func qualifiedLead() lead.Lead {
	return lead.Lead{
		ID: 7, Name: "Sofia Berg", Email: "sofia@polarfoods.example",
		Phone: "+47 903 55 666", Company: "Polar Foods",
		Status: lead.StatusQualified, AssignedUserID: uintPtr(42), TeamID: uintPtr(3),
	}
}

func viewer() agent.Agent {
	return agent.Agent{ID: 42, Role: agent.RoleAgent, TeamID: uintPtr(3)}
}

func TestConvertMapsLeadToNewAccountAndContact(t *testing.T) {
	repo := &fakeRepo{lead: qualifiedLead()}
	svc := NewService(repo)

	got, err := svc.Convert(context.Background(), viewer(), 7)
	if err != nil {
		t.Fatalf("Convert returned unexpected error: %v", err)
	}
	if got.AccountID != 100 || got.ContactID != 200 {
		t.Fatalf("Result = %+v, want the ids the repo minted (100/200)", got)
	}
	if !repo.persistHit || repo.persisted == nil {
		t.Fatal("Convert did not hand a plan to Persist")
	}
	p := repo.persisted
	if p.LeadID != 7 {
		t.Errorf("plan.LeadID = %d, want 7", p.LeadID)
	}
	// Account: company → name; email/phone/owner/team carried; active.
	if p.Account.Name != "Polar Foods" {
		t.Errorf("account name = %q, want company 'Polar Foods'", p.Account.Name)
	}
	if p.Account.Email != "sofia@polarfoods.example" || p.Account.Phone != "+47 903 55 666" {
		t.Errorf("account email/phone not carried: %+v", p.Account)
	}
	if p.Account.Status != customer.StatusActive {
		t.Errorf("account status = %q, want active", p.Account.Status)
	}
	if p.Account.AssignedUserID == nil || *p.Account.AssignedUserID != 42 ||
		p.Account.TeamID == nil || *p.Account.TeamID != 3 {
		t.Errorf("account owner/team not carried: %+v", p.Account)
	}
	// Contact: lead name/email/phone; owner/team carried (AccountID wired in store).
	if p.Contact.Name != "Sofia Berg" || p.Contact.Email != "sofia@polarfoods.example" ||
		p.Contact.Phone != "+47 903 55 666" {
		t.Errorf("contact fields not carried: %+v", p.Contact)
	}
	if p.Contact.AssignedUserID == nil || *p.Contact.AssignedUserID != 42 ||
		p.Contact.TeamID == nil || *p.Contact.TeamID != 3 {
		t.Errorf("contact owner/team not carried: %+v", p.Contact)
	}
}

func TestConvertRejectsNonQualifiedLeadWithoutPersisting(t *testing.T) {
	l := qualifiedLead()
	l.Status = lead.StatusWorking
	repo := &fakeRepo{lead: l}
	svc := NewService(repo)

	_, err := svc.Convert(context.Background(), viewer(), 7)
	if !errors.Is(err, ErrNotQualified) {
		t.Fatalf("Convert error = %v, want ErrNotQualified", err)
	}
	if repo.persistHit {
		t.Error("a non-qualified lead must not be persisted")
	}
}

func TestConvertRejectsAlreadyConvertedLead(t *testing.T) {
	l := qualifiedLead()
	l.ConvertedAccountID = uintPtr(99) // already converted
	repo := &fakeRepo{lead: l}
	svc := NewService(repo)

	_, err := svc.Convert(context.Background(), viewer(), 7)
	if !errors.Is(err, ErrAlreadyConverted) {
		t.Fatalf("Convert error = %v, want ErrAlreadyConverted", err)
	}
	if repo.persistHit {
		t.Error("an already-converted lead must not be persisted again")
	}
}

func TestConvertHidesLeadOutsideViewerScopeAsNotFound(t *testing.T) {
	l := qualifiedLead()
	// Owned by someone else, on another team — invisible to the viewer.
	l.AssignedUserID = uintPtr(999)
	l.TeamID = uintPtr(888)
	repo := &fakeRepo{lead: l}
	svc := NewService(repo)

	_, err := svc.Convert(context.Background(), viewer(), 7)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Convert error = %v, want ErrNotFound (not-visible is hidden)", err)
	}
	if repo.persistHit {
		t.Error("a lead outside the viewer's scope must not be persisted")
	}
}

func TestConvertUnknownLeadReturnsNotFound(t *testing.T) {
	repo := &fakeRepo{lead: qualifiedLead()} // GetLead(404) yields lead.ErrNotFound
	svc := NewService(repo)

	_, err := svc.Convert(context.Background(), viewer(), 404)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Convert error = %v, want ErrNotFound", err)
	}
}

func TestConvertRejectsLeadWithNoCompany(t *testing.T) {
	// Company maps to the account's Name, which every account must have. A
	// qualified lead with no company can't form a valid account, so converting
	// it is rejected rather than minting a nameless account.
	l := qualifiedLead()
	l.Company = "   " // blank after trimming
	repo := &fakeRepo{lead: l}
	svc := NewService(repo)

	_, err := svc.Convert(context.Background(), viewer(), 7)
	if !errors.Is(err, ErrCompanyRequired) {
		t.Fatalf("Convert error = %v, want ErrCompanyRequired", err)
	}
	if repo.persistHit {
		t.Error("a lead with no company must not be persisted")
	}
}
