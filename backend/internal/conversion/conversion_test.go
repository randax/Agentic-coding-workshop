package conversion

import (
	"context"
	"errors"
	"testing"

	"saltcrm/internal/lead"
)

// fakeRepo records the plan it was asked to persist and returns canned IDs.
type fakeRepo struct {
	leads     map[uint]lead.Lead
	persisted *Plan
}

func (f *fakeRepo) GetLead(ctx context.Context, id uint) (lead.Lead, error) {
	if l, ok := f.leads[id]; ok {
		return l, nil
	}
	return lead.Lead{}, lead.ErrNotFound
}
func (f *fakeRepo) Persist(ctx context.Context, p Plan) (Result, error) {
	f.persisted = &p
	return Result{AccountID: 11, ContactID: 22, OpportunityID: 33}, nil
}

func TestConvertBuildsAccountContactOpportunityFromLead(t *testing.T) {
	team := uint(5)
	repo := &fakeRepo{leads: map[uint]lead.Lead{
		1: {ID: 1, Name: "Priya Patel", Company: "Fjord Logistics", Email: "priya@fjord.example", Phone: "555", Status: lead.StatusQualified, TeamID: &team},
	}}
	svc := NewService(repo)

	res, err := svc.Convert(context.Background(), 1)
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}
	if res.AccountID != 11 || res.ContactID != 22 || res.OpportunityID != 33 {
		t.Errorf("result = %+v, want the persisted ids", res)
	}

	p := repo.persisted
	if p == nil {
		t.Fatal("Persist was not called")
	}
	if p.Account.Name != "Fjord Logistics" {
		t.Errorf("account name = %q, want the lead's company", p.Account.Name)
	}
	if p.Contact.Name != "Priya Patel" || p.Contact.Email != "priya@fjord.example" {
		t.Errorf("contact = %+v, want built from the lead's person details", p.Contact)
	}
	if p.Opportunity.Name == "" {
		t.Errorf("opportunity should have a name derived from the lead")
	}
	// Visibility carries over from the lead.
	if p.Account.TeamID == nil || *p.Account.TeamID != team {
		t.Errorf("account team = %v, want lead's team %d", p.Account.TeamID, team)
	}
}

func TestConvertFallsBackToLeadNameWhenNoCompany(t *testing.T) {
	repo := &fakeRepo{leads: map[uint]lead.Lead{
		1: {ID: 1, Name: "Solo Founder", Status: lead.StatusNew},
	}}
	svc := NewService(repo)

	if _, err := svc.Convert(context.Background(), 1); err != nil {
		t.Fatalf("Convert error: %v", err)
	}
	if repo.persisted.Account.Name != "Solo Founder" {
		t.Errorf("account name = %q, want the lead name when there's no company", repo.persisted.Account.Name)
	}
}

func TestConvertAlreadyConvertedIsRejected(t *testing.T) {
	repo := &fakeRepo{leads: map[uint]lead.Lead{
		1: {ID: 1, Name: "Priya", Status: lead.StatusConverted},
	}}
	svc := NewService(repo)

	_, err := svc.Convert(context.Background(), 1)
	if !errors.Is(err, ErrAlreadyConverted) {
		t.Fatalf("Convert error = %v, want ErrAlreadyConverted", err)
	}
}

func TestConvertUnknownLeadReturnsNotFound(t *testing.T) {
	repo := &fakeRepo{leads: map[uint]lead.Lead{}}
	svc := NewService(repo)

	_, err := svc.Convert(context.Background(), 999)
	if !errors.Is(err, lead.ErrNotFound) {
		t.Fatalf("Convert error = %v, want lead.ErrNotFound", err)
	}
}
