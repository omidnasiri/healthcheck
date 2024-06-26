package v1

import (
	"healthcheck/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EndpointController struct {
	endpointService service.EndpointService
}

func NewEndpointController(endpointService service.EndpointService) *EndpointController {
	return &EndpointController{endpointService}
}

func (c *EndpointController) CreateEndpoint(ctx *gin.Context) {
	req := struct {
		URL      string `json:"url" binding:"required"`
		Interval int    `json:"interval" binding:"required"`
		Retries  int    `json:"retries" binding:"required"`
	}{}
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = c.endpointService.CreateEndpoint(req.URL, req.Interval, req.Retries)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "endpoint registered successfully"})
}

func (c *EndpointController) FetchAllEndpoints(ctx *gin.Context) {
	endpoints, err := c.endpointService.FetchAllEndpoints()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"endpoints": endpoints})
}

func (c *EndpointController) UpdateEndpointActivationStatus(ctx *gin.Context) {
	idStr := ctx.Param("id")
	req := struct {
		IsActive bool `json:"is_active" binding:"required"`
	}{}
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	err = c.endpointService.UpdateEndpointActivationStatus(uint(id), req.IsActive)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "endpoint updated successfully"})
}

func (c *EndpointController) DeleteEndpoint(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	err = c.endpointService.DeleteEndpoint(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "endpoint deleted successfully"})
}
