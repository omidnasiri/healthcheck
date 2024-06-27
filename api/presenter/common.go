package presenter

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type GenericResponse struct {
	Data   any    `json:"data,omitempty"`
	Error  string `json:"error,omitempty"`
	Result bool   `json:"result"`
}

func newGenericResponse(data any, err string, result bool) GenericResponse {
	return GenericResponse{
		Data:   data,
		Error:  err,
		Result: result,
	}
}

func Failure(ctx *gin.Context, statusCode int, err error) {
	ctx.JSON(statusCode, newGenericResponse(nil, err.Error(), false))
}

func Success(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, newGenericResponse(data, "", true))
}
