package store

import (
	"context"
	"errors"

	"saltcrm/internal/agent"

	"gorm.io/gorm"
)

// AgentRepository is the GORM-backed implementation of agent.Repository.
type AgentRepository struct {
	db *gorm.DB
}

// NewAgentRepository wires a repository to a GORM database handle.
func NewAgentRepository(db *gorm.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

// List returns all agents ordered by name.
func (r *AgentRepository) List(ctx context.Context) ([]agent.Agent, error) {
	var agents []agent.Agent
	if err := r.db.WithContext(ctx).Order("name").Find(&agents).Error; err != nil {
		return nil, err
	}
	return agents, nil
}

// FindByEmail returns the agent with the given email, or agent.ErrNotFound.
// It also satisfies identity.UserRepository.
func (r *AgentRepository) FindByEmail(ctx context.Context, email string) (agent.Agent, error) {
	var a agent.Agent
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return agent.Agent{}, agent.ErrNotFound
		}
		return agent.Agent{}, err
	}
	return a, nil
}

// FindByID returns the agent with the given id, or agent.ErrNotFound.
func (r *AgentRepository) FindByID(ctx context.Context, id uint) (agent.Agent, error) {
	var a agent.Agent
	if err := r.db.WithContext(ctx).First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return agent.Agent{}, agent.ErrNotFound
		}
		return agent.Agent{}, err
	}
	return a, nil
}
