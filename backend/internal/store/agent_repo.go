package store

import (
	"context"

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
