package repository

import (
	"healthcheck/internal/model"
	"log"

	"gorm.io/gorm"
)

type CheckLogRepository interface {
	Create(endpointID uint, statusCode int, body string) error
	FetchByEndpointID(endpointID uint) ([]*model.Endpoint, error)
}

type checkLogRepository struct {
	db *gorm.DB
}

func NewCheckLogRepository(db *gorm.DB) CheckLogRepository {
	return &checkLogRepository{db}
}

func (r *checkLogRepository) Create(endpointID uint, statusCode int, body string) error {
	model := &model.CheckLog{EndpointID: endpointID, ResultStatusCode: statusCode, ResultBody: body}
	if err := r.db.Create(model).Error; err != nil {
		log.Printf("error creating check log => %v", err)
		return ErrCreate
	}
	return nil
}

func (r *checkLogRepository) FetchByEndpointID(endpointID uint) ([]*model.Endpoint, error) {
	var checkLogs []*model.Endpoint
	if err := r.db.Where("endpoint_id = ?", endpointID).Find(&checkLogs).Error; err != nil {
		log.Printf("error fetching check logs => %v", err)
		return nil, ErrFetch
	}
	return checkLogs, nil
}
