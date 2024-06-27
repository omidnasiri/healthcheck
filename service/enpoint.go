package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"healthcheck/internal/model"
	"healthcheck/internal/repository"
	"log"
	"net/http"
	"sync"
	"time"
)

type EndpointService interface {
	CreateEndpoint(url string, interval, retries int) error
	FetchAllEndpoints() ([]*model.Endpoint, error)
	UpdateEndpointActivationStatus(id uint, isActive bool) error
	DeleteEndpoint(id uint) error
}

type endpointService struct {
	webhookURL           string
	wg                   *sync.WaitGroup
	endpointRepo         repository.EndpointRepository
	healthCheckAgentRepo repository.HealthCheckAgentRepository
}

func NewEndpointService(
	webhookURL string,
	wg *sync.WaitGroup,
	endpointRepo repository.EndpointRepository,
	healthCheckAgentRepo repository.HealthCheckAgentRepository,
) EndpointService {
	return &endpointService{webhookURL, wg, endpointRepo, healthCheckAgentRepo}
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

	if err := s.healthCheckAgentRepo.Create(model, agentBuilder(s.webhookURL)); err != nil {
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
	if isActive {
		if err := s.healthCheckAgentRepo.Start(id, s.wg); err != nil {
			return err
		}
	} else {
		if err := s.healthCheckAgentRepo.Stop(id); err != nil {
			return err
		}
	}

	if err := s.endpointRepo.UpdateActivationStatus(id, isActive); err != nil {
		return err
	}

	return nil
}

func (s *endpointService) DeleteEndpoint(id uint) error {
	if err := s.healthCheckAgentRepo.Delete(id); err != nil {
		return err
	}

	if err := s.endpointRepo.Delete(id); err != nil {
		return err
	}

	return nil
}

func agentBuilder(webhookURL string) model.HealthCheckAgentFunctionSignature {
	healthCheck := func(url string) error {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return errors.New("unhealthy")
		}
		return nil
	}

	webhook := func(endpointID uint, status bool) error {
		payload := struct {
			Status bool `json:"status"`
		}{
			Status: status,
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		resp, err := http.Post(fmt.Sprintf("%s/%v", webhookURL, endpointID), "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return errors.New("webhook failed")
		}

		return nil
	}

	return func(ctx context.Context, wg *sync.WaitGroup, endpoint *model.Endpoint) {
		defer wg.Done()
		tries := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(endpoint.Interval) * time.Second):
				err := healthCheck(endpoint.URL)
				if err != nil {
					tries++
					log.Println(endpoint.URL, "health check failed, try ", tries, ", err:", err.Error())
					if tries >= endpoint.Retries {
						tries = 0
						log.Println(endpoint.URL, "endpoint is unhealthy")
						if endpoint.LastStatus {
							endpoint.LastStatus = false
							webhook(endpoint.ID, false)
						}
					}
					continue
				}
				tries = 0
				if !endpoint.LastStatus {
					endpoint.LastStatus = true
					webhook(endpoint.ID, true)
				}
				log.Println(endpoint.URL, "endpoint is healthy")
			}
		}
	}
}
