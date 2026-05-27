// Package team holds the Team model. Teams scope record visibility: a user
// belongs to a team, and (in a later slice) records carry a team so that users
// see records owned by them or shared with their team.
package team

// Team is a group of users that share record visibility.
type Team struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `json:"name"`
}
