// Package customer holds the customer domain model and the service that owns
// customer-related business logic. The service depends on the Repository
// interface, not on any database, so it can be unit-tested in isolation.
package customer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrNotFound is returned when a requested customer does not exist. Repository
// implementations translate storage-specific "not found" errors into this so
// the rest of the app stays decoupled from the persistence layer.
var ErrNotFound = errors.New("customer not found")

// Validation errors returned by Create and Update when required fields are
// missing or the account status is not a known value.
var (
	ErrNameRequired          = errors.New("customer name is required")
	ErrEmailRequired         = errors.New("customer email is required")
	ErrAccountNumberRequired = errors.New("customer account number is required")
	ErrInvalidStatus         = errors.New("invalid customer status")
)

// FormatAccountNumber renders a sequence number as a server-minted account
// number in the seed format ISP-#### (zero-padded to four digits). The next
// sequence is computed where uniqueness must hold (inside the persisting
// transaction); this function owns only the format rule.
func FormatAccountNumber(seq int) string {
	return fmt.Sprintf("ISP-%04d", seq)
}

// Status is a customer's account standing.
type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
)

// Valid reports whether s is a known account status.
func (s Status) Valid() bool {
	switch s {
	case StatusActive, StatusSuspended:
		return true
	default:
		return false
	}
}

// Customer is a residential ISP customer. Struct tags configure GORM and JSON
// but the package itself does not import GORM, keeping the domain layer clean.
type Customer struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	ServiceAddress string    `json:"serviceAddress"`
	AccountNumber  string    `gorm:"uniqueIndex" json:"accountNumber"`
	CustomerSince  time.Time `json:"customerSince"`
	Status         Status    `json:"status"`

	// Ownership for record-level visibility (own-or-team). Nullable.
	AssignedUserID *uint `json:"assignedUserId,omitempty"`
	TeamID         *uint `json:"teamId,omitempty"`

	// CustomFields holds runtime-defined (Studio) field values. It is flattened
	// to the top level in JSON (see MarshalJSON) so the generic views render it
	// like any other field, and unknown keys on input are captured into it.
	CustomFields map[string]any `gorm:"serializer:json" json:"-"`
}

// knownCustomerFields are the JSON keys that map to real struct fields; anything
// else on input is a custom field value.
var knownCustomerFields = map[string]bool{
	"id": true, "name": true, "email": true, "phone": true, "serviceAddress": true,
	"accountNumber": true, "customerSince": true, "status": true,
	"assignedUserId": true, "teamId": true,
}

// MarshalJSON flattens CustomFields to the top level alongside the struct fields.
func (c Customer) MarshalJSON() ([]byte, error) {
	type alias Customer
	b, err := json.Marshal(alias(c))
	if err != nil {
		return nil, err
	}
	if len(c.CustomFields) == 0 {
		return b, nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	for k, v := range c.CustomFields {
		m[k] = v
	}
	return json.Marshal(m)
}

// UnmarshalJSON populates the struct fields and captures any unknown top-level
// keys into CustomFields (so custom field values sent by the generic edit form
// are routed to the JSON column).
func (c *Customer) UnmarshalJSON(b []byte) error {
	type alias Customer
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*c = Customer(a)
	var all map[string]json.RawMessage
	if err := json.Unmarshal(b, &all); err != nil {
		return err
	}
	for k, raw := range all {
		if knownCustomerFields[k] {
			continue
		}
		if c.CustomFields == nil {
			c.CustomFields = map[string]any{}
		}
		var v any
		if err := json.Unmarshal(raw, &v); err != nil {
			return err
		}
		c.CustomFields[k] = v
	}
	return nil
}

// Repository is the persistence seam the service depends on. The real
// implementation lives in the store package; tests use an in-memory fake.
type Repository interface {
	List(ctx context.Context) ([]Customer, error)
	// Get returns the customer with the given ID, or ErrNotFound.
	Get(ctx context.Context, id uint) (Customer, error)
	// Create inserts a new customer, assigning its ID.
	Create(ctx context.Context, c *Customer) error
	// Update persists all fields of an existing customer.
	Update(ctx context.Context, c *Customer) error
}

// Filter narrows a customer list query. The zero value matches every customer.
type Filter struct {
	// Search is a partial, case-insensitive term matched against a customer's
	// name, email, or account number. Empty matches all.
	Search string
	// Status restricts results to a single account status. Empty matches all.
	Status Status
}

// Service owns customer business logic.
type Service struct {
	repo Repository
}

// NewService wires a Service to its repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// List returns the customers matching the filter. Filtering happens in the
// service (not the database) so the matching logic is unit-testable in isolation.
func (s *Service) List(ctx context.Context, f Filter) ([]Customer, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	var matched []Customer
	for _, c := range all {
		if f.matches(c) {
			matched = append(matched, c)
		}
	}
	return matched, nil
}

// matches reports whether a customer satisfies the filter.
func (f Filter) matches(c Customer) bool {
	if f.Status != "" && c.Status != f.Status {
		return false
	}
	if term := strings.ToLower(f.Search); term != "" {
		if !strings.Contains(strings.ToLower(c.Name), term) &&
			!strings.Contains(strings.ToLower(c.Email), term) &&
			!strings.Contains(strings.ToLower(c.AccountNumber), term) {
			return false
		}
	}
	return true
}

// Get returns a single customer by ID, or ErrNotFound.
func (s *Service) Get(ctx context.Context, id uint) (Customer, error) {
	return s.repo.Get(ctx, id)
}

// validate enforces the required-field and status rules shared by Create and Update.
func (c Customer) validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return ErrNameRequired
	}
	if strings.TrimSpace(c.Email) == "" {
		return ErrEmailRequired
	}
	if strings.TrimSpace(c.AccountNumber) == "" {
		return ErrAccountNumberRequired
	}
	if !c.Status.Valid() {
		return ErrInvalidStatus
	}
	return nil
}

// Create adds a new customer. A new customer defaults to active status unless
// another valid status is given.
func (s *Service) Create(ctx context.Context, c Customer) (Customer, error) {
	if c.Status == "" {
		c.Status = StatusActive
	}
	if err := c.validate(); err != nil {
		return Customer{}, err
	}
	if c.CustomerSince.IsZero() {
		c.CustomerSince = time.Now()
	}
	c.ID = 0 // the repository assigns identifiers
	if err := s.repo.Create(ctx, &c); err != nil {
		return Customer{}, err
	}
	return c, nil
}

// Update edits an existing customer's profile fields and status. Server-managed
// fields (CustomerSince) are preserved by loading the existing record first;
// an unknown ID yields ErrNotFound.
func (s *Service) Update(ctx context.Context, c Customer) (Customer, error) {
	existing, err := s.repo.Get(ctx, c.ID)
	if err != nil {
		return Customer{}, err
	}
	existing.Name = c.Name
	existing.Email = c.Email
	existing.Phone = c.Phone
	existing.ServiceAddress = c.ServiceAddress
	existing.AccountNumber = c.AccountNumber
	existing.Status = c.Status
	existing.AssignedUserID = c.AssignedUserID // owner reassignment
	existing.TeamID = c.TeamID
	if c.CustomFields != nil {
		existing.CustomFields = c.CustomFields
	}
	if err := existing.validate(); err != nil {
		return Customer{}, err
	}
	if err := s.repo.Update(ctx, &existing); err != nil {
		return Customer{}, err
	}
	return existing, nil
}
