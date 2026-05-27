package access

import (
	"testing"

	"saltcrm/internal/agent"
)

func ptr(u uint) *uint { return &u }

func TestAdminSeesEverything(t *testing.T) {
	admin := agent.Agent{ID: 1, Role: agent.RoleAdmin, TeamID: ptr(10)}

	if !Visible(admin, nil, nil) {
		t.Error("admin should see a record with no owner/team")
	}
	if !Visible(admin, ptr(999), ptr(999)) {
		t.Error("admin should see records owned by anyone on any team")
	}
}

func TestOwnerSeesOwnRecord(t *testing.T) {
	user := agent.Agent{ID: 5, Role: agent.RoleAgent, TeamID: ptr(10)}

	if !Visible(user, ptr(5), nil) {
		t.Error("a user should see a record assigned to them")
	}
	if Visible(user, ptr(6), nil) {
		t.Error("a user should not see a record owned by someone else on no team")
	}
}

func TestSameTeamIsVisible(t *testing.T) {
	user := agent.Agent{ID: 5, Role: agent.RoleAgent, TeamID: ptr(10)}

	if !Visible(user, ptr(6), ptr(10)) {
		t.Error("a user should see another's record on their own team")
	}
	if Visible(user, ptr(6), ptr(20)) {
		t.Error("a user should not see a record on a different team they don't own")
	}
}

func TestNoTeamUserOnlySeesOwn(t *testing.T) {
	user := agent.Agent{ID: 5, Role: agent.RoleAgent, TeamID: nil}

	if Visible(user, ptr(6), ptr(10)) {
		t.Error("a teamless user should not see team records they don't own")
	}
	if !Visible(user, ptr(5), ptr(10)) {
		t.Error("a teamless user should still see records they own")
	}
}
