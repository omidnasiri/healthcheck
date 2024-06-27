package v1

import (
	"errors"
	"healthcheck/api/presenter"
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
		presenter.Failure(ctx, http.StatusBadRequest, err)
		return
	}

	err = c.endpointService.CreateEndpoint(req.URL, req.Interval, req.Retries)
	if err != nil {
		// ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		presenter.Failure(ctx, http.StatusBadRequest, err)
		return
	}

	presenter.Success(ctx, "endpoint registered successfully")
}

func (c *EndpointController) FetchAllEndpoints(ctx *gin.Context) {
	endpoints, err := c.endpointService.FetchAllEndpoints()
	if err != nil {
		presenter.Failure(ctx, http.StatusBadRequest, err)
		// ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	presenter.Success(ctx, endpoints)
}

func (c *EndpointController) UpdateEndpointActivationStatus(ctx *gin.Context) {
	idStr := ctx.Param("id")
	req := struct {
		Check string `json:"check" binding:"required"`
	}{}
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		presenter.Failure(ctx, http.StatusBadRequest, err)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		presenter.Failure(ctx, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var status bool
	switch req.Check {
	case "activate":
		status = true
	case "deactivate":
		status = false
	default:
		presenter.Failure(ctx, http.StatusBadRequest, errors.New("invalid check"))
		return
	}

	err = c.endpointService.UpdateEndpointActivationStatus(uint(id), status)
	if err != nil {
		presenter.Failure(ctx, http.StatusBadRequest, err)
		// ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	presenter.Success(ctx, "endpoint updated successfully")
}

func (c *EndpointController) DeleteEndpoint(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		presenter.Failure(ctx, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	err = c.endpointService.DeleteEndpoint(uint(id))
	if err != nil {
		// ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		presenter.Failure(ctx, http.StatusBadRequest, err)
		return
	}

	presenter.Success(ctx, "endpoint deleted successfully")
}
