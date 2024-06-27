package model

import (
	"context"
	"sync"
)

type HealthCheckAgentFunctionSignature func(ctx context.Context, wg *sync.WaitGroup, endpoint *Endpoint)

type HealthCheckAgent struct {
	ID        uint
	IsActive  bool
	Endpoint  *Endpoint
	Context   context.Context
	Cancel    context.CancelFunc
	AgentFunc HealthCheckAgentFunctionSignature
}
