package boot

import (
	"healthcheck/api"
	controllerV1 "healthcheck/api/controller/v1"
	"healthcheck/config"
	"healthcheck/internal/repository"
	"healthcheck/service"
	"sync"

	"gorm.io/gorm"
)

func Inject(db *gorm.DB, wg *sync.WaitGroup, cfg *config.Config) (*api.ControllerContainer, error) {

	// Repositories
	endpointRepo := repository.NewEndpointRepository(db)
	checkLogRepo := repository.NewCheckLogRepository(db)
	healthCheckAgentRepo := repository.NewAgentInMemoryRepository()

	// Services
	endpointService, err := service.NewEndpointService(cfg.WebhookURL, wg, checkLogRepo, endpointRepo, healthCheckAgentRepo)
	if err != nil {
		return nil, err
	}

	// Controllers
	endpointController := controllerV1.NewEndpointController(endpointService)

	return api.NewControllerContainer(endpointController), nil
}
