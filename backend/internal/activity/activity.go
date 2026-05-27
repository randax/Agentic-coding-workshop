// Package activity holds the activity domain model and service. An activity is
// a call, meeting, or task logged against any record (account, contact,
// opportunity, …) and shown on that record's timeline. The service depends on
// the Repository interface, not a database, so it is unit-testable in isolation.
package activity

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"
)

// ErrNotFound is returned when a requested activity does not exist.
var ErrNotFound = errors.New("activity not found")

// ErrSubjectRequired is returned when an activity has no subject.
var ErrSubjectRequired = errors.New("activity subject is required")

// ErrInvalidType is returned when an activity's type is not a known value.
var ErrInvalidType = errors.New("invalid activity type")

// Type is the kind of activity.
type Type string

const (
	TypeCall    Type = "call"
	TypeMeeting Type = "meeting"
	TypeTask    Type = "task"
)

// Valid reports whether t is a known activity type.
func (t Type) Valid() bool {
	switch t {
	case TypeCall, TypeMeeting, TypeTask:
		return true
	default:
		return false
	}
}

// Status tracks task completion (calls/meetings are typically logged as done).
type Status string

const (
	StatusOpen Status = "open"
	StatusDone Status = "done"
)

// Activity is a call/meeting/task logged against a parent record. ParentType +
// ParentID are a lightweight polymorphic link (e.g. "account"/42).
// AssignedUserID/TeamID scope visibility.
type Activity struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	Type       Type       `json:"type"`
	Subject    string     `json:"subject"`
	Notes      string     `json:"notes"`
	Status     Status     `json:"status"`
	DueDate    *time.Time `json:"dueDate,omitempty"`
	ParentType string     `gorm:"index" json:"parentType"`
	ParentID   uint       `gorm:"index" json:"parentId"`
	OccurredAt time.Time  `json:"occurredAt"`

	AssignedUserID *uint `json:"assignedUserId,omitempty"`
	TeamID         *uint `json:"teamId,omitempty"`
}

// Repository is the persistence seam the service depends on.
type Repository interface {
	List(ctx context.Context) ([]Activity, error)
	ListForParent(ctx context.Context, parentType string, parentID uint) ([]Activity, error)
	Get(ctx context.Context, id uint) (Activity, error)
	Create(ctx context.Context, a *Activity) error
	Update(ctx context.Context, a *Activity) error
}

// Service owns activity business logic.
type Service struct {
	repo Repository
}

// NewService wires a Service to its repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// List returns all activities.
func (s *Service) List(ctx context.Context) ([]Activity, error) {
	return s.repo.List(ctx)
}

// ListForParent returns the activities logged against one record.
func (s *Service) ListForParent(ctx context.Context, parentType string, parentID uint) ([]Activity, error) {
	return s.repo.ListForParent(ctx, parentType, parentID)
}

// OpenTasksForUser returns a user's still-open tasks, soonest first — the "My
// Tasks" dashlet. Calls/meetings and completed tasks are excluded.
func (s *Service) OpenTasksForUser(ctx context.Context, userID uint) ([]Activity, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	var tasks []Activity
	for _, a := range all {
		if a.Type == TypeTask && a.Status == StatusOpen && a.AssignedUserID != nil && *a.AssignedUserID == userID {
			tasks = append(tasks, a)
		}
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].OccurredAt.Before(tasks[j].OccurredAt) })
	return tasks, nil
}

// Log records a new activity. New activities default to open status and to the
// current time if no occurrence time is given.
func (s *Service) Log(ctx context.Context, a Activity) (Activity, error) {
	if strings.TrimSpace(a.Subject) == "" {
		return Activity{}, ErrSubjectRequired
	}
	if !a.Type.Valid() {
		return Activity{}, ErrInvalidType
	}
	if a.Status == "" {
		a.Status = StatusOpen
	}
	if a.OccurredAt.IsZero() {
		a.OccurredAt = time.Now()
	}
	a.ID = 0
	if err := s.repo.Create(ctx, &a); err != nil {
		return Activity{}, err
	}
	return a, nil
}

// Complete marks a task activity done.
func (s *Service) Complete(ctx context.Context, id uint) (Activity, error) {
	a, err := s.repo.Get(ctx, id)
	if err != nil {
		return Activity{}, err
	}
	a.Status = StatusDone
	if err := s.repo.Update(ctx, &a); err != nil {
		return Activity{}, err
	}
	return a, nil
}
