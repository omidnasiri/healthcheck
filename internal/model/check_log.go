package model

import "gorm.io/gorm"

type CheckLog struct {
	gorm.Model
	EndpointID       uint
	ResultStatusCode int
	ResultBody       string
}
