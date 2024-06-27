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
	UpdateCheckActivation(id uint, isActive bool) error
	UpdateLastStatus(id uint, status bool) error
	Delete(id uint) error
}

type endpointGormRepository struct {
	db *gorm.DB
}

func NewEndpointRepository(db *gorm.DB) EndpointRepository {
	return &endpointGormRepository{db}
}

func (r *endpointGormRepository) Create(model *model.Endpoint) error {
	if err := r.db.Create(model).Error; err != nil {
		log.Printf("error creating endpoint => %v", err)
		return ErrCreate
	}
	return nil
}

func (r *endpointGormRepository) FetchAll() ([]*model.Endpoint, error) {
	var model []*model.Endpoint
	if err := r.db.Find(&model).Error; err != nil {
		log.Printf("error fetching endpoints => %v", err)
		return nil, ErrFetch
	}
	return model, nil
}

func (r *endpointGormRepository) UpdateCheckActivation(id uint, isActive bool) error {
	if err := r.db.Model(&model.Endpoint{}).Where("id = ?", id).Update("active_check", isActive).Error; err != nil {
		log.Printf("error updating endpoint activation status => %v", err)
		return ErrUpdate
	}
	return nil
}

func (r *endpointGormRepository) UpdateLastStatus(id uint, status bool) error {
	if err := r.db.Model(&model.Endpoint{}).Where("id = ?", id).Update("last_status", status).Error; err != nil {
		log.Printf("error updating endpoint activation status => %v", err)
		return ErrUpdate
	}
	return nil
}

func (r *endpointGormRepository) Delete(id uint) error {
	if err := r.db.Delete(&model.Endpoint{}, id).Error; err != nil {
		log.Printf("error deleting endpoint => %v", err)
		return ErrDelete
	}
	return nil
}
