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
) (EndpointService, error) {
	endpointService := &endpointService{webhookURL, wg, endpointRepo, healthCheckAgentRepo}
	if err := endpointService.bootstrap(); err != nil {
		log.Println("failed to bootstrap endpoint service, err:", err.Error())
		return nil, err
	}
	return endpointService, nil
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

	if err := s.healthCheckAgentRepo.Create(model, s.agentFactory()); err != nil {
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

	if err := s.endpointRepo.UpdateCheckActivation(id, isActive); err != nil {
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

func (s *endpointService) bootstrap() error {
	models, err := s.FetchAllEndpoints()
	if err != nil {
		return err
	}

	for _, model := range models {
		if model.ActiveCheck {
			if err := s.healthCheckAgentRepo.Create(model, s.agentFactory()); err != nil {
				log.Println("failed to create health check agent for endpoint ", model.ID, ", err:", err.Error())
				continue
			}

			if err := s.healthCheckAgentRepo.Start(model.ID, s.wg); err != nil {
				log.Println("failed to start health check agent for endpoint ", model.ID, ", err:", err.Error())
			}
		}
	}

	log.Println("all health check agents started")
	return nil
}

func (s *endpointService) agentFactory() model.HealthCheckAgentFunctionSignature {
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

	webhook := func(endpointID uint, status bool) {
		payload := struct {
			Status bool `json:"status"`
		}{
			Status: status,
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			log.Println("failed to marshal json payload, err:", err.Error())
			return
		}

		resp, err := http.Post(fmt.Sprintf("%s/%v", s.webhookURL, endpointID), "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			log.Println("failed to send webhook, err:", err.Error())
			return
		}

		if resp.StatusCode != http.StatusOK {
			log.Println("webhook failed, status:", resp.StatusCode)
			return
		}
	}

	updateStatus := func(endpoint *model.Endpoint, status bool) {
		endpoint.LastStatus = status

		if err := s.endpointRepo.UpdateLastStatus(endpoint.ID, status); err != nil {
			log.Println("failed to update last status for endpoint ", endpoint.ID, ", err:", err.Error())
		}

		webhook(endpoint.ID, status)
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
							updateStatus(endpoint, false)
						}
					}
					continue
				}
				tries = 0
				if !endpoint.LastStatus {
					updateStatus(endpoint, true)
				}
				log.Println(endpoint.URL, "endpoint is healthy")
			}
		}
	}
}
