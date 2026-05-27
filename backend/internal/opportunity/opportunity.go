// Package opportunity holds the sales-opportunity domain model and service. An
// opportunity is a potential deal tied to an account, moving through sales
// stages toward Closed Won/Lost. The service depends on the Repository
// interface, not a database, so it is unit-testable in isolation.
package opportunity

import (
	"context"
	"errors"
	"strings"
	"time"
)

// ErrNotFound is returned when a requested opportunity does not exist.
var ErrNotFound = errors.New("opportunity not found")

// ErrNameRequired is returned when an opportunity has no name.
var ErrNameRequired = errors.New("opportunity name is required")

// ErrInvalidAmount is returned when an opportunity's amount is negative.
var ErrInvalidAmount = errors.New("opportunity amount must not be negative")

// ErrInvalidStage is returned when an opportunity's stage is not a known value.
var ErrInvalidStage = errors.New("invalid opportunity stage")

// Stage is a sales stage in the pipeline.
type Stage string

const (
	StageProspecting   Stage = "prospecting"
	StageQualification Stage = "qualification"
	StageProposal      Stage = "proposal"
	StageNegotiation   Stage = "negotiation"
	StageClosedWon     Stage = "closed_won"
	StageClosedLost    Stage = "closed_lost"
)

// Stages is the ordered pipeline, used for grouping and display order.
var Stages = []Stage{
	StageProspecting, StageQualification, StageProposal,
	StageNegotiation, StageClosedWon, StageClosedLost,
}

// defaultProbability is the win probability (%) each stage implies.
var defaultProbability = map[Stage]int{
	StageProspecting:   10,
	StageQualification: 25,
	StageProposal:      50,
	StageNegotiation:   75,
	StageClosedWon:     100,
	StageClosedLost:    0,
}

// Valid reports whether s is a known stage.
func (s Stage) Valid() bool {
	_, ok := defaultProbability[s]
	return ok
}

// Probability returns the default win probability for the stage.
func (s Stage) Probability() int { return defaultProbability[s] }

// Opportunity is a potential deal tied to an account. Probability tracks the
// stage's default. AssignedUserID/TeamID scope visibility.
type Opportunity struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Name              string    `json:"name"`
	AccountID         uint      `gorm:"index" json:"accountId"`
	Amount            float64   `json:"amount"`
	Stage             Stage     `json:"stage"`
	Probability       int       `json:"probability"`
	ExpectedCloseDate time.Time `json:"expectedCloseDate"`

	AssignedUserID *uint `json:"assignedUserId,omitempty"`
	TeamID         *uint `json:"teamId,omitempty"`
}

// Repository is the persistence seam the service depends on.
type Repository interface {
	List(ctx context.Context) ([]Opportunity, error)
	Get(ctx context.Context, id uint) (Opportunity, error)
	Create(ctx context.Context, o *Opportunity) error
	Update(ctx context.Context, o *Opportunity) error
}

// Service owns opportunity business logic.
type Service struct {
	repo Repository
}

// NewService wires a Service to its repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// List returns all opportunities.
func (s *Service) List(ctx context.Context) ([]Opportunity, error) {
	return s.repo.List(ctx)
}

// Get returns a single opportunity by ID, or ErrNotFound.
func (s *Service) Get(ctx context.Context, id uint) (Opportunity, error) {
	return s.repo.Get(ctx, id)
}

// validate enforces the rules shared by Create and Update, and syncs the
// probability to the stage's default.
func (o *Opportunity) validateAndSync() error {
	if o.Stage == "" {
		o.Stage = StageProspecting
	}
	if strings.TrimSpace(o.Name) == "" {
		return ErrNameRequired
	}
	if o.Amount < 0 {
		return ErrInvalidAmount
	}
	if !o.Stage.Valid() {
		return ErrInvalidStage
	}
	o.Probability = o.Stage.Probability()
	return nil
}

// Create adds a new opportunity, defaulting the stage and syncing probability.
func (s *Service) Create(ctx context.Context, o Opportunity) (Opportunity, error) {
	if err := o.validateAndSync(); err != nil {
		return Opportunity{}, err
	}
	o.ID = 0
	if err := s.repo.Create(ctx, &o); err != nil {
		return Opportunity{}, err
	}
	return o, nil
}

// Update edits an opportunity, loading the existing record first so an unknown
// ID yields ErrNotFound; the probability is resynced to the (possibly new) stage.
func (s *Service) Update(ctx context.Context, o Opportunity) (Opportunity, error) {
	existing, err := s.repo.Get(ctx, o.ID)
	if err != nil {
		return Opportunity{}, err
	}
	o.ID = existing.ID
	if err := o.validateAndSync(); err != nil {
		return Opportunity{}, err
	}
	existing.Name = o.Name
	existing.AccountID = o.AccountID
	existing.Amount = o.Amount
	existing.Stage = o.Stage
	existing.Probability = o.Probability
	existing.ExpectedCloseDate = o.ExpectedCloseDate
	existing.AssignedUserID = o.AssignedUserID
	existing.TeamID = o.TeamID
	if err := s.repo.Update(ctx, &existing); err != nil {
		return Opportunity{}, err
	}
	return existing, nil
}
