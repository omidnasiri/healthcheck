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

func Inject(db *gorm.DB, wg *sync.WaitGroup, cfg *config.Config) *api.ControllerContainer {

	// Repositories
	endpointRepo := repository.NewEndpointRepository(db)
	healthCheckAgentRepo := repository.NewAgentInMemoryRepository()

	// Services
	endpointService := service.NewEndpointService(cfg.WebhookURL, wg, endpointRepo, healthCheckAgentRepo)

	// Controllers
	endpointController := controllerV1.NewEndpointController(endpointService)

	return api.NewControllerContainer(endpointController)
}
