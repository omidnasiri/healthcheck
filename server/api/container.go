package api

import controllerV1 "healthcheck/api/controller/v1"

type ControllerContainer struct {
	V1 v1
}

type v1 struct {
	EndpointController *controllerV1.EndpointController
}

func NewControllerContainer(endpointController *controllerV1.EndpointController) *ControllerContainer {
	return &ControllerContainer{
		V1: v1{
			endpointController,
		},
	}
}
