package repository

import (
	"context"
	"errors"
	model "healthcheck/internal/model"
	"sync"
)

var (
	ErrActiveAgent   = errors.New("agent is active")
	ErrInActiveAgent = errors.New("agent is not active")
)

type HealthCheckAgentRepository interface {
	Create(endpoint *model.Endpoint, fn model.HealthCheckAgentFunctionSignature) error
	Delete(id uint) error
	Start(id uint, wg *sync.WaitGroup) error
	Stop(id uint) error
}

type agentInMemoryRepository struct {
	agents map[uint]*model.HealthCheckAgent
}

func NewAgentInMemoryRepository() HealthCheckAgentRepository {
	return &agentInMemoryRepository{
		agents: make(map[uint]*model.HealthCheckAgent),
	}
}

func (r *agentInMemoryRepository) Create(endpoint *model.Endpoint, fn model.HealthCheckAgentFunctionSignature) error {
	_, ok := r.agents[endpoint.ID]
	if ok {
		return ErrCreate
	}
	ctx, cancel := context.WithCancel(context.Background())
	agent := &model.HealthCheckAgent{
		ID:        endpoint.ID,
		Context:   ctx,
		Cancel:    cancel,
		Endpoint:  endpoint,
		AgentFunc: fn,
	}
	r.agents[agent.ID] = agent
	return nil
}

func (r *agentInMemoryRepository) Delete(id uint) error {
	agent, ok := r.agents[id]
	if !ok {
		return ErrFetch
	}
	if agent.IsActive {
		return ErrActiveAgent
	}
	delete(r.agents, id)
	return nil
}

func (r *agentInMemoryRepository) Start(id uint, wg *sync.WaitGroup) error {
	agent, ok := r.agents[id]
	if !ok {
		return ErrFetch
	}
	if agent.IsActive {
		return ErrActiveAgent
	}
	agent.IsActive = true
	wg.Add(1)
	go agent.AgentFunc(agent.Context, wg, agent.Endpoint)
	return nil
}

func (r *agentInMemoryRepository) Stop(id uint) error {
	agent, ok := r.agents[id]
	if !ok {
		return ErrFetch
	}
	if !agent.IsActive {
		return ErrInActiveAgent
	}
	agent.IsActive = false
	agent.Cancel()
	return nil
}
