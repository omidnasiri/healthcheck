package service

import (
	"healthcheck/internal/model"
	"healthcheck/internal/repository"
)

type EndpointService interface {
	CreateEndpoint(url string, interval, retries int) error
	FetchAllEndpoints() ([]*model.Endpoint, error)
	UpdateEndpointActivationStatus(id uint, isActive bool) error
	DeleteEndpoint(id uint) error
}

type endpointService struct {
	endpointRepo repository.EndpointRepository
}

func NewEndpointService(endpointRepo repository.EndpointRepository) EndpointService {
	return &endpointService{endpointRepo}
}

func (s *endpointService) CreateEndpoint(url string, interval, retries int) error {
	model := &model.Endpoint{
		URL:      url,
		Interval: interval,
		Retries:  retries,
	}

	if err := s.endpointRepo.Create(model); err != nil {
		return err
	}

	return nil
}

func (s *endpointService) FetchAllEndpoints() ([]*model.Endpoint, error) {
	models, err := s.endpointRepo.FetchAll()
	if err != nil {
		return nil, err
	}

	return models, nil
}

func (s *endpointService) UpdateEndpointActivationStatus(id uint, isActive bool) error {
	if err := s.endpointRepo.UpdateActivationStatus(id, isActive); err != nil {
		return err
	}

	// todo: stop or start agent
	// wg.Add(1)
	// go Agent(ctx, &wg, alpha)

	return nil
}

func (s *endpointService) DeleteEndpoint(id uint) error {
	if err := s.endpointRepo.Delete(id); err != nil {
		return err
	}

	return nil
}
