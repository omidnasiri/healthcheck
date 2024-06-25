package repository

import (
	"errors"
	"healthcheck/internal/model"
	"log"

	"gorm.io/gorm"
)

var (
	ErrCreate = errors.New("error creating model")
	ErrFetch  = errors.New("error fetching model")
	ErrUpdate = errors.New("error updating model")
	ErrDelete = errors.New("error deleting model")
)

type EndpointRepository interface {
	Create(model *model.Endpoint) error
	FetchAll() ([]*model.Endpoint, error)
	UpdateActivationStatus(id uint, isActive bool) error
	Delete(id uint) error
}

type endpointRepository struct {
	db *gorm.DB
}

func NewEndpointRepository(db *gorm.DB) EndpointRepository {
	return &endpointRepository{db: db}
}

func (r *endpointRepository) Create(model *model.Endpoint) error {
	if err := r.db.Create(model).Error; err != nil {
		log.Printf("error creating endpoint => %v", err)
		return ErrCreate
	}
	return nil
}

func (r *endpointRepository) FetchAll() ([]*model.Endpoint, error) {
	var models []*model.Endpoint
	if err := r.db.Find(&models).Error; err != nil {
		log.Printf("error fetching endpoints => %v", err)
		return nil, ErrFetch
	}
	return models, nil
}

func (r *endpointRepository) UpdateActivationStatus(id uint, isActive bool) error {
	if err := r.db.Model(&model.Endpoint{}).Where("id = ?", id).Update("is_active", isActive).Error; err != nil {
		log.Printf("error updating endpoint activation status => %v", err)
		return ErrUpdate
	}
	return nil
}

func (r *endpointRepository) Delete(id uint) error {
	if err := r.db.Delete(&model.Endpoint{}, id).Error; err != nil {
		log.Printf("error deleting endpoint => %v", err)
		return ErrDelete
	}
	return nil
}
