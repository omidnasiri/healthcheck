package api

import "github.com/gin-gonic/gin"

func SetupRoutes(container *ControllerContainer) *gin.Engine {
	routes := gin.Default()
	api := routes.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			endpoints := v1.Group("/endpoints")
			{
				endpoints.POST("/", container.V1.EndpointController.CreateEndpoint)
				endpoints.GET("/", container.V1.EndpointController.FetchAllEndpoints)
				endpoints.PATCH("/:id", container.V1.EndpointController.UpdateEndpointActivationStatus)
				endpoints.GET("/:id", container.V1.EndpointController.DeleteEndpoint)
			}
		}
	}

	return routes
}
