package boot

import (
	"healthcheck/api"
	controllerV1 "healthcheck/api/controller/v1"
	"healthcheck/internal/repository"
	"healthcheck/service"

	"gorm.io/gorm"
)

func Inject(db *gorm.DB) *api.ControllerContainer {

	// Repositories
	endpointRepo := repository.NewEndpointRepository(db)

	// Services
	endpointService := service.NewEndpointService(endpointRepo)

	// Controllers
	endpointController := controllerV1.NewEndpointController(endpointService)

	return api.NewControllerContainer(endpointController)
}
