package services

import (
	"errors"
	"tridorian-ztna/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BackofficeService struct {
	db *gorm.DB
}

func NewBackofficeService(db *gorm.DB) *BackofficeService {
	return &BackofficeService{db: db}
}

func (s *BackofficeService) Authenticate(email, password string) (*models.BackofficeUser, error) {
	var user models.BackofficeUser
	err := s.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !user.CheckPassword(password) {
		return nil, errors.New("invalid email or password")
	}

	return &user, nil
}

func (s *BackofficeService) GetByID(id uuid.UUID) (*models.BackofficeUser, error) {
	var user models.BackofficeUser
	err := s.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
