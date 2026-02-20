package repository

import (
	"wireless_gallery/internal/model"

	"gorm.io/gorm"
)

type MediaRepository interface {
	Create(entry *model.Media) error
	FindByID(id uint) (*model.Media, error)
	FindByOwnerID(ownerID uint) ([]model.Media, error)
	Update(entry *model.Media) error
	Delete(id uint) error
}

type mediaRepo struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepo{db}
}

func (r *mediaRepo) Create(entry *model.Media) error {
	return r.db.Create(entry).Error
}

func (r *mediaRepo) FindByID(id uint) (*model.Media, error) {
	var media model.Media
	result := r.db.First(&media, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &media, nil
}

func (r *mediaRepo) FindByOwnerID(ownerID uint) ([]model.Media, error) {
	var medias []model.Media
	result := r.db.Where("owner_id = ?", ownerID).Find(&medias)
	if result.Error != nil {
		return nil, result.Error
	}
	return medias, nil
}

func (r *mediaRepo) Update(entry *model.Media) error {
	return r.db.Save(entry).Error
}

func (r *mediaRepo) Delete(id uint) error {
	return r.db.Delete(&model.Media{}, id).Error
}
