// Package access holds the record-visibility rule shared by every module's
// list endpoint: a user sees records they own or that belong to their team;
// admins see everything. It depends only on the agent model, so the rule lives
// in one place and is unit-tested in isolation.
package access

import "saltcrm/internal/agent"

// Visible reports whether viewer may see a record with the given assigned-user
// and team. Admins see all records. Otherwise a record is visible if the viewer
// owns it (assigned to them) or it belongs to the viewer's team.
func Visible(viewer agent.Agent, assignedUserID *uint, teamID *uint) bool {
	if viewer.Role == agent.RoleAdmin {
		return true
	}
	if assignedUserID != nil && *assignedUserID == viewer.ID {
		return true
	}
	if teamID != nil && viewer.TeamID != nil && *teamID == *viewer.TeamID {
		return true
	}
	return false
}
